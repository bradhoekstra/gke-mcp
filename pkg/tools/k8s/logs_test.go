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

package k8s

import (
	"context"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetK8SLogsArgs_Fields(t *testing.T) {
	args := getK8SLogsArgs{
		Name:          "my-pod",
		Namespace:     "default",
		AllContainers: true,
		Container:     "my-container",
		Previous:      true,
		Timestamps:    true,
		Since:         "1h",
		Tail:          10,
	}
	args.ProjectID = "p"
	args.Location = "l"
	args.ClusterName = "c"

	if args.Name != "my-pod" {
		t.Errorf("Name = %s, want my-pod", args.Name)
	}
	if args.Namespace != "default" {
		t.Errorf("Namespace = %s, want default", args.Namespace)
	}
	if !args.AllContainers {
		t.Errorf("AllContainers = %v, want true", args.AllContainers)
	}
	if args.Container != "my-container" {
		t.Errorf("Container = %s, want my-container", args.Container)
	}
	if !args.Previous {
		t.Errorf("Previous = %v, want true", args.Previous)
	}
	if !args.Timestamps {
		t.Errorf("Timestamps = %v, want true", args.Timestamps)
	}
	if args.Since != "1h" {
		t.Errorf("Since = %s, want 1h", args.Since)
	}
	if args.Tail != 10 {
		t.Errorf("Tail = %d, want 10", args.Tail)
	}
}

func TestGetK8SLogs_PodNotFound(t *testing.T) {
	ctx := context.Background()

	fakeClientset := fake.NewSimpleClientset()
	mockProvider := &mockClientProvider{
		kubernetesClient: fakeClientset,
	}

	h := &handlers{
		c:        &config.Config{},
		provider: mockProvider,
	}

	args := &getK8SLogsArgs{
		Name:      "non-existent-pod",
		Namespace: "default",
	}
	args.ProjectID = "p"
	args.Location = "l"
	args.ClusterName = "c"

	result, _, err := h.getK8SLogs(ctx, &mcp.CallToolRequest{}, args)
	if err != nil {
		t.Fatalf("getK8SLogs failed: %v", err)
	}

	if !result.IsError {
		t.Fatalf("getK8SLogs expected error result, got success")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("result.Content[0] is not TextContent")
	}

	if !strings.Contains(textContent.Text, "failed to get pod") {
		t.Errorf("expected error message to contain 'failed to get pod', got %q", textContent.Text)
	}
}
