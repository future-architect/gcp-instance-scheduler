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
	"time"

	"github.com/future-architect/gcp-instance-scheduler/model"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/net/context"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

type SQLCall struct {
	Service   *sqladmin.Service
	Call      *sqladmin.InstancesListCall
	ProjectID string
	Error     error
}

func SQL(ctx context.Context, projectID string) *SQLCall {
	s, err := sqladmin.NewService(ctx)
	if err != nil {
		return &SQLCall{Error: err}
	}

	return &SQLCall{
		Service:   s,
		ProjectID: projectID,
		Call:      sqladmin.NewInstancesService(s).List(projectID),
	}
}

func (r *SQLCall) Filter(labelName string, flag bool) *SQLCall {
	if r.Error != nil {
		return r
	}
	return &SQLCall{
		ProjectID: r.ProjectID,
		Call:      r.Call.Filter("userLabels." + labelName + "=true"),
	}
}

func (r *SQLCall) Do(ctx context.Context, interval time.Duration) (*model.ShutdownReport, error) {
	if r.Error != nil {
		return nil, r.Error
	}

	targets, err := r.Call.Do()
	if err != nil {
		return nil, err
	}

	var res = r.Error
	var doneRes []string
	var alreadyRes []string

	for _, instance := range targets.Items {
		// do not change replica instance's activation policy
		if instance.InstanceType == "READ_REPLICA_INSTANCE" {
			continue
		}

		// do not change instance's activation policy which is already "NEVER"
		if instance.Settings.ActivationPolicy == "NEVER" {
			alreadyRes = append(alreadyRes, instance.Name)
			continue
		}

		// stop the target instance
		instance.Settings.ActivationPolicy = "NEVER"

		// apply the settings
		_, err := sqladmin.NewInstancesService(r.Service).Patch(r.ProjectID, instance.Name, instance).Do()
		if err != nil {
			res = multierror.Append(res, err)
		}
		doneRes = append(doneRes, instance.Name)
		time.Sleep(interval)
	}

	log.Printf("Success in stopping SQL instances: Done.")

	return &model.ShutdownReport{
		InstanceType:             model.SQL,
		DoneResources:            doneRes,
		AlreadyShutdownResources: alreadyRes,
	}, res
}
