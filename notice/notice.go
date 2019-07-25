package notice

import (
	"fmt"
	"time"
	"reflect"

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

	for i, fieldName := range getFieldNameList(&model.ShutdownReport{}) {
		text += fmt.Sprintf("%*v", pad[fieldName], fieldName)
		if i != 0 {
			text += " | "
		}
	}
	text += "\n"

//	// to get status name
//	statusName := reflect.ValueOf(&model.ShutdownReport{}).Type()
//
//	for i := 1; i < statusName.NumField() - 1; i++ {
//		status := statusName.Field(i).Name
//		text += fmt.Sprintf("%*s | ", pad[status], status)
//	}
//	lastStatus := statusName.Field(statusName.NumField()-1).Name
//	text += fmt.Sprintf("%s\n", pad[lastStatus], lastStatus)

	text += fmt.Sprintf("-----------------------------------------------------------------\n")
	return text
}

func createDetailReport(r *model.ShutdownReport) string {
	var text string
	pad := -25

	// fiels values of model.ShutdownReport
	resultList := reflect.ValueOf(r)
	// field names of model.ShutdownReport
	statusName := resultList.Type()

	text += fmt.Sprintf("%s\n", statusName.Field(0).Name)

	for i := 1; i < statusName.NumField(); i++ {
		status := statusName.Field(i).Name
		for _, resource := range resultList.FieldByName(status).Interface().([]string) {
			text += fmt.Sprintf("%*s | %s\n", pad, status, resource)
		}
	}

	return text
}

// return struct 
func getFieldNameList(i interface{}) []string {
	var fieldName []string

	fieldTypes := reflect.ValueOf(i).Type()
	for i := 0; i < fieldTypes.NumField(); i++ {
		fieldName = append(fieldName, fieldTypes.Field(i).Name)
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
	pad := map[string]int{
		"InstanceType": -15,
		"DoneResources": -15,
		"AlreadyShutdownResources": -25,
		"SkipResources": -15,
	}

	text := createHeader(pad)

	for _, execResult := range report {
		sum := execResult.CountResource()
		statusList := getFieldNameList(&model.ShutdownReport{})
		for i := 0; i < len(statusList); i++ {
			status := statusList[i]
			if i == 0 {
				text += fmt.Sprintf("%*s | ", pad[status], status)
			} else {
				text += fmt.Sprintf(" %*d", pad[status], sum[i])
			}
		}
		text += "\n"
	}

	text += fmt.Sprintf("-----------------------------------------------------------------\n")

	return n.postInline(text)
}

func (n *slackNotifier) PostReportThread(parentTS string, report *model.ShutdownReport) error {
	var text string
	text += fmt.Sprintf("[ %s ]\n", report.InstanceType)
	text += createDetailReport(report)

	return n.postThreadInline(text, parentTS)
}
