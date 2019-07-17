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
package report

import (
	"log"

	"github.com/future-architect/gcp-instance-scheduler/model"
)

type ShutdownReport struct {
	model.ShutdownReport
}

const (
	ComputeEngine = "ComputeEngine"
	InstanceGroup = "InstanceGroup"
	SQL           = "SQL"
)

func (r *ShutdownReport) Show() {
	log.Println("<<<<< " + r.InstanceType + " >>>>>")
	if r == nil {
		log.Printf("There are no instances in %s. Skip.\n", r.InstanceType)
		return
	}
	log.Println("!REPORT!")
	log.Println("[Shutdown Resource]")

	for i, resource := range r.DoneResources {
		log.Printf(">> Resouce(%v): %v\n", i+1, resource)
	}

	log.Println("[Already Shutdown Resource]")
	for i, resource := range r.AlreadyShutdownResources {
		log.Printf(">> Resouce(%v): %v\n", i+1, resource)
	}
}
