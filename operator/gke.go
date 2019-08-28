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
	"log"
	"strconv"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/api/container/v1"
)

func ShowClusterStatus(ctx context.Context, projectID string, targetLabel string) error {
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
	scheduled := filter(clusters.Clusters, targetLabel, "true")
	for _, cluster := range scheduled {
		log.Printf("Cluster [Name]%v [Status]%v", cluster.Name, cluster.Status)
	}

	return nil
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
func GetOriginalNodePoolSize(ctx context.Context, projectID string, targetLabel string) (map[string]int64, error) {
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

	for _, cluster := range filter(clusters.Clusters, targetLabel, "true") {
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
