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
const Label = "state-scheduler"

// API call interval
const CallInterval = 50 * time.Millisecond

type Options struct {
	Project      string
	SlackEnable  bool
	SlackToken   string
	SlackChannel string
}

func NewOptions(projectID, slackToken, slackChannel string, slackEnable bool) *Options {
	return &Options{
		Project:      projectID,
		SlackEnable:  slackEnable,
		SlackToken:   slackToken,
		SlackChannel: slackChannel,
	}
}

func Shutdown(ctx context.Context, op *Options) error {
	projectID := op.Project
	log.Printf("Project ID: %v", projectID)

	var errorLog error
	var result []*model.ShutdownReport

	if err := operator.SetLabelNodePoolSize(ctx, projectID, Label, CallInterval); err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Printf("Error in setting labels on GKE cluster: %v", err)
	}

	if err := operator.ShowClusterStatus(ctx, projectID, Label); err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Printf("Error in stopping GKE: %v", err)
	}

	rpt, err := operator.InstanceGroup(ctx, projectID).Filter(Label, true).Do(ctx, CallInterval)
	if err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Printf("Some error occured in stopping gce instances: %v", err)
	}
	result = append(result, rpt)
	rpt.Show()

	rpt, err = operator.ComputeEngine(ctx, projectID).Filter(Label, true).Do(ctx, CallInterval)
	if err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Printf("Some error occured in stopping gce instances: %v", err)
	}
	result = append(result, rpt)
	rpt.Show()

	rpt, err = operator.SQL(ctx, projectID).Filter(Label, true).Do(ctx, CallInterval)
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

	n := notice.NewSlackNotifier(op.SlackToken, op.SlackChannel)
	parentTS, err := n.PostReport(report.NewResourceCountReport(result, projectID))
	if err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Println("Error in Slack notification:", err)
	}

	detailReport := report.NewDetailReports(result)
	for _, r := range detailReport {
		if err := n.PostReportThread(parentTS, r); err != nil {
			errorLog = multierror.Append(errorLog, err)
			log.Println("Error in Slack notification (thread):", err)
		}
	}

	log.Printf("done.")
	return errorLog
}

func Reboot(ctx context.Context, op *Options) error {
	projectID := op.Project
	log.Printf("Project ID: %v", projectID)

	var errorLog error
	var result []*model.ShutdownReport

	rpt, err := operator.InstanceGroup(ctx, projectID).Filter(Label, true).Do(ctx, CallInterval)
	if err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Printf("Some error occured in stopping gce instances: %v", err)
	}
	result = append(result, rpt)

	rpt, err = operator.ComputeEngine(ctx, projectID).Filter(Label, true).Do(ctx, CallInterval)
	if err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Printf("Some error occured in stopping gce instances: %v", err)
	}
	result = append(result, rpt)

	rpt, err = operator.SQL(ctx, projectID).Filter(Label, true).Do(ctx, CallInterval)
	if err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Printf("Some error occured in stopping sql instances: %v", err)
	}
	result = append(result, rpt)

	if !op.SlackEnable {
		log.Printf("done.")
		return errorLog
	}

	n := notice.NewSlackNotifier(op.SlackToken, op.SlackChannel)
	parentTS, err := n.PostReport(report.NewResourceCountReport(result, projectID))
	if err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Println("Error in Slack notification:", err)
	}

	detailReport := report.NewDetailReports(result)
	for _, r := range detailReport {
		if err := n.PostReportThread(parentTS, r); err != nil {
			errorLog = multierror.Append(errorLog, err)
			log.Println("Error in Slack notification (thread):", err)
		}
	}

	log.Printf("done.")
	return errorLog
}
