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

type Report struct {
	// InstanceGroup, ComputeEngine, SQL
	InstanceType string
	// shutdown resource names
	DoneResources []string
	// already stopped resource names
	AlreadyShutdownResources []string
	// skipped resource name
	SkipResources []string
}

func (r *Report) Show() {
	log.Println("[" + r.InstanceType + "]")

	log.Printf("└- Shutdown Resource: %v", len(r.DoneResources))
	for _, resource := range r.DoneResources {
		log.Printf("  └-- %v\n", resource)
	}

	log.Printf("└- Already Shutdown Resource: %v", len(r.AlreadyShutdownResources))
	for _, resource := range r.AlreadyShutdownResources {
		log.Printf("    └-- %v\n", resource)
	}
}
