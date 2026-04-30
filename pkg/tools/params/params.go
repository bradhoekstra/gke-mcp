// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package params provide common tool parameter types.
package params

import "fmt"

type Project struct {
	ProjectID string `json:"project_id" jsonschema:"Required. GCP project ID."`
}

func (p *Project) ProjectIDPath() string {
	return fmt.Sprintf("projects/%s", p.ProjectID)
}

type Location struct {
	Project
	Location string `json:"location" jsonschema:"Required. GKE cluster location."`
}

func (l *Location) LocationPath() string {
	return fmt.Sprintf("%s/locations/%s", l.ProjectIDPath(), l.Location)
}

type LocationOptional struct {
	Project
	Location string `json:"location,omitempty" jsonschema:"Optional. GKE cluster location."`
}

func (l *LocationOptional) LocationPath() string {
	if l.Location == "" {
		return fmt.Sprintf("%s/locations/-", l.ProjectIDPath())
	}
	return fmt.Sprintf("%s/locations/%s", l.ProjectIDPath(), l.Location)
}

type Cluster struct {
	Location
	ClusterName string `json:"cluster_name" jsonschema:"Required. GKE cluster name."`
}

func (c *Cluster) ClusterPath() string {
	return fmt.Sprintf("%s/clusters/%s", c.LocationPath(), c.ClusterName)
}
