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
	set "github.com/deckarep/golang-set"
	"github.com/future-architect/gcp-instance-scheduler/model"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/net/context"
	"google.golang.org/api/compute/v1"
	"strings"
	"time"
)

type InstanceGroupCall struct {
	instanceGroupList *compute.InstanceGroupManagerAggregatedList
	templateListCall  *compute.InstanceTemplatesListCall
	targetLabel       string
	targetLabelValue  string
	projectID         string
	error             error
	s                 *compute.Service
	ctx               context.Context
}

func InstanceGroup(ctx context.Context, projectID string) *InstanceGroupCall {
	s, err := compute.NewService(ctx)
	if err != nil {
		return &InstanceGroupCall{error: err}
	}

	// get all instance group managers list
	managerList, err := compute.NewInstanceGroupManagersService(s).AggregatedList(projectID).Do()
	if err != nil {
		return &InstanceGroupCall{error: err}
	}

	// get all templates list
	return &InstanceGroupCall{
		s:                 s,
		templateListCall:  compute.NewInstanceTemplatesService(s).List(projectID),
		instanceGroupList: managerList,
		projectID:         projectID,
		ctx:               ctx,
	}
}

func (r *InstanceGroupCall) Filter(labelName, value string) *InstanceGroupCall {
	if r.error != nil {
		return r
	}
	r.targetLabel = labelName
	r.targetLabelValue = value
	r.templateListCall = r.templateListCall.Filter("properties.labels." + labelName + "=" + value)
	return r
}

func (r *InstanceGroupCall) Resize(size int64) (*model.Report, error) {
	if r.error != nil {
		return nil, r.error
	}

	templateList, err := r.templateListCall.Do()
	if err != nil {
		return nil, err
	}

	var res = r.error
	var doneRes []string
	var alreadyRes []string

	for _, manager := range valuesIG(r.instanceGroupList.Items) {
		// get manager zone name
		zoneUrlElements := strings.Split(manager.Zone, "/")
		zone := zoneUrlElements[len(zoneUrlElements)-1]

		// get manager's template name
		tmpUrlElements := strings.Split(manager.InstanceTemplate, "/")
		managerTemplate := tmpUrlElements[len(tmpUrlElements)-1]

		// add instance group name to Set
		instanceGroupSet := set.NewSet()
		for _, t := range templateList.Items {
			instanceGroupSet.Add(t.Name)
		}

		// compare filtered instance template name and manager which is created by template
		if instanceGroupSet.Contains(managerTemplate) {
			if !manager.Status.IsStable {
				continue
			}

			if manager.TargetSize == 0 {
				alreadyRes = append(alreadyRes, manager.Name)
				continue
			}

			ms := compute.NewInstanceGroupManagersService(r.s)
			if _, err := ms.Resize(r.projectID, zone, manager.Name, size).Do(); err != nil {
				res = multierror.Append(res, err)
				continue
			}
			doneRes = append(doneRes, manager.Name)
		}

		time.Sleep(CallInterval)
	}

	return &model.Report{
		InstanceType: model.InstanceGroup,
		Dones:        doneRes,
		Alreadies:    alreadyRes,
	}, res
}

func (r *InstanceGroupCall) Recovery() (*model.Report, error) {
	if r.error != nil {
		return nil, r.error
	}

	templateList, err := r.templateListCall.Do()
	if err != nil {
		return nil, err
	}

	// add instance group name to Set
	// NodePool that belong to GKE which has target label
	targetInstanceGroupSet := set.NewSet()
	for _, t := range templateList.Items {
		targetInstanceGroupSet.Add(t.Name)
	}

	sizeMap, err := GetOriginalNodePoolSize(r.ctx, r.projectID, r.targetLabel, r.targetLabelValue)
	if err != nil {
		return nil, err
	}

	var res = r.error
	var doneRes []string
	var alreadyRes []string

	for _, manager := range valuesIG(r.instanceGroupList.Items) {
		// get manager zone name
		zoneUrlElements := strings.Split(manager.Zone, "/")
		zone := zoneUrlElements[len(zoneUrlElements)-1] // ex) us-central1-a

		// get manager's template name
		tmpUrlElements := strings.Split(manager.InstanceTemplate, "/")
		instanceTemplateName := tmpUrlElements[len(tmpUrlElements)-1] // ex) gke-standard-cluster-1-default-pool-f789c8df

		// compare filtered instance template name and manager which is created by template
		if targetInstanceGroupSet.Contains(instanceTemplateName) {
			if !manager.Status.IsStable {
				continue
			}

			split := strings.Split(manager.InstanceGroup, "/")
			instanceGroupName := split[len(split)-1]

			originalSize := sizeMap[instanceGroupName]

			if manager.TargetSize == originalSize {
				alreadyRes = append(alreadyRes, manager.Name)
				continue
			}

			ms := compute.NewInstanceGroupManagersService(r.s)
			if _, err := ms.Resize(r.projectID, zone, manager.Name, originalSize).Do(); err != nil {
				res = multierror.Append(res, err)
				continue
			}
			doneRes = append(doneRes, manager.Name)
		}

		time.Sleep(CallInterval)
	}

	return &model.Report{
		InstanceType: model.InstanceGroup,
		Dones:        doneRes,
		Alreadies:    alreadyRes,
	}, res
}

// create instance group manager list
func valuesIG(m map[string]compute.InstanceGroupManagersScopedList) []*compute.InstanceGroupManager {
	var res []*compute.InstanceGroupManager
	for _, managerList := range m {
		if len(managerList.InstanceGroupManagers) == 0 {
			continue
		}
		res = append(res, managerList.InstanceGroupManagers...)
	}
	return res
}
