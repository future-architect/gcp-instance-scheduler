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
	"errors"
	"log"

	"cloud.google.com/go/pubsub"
	"github.com/future-architect/gcp-instance-scheduler/scheduler"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/net/context"
)

// Env is cloud function environment variables
type Env struct {
	ProjectID    string `envconfig:"GCP_PROJECT" required:"true"`
	SlackToken   string `envconfig:"SLACK_API_TOKEN" required:"true"`
	SlackChannel string `envconfig:"SLACK_CHANNEL"`
	SlackNotify  bool   `envconfig:"SLACK_ENABLE"`
}

func ReceiveEvent(ctx context.Context, msg *pubsub.Message) error {

	var e Env
	if err := envconfig.Process("", &e); err != nil {
		log.Printf("Error at the fucntion 'DecodeMessage': %v", err)
		return err
	}
	if e.SlackNotify && (e.SlackToken == "" || e.SlackChannel == "") {
		return errors.New("missing environment variable")
	}

	payload, err := decode(msg.Data)
	if err != nil {
		log.Printf("Error at the fucntion 'DecodeMessage': %v", err)
		return err
	}
	log.Printf("Subscribed message(Command): %v", payload.Command)

	log.Printf("Project ID: %v", e.ProjectID)
	opts := scheduler.NewOptions(e.ProjectID, e.SlackToken, e.SlackChannel, e.SlackNotify)

	switch payload.Command {
	case "start":
		if err := scheduler.Restart(ctx, opts); err != nil {
			return err
		}
	case "stop":
		if err := scheduler.Shutdown(ctx, opts); err != nil {
			return err
		}
	default:
		return errors.New("unknown command type")
	}

	return nil
}

type Payload struct {
	Command string `json:"command"`
}

func decode(payload []byte) (p Payload, err error) {
	if err = json.Unmarshal(payload, &p); err != nil {
		log.Printf("Message[%v] ... Could not decode subscribing data: %v", payload, err)
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		return
	}
	return
}
