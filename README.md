# gcp-instance-scheduler
[![Apache 2.0 License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/future-architect/gcp-instance-scheduler)](https://goreportcard.com/report/github.com/future-architect/gcp-instance-scheduler)

Tools that shutdown GCP Instance on your schedule.

## Abstract

* Shutdown target
   * GCE, GKE, SQL
   * The label `state-scheduler: true` is required to stop / resume the instance.
   * In order to be processed, it is necessary to assign a label to Instance, InstanceGroup or Cluster.
   If a label is assigned to Cluster or InstanceGroup, this tool will reduce the size of InstanceGroup to 0.   
* Architecture
  * Cloud Scheduler --> Pub/Sub --> CloudFunction
    * https://cloud.google.com/scheduler/docs/start-and-stop-compute-engine-instances-on-a-schedule
* 🚧Limitation🚧
   * Stop only. This tool does not support restart yet
     * If you need to recover, you should restart instances manually


## Config

* nothing special
  * [GCP_PROJECT is automation set by CloudFunction](https://cloud.google.com/functions/docs/concepts/go-runtime#contextcontext)


## Getting Started

* install [gcloud](https://cloud.google.com/sdk/gcloud/)

### Required variables
When you want to get slack notification, please set these environment variables.
You can get slack notification if and only if these three variables are setted.

|#  |variables       |Note                               |
|---|----------------|-----------------------------------|
| 1 |SLACK_ENABLE    |Slack notification enable ("true") |
| 2 |SLACK_API_TOKEN |Slack api token                    |
| 3 |SLACK_CHANNEL   |Slack channel name                 |

```sh
# Deploy Cloud Function: slack notification enable
gcloud functions deploy ReceiveEvent --project <project-id> \
  --runtime go111 \
  --trigger-topic instance-scheduler-event \
  --set-env-vars SLACK_ENABLE=true \
  --set-env-vars SLACK_API_TOKEN=<slack-api-token> \
  --set-env-vars SLACK_CHANNEL=<slack-channel-name>

# Create Cloud Scheduler Job
gcloud beta scheduler jobs create pubsub shutdown-workday \
  --project <project-id> \
  --schedule '0 22 * * *' \
  --topic instance-scheduler-event \
  --message-body '{"command":"stop"}' \
  --time-zone 'Asia/Tokyo' \
  --description 'automatically stop instances'
```

## Example: create target resources

Set label for target instance

```sh
# GCE
gcloud compute instances update <insntance-name> \
  --project <project-id> \
  --update-labels state-scheduler=true

# Instance Group
gcloud compute instance-templates create <tmeplate-name> ... \
  --project <project-id> \
  --labels state-scheduler=true

# Cloud SQL (master must be running)
gcloud beta sql instances patch <insntance-name> \
  --project <project-id> \
  --update-labels state-scheduler=true

# GKE
gcloud container clusters update <cluster-name> \
  --project <project-id> \
  --zone <cluster-master-node-zone> \
  --update-labels state-scheduler=true
```

## Local Execution Tool

gscheduler
====

Shutdown instances with executing functinos from your console.

### Install

`go get -v github.com/future-architect/gcp-instance-scheduler/cmd/gscheduler`

### Usage

You can designate project id and timeout length by using flags.
If you use slack notification, you have to enable slack notification by adding the flag `--slackNotification`.

```
gcp-instance-scheduler local execution entroy porint

Usage:
  gscheduler [flags]

Flags:
      --config string         config file (default is $HOME/.gscheduler.yaml)
  -h, --help                  help for gscheduler
  -p, --project string        project id (defautl $GCP_PROJECT)
      --slackChannel string   Slack Channel name (should enable slack notify)
  -s, --slackNotifyEnable     Enable slack notification
      --slackToken string     SlackAPI token (should enable slack notify)
      --timeout string        set timeout seconds (default "60")
  -t, --toggle                Help message for toggle
``` 
Following valiables are used when you didn't designate these flags.

|#  |flags                  |variables       |
|---|-----------------------|----------------|
| 1 |project(p)             |GCP_PROEJCT     |
| 2 |slackAPIToken          |SLACK_API_TOKEN |
| 3 |slackChannel           |SLACK_CHANNEL   |

## Tips: Debug Function

* publish message to pub/sub
  * `gcloud pubsub topics publish stop-instance-event --project <project-id> --message "{"command":"stop"}"`
* confirm Functions log
  * `gcloud functions logs read --project <project-id> --limit 50`
* manual launch for job of scheduler
  * `gcloud beta scheduler jobs run shutdown-workday-instance`

## License

This project is licensed under the Apache License 2.0 License - see the [LICENSE](LICENSE) file for details
