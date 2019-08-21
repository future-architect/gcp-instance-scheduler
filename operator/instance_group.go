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
	"strings"
	"time"

	set "github.com/deckarep/golang-set"
	"github.com/future-architect/gcp-instance-scheduler/model"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/net/context"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/container/v1"
)

type InstanceGroupCall struct {
	Service           *compute.Service
	InstanceGroupList *compute.InstanceGroupManagerAggregatedList
	TemplateListCall  *compute.InstanceTemplatesListCall
	TargetLabel       string
	ProjectID         string
	Error             error
}

func InstanceGroup(ctx context.Context, projectID string) *InstanceGroupCall {
	s, err := compute.NewService(ctx)
	if err != nil {
		return &InstanceGroupCall{Error: err}
	}

	// get all instance group mangers list
	managerList, err := compute.NewInstanceGroupManagersService(s).AggregatedList(projectID).Do()
	if err != nil {
		return &InstanceGroupCall{Error: err}
	}

	// get all templates list
	return &InstanceGroupCall{
		Service:           s,
		TemplateListCall:  compute.NewInstanceTemplatesService(s).List(projectID),
		InstanceGroupList: managerList,
		ProjectID:         projectID,
	}
}

func (r *InstanceGroupCall) Filter(labelName string, flag bool) *InstanceGroupCall {
	if r.Error != nil {
		return r
	}

	return &InstanceGroupCall{
		TemplateListCall:  r.TemplateListCall.Filter("properties.labels." + labelName + "=true"),
		InstanceGroupList: r.InstanceGroupList,
		TargetLabel:       labelName,
		ProjectID:         r.ProjectID,
	}
}

func (r *InstanceGroupCall) Do(ctx context.Context, interval time.Duration) (*model.ShutdownReport, error) {
	if r.Error != nil {
		return nil, r.Error
	}

	templateList, err := r.TemplateListCall.Do()
	if err != nil {
		return nil, err
	}

	var res = r.Error
	var doneRes []string
	var alreadyRes []string

	for _, manager := range valuesIG(r.InstanceGroupList.Items) {
		// get manager zone name
		zoneUrlElements := strings.Split(manager.Zone, "/")
		zone := zoneUrlElements[len(zoneUrlElements)-1]

		// get manager's template name
		tmpUrlElements := strings.Split(manager.InstanceTemplate, "/")
		managerTemplate := tmpUrlElements[len(tmpUrlElements)-1]

		// add instance group name of cluster node pool to Set
		instanceGroupSet, err := getGKEInstanceGroup(ctx, r.TargetLabel, r.ProjectID)
		if err != nil {
			res = multierror.Append(res, err)
			continue
		}

		// add instance group name to Set
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

			ms := compute.NewInstanceGroupManagersService(r.Service)
			if _, err := ms.Resize(r.ProjectID, zone, manager.Name, 0).Do(); err != nil {
				res = multierror.Append(res, err)
				continue
			}
			doneRes = append(doneRes, manager.Name)
		}

		time.Sleep(interval)
	}
	log.Printf("Success in stopping InstanceGroup: Done.")

	return &model.ShutdownReport{
		InstanceType:             model.InstanceGroup,
		DoneResources:            doneRes,
		AlreadyShutdownResources: alreadyRes,
	}, res
}

// get target GKE instance group Set
func getGKEInstanceGroup(ctx context.Context, targetLabel, projectID string) (set.Set, error) {
	s, err := container.NewService(ctx)
	if err != nil {
		return nil, err
	}

	// get all clusters list
	clusters, err := container.NewProjectsLocationsClustersService(s).List("projects/" + projectID + "/locations/-").Do()
	if err != nil {
		return nil, err
	}

	res := set.NewSet()
	for _, cluster := range filter(clusters.Clusters, targetLabel, "true") {
		for _, nodePool := range cluster.NodePools {
			for _, gkeInstanceGroup := range nodePool.InstanceGroupUrls {
				tmpUrlElements := strings.Split(gkeInstanceGroup, "/")
				managerTemplate := tmpUrlElements[len(tmpUrlElements)-1]
				// remove suffix(*-grp)
				res.Add(managerTemplate[:len(managerTemplate)-4])
			}
		}
	}
	return res, nil
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

// grep target cluster and create target cluster list
func filter(l []*container.Cluster, label string, value string) []*container.Cluster {
	if label == "" { //TODO Temp impl
		return l
	}

	var res []*container.Cluster
	for _, cluster := range l {
		if cluster.ResourceLabels[label] == value {
			res = append(res, cluster)
		}
	}
	return res
}
