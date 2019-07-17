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
	"github.com/future-architect/gcp-instance-scheduler/report"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/net/context"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/container/v1"
)

type InstanceGroupListCall struct {
	InstanceGroupList *compute.InstanceGroupManagerAggregatedList
	TemplateListCall  *compute.InstanceTemplatesListCall
	ProjectID         string
	Error             error
}

type InstanceGroupShutdownCall struct {
	InstanceGroupList *compute.InstanceGroupManagerAggregatedList
	TemplateList      *compute.InstanceTemplateList
	TargetLabel       string
	ProjectID         string
	Error             error
}

// search templates name in Set of instance group name
func contains(instanceGroup set.Set, instanceTemplateName string) bool {
	return instanceGroup.Contains(instanceTemplateName)
}

// create instance group manager list
func valuesIG(m map[string]compute.InstanceGroupManagersScopedList) []*compute.InstanceGroupManager {
	var res []*compute.InstanceGroupManager
	if len(m) == 0 {
		return nil
	}
	for _, managerList := range m {
		res = append(res, managerList.InstanceGroupManagers...)
	}
	return res
}

// grep target cluster and create target cluster list
func filter(l []*container.Cluster, label string, value string) []*container.Cluster {
	var res []*container.Cluster
	for _, cluster := range l {
		if cluster.ResourceLabels[label] == value {
			res = append(res, cluster)
		}
	}
	return res
}

// get target GKE instance group Set
func getGKEInstanceGroup(ctx context.Context, targetLabel string, projectID string) (set.Set, error) {
	// create service to operate container
	s, err := container.NewService(ctx)
	if err != nil {
		return nil, err
	}

	// get all clusters list
	clusters, err := container.NewProjectsLocationsClustersService(s).List("projects/" + projectID + "/locations/-").Do()
	if err != nil {
		return nil, err
	}

	// filtering with target label
	scheduled := filter(clusters.Clusters, targetLabel, "true")
	res := set.NewSet()
	for _, cluster := range scheduled {
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

func InstanceGroupResource(ctx context.Context, projectID string) *InstanceGroupListCall {
	// reporting error list
	var res error

	// create service to operate instances
	s, err := compute.NewService(ctx)
	if err != nil {
		res = multierror.Append(res, err)
	}

	// get all instance group mangers list
	managerList, err := compute.NewInstanceGroupManagersService(s).AggregatedList(projectID).Do()
	if err != nil {
		res = multierror.Append(res, err)
	}

	// get all templates list
	list := compute.NewInstanceTemplatesService(s).List(projectID)

	return &InstanceGroupListCall{
		TemplateListCall:  list,
		InstanceGroupList: managerList,
		ProjectID:         projectID,
		Error:             res,
	}
}

func (r *InstanceGroupListCall) FilterLabel(targetLabel string, flag bool) *InstanceGroupShutdownCall {
	// reporting error list
	var res = r.Error

	templateList, err := r.TemplateListCall.Filter("properties.labels." + targetLabel + "=true").Do()
	if err != nil {
		res = multierror.Append(res, err)
	}

	return &InstanceGroupShutdownCall{
		TemplateList:      templateList,
		InstanceGroupList: r.InstanceGroupList,
		TargetLabel:       targetLabel,
		ProjectID:         r.ProjectID,
		Error:             res,
	}
}

func (r *InstanceGroupShutdownCall) ShutdownWithInterval(ctx context.Context, interval time.Duration) (*report.ShutdownReport, error) {
	var res = r.Error
	var doneRes []string
	var alreadyRes []string

	// create service to operate instances
	s, err := compute.NewService(ctx)
	if err != nil {
		res = multierror.Append(res, err)
	}

	// instance group manager service
	ms := compute.NewInstanceGroupManagersService(s)

	// error bundle before executing stop call
	if res != nil {
		return nil, res
	}

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
		}

		// add instance group name to Set
		for _, template := range r.TemplateList.Items {
			instanceGroupSet.Add(template.Name)
		}

		// compare filtered instance template name and manager which is created by template
		if contains(instanceGroupSet, managerTemplate) {
			if !manager.Status.IsStable {
				continue
			}
			if manager.TargetSize == 0 {
				alreadyRes = append(alreadyRes, manager.Name)
				continue
			}
			if _, err := ms.Resize(r.ProjectID, zone, manager.Name, 0).Do(); err != nil {
				res = multierror.Append(res, err)
			}
			doneRes = append(doneRes, manager.Name)
		}
		time.Sleep(interval)
	}
	log.Printf("Success in stopping InstanceGroup: Done.")

	return &report.ShutdownReport{
		ShutdownReport: model.ShutdownReport{
			InstanceType:             report.InstanceGroup,
			DoneResources:            doneRes,
			AlreadyShutdownResources: alreadyRes,
		},
	}, res
}
