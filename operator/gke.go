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

func SetLabelNodePoolSize(ctx context.Context, projectID string, targetLabel string, interval time.Duration) error {
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

		parseRegion := strings.Split(cluster.Location, "/")
		region := parseRegion[len(parseRegion)-1]
		labels := cluster.ResourceLabels

		for _, nodePool := range cluster.NodePools {
			name := "projects/" + projectID + "/locations/" + region + "/clusters/" + cluster.Name
			nodeSizeLabel := "restore-size-" + nodePool.Name
			labels[nodeSizeLabel] = strconv.FormatInt(nodePool.InitialNodeCount, 10)

			rb := &container.SetLabelsRequest{
				ResourceLabels: labels,
			}

			_, err := container.NewProjectsLocationsClustersService(s).SetResourceLabels(name, rb).Do()
			if err != nil {
				log.Printf("Could not add lable to cluster: %v", err)
				return err
			}
			time.Sleep(interval)
		}
	}

	return nil
}
