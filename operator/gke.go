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
	"fmt"
	set "github.com/deckarep/golang-set"
	"github.com/future-architect/gcp-instance-scheduler/model"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/net/context"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/container/v1"
	"strconv"
	"strings"
	"time"
)

type GKENodePoolCall struct {
	targetLabel      string
	projectID        string
	error            error
	s                *compute.Service
	ctx              context.Context
	targetLabelValue string
}

func GKENodePool(ctx context.Context, projectID string) *GKENodePoolCall {
	s, err := compute.NewService(ctx)
	if err != nil {
		return &GKENodePoolCall{error: err}
	}

	// get all templates list
	return &GKENodePoolCall{
		s:         s,
		projectID: projectID,
		ctx:       ctx,
	}
}

func (r *GKENodePoolCall) Filter(labelName, value string) *GKENodePoolCall {
	if r.error != nil {
		return r
	}
	r.targetLabel = labelName
	r.targetLabelValue = value
	return r
}

func (r *GKENodePoolCall) Resize(size int64) (*model.Report, error) {
	if r.error != nil {
		return nil, r.error
	}

	// get all instance group mangers list
	managerList, err := compute.NewInstanceGroupManagersService(r.s).AggregatedList(r.projectID).Do()
	if err != nil {
		return nil, err
	}

	// add instance group name of cluster node pool to Set
	gkeNodePoolInstanceGroupSet, err := r.getGKEInstanceGroup()
	if err != nil {
		return nil, err
	}

	fmt.Println("gkeNodePoolInstanceGroupSet:", gkeNodePoolInstanceGroupSet.ToSlice())

	var res = r.error
	var alreadyRes []string
	var doneRes []string

	for _, manager := range valuesIG(managerList.Items) {

		fmt.Println("manager.InstanceTemplate:", manager.InstanceTemplate)
		fmt.Println("manager.Name:", manager.Name)

		// Check GKE NodePool InstanceGroup
		if gkeNodePoolInstanceGroupSet.Contains(manager.Name) {
			if !manager.Status.IsStable {
				continue
			}

			if manager.TargetSize == size {
				alreadyRes = append(alreadyRes, manager.Name)
				continue
			}

			// get manager zone name
			zoneUrlElements := strings.Split(manager.Zone, "/")
			zone := zoneUrlElements[len(zoneUrlElements)-1]

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
		InstanceType: model.GKENodePool,
		Dones:        doneRes,
		Alreadies:    alreadyRes,
	}, res
}

func (r *GKENodePoolCall) Recovery() (*model.Report, error) {
	if r.error != nil {
		return nil, r.error
	}

	managerList, err := compute.NewInstanceGroupManagersService(r.s).AggregatedList(r.projectID).Do()
	if err != nil {
		return nil, err
	}

	// add instance group name of cluster node pool to Set
	gkeNodePoolInstanceGroupSet, err := r.getGKEInstanceGroup()
	if err != nil {
		return nil, err
	}

	sizeMap, err := GetOriginalNodePoolSize(r.ctx, r.projectID, r.targetLabel, r.targetLabelValue)
	if err != nil {
		return nil, err
	}

	var res = r.error
	var doneRes []string
	var alreadyRes []string

	for _, manager := range valuesIG(managerList.Items) {

		// check instance group of gke node pool
		if gkeNodePoolInstanceGroupSet.Contains(manager.Name) {
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

			// get manager zone name
			zoneUrlElements := strings.Split(manager.Zone, "/")
			zone := zoneUrlElements[len(zoneUrlElements)-1] // ex) us-central1-a

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
		InstanceType: model.GKENodePool,
		Dones:        doneRes,
		Alreadies:    alreadyRes,
	}, res
}

// get target GKE instance group Set
func (r *GKENodePoolCall) getGKEInstanceGroup() (set.Set, error) {
	s, err := container.NewService(r.ctx)
	if err != nil {
		return nil, err
	}

	// get all clusters list
	clusters, err := container.NewProjectsLocationsClustersService(s).List("projects/" + r.projectID + "/locations/-").Do()
	if err != nil {
		return nil, err
	}

	res := set.NewSet()
	for _, cluster := range filter(clusters.Clusters, r.targetLabel, r.targetLabelValue) {
		for _, nodePool := range cluster.NodePools {
			for _, gkeInstanceGroup := range nodePool.InstanceGroupUrls {
				tmpUrlElements := strings.Split(gkeInstanceGroup, "/")
				managerTemplate := tmpUrlElements[len(tmpUrlElements)-1]
				res.Add(managerTemplate) // e.g. gke-tky-cluster-default-pool-cb765a7d-grp
			}
		}
	}
	return res, nil
}

func SetLabelNodePoolSize(ctx context.Context, projectID string, targetLabel string) error {
	s, err := container.NewService(ctx)
	if err != nil {
		return err
	}

	// get all clusters list
	clusters, err := container.NewProjectsLocationsClustersService(s).List("projects/" + projectID + "/locations/-").Do()
	if err != nil {
		return err
	}

	// filtering with target label
	for _, cluster := range filter(clusters.Clusters, targetLabel, "true") {
		labels := cluster.ResourceLabels
		for _, nodePool := range cluster.NodePools {
			nodeSizeLabel := "restore-size-" + nodePool.Name
			labels[nodeSizeLabel] = strconv.FormatInt(nodePool.InitialNodeCount, 10)
		}

		parseRegion := strings.Split(cluster.Location, "/")
		region := parseRegion[len(parseRegion)-1]
		name := "projects/" + projectID + "/locations/" + region + "/clusters/" + cluster.Name
		req := &container.SetLabelsRequest{
			ResourceLabels: labels,
		}

		_, err := container.NewProjectsLocationsClustersService(s).SetResourceLabels(name, req).Do()
		if err != nil {
			return err
		}
	}

	return nil
}

// GetOriginalNodePoolSize returns map that key=instanceGroupName and value=originalSize
func GetOriginalNodePoolSize(ctx context.Context, projectID, targetLabel, labelValue string) (map[string]int64, error) {
	s, err := container.NewService(ctx)
	if err != nil {
		return nil, err
	}

	// get all clusters list
	clusters, err := container.NewProjectsLocationsClustersService(s).List("projects/" + projectID + "/locations/-").Do()
	if err != nil {
		return nil, err
	}

	result := make(map[string]int64)

	for _, cluster := range filter(clusters.Clusters, targetLabel, labelValue) {
		labels := cluster.ResourceLabels
		for _, nodePool := range cluster.NodePools {
			restoreSize, ok := labels["restore-size-"+nodePool.Name]
			if !ok {
				continue
			}

			size, err := strconv.Atoi(restoreSize)
			if err != nil {
				return nil, errors.New("label: " + "restore-size-" + nodePool.Name + " value is not number format?")
			}

			for _, url := range nodePool.InstanceGroupUrls {
				// u;rl is below format
				// e.g. https://www.googleapis.com/compute/v1/projects/{ProjectID}/zones/us-central1-a/instanceGroupManagers/gke-standard-cluster-1-default-pool-1234abcd-grp
				urlSplit := strings.Split(url, "/")
				instanceGroupName := urlSplit[len(urlSplit)-1]
				result[instanceGroupName] = int64(size)
			}
		}
	}

	return result, nil
}

// grep target cluster and create target cluster list
func filter(l []*container.Cluster, label, value string) []*container.Cluster {
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
