package notice

import (
	"fmt"
	"reflect"
	"time"

	"github.com/future-architect/gcp-instance-scheduler/model"
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

func createHeader(pad map[string]int) string {
	text := fmt.Sprintf("Instances Shutdown Report <%s>\n", getDate())

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

func createDetailReport(r *model.ShutdownReport) string {
	var text string
	pad := -25

	// fiels values of model.ShutdownReport
	resultVal := reflect.Indirect(reflect.ValueOf(r))
	// field names of model.ShutdownReport
	statusType := getFieldNameList(&model.ShutdownReport{})

	for i := 1; i < len(statusType); i++ {
		status := statusType[i]
		for _, resource := range resultVal.FieldByName(status).Interface().([]string) {
			text += fmt.Sprintf("%*s | %s\n", pad, status, resource)
		}
	}

	return text
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
func (n *slackNotifier) PostReport(report []*model.ShutdownReport) (string, error) {
	padDegree := map[string]int{
		"InstanceType":             -15,
		"DoneResources":            -15,
		"AlreadyShutdownResources": -25,
		"SkipResources":            -15,
	}

	text := createHeader(padDegree)

	for _, execResult := range report {
		sum := execResult.CountResource()

		for resourceState, pad := range padDegree {
			if resourceState == "InstanceType" {
				text += fmt.Sprintf("%*s ", pad, execResult.InstanceType)
			} else {
				text += fmt.Sprintf("| %*d", pad, sum[resourceState])
			}
		}
		text += "\n"
	}

	text += fmt.Sprintf("-------------------------------------------------------------------------------\n")

	return n.postInline(text)
}

func (n *slackNotifier) PostReportThread(parentTS string, report *model.ShutdownReport) error {
	var text string
	text += fmt.Sprintf("[ %s ]\n", report.InstanceType)
	text += createDetailReport(report)

	return n.postThreadInline(text, parentTS)
}
