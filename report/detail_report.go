package report

import (
	"sort"
	"strings"

	"github.com/future-architect/gcp-instance-scheduler/model"
)

type DetailReport struct {
	InstanceType             string
	DoneResources            []string
	AlreadyShutdownResources []string
	SkipResources            []string
}

func NewDetailReports(result []*model.Report) []*DetailReport {
	var drList []*DetailReport

	for _, r := range result {
		drList = append(drList, &DetailReport{
			InstanceType:             r.InstanceType,
			DoneResources:            r.DoneResources,
			AlreadyShutdownResources: r.AlreadyShutdownResources,
			SkipResources:            r.SkipResources,
		})
	}

	sort.Slice(drList,
		func(i, j int) bool {
			return strings.ToLower((*drList[i]).InstanceType) < strings.ToLower((*drList[j]).InstanceType)
		})

	return drList
}
