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

package config

import (
	"log"
	"os/exec"
	"strings"
)

type Config struct {
	userAgent        string
	defaultProjectID string
}

func (c *Config) UserAgent() string {
	return c.userAgent
}

func (c *Config) DefaultProjectID() string {
	return c.defaultProjectID
}

func New(version string) *Config {
	return &Config{
		userAgent:        "gke-mcp/" + version,
		defaultProjectID: getDefaultProjectID(),
	}
}

func getDefaultProjectID() string {
	out, err := exec.Command("gcloud", "config", "get", "core/project").Output()
	if err != nil {
		log.Printf("Failed to get default project: %v", err)
		return ""
	}
	projectID := strings.TrimSpace(string(out))
	log.Printf("Using default project ID: %s", projectID)
	return projectID
}
