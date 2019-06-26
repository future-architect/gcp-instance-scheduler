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

type GCEListCall struct {
	AggregatedListCall *compute.InstancesAggregatedListCall
	ProjectID          string
	Error              error
}

type GCEShutdownCall struct {
	TargetList *compute.InstanceAggregatedList
	ProjectID  string
	Error      error
}

// create instance list
func valuesGCE(m map[string]compute.InstancesScopedList) []*compute.Instance {
	var res []*compute.Instance
	for _, instanceList := range m {
		res = append(res, instanceList.Instances...)
	}
	return res
}

func ComputeEngineResource(ctx context.Context, projectID string) *GCEListCall {
	// reporting error list
	var res error

	// create service to operate instances
	computeService, err := compute.NewService(ctx)
	if err != nil {
		res = multierror.Append(res, err)
	}

	// get all instances in each zone at this project
	list := compute.NewInstancesService(computeService).AggregatedList(projectID)

	return &GCEListCall{
		AggregatedListCall: list,
		ProjectID:          projectID,
		Error:              res,
	}
}

func (r *GCEListCall) FilterLabel(targetLabel string, flag bool) *GCEShutdownCall {
	// reporting error list
	var res = r.Error

	list, err := r.AggregatedListCall.Filter("labels." + targetLabel + "=" + strconv.FormatBool(flag)).Do()
	if err != nil {
		res = multierror.Append(r.Error, err)
	}

	return &GCEShutdownCall{
		TargetList: list,
		ProjectID:  r.ProjectID,
		Error:      res,
	}
}

func (r *GCEShutdownCall) ShutdownWithInterval(ctx context.Context, interval time.Duration) (*model.ShutdownReport, error) {
	var res = r.Error
	var doneRes []string
	var alreadyRes []string

	// create service to operate instances
	computeService, err := compute.NewService(ctx)
	if err != nil {
		res = multierror.Append(res, err)
	}

	for _, instance := range valuesGCE(r.TargetList.Items) {
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
		_, err = compute.NewInstancesService(computeService).Stop(r.ProjectID, zone, instance.Name).Do()
		if err != nil {
			res = multierror.Append(res, err)
		}
		doneRes = append(doneRes, instance.Name)
		time.Sleep(interval)
	}
	log.Printf("Success in stopping GCE instances: Done.")

	return &model.ShutdownReport{
		DoneResources:            doneRes,
		AlreadyShutdownResources: alreadyRes,
	}, res
}
