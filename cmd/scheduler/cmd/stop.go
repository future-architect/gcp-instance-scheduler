package cmd

import (
	"context"
	"errors"
	"github.com/future-architect/gcp-instance-scheduler/scheduler"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop is execution command that shutdown all gcp resources that assigned target label",
	Long:  `stop is execution command that shutdown gcp resources that assigned target label.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, slackToken, slackChannel, timeout, slackEnable, err := getFlags(cmd)
		if err != nil {
			return err
		}

		if projectID == "" {
			return errors.New("not found project variable")
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancel()

		return scheduler.Shutdown(ctx, scheduler.NewOptions(projectID, slackToken, slackChannel, slackEnable))
	},
}

func init() {
	stopCmd.PersistentFlags().StringP("project", "p", os.Getenv("GCP_PROJECT"), "project id (default $GCP_PROJECT)")
	stopCmd.PersistentFlags().Int("timeout", 60, "set timeout seconds")
	stopCmd.PersistentFlags().String("slackToken", os.Getenv("SLACK_API_TOKEN"), "SlackAPI token (should enable slack notify)")
	stopCmd.PersistentFlags().String("slackChannel", os.Getenv("SLACK_CHANNEL"), "Slack Channel name (should enable slack notify)")
	stopCmd.PersistentFlags().BoolP("slackNotifyEnable", "s", false, "Enable slack notification")

	rootCmd.AddCommand(stopCmd)
}
