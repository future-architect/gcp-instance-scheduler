/**
 * Copyright (c) 2019-present Future Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package operator

import (
	"errors"
	"strings"
	"time"

	"github.com/future-architect/gcp-instance-scheduler/model"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/net/context"
	"google.golang.org/api/compute/v1"
)

type ComputeEngineCall struct {
	s         *compute.Service
	call      *compute.InstancesAggregatedListCall
	projectID string
	error     error
}

func ComputeEngine(ctx context.Context, projectID string) *ComputeEngineCall {
	s, err := compute.NewService(ctx)
	if err != nil {
		return &ComputeEngineCall{error: err}
	}

	// get all instances in each zone at this project
	return &ComputeEngineCall{
		s:         s,
		projectID: projectID,
		call:      compute.NewInstancesService(s).AggregatedList(projectID),
	}
}

func (r *ComputeEngineCall) Filter(labelName, value string) *ComputeEngineCall {
	if r.error != nil {
		return r
	}
	r.call = r.call.Filter("labels." + labelName + "=" + value)
	return r
}

func (r *ComputeEngineCall) Stop() (*model.Report, error) {
	if r.error != nil {
		return nil, r.error
	}

	list, err := r.call.Do()
	if err != nil {
		return nil, err
	}

	var res = r.error
	var doneRes []string
	var alreadyRes []string

	for _, instance := range valuesGCE(list.Items) {
		// check a instance which was already stopped
		if instance.Status == "STOPPED" || instance.Status == "STOPPING" || instance.Status == "TERMINATED" ||
			instance.Status == "PROVISIONING" || instance.Status == "REPAIRING" {
			alreadyRes = append(alreadyRes, instance.Name)
			continue
		}

		// get zone name
		urlElements := strings.Split(instance.Zone, "/")
		zone := urlElements[len(urlElements)-1]

		_, err = compute.NewInstancesService(r.s).Stop(r.projectID, zone, instance.Name).Do()
		if err != nil {
			res = multierror.Append(res, errors.New(instance.Name+" stopping failed: %v"+err.Error()))
		}

		doneRes = append(doneRes, instance.Name)
		time.Sleep(CallInterval)
	}

	return &model.Report{
		InstanceType: model.ComputeEngine,
		Dones:        doneRes,
		Alreadies:    alreadyRes,
	}, res
}

func (r *ComputeEngineCall) Start() (*model.Report, error) {
	if r.error != nil {
		return nil, r.error
	}

	list, err := r.call.Do()
	if err != nil {
		return nil, err
	}

	var res = r.error
	var doneRes []string
	var alreadyRes []string

	for _, instance := range valuesGCE(list.Items) {
		// check a instance which was already running
		if instance.Status == "RUNNING" || instance.Status == "PROVISIONING" || instance.Status == "REPAIRING" {
			alreadyRes = append(alreadyRes, instance.Name)
			continue
		}

		// get zone name
		urlElements := strings.Split(instance.Zone, "/")
		zone := urlElements[len(urlElements)-1]

		_, err = compute.NewInstancesService(r.s).Start(r.projectID, zone, instance.Name).Do()
		if err != nil {
			res = multierror.Append(res, err)
		}

		doneRes = append(doneRes, instance.Name)
		time.Sleep(CallInterval)
	}

	return &model.Report{
		InstanceType: model.ComputeEngine,
		Dones:        doneRes,
		Alreadies:    alreadyRes,
	}, res
}

// create instance list
func valuesGCE(m map[string]compute.InstancesScopedList) []*compute.Instance {
	var res []*compute.Instance
	for _, instanceList := range m {
		if len(instanceList.Instances) == 0 {
			continue
		}
		res = append(res, instanceList.Instances...)
	}
	return res
}
