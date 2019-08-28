package cmd

import (
	"context"
	"errors"
	"github.com/future-architect/gcp-instance-scheduler/scheduler"
	"github.com/spf13/cobra"
	"log"
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

		log.Printf("Project ID: %v", projectID)
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
	stopCmd.PersistentFlags().StringP("slackToken", "t", os.Getenv("SLACK_API_TOKEN"), "SlackAPI token (should enable slack notify) (default $SLACK_API_TOKEN)")
	stopCmd.PersistentFlags().StringP("slackChannel", "c", os.Getenv("SLACK_CHANNEL"), "Slack Channel name (should enable slack notify) (default SLACK_CHANNEL)")
	stopCmd.PersistentFlags().BoolP("slackNotifyEnable", "s", false, "Enable slack notification")
	stopCmd.PersistentFlags().Int("timeout", 60, "set timeout seconds")

	rootCmd.AddCommand(stopCmd)
}
