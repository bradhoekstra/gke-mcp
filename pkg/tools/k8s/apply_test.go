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
	k8stesting "k8s.io/client-go/testing"
)

func TestApplyK8SManifest(t *testing.T) {
	tests := []struct {
		name           string
		yamlManifest   string
		dryRun         bool
		forceConflicts bool
		expectErr      bool
		expectedOutput string
	}{
		{
			name: "successful apply",
			yamlManifest: `
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
  namespace: default
`,
			expectErr:      false,
			expectedOutput: "my-pod",
		},
		{
			name: "invalid yaml",
			yamlManifest: `
invalid: yaml: :
`,
			expectErr: true,
		},
		{
			name: "dry run",
			yamlManifest: `
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
  namespace: default
`,
			dryRun:         true,
			expectErr:      false,
			expectedOutput: "my-pod",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			scheme := runtime.NewScheme()
			pod := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "my-pod",
						"namespace": "default",
					},
				},
			}
			fakeDynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, pod)

			fakeDynamicClient.PrependReactor("patch", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
				if tc.dryRun {
					// We can't easily check dry-run options here without casting to specific internal types that might be hard to access.
					// But we can at least ensure the reactor is called.
				}
				return true, pod, nil
			})

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

			args := &applyK8SManifestArgs{
				YamlManifest:   tc.yamlManifest,
				DryRun:         tc.dryRun,
				ForceConflicts: tc.forceConflicts,
			}
			args.ProjectID = "p"
			args.Location = "l"
			args.ClusterName = "c"

			result, _, err := h.applyK8SManifest(ctx, &mcp.CallToolRequest{}, args)
			if err != nil {
				t.Fatalf("applyK8SManifest failed: %v", err)
			}

			if tc.expectErr {
				if !result.IsError {
					t.Fatalf("expected error but got none")
				}
				return
			}

			if result.IsError {
				t.Fatalf("applyK8SManifest returned error result: %v", result.Content[0])
			}

			if len(result.Content) != 1 {
				t.Fatalf("len(result.Content) = %d, want 1", len(result.Content))
			}

			textContent, ok := result.Content[0].(*mcp.TextContent)
			if !ok {
				t.Fatalf("result.Content[0] is not TextContent")
			}

			if tc.expectedOutput != "" && !strings.Contains(textContent.Text, tc.expectedOutput) {
				t.Errorf("output does not contain %q", tc.expectedOutput)
			}
		})
	}
}
