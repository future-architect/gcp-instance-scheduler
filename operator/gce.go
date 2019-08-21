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
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/future-architect/gcp-instance-scheduler/model"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/net/context"
	"google.golang.org/api/compute/v1"
)

type ComputeEngineCall struct {
	Service   *compute.Service
	Call      *compute.InstancesAggregatedListCall
	ProjectID string
	Error     error
}

func ComputeEngine(ctx context.Context, projectID string) *ComputeEngineCall {
	s, err := compute.NewService(ctx)
	if err != nil {
		return &ComputeEngineCall{Error: err}
	}

	// get all instances in each zone at this project
	return &ComputeEngineCall{
		Service:   s,
		ProjectID: projectID,
		Call:      compute.NewInstancesService(s).AggregatedList(projectID),
	}
}

func (r *ComputeEngineCall) Filter(labelName string, flag bool) *ComputeEngineCall {
	if r.Error != nil {
		return r
	}
	return &ComputeEngineCall{
		ProjectID: r.ProjectID,
		Call:      r.Call.Filter("labels." + labelName + "=" + strconv.FormatBool(flag)),
	}
}

func (r *ComputeEngineCall) Do(ctx context.Context, interval time.Duration) (*model.ShutdownReport, error) {
	if r.Error != nil {
		return nil, r.Error
	}

	list, err := r.Call.Do()
	if err != nil {
		return nil, err
	}

	var res = r.Error
	var doneRes []string
	var alreadyRes []string

	for _, instance := range valuesGCE(list.Items) {
		// check a instance which was already stopped
		if instance.Status == "STOPPED" ||
			instance.Status == "STOPPING" ||
			instance.Status == "TERMINATED" {
			alreadyRes = append(alreadyRes, instance.Name)
			continue
		}

		// get zone name
		urlElements := strings.Split(instance.Zone, "/")
		zone := urlElements[len(urlElements)-1]

		// shutdown an instance
		_, err = compute.NewInstancesService(r.Service).Stop(r.ProjectID, zone, instance.Name).Do()
		if err != nil {
			res = multierror.Append(res, err)
		}

		doneRes = append(doneRes, instance.Name)
		time.Sleep(interval)
	}

	log.Printf("Success in stopping GCE instances: Done.")

	return &model.ShutdownReport{
		InstanceType:             model.ComputeEngine,
		DoneResources:            doneRes,
		AlreadyShutdownResources: alreadyRes,
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
