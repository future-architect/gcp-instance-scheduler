package cmd

import (
	"context"
	"errors"
	"github.com/future-architect/gcp-instance-scheduler/scheduler"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "restart is launch shutdown gcp resource",
	Long:  `restart is launch shutdown gcp resource.`,
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

		return scheduler.Restart(ctx, scheduler.NewOptions(projectID, slackToken, slackChannel, slackEnable))
	},
}

func init() {
	restartCmd.PersistentFlags().StringP("project", "p", os.Getenv("GCP_PROJECT"), "project id (default $GCP_PROJECT)")
	restartCmd.PersistentFlags().Int("timeout", 60, "set timeout seconds")
	restartCmd.PersistentFlags().String("slackToken", os.Getenv("SLACK_API_TOKEN"), "SlackAPI token (should enable slack notify)")
	restartCmd.PersistentFlags().String("slackChannel", os.Getenv("SLACK_CHANNEL"), "Slack Channel name (should enable slack notify)")
	restartCmd.PersistentFlags().BoolP("slackNotifyEnable", "s", false, "Enable slack notification")

	rootCmd.AddCommand(restartCmd)
}
