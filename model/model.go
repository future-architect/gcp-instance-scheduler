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
package model

import (
	"log"
)

const (
	ComputeEngine = "ComputeEngine"
	InstanceGroup = "InstanceGroup"
	SQL           = "SQL"
)

type ResourceState int

const (
	Done ResourceState = iota
	Already
	Skip
)

type ShutdownReport struct {
	// InstanceGroup, ComputeEngine, SQL
	InstanceType string
	// shutdown resource names
	DoneResources []string
	// already stopped resource names
	AlreadyShutdownResources []string
	// skipped resource name
	SkipResources []string
}

func (r *ShutdownReport) Show() {
	log.Println("<<<<< " + r.InstanceType + " >>>>>")

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

func (r *ShutdownReport) CountResource() [3]int {
	var counts [3]int

	counts[Done] = len(r.DoneResources)
	counts[Already] = len(r.AlreadyShutdownResources)
	counts[Skip] = len(r.SkipResources)

	return counts
}
