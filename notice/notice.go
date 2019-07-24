package notice

import (
	"fmt"
	"github.com/future-architect/gcp-instance-scheduler/model"
	"github.com/nlopes/slack"
	"time"
)

type slackNotifier struct {
	slackAPIToken string
	slackChannel  string
}

func getDate() string {
	year, month, day := time.Now().Date()
	return fmt.Sprintf("%d/%d/%d", year, month, day)
}

func createHeader() string {
	text := fmt.Sprintf("Instances Shutdown Report <%s>\n", getDate())
	text += fmt.Sprintf("InstanceType    | Done Shutdowned | Already Shutdowned | Skipped\n")
	text += fmt.Sprintf("-----------------------------------------------------------------\n")
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
func (n *slackNotifier) PostReport(report []model.ShutdownReport) (string, error) {
	pad := 13

	text := createHeader()

	for _, executionResult := range report {
		sum := executionResult.CountResource()
		text += fmt.Sprintf("%*s | %d | %d | %d\n",
			pad,
			executionResult.InstanceType,
			sum[model.Done],
			sum[model.Already],
			sum[model.Skipped])
	}

	text += fmt.Sprintf("―――――――――――――――――――――――――――――――――――――――――――――――――――――\n")

	return n.postInline(text)
}