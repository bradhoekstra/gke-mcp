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
	"time"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

type mockClientProvider struct {
	restConfig       *rest.Config
	dynamicClient    dynamic.Interface
	discoveryClient  discovery.DiscoveryInterface
	kubernetesClient kubernetes.Interface
	err              error
}

func (m *mockClientProvider) RESTConfig(_ context.Context, _ string) (*rest.Config, error) {
	return m.restConfig, m.err
}

func (m *mockClientProvider) DynamicClient(_ context.Context, _ string) (dynamic.Interface, error) {
	return m.dynamicClient, m.err
}

func (m *mockClientProvider) DynamicClientWithHeaders(_ context.Context, _ string, _, _ string) (dynamic.Interface, error) {
	return m.dynamicClient, m.err
}

func (m *mockClientProvider) DiscoveryClient(_ context.Context, _ string) (discovery.DiscoveryInterface, error) {
	return m.discoveryClient, m.err
}

func (m *mockClientProvider) KubernetesClient(_ context.Context, _ string) (kubernetes.Interface, error) {
	return m.kubernetesClient, m.err
}

func TestListK8SEventsArgs_Fields(t *testing.T) {
	args := listK8SEventsArgs{
		ResourceType:  "pod",
		Name:          "my-pod",
		Namespace:     "default",
		AllNamespaces: true,
		Limit:         100,
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
	if !args.AllNamespaces {
		t.Errorf("AllNamespaces = %v, want true", args.AllNamespaces)
	}
	if args.Limit != 100 {
		t.Errorf("Limit = %d, want 100", args.Limit)
	}
	if args.ClusterPath() != "projects/p/locations/l/clusters/c" {
		t.Errorf("ClusterPath = %s, want projects/p/locations/l/clusters/c", args.ClusterPath())
	}
}

func TestListK8SEvents(t *testing.T) {
	ctx := context.Background()

	// Create fake clientset with some events
	event1 := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "event1",
			Namespace: "default",
		},
		InvolvedObject: corev1.ObjectReference{
			Kind: "Pod",
			Name: "my-pod",
		},
		Reason:        "Scheduled",
		Message:       "Successfully assigned default/my-pod to node1",
		LastTimestamp: metav1.Time{Time: time.Now().Add(-10 * time.Minute)},
	}
	event2 := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "event2",
			Namespace: "default",
		},
		InvolvedObject: corev1.ObjectReference{
			Kind: "Pod",
			Name: "my-pod",
		},
		Reason:        "Pulled",
		Message:       "Container image already present on machine",
		LastTimestamp: metav1.Time{Time: time.Now().Add(-5 * time.Minute)},
	}

	fakeClientset := fake.NewSimpleClientset(event1, event2)

	mockProvider := &mockClientProvider{
		kubernetesClient: fakeClientset,
	}

	h := &handlers{
		c:        &config.Config{},
		provider: mockProvider,
	}

	args := &listK8SEventsArgs{
		Namespace: "default",
	}
	args.ProjectID = "p"
	args.Location = "l"
	args.ClusterName = "c"

	result, _, err := h.listK8SEvents(ctx, &mcp.CallToolRequest{}, args)
	if err != nil {
		t.Fatalf("listK8SEvents failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("listK8SEvents returned error result: %v", result.Content[0])
	}

	if len(result.Content) != 1 {
		t.Fatalf("len(result.Content) = %d, want 1", len(result.Content))
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("result.Content[0] is not TextContent")
	}

	// Verify output contains events
	if !strings.Contains(textContent.Text, "Scheduled") {
		t.Errorf("output does not contain 'Scheduled'")
	}
	if !strings.Contains(textContent.Text, "Pulled") {
		t.Errorf("output does not contain 'Pulled'")
	}

	// Verify sorting (newest first)
	lines := strings.Split(strings.TrimSpace(textContent.Text), "\n")
	if len(lines) < 3 {
		t.Fatalf("len(lines) = %d, want at least 3 (header + 2 events)", len(lines))
	}
	// Line 0 is header
	// Line 1 should be event2 (newer)
	// Line 2 should be event1 (older)
	if !strings.Contains(lines[1], "Pulled") {
		t.Errorf("line 1 does not contain 'Pulled'")
	}
	if !strings.Contains(lines[2], "Scheduled") {
		t.Errorf("line 2 does not contain 'Scheduled'")
	}
}

func TestListK8SEvents_WithResourceType(t *testing.T) {
	ctx := context.Background()

	fakeClientset := fake.NewSimpleClientset()

	// Mock discovery
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
		kubernetesClient: fakeClientset,
		discoveryClient:  fakeDiscovery,
	}

	h := &handlers{
		c:        &config.Config{},
		provider: mockProvider,
	}

	args := &listK8SEventsArgs{
		Namespace:    "default",
		ResourceType: "pod",
		Name:         "my-pod",
	}
	args.ProjectID = "p"
	args.Location = "l"
	args.ClusterName = "c"

	result, _, err := h.listK8SEvents(ctx, &mcp.CallToolRequest{}, args)
	if err != nil {
		t.Fatalf("listK8SEvents failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("listK8SEvents returned error result: %v", result.Content[0])
	}
}

func TestListK8SEvents_AllNamespacesOverridesNamespace(t *testing.T) {
	ctx := context.Background()

	fakeClientset := fake.NewSimpleClientset()

	mockProvider := &mockClientProvider{
		kubernetesClient: fakeClientset,
	}

	h := &handlers{
		c:        &config.Config{},
		provider: mockProvider,
	}

	args := &listK8SEventsArgs{
		Namespace:     "some-namespace",
		AllNamespaces: true,
	}
	args.ProjectID = "p"
	args.Location = "l"
	args.ClusterName = "c"

	_, _, err := h.listK8SEvents(ctx, &mcp.CallToolRequest{}, args)
	if err != nil {
		t.Fatalf("listK8SEvents failed: %v", err)
	}

	actions := fakeClientset.Actions()
	if len(actions) < 1 {
		t.Fatalf("expected at least 1 action, got %d", len(actions))
	}
	action := actions[0]
	if action.GetVerb() != "list" {
		t.Errorf("expected list action, got %s", action.GetVerb())
	}
	if action.GetNamespace() != "" {
		t.Errorf("expected all namespaces (empty string), got %s", action.GetNamespace())
	}
}
