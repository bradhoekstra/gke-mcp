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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	fakediscovery "k8s.io/client-go/discovery/fake"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetK8SResourceArgs_Fields(t *testing.T) {
	args := getK8SResourceArgs{
		ResourceType:  "pod",
		Name:          "my-pod",
		Namespace:     "default",
		LabelSelector: "app=myapp",
		FieldSelector: "status.phase=Running",
		OutputFormat:  "yaml",
	}
	args.ProjectID = "p"
	args.Location = "l"
	args.ClusterName = "c"

	if args.ResourceType != "pod" {
		t.Errorf("ResourceType = %s, want pod", args.ResourceType)
	}
	if args.Name != "my-pod" {
		t.Errorf("Name = %s, want my-pod", args.Name)
	}
	if args.Namespace != "default" {
		t.Errorf("Namespace = %s, want default", args.Namespace)
	}
	if args.LabelSelector != "app=myapp" {
		t.Errorf("LabelSelector = %s, want app=myapp", args.LabelSelector)
	}
	if args.FieldSelector != "status.phase=Running" {
		t.Errorf("FieldSelector = %s, want status.phase=Running", args.FieldSelector)
	}
	if args.OutputFormat != "yaml" {
		t.Errorf("OutputFormat = %s, want yaml", args.OutputFormat)
	}
	if args.ClusterPath() != "projects/p/locations/l/clusters/c" {
		t.Errorf("ClusterPath = %s, want projects/p/locations/l/clusters/c", args.ClusterPath())
	}
}

func TestGetK8SResource(t *testing.T) {
	ctx := context.Background()

	// Create a fake unstructured pod
	pod := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "my-pod",
				"namespace": "default",
				"labels": map[string]interface{}{
					"app": "myapp",
				},
			},
			"status": map[string]interface{}{
				"phase": "Running",
			},
		},
	}

	scheme := runtime.NewScheme()
	fakeDynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, pod)

	// We need to mock discovery for ResolveGVR
	fakeClientset := fake.NewSimpleClientset()
	fakeDiscovery := fakeClientset.Discovery().(*fakediscovery.FakeDiscovery)
	fakeDiscovery.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{Name: "pods", Namespaced: true, Kind: "Pod"},
			},
		},
	}

	mockProvider := &mockClientProvider{
		dynamicClient:   fakeDynamicClient,
		discoveryClient: fakeDiscovery,
	}

	h := &handlers{
		c:        &config.Config{},
		provider: mockProvider,
	}

	args := &getK8SResourceArgs{
		ResourceType: "pod",
		Name:         "my-pod",
		Namespace:    "default",
		OutputFormat: "yaml",
	}
	args.ProjectID = "p"
	args.Location = "l"
	args.ClusterName = "c"

	result, _, err := h.getK8SResource(ctx, &mcp.CallToolRequest{}, args)
	if err != nil {
		t.Fatalf("getK8SResource failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("getK8SResource returned error result: %v", result.Content[0])
	}

	if len(result.Content) != 1 {
		t.Fatalf("len(result.Content) = %d, want 1", len(result.Content))
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("result.Content[0] is not TextContent")
	}

	// Verify output contains pod name
	if !strings.Contains(textContent.Text, "my-pod") {
		t.Errorf("output does not contain 'my-pod'")
	}
}
