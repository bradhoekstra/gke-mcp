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
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
)

type mockDiscovery struct {
	discovery.DiscoveryInterface
	version *version.Info
	err     error
}

func (m *mockDiscovery) ServerVersion() (*version.Info, error) {
	return m.version, m.err
}

func TestGetK8SVersion(t *testing.T) {
	ctx := context.Background()

	mockDisc := &mockDiscovery{
		version: &version.Info{
			GitVersion: "v1.27.3",
		},
	}

	mockProvider := &mockClientProvider{
		discoveryClient: mockDisc,
	}

	h := &handlers{
		c:        &config.Config{},
		provider: mockProvider,
	}

	args := &getK8SVersionArgs{}
	args.ProjectID = "p"
	args.Location = "l"
	args.ClusterName = "c"

	result, _, err := h.getK8SVersion(ctx, &mcp.CallToolRequest{}, args)
	if err != nil {
		t.Fatalf("getK8SVersion failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("getK8SVersion returned error result: %v", result.Content[0])
	}

	if len(result.Content) != 1 {
		t.Fatalf("len(result.Content) = %d, want 1", len(result.Content))
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("result.Content[0] is not TextContent")
	}

	if !strings.Contains(textContent.Text, "v1.27.3") {
		t.Errorf("output %q does not contain %q", textContent.Text, "v1.27.3")
	}
}
