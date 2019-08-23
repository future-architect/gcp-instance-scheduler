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
	"fmt"
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
	Dones []string
	// already stopped resource names
	Alreadies []string
	// skipped resource name
	Skips []string
}

func (r *Report) Show() []string {
	var lines []string
	lines = append(lines, "."+r.InstanceType)

	lines = append(lines, fmt.Sprintf("  └- Done: %v", len(r.Dones)))
	for _, resource := range r.Dones {
		lines = append(lines, fmt.Sprintf("    └-- %v", resource))
	}

	lines = append(lines, fmt.Sprintf("  └- AlreadyDone: %v", len(r.Alreadies)))
	for _, resource := range r.Alreadies {
		lines = append(lines, fmt.Sprintf("    └-- %v", resource))
	}

	lines = append(lines, fmt.Sprintf("  └- Skip: %v", len(r.Alreadies)))
	for _, resource := range r.Alreadies {
		lines = append(lines, fmt.Sprintf("    └-- %v", resource))
	}

	return lines
}
