/**
 * Copyright (c) 2019-present Future Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package cmd

import (
	"fmt"
	"os"

	"github.com/future-architect/gcp-instance-scheduler/scheduler"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

var cfgFile string

func getFlags(c *cobra.Command) (project, timeout, slackToken, slackChannel string, slackEnable bool, err error) {
	if project, err = c.PersistentFlags().GetString("project"); err != nil {
		return
	}
	if timeout, err = c.PersistentFlags().GetString("timeout"); err != nil {
		return
	}
	if slackToken, err = c.PersistentFlags().GetString("slackToken"); err != nil {
		return
	}
	if slackChannel, err = c.PersistentFlags().GetString("slackChannel"); err != nil {
		return
	}
	if slackEnable, err = c.PersistentFlags().GetBool("slackNotifyEnable"); err != nil {
		return
	}
	return
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gscheduler",
	Short: "gcp-instance-scheduler local execution entroy porint",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, timeout, slackAPIToken, slackChannel, slackEnable, err := getFlags(cmd)
		if err != nil {
			return err
		}
		subMessage := scheduler.SubscribedMessage{
			Command: "stop",
		}
		slackEnableFlag := "false"
		if slackEnable {
			slackEnableFlag = "true"
		}

		opts, err := scheduler.NewSchedulerOptions(subMessage, projectID, timeout, slackAPIToken, slackChannel, slackEnableFlag)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
		defer cancel()

		return scheduler.Shutdown(ctx, opts)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gscheduler.yaml)")

	rootCmd.PersistentFlags().StringP("project", "p", "", "project id (defautl $GCP_PROJECT)")
	rootCmd.PersistentFlags().String("timeout", "60", "set timeout seconds")
	rootCmd.PersistentFlags().String("slackToken", "", "SlackAPI token (should enable slack notify)")
	rootCmd.PersistentFlags().String("slackChannel", "", "Slack Channel name (should enable slack notify)")
	rootCmd.PersistentFlags().BoolP("slackNotifyEnable", "s", false, "Enable slack notification")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".gscheduler" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".gscheduler")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
