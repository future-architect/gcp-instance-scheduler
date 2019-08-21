package report

import (
	"sort"
	"strings"

	"github.com/future-architect/gcp-instance-scheduler/model"
)

type InstanceCount struct {
	InstanceType string
	DoneCount    int
	AlreadyCount int
	SkipCount    int
}

type ResourceCountReport struct {
	Project           string
	InstanceCountList []InstanceCount
	Padding           map[string]int
}

func NewResourceCountReport(result []*model.Report, projectID string) *ResourceCountReport {
	var icList []InstanceCount

	for _, r := range result {
		icList = append(icList, InstanceCount{
			InstanceType: r.InstanceType,
			DoneCount:    len(r.DoneResources),
			AlreadyCount: len(r.AlreadyShutdownResources),
			SkipCount:    len(r.SkipResources),
		})
	}

	sort.Slice(icList,
		func(i, j int) bool {
			return strings.ToLower(icList[i].InstanceType) < strings.ToLower(icList[j].InstanceType)
		})

	pad := map[string]int{
		"InstanceType":             -15,
		"DoneResources":            -15,
		"AlreadyShutdownResources": -25,
		"SkipResources":            -15,
	}

	return &ResourceCountReport{
		Project:           projectID,
		InstanceCountList: icList,
		Padding:           pad,
	}
}
