// Copyright 2026 Google LLC
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

package gkeagent

import (
	"context"
	"testing"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
)

func TestNewAgent_MissingProject(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{} // Empty config

	_, err := NewAgent(ctx, cfg, "skills")
	if err == nil {
		t.Errorf("Expected error for missing project, got nil")
	}
}

func TestListSkills(t *testing.T) {
	agent := &Agent{skillsDir: "../../../skills"} // Minimal agent for testing this method

	skills, err := agent.listSkills()
	if err != nil {
		t.Fatalf("Failed to list skills: %v", err)
	}

	if len(skills) == 0 {
		t.Errorf("Expected to find some skills, got 0")
	}

	// We can check for a specific known skill to be safe.
	found := false
	for _, skill := range skills {
		if skill == "gke-app-onboarding" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find 'gke-app-onboarding' skill, but it was missing.")
	}
}
