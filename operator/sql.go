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
	"time"

	"github.com/future-architect/gcp-instance-scheduler/model"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/net/context"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

type SQLCall struct {
	s         *sqladmin.Service
	call      *sqladmin.InstancesListCall
	projectID string
	error     error
}

func SQL(ctx context.Context, projectID string) *SQLCall {
	s, err := sqladmin.NewService(ctx)
	if err != nil {
		return &SQLCall{error: err}
	}

	return &SQLCall{
		s:         s,
		projectID: projectID,
		call:      sqladmin.NewInstancesService(s).List(projectID),
	}
}

func (r *SQLCall) Filter(labelName string, flag bool) *SQLCall {
	if r.error != nil {
		return r
	}
	r.call = r.call.Filter("userLabels." + labelName + "=true")
	return r
}

func (r *SQLCall) Stop() (*model.Report, error) {
	if r.error != nil {
		return nil, r.error
	}

	targets, err := r.call.Do()
	if err != nil {
		return nil, err
	}

	var res = r.error
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

		// update policy
		instance.Settings.ActivationPolicy = "NEVER"

		// apply the settings
		_, err := sqladmin.NewInstancesService(r.s).Patch(r.projectID, instance.Name, instance).Do()
		if err != nil {
			res = multierror.Append(res, err)
		}
		doneRes = append(doneRes, instance.Name)
		time.Sleep(CallInterval)
	}

	return &model.Report{
		InstanceType: model.SQL,
		Dones:        doneRes,
		Alreadies:    alreadyRes,
	}, res
}

func (r *SQLCall) Start() (*model.Report, error) {
	if r.error != nil {
		return nil, r.error
	}

	targets, err := r.call.Do()
	if err != nil {
		return nil, err
	}

	var res = r.error
	var doneRes []string
	var alreadyRes []string

	for _, instance := range targets.Items {
		// do not change replica instance's activation policy
		if instance.InstanceType == "READ_REPLICA_INSTANCE" {
			continue
		}

		// do not change instance's activation policy which is already "ALWAYS"
		if instance.Settings.ActivationPolicy == "ALWAYS" {
			alreadyRes = append(alreadyRes, instance.Name)
			continue
		}

		// Update policy
		instance.Settings.ActivationPolicy = "ALWAYS"

		// apply the settings
		_, err := sqladmin.NewInstancesService(r.s).Patch(r.projectID, instance.Name, instance).Do()
		if err != nil {
			res = multierror.Append(res, err)
		}
		doneRes = append(doneRes, instance.Name)
		time.Sleep(CallInterval)
	}

	return &model.Report{
		InstanceType: model.SQL,
		Dones:        doneRes,
		Alreadies:    alreadyRes,
	}, res
}
