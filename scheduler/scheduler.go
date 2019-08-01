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
package scheduler

import (
	"log"
	"time"

	"github.com/future-architect/gcp-instance-scheduler/model"
	"github.com/future-architect/gcp-instance-scheduler/notice"
	"github.com/future-architect/gcp-instance-scheduler/operator"
	"github.com/future-architect/gcp-instance-scheduler/report"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/net/context"
)

// Operation target label name
const TargetLabel = "state-scheduler"

// API call interval
const ShutdownInterval = 50 * time.Millisecond

type SubscribedMessage struct {
	Command string `json:"command"`
}

type ShutdownOptions struct {
	Project       string
	SlackAPIToken string
	SlackChannel  string
	SlackEnable   bool
}

func NewSchedulerOptions(projectID string, slackToken, slackChannel string, slackEnable bool) *ShutdownOptions {
	return &ShutdownOptions{
		Project:       projectID,
		SlackAPIToken: slackToken,
		SlackChannel:  slackChannel,
		SlackEnable:   slackEnable,
	}
}

func Shutdown(ctx context.Context, op *ShutdownOptions) error {
	projectID := op.Project
	slackAPIToken := op.SlackAPIToken
	slackChannel := op.SlackChannel

	log.Printf("Project ID: %v", projectID)

	// for multierror
	var errorLog error

	var result []*model.ShutdownReport

	if err := operator.SetLabelNodePoolSize(ctx, projectID, TargetLabel, ShutdownInterval); err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Printf("Error in setting labels on GKE cluster: %v", err)
	}

	// show cluster status
	if err := operator.ShowClusterStatus(ctx, projectID, TargetLabel); err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Printf("Error in stopping GKE: %v", err)
	}

	rpt, err := operator.InstanceGroupResource(ctx, projectID).
		FilterLabel(TargetLabel, true).
		ShutdownWithInterval(ctx, ShutdownInterval)
	if err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Printf("Some error occured in stopping gce instances: %v", err)
	}
	result = append(result, rpt)
	rpt.Show()

	rpt, err = operator.ComputeEngineResource(ctx, projectID).
		FilterLabel(TargetLabel, true).
		ShutdownWithInterval(ctx, ShutdownInterval)
	if err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Printf("Some error occured in stopping gce instances: %v", err)
	}
	result = append(result, rpt)
	rpt.Show()

	rpt, err = operator.SQLResource(ctx, projectID).
		FilterLabel(TargetLabel, true).
		ShutdownWithInterval(ctx, ShutdownInterval)
	if err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Printf("Some error occured in stopping sql instances: %v", err)
	}
	result = append(result, rpt)
	rpt.Show()

	if !op.SlackEnable {
		log.Printf("done.")
		return errorLog
	}

	notifier := notice.NewSlackNotifier(slackAPIToken, slackChannel)

	countReport := report.NewResourceCountReport(result, projectID)
	parentTS, err := notifier.PostReport(countReport)
	if err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Fatal("Error in Slack notification:", err)
	}

	detailReport := report.NewDetailReportList(result)
	for _, r := range detailReport {
		if err := notifier.PostReportThread(parentTS, r); err != nil {
			errorLog = multierror.Append(errorLog, err)
			log.Fatal("Error in Slack notification (thread):", err)
		}
	}

	log.Printf("done.")
	return errorLog
}
