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

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any sub commands
var rootCmd = &cobra.Command{
	Use:   "scheduler",
	Short: "gcp-instance-scheduler local execution entry point",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getFlags(c *cobra.Command) (project, slackToken, slackChannel string, timeout int, slackEnable bool, err error) {
	if project, err = c.PersistentFlags().GetString("project"); err != nil {
		return
	}
	if timeout, err = c.PersistentFlags().GetInt("timeout"); err != nil {
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
