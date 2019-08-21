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
	"encoding/json"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/future-architect/gcp-instance-scheduler/scheduler"
	"golang.org/x/net/context"
)

func ReceiveEvent(ctx context.Context, msg *pubsub.Message) error {

	projectID := os.Getenv("GCP_PROJECT")
	slackAPIToken := os.Getenv("SLACK_API_TOKEN")
	slackChannel := os.Getenv("SLACK_CHANNEL")
	slackNotify := os.Getenv("SLACK_ENABLE")
	if projectID == "" || (slackNotify == "true" && (slackAPIToken == "" || slackChannel == "")) {
		return fmt.Errorf("missing environment variable")
	}

	// decode the json message from Pub/Sub
	message, err := decode(msg.Data)
	if err != nil {
		log.Printf("Error at the fucntion 'DecodeMessage': %v", err)
		return err
	}
	log.Printf("Subscribed message(Command): %v", message.Command)

	slackEnable := false
	if slackNotify == "true" {
		slackEnable = true
	}

	opts := scheduler.NewOptions(projectID, slackAPIToken, slackChannel, slackEnable)

	return scheduler.Shutdown(ctx, opts)
}

type SubscribedMessage struct {
	Command string `json:"command"`
}

func decode(payload []byte) (msgData SubscribedMessage, err error) {
	if err = json.Unmarshal(payload, &msgData); err != nil {
		log.Printf("Message[%v] ... Could not decode subscribing data: %v", payload, err)
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		return
	}
	return
}
