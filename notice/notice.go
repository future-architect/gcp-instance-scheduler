package notice

import (
	"fmt"
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

func createHeader(pad [4]int) string {
	text := fmt.Sprintf("Instances Shutdown Report <%s>\n", getDate())
	text += fmt.Sprintf("%*s | %*s | %*s | %*s\n",
		pad[0], "InstanceType",
		pad[1], "Done Shutdown",
		pad[2], "Already Shutdowned",
		pad[3], "Skipped")
	text += fmt.Sprintf("-----------------------------------------------------------------\n")
	return text
}

func createDetailReport(r *model.ShutdownReport) string {
	var text string
	pad := -18

	for _, d := range r.DoneResources {
		text += fmt.Sprintf("%*s | %s\n", pad, "DoneResource", d)
	}
	for _, a := range r.AlreadyShutdownResources {
		text += fmt.Sprintf("%*s | %s\n", pad, "AlreadyShutdownResource", a)
	}
	for _, s := range r.SkipResources {
		text += fmt.Sprintf("%*s | %s\n", pad, "SkipResources", s)
	}

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
func (n *slackNotifier) PostReport(report []*model.ShutdownReport) (string, error) {
	pad := [4]int{-15, -15, -18, -15}

	text := createHeader(pad)

	for _, execResult := range report {
		sum := execResult.CountResource()
		text += fmt.Sprintf("%*s | %*d | %*d | %*d\n",
			pad[0], execResult.InstanceType,
			pad[1], sum[model.Done],
			pad[2], sum[model.Already],
			pad[3], sum[model.Skip])
	}

	text += fmt.Sprintf("-----------------------------------------------------------------\n")

	return n.postInline(text)
}

func (n *slackNotifier) PostReportThread(parentTS string, report []*model.ShutdownReport) error {
	var text string
	for _, execResult := range report {
		text += fmt.Sprintf("[ %s ]\n", execResult.InstanceType)
		text += createDetailReport(execResult)
	}

	return n.postThreadInline(text, parentTS)
}
