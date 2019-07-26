package notice

import (
	"fmt"
	"reflect"
	"time"

	"github.com/future-architect/gcp-instance-scheduler/model"
	"github.com/future-architect/gcp-instance-scheduler/report"
	"github.com/nlopes/slack"
)

type slackNotifier struct {
	slackAPIToken string
	slackChannel  string
}

func getDate() string {
	year, month, day := time.Now().Date()
	return fmt.Sprintf("%d/%d/%d", year, month, day)
}

// return struct
func getFieldNameList(e interface{}) []string {
	var fieldName []string

	fieldVals := reflect.Indirect(reflect.ValueOf(e))
	for i := 0; i < fieldVals.NumField(); i++ {
		fieldName = append(fieldName, fieldVals.Type().Field(i).Name)
	}

	return fieldName
}

// return field value greped by name
func getFieldValue(e interface{}, field string) interface{} {
	values := reflect.Indirect(reflect.ValueOf(e))
	return values.FieldByName(field).Interface()
}

func createHeader(pad map[string]int, project string) string {
	text := fmt.Sprintf("[Project: %s] Instances Shutdown Report <%s>\n", project, getDate())

	fieldList := getFieldNameList(&model.ShutdownReport{})
	for i := 0; i < len(fieldList); i++ {
		fieldName := fieldList[i]
		if i == len(fieldList)-1 {
			text += fmt.Sprintf("%*v\n", pad[fieldName], fieldName)
			break
		}
		text += fmt.Sprintf("%*v | ", pad[fieldName], fieldName)
	}

	text += fmt.Sprintf("-------------------------------------------------------------------------------\n")

	return text
}

func createHeaderDetail(instanceType string, pad int) string {
	text := fmt.Sprintf("[ %s ]\n", instanceType)
	text += fmt.Sprintf("%*s | %s\n", pad, "Status", "Instance Name")
	text += fmt.Sprintf("-------------------------------------------------------------------------------\n")
	return text
}

func NewSlackNotifier(slackAPIToken, slackChannel string) *slackNotifier {
	return &slackNotifier{
		slackAPIToken: slackAPIToken,
		slackChannel:  slackChannel,
	}
}

func (n *slackNotifier) postInline(text string) (string, error) {
	_, ts, err := slack.New(n.slackAPIToken).PostMessage(
		n.slackChannel,
		slack.MsgOptionText("```"+text+"```", false),
	)

	return ts, err
}

func (n *slackNotifier) postThreadInline(text, ts string) error {
	_, _, err := slack.New(n.slackAPIToken).PostMessage(
		n.slackChannel,
		slack.MsgOptionText("```"+text+"```", false),
		slack.MsgOptionTS(ts),
	)

	return err
}

// return timestamp to make thread bellow this message
func (n *slackNotifier) PostReport(report *report.ResourceCountReport) (string, error) {
	pad := report.Padding
	text := createHeader(pad, report.Project)

	for _, resourceResult := range report.InstanceCountList {
		text += fmt.Sprintf("%*s | %*d | %*d | %*d",
			pad["InstanceType"], resourceResult.InstanceType,
			pad["DoneResources"], resourceResult.DoneCount,
			pad["AlreadyShutdownResources"], resourceResult.AlreadyCount,
			pad["SkipResources"], resourceResult.SkipCount)
		text += "\n"
	}

	return n.postInline(text)
}

func (n *slackNotifier) PostReportThread(parentTS string, report *report.DetailReport) error {
	// align to left
	pad := -25

	text := createHeaderDetail(report.InstanceType, pad)

	// field names of model.ShutdownReport
	statusType := getFieldNameList(*report)

	for i := 1; i < len(statusType); i++ {
		status := statusType[i]
		// pick up instance value from field name
		for _, resource := range getFieldValue(*report, status).([]string) {
			text += fmt.Sprintf("%*s | %s\n", pad, status, resource)
		}
	}

	return n.postThreadInline(text, parentTS)
}
