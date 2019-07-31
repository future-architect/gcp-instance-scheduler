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
package function

import (
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/future-architect/gcp-instance-scheduler/scheduler"
	"golang.org/x/net/context"
)

// Operation target label name
const TargetLabel = "state-scheduler"

// API call interval
const ShutdownInterval = 50 * time.Millisecond

type SubscribedMessage struct {
	Command string `json:"command"`
}

func ReceiveEvent(ctx context.Context, msg *pubsub.Message) error {

	projectID := os.Getenv("GCP_PROJECT")
	slackAPIToken := os.Getenv("SLACK_API_TOKEN")
	slackChannel := os.Getenv("SLACK_CHANNEL_NAME")

	opts, err := scheduler.NewSchedulerOptions(projectID, "60", slackAPIToken, slackChannel, true)
	if err != nil {
		return err
	}

	return scheduler.Shutdown(ctx, msg, opts)
}
