package report

import (
	"fmt"
	"github.com/future-architect/gcp-instance-scheduler/model"
	"github.com/nlopes/slack"
)

type Report struct {
	ProjectID string
	Command   string
	Reports   []*model.Report
}

type slackNotifier struct {
	slackAPIToken string
	slackChannel  string
}

func NewSlackNotifier(slackAPIToken, slackChannel string) *slackNotifier {
	return &slackNotifier{
		slackAPIToken: slackAPIToken,
		slackChannel:  slackChannel,
	}
}

func (n *slackNotifier) Post(r Report) (string, error) {
	text := fmt.Sprintf("Project(%s) %s Report\n", r.ProjectID, r.Command)

	for _, detail := range r.Reports {
		lines := detail.Show()
		for _, line := range lines {
			text += line + "\n"
		}
	}

	return n.postInline(text)
}

func (n *slackNotifier) postInline(text string) (string, error) {
	_, ts, err := slack.New(n.slackAPIToken).PostMessage(
		n.slackChannel,
		slack.MsgOptionText("```"+text+"```", false),
	)
	return ts, err
}
