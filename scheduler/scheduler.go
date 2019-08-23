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
	"strings"

	"github.com/future-architect/gcp-instance-scheduler/model"
	"github.com/future-architect/gcp-instance-scheduler/operator"
	"github.com/future-architect/gcp-instance-scheduler/report"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/net/context"
)

// Operation target label name
const Label = "state-scheduler"

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

	var errorLog error
	var result []*model.Report

	if err := operator.SetLabelNodePoolSize(ctx, projectID, Label); err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Printf("Error in setting labels on GKE cluster: %v", err)
	}

	if err := operator.ShowClusterStatus(ctx, projectID, Label); err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Printf("Error in stopping GKE: %v", err)
	}

	rpt, err := operator.InstanceGroup(ctx, projectID).Filter(Label, true).Resize(0)
	if err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Printf("Some error occured in stopping gce instances: %v", err)
	}
	result = append(result, rpt)
	log.Println(strings.Join(rpt.Show(), "\n"))

	rpt, err = operator.ComputeEngine(ctx, projectID).Filter(Label, true).Stop()
	if err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Printf("Some error occured in stopping gce instances: %v", err)
	}
	result = append(result, rpt)
	log.Println(strings.Join(rpt.Show(), "\n"))

	rpt, err = operator.SQL(ctx, projectID).Filter(Label, true).Stop()
	if err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Printf("Some error occured in stopping sql instances: %v", err)
	}
	result = append(result, rpt)
	log.Println(strings.Join(rpt.Show(), "\n"))

	if !op.SlackEnable {
		log.Printf("done.")
		return errorLog
	}

	_, err = report.NewSlackNotifier(op.SlackToken, op.SlackChannel).Post(report.Report{
		ProjectID: projectID,
		Reports:   result,
		Command:   "Shutdown",
	})
	if err != nil {
		log.Println("error in Slack notification:", err)
		return multierror.Append(errorLog, err)
	}

	log.Printf("done.")
	return errorLog
}

func Restart(ctx context.Context, op *Options) error {
	projectID := op.Project

	var errorLog error
	var result []*model.Report

	rpt, err := operator.InstanceGroup(ctx, projectID).Filter(Label, true).Recovery()
	if err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Printf("Some error occurred in starting instances group: %v\n", err)
	}
	result = append(result, rpt)
	log.Println(strings.Join(rpt.Show(), "\n"))

	rpt, err = operator.ComputeEngine(ctx, projectID).Filter(Label, true).Start()
	if err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Printf("Some error occurred in starting compute engine: %v\n", err)
	}
	result = append(result, rpt)
	log.Println(strings.Join(rpt.Show(), "\n"))

	rpt, err = operator.SQL(ctx, projectID).Filter(Label, true).Start()
	if err != nil {
		errorLog = multierror.Append(errorLog, err)
		log.Printf("Some error occurred in starting SQL: %v\n", err)
	}
	result = append(result, rpt)
	log.Println(strings.Join(rpt.Show(), "\n"))

	if !op.SlackEnable {
		log.Printf("done.")
		return errorLog
	}

	_, err = report.NewSlackNotifier(op.SlackToken, op.SlackChannel).Post(report.Report{
		ProjectID: projectID,
		Reports:   result,
		Command:   "Restart",
	})
	if err != nil {
		log.Println("error in Slack notification:", err)
		return multierror.Append(errorLog, err)
	}

	log.Printf("done.")
	return errorLog
}
