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
	"fmt"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/params"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type getK8SVersionArgs struct {
	params.Cluster
}

func (h *handlers) getK8SVersion(ctx context.Context, _ *mcp.CallToolRequest, args *getK8SVersionArgs) (*mcp.CallToolResult, any, error) {
	clusterPath := args.ClusterPath()

	discoveryClient, err := h.provider.DiscoveryClient(ctx, clusterPath)
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to get discovery client: %w", err)), nil, nil
	}

	version, err := discoveryClient.ServerVersion()
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to get server version: %w", err)), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Server Version: %s", version.GitVersion)},
		},
	}, nil, nil
}
