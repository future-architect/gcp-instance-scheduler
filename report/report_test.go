package report

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/future-architect/gcp-instance-scheduler/model"
)

var input = `[
		{"InstanceType": "ComputeEngine", "DoneResources": [ "standard-1", "standard-2" ], "AlreadyShutdownResources": [ ], "SkipResource": [ "standard-5" ]},
		{"InstanceType": "SQL", "DoneResources": [], "AlreadyShutdownResources": [], "SkipResource": []},
		{"InstanceType": "InstanceGroup", "DoneResources": [ "gke-cluster-1" ], "AlreadyShutdownResources": [ "gke-cluster-2", "gke-cluster-3", "instance-group-1" ], "SkipResource": []},
		{"InstanceType": "SQL", "DoneResources": [ "sql-instance-1" ], "AlreadyShutdownResources": [], "SkipResource": [ "sql-instance-2" ]}
	]`

func TestCount(t *testing.T) {
	var shutdownReport []model.Report
	if err := json.Unmarshal([]byte(input), &shutdownReport); err != nil {
		t.Fatal("test data is invalid: parse failed:", err)
	}

	var execReport []*model.Report
	for i := 0; i < len(shutdownReport); i++ {
		execReport = append(execReport, &shutdownReport[i])
	}

	report := NewResourceCountReport(execReport, "dev-pj")

	if report.Project != "dev-pj" {
		t.Fatalf("project name is expected %s, got=%s", "dev-jp", report.Project)
	}
	// sort test
	if report.InstanceCountList[0].InstanceType != "ComputeEngine" &&
		report.InstanceCountList[1].InstanceType != "InstanceGroup" &&
		report.InstanceCountList[2].InstanceType != "SQL" {
		t.Fatal("sort test error", report.InstanceCountList[0].InstanceType, report.InstanceCountList[1].InstanceType, report.InstanceCountList[2].InstanceType)
	}

	if report.InstanceCountList[0].InstanceType != "ComputeEngine" &&
		report.InstanceCountList[0].DoneCount != 2 &&
		report.InstanceCountList[0].AlreadyCount != 0 &&
		report.InstanceCountList[0].SkipCount != 1 {
		t.Fatalf("cout value test error: InstanceType: %s, DoneCount: %d, AlreadyCount: %d, SkipCount: %d\n", report.InstanceCountList[0].InstanceType, report.InstanceCountList[0].DoneCount, report.InstanceCountList[0].AlreadyCount, report.InstanceCountList[0].SkipCount)
	}

	if report.InstanceCountList[1].InstanceType != "InstanceGroup" &&
		report.InstanceCountList[1].DoneCount != 1 &&
		report.InstanceCountList[1].AlreadyCount != 3 &&
		report.InstanceCountList[1].SkipCount != 0 {
		t.Fatalf("cout value test error: InstanceType: %s, DoneCount: %d, AlreadyCount: %d, SkipCount: %d\n", report.InstanceCountList[1].InstanceType, report.InstanceCountList[1].DoneCount, report.InstanceCountList[1].AlreadyCount, report.InstanceCountList[1].SkipCount)
	}

	if report.InstanceCountList[2].InstanceType != "SQL" &&
		report.InstanceCountList[2].DoneCount != 0 &&
		report.InstanceCountList[2].AlreadyCount != 0 &&
		report.InstanceCountList[2].SkipCount != 0 {
		t.Fatalf("cout value test error: InstanceType: %s, DoneCount: %d, AlreadyCount: %d, SkipCount: %d\n", report.InstanceCountList[2].InstanceType, report.InstanceCountList[2].DoneCount, report.InstanceCountList[2].AlreadyCount, report.InstanceCountList[2].SkipCount)
	}

	fmt.Printf("%+v", report)
}

func TestDetail(t *testing.T) {
	var shutdownReport []model.Report
	if err := json.Unmarshal([]byte(input), &shutdownReport); err != nil {
		t.Fatal("test data is invalid parse failed:", err)
	}

	var execReport []*model.Report
	for i := 0; i < len(shutdownReport); i++ {
		execReport = append(execReport, &shutdownReport[i])
	}

	report := NewDetailReports(execReport)

	if len(report) != 4 {
		t.Fatalf("report size is expected %d, got=%d", 3, len(report))
	}

	if report[0].InstanceType != "ComputeEngine" &&
		report[0].DoneResources[0] != "standard-1" &&
		report[0].DoneResources[1] != "standard-2" &&
		report[0].SkipResources[0] != "standard-5" {
		t.Fatalf("Detail report test error: DoneResources[0]: %s, DoneResources[1]: %s, SkipResources[0]: %s", report[0].DoneResources[0], report[0].DoneResources[1], report[0].SkipResources[0])
	}

	if report[3].InstanceType != "SQL" &&
		report[3].DoneResources[0] != "sql-instance-1" &&
		report[3].SkipResources[0] != "sql-instance-2" {
		t.Fatalf("Detail report test error: DoneResources[3]: %s, DoneResources[1]: %s, SkipResources[3]: %s", report[3].DoneResources[3], report[3].DoneResources[1], report[3].SkipResources[3])
	}

	// sort test
	if report[0].InstanceType != "ComputeEngine" &&
		report[1].InstanceType != "InstanceGroup" &&
		report[2].InstanceType != "SQL" {
		t.Fatal("sort test error", report[0].InstanceType, report[1].InstanceType, report[2].InstanceType)
	}

	fmt.Printf("%+v\n", report)
}
