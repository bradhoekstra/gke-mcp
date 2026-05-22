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

package k8s

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func TestListK8SAPIResources(t *testing.T) {
	ctx := context.Background()

	fakeClientset := fake.NewSimpleClientset()
	fakeDiscovery := fakeClientset.Discovery().(*fakediscovery.FakeDiscovery)
	fakeDiscovery.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{Name: "pods", Namespaced: true, Kind: "Pod"},
			},
		},
		{
			GroupVersion: "apps/v1",
			APIResources: []metav1.APIResource{
				{Name: "deployments", Namespaced: true, Kind: "Deployment"},
			},
		},
	}

	mockProvider := &mockClientProvider{
		discoveryClient: fakeDiscovery,
	}

	h := &handlers{
		c:        &config.Config{},
		provider: mockProvider,
	}

	args := &listK8SAPIResourcesArgs{}
	args.ProjectID = "p"
	args.Location = "l"
	args.ClusterName = "c"

	result, _, err := h.listK8SAPIResources(ctx, &mcp.CallToolRequest{}, args)
	if err != nil {
		t.Fatalf("listK8SAPIResources failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("listK8SAPIResources returned error result: %v", result.Content[0])
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("result.Content[0] is not TextContent")
	}

	var resources []APIGroupDiscovery
	if err := json.Unmarshal([]byte(textContent.Text), &resources); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if len(resources) != 2 {
		t.Fatalf("len(resources) = %d, want 2", len(resources))
	}

	// Verify content
	foundPods := false
	foundDeployments := false
	for _, r := range resources {
		if r.Name == "pods" {
			foundPods = true
			if len(r.Versions) != 1 || r.Versions[0] != "v1" {
				t.Errorf("pods versions = %v, want [v1]", r.Versions)
			}
			if r.PreferredVersion != "v1" {
				t.Errorf("pods preferredVersion = %s, want v1", r.PreferredVersion)
			}
		}
		if r.Name == "deployments" {
			foundDeployments = true
			if len(r.Versions) != 1 || r.Versions[0] != "apps/v1" {
				t.Errorf("deployments versions = %v, want [apps/v1]", r.Versions)
			}
			if r.PreferredVersion != "apps/v1" {
				t.Errorf("deployments preferredVersion = %s, want apps/v1", r.PreferredVersion)
			}
		}
	}

	if !foundPods {
		t.Errorf("pods not found in result")
	}
	if !foundDeployments {
		t.Errorf("deployments not found in result")
	}
}
