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
	// create service to operate container
	s, err := container.NewService(ctx)
	if err != nil {
		log.Printf("Could not create service: %v", err)
		return err
	}

	// get all clusters list
	clusters, err := container.NewProjectsLocationsClustersService(s).List("projects/" + projectID + "/locations/-").Do()
	if err != nil {
		log.Printf("Could not get GKE clusters list: %v", err)
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
	// create service to operate container
	s, err := container.NewService(ctx)
	if err != nil {
		log.Printf("Could not create service: %v", err)
		return err
	}

	// get all clusters list
	clusters, err := container.NewProjectsLocationsClustersService(s).List("projects/" + projectID + "/locations/-").Do()
	if err != nil {
		log.Printf("Could not get GKE clusters list: %v", err)
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
				ResourceLabels: make(map[string]string),
			}

			for key, value := range labels {
				rb.ResourceLabels[key] = value
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
