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
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/params"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listK8SAPIResourcesArgs struct {
	params.Cluster
}

type APIGroupDiscovery struct {
	Name             string   `json:"name"`
	Versions         []string `json:"versions"`
	PreferredVersion string   `json:"preferred_version"`
}

func (h *handlers) listK8SAPIResources(ctx context.Context, _ *mcp.CallToolRequest, args *listK8SAPIResourcesArgs) (*mcp.CallToolResult, any, error) {
	clusterPath := args.ClusterPath()

	discoveryClient, err := h.provider.DiscoveryClient(ctx, clusterPath)
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to get discovery client: %w", err)), nil, nil
	}

	groups, resourceLists, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to get server groups and resources: %w", err)), nil, nil
	}

	resourceVersions := make(map[string][]string)
	resourcePreferredVersion := make(map[string]string)

	// Map group name to preferred version
	prefVersions := make(map[string]string)
	for _, g := range groups {
		prefVersions[g.Name] = g.PreferredVersion.GroupVersion
	}

	for _, rl := range resourceLists {
		gv := rl.GroupVersion
		for _, r := range rl.APIResources {
			if strings.Contains(r.Name, "/") {
				// Skip subresources like pods/log
				continue
			}
			resourceVersions[r.Name] = append(resourceVersions[r.Name], gv)

			// Find group name
			parts := strings.Split(gv, "/")
			var group string
			if len(parts) == 2 {
				group = parts[0]
			}

			pref := prefVersions[group]
			if pref == "" {
				pref = "v1" // fallback for core group
			}
			resourcePreferredVersion[r.Name] = pref
		}
	}

	var resources []APIGroupDiscovery
	for name, versions := range resourceVersions {
		resources = append(resources, APIGroupDiscovery{
			Name:             name,
			Versions:         versions,
			PreferredVersion: resourcePreferredVersion[name],
		})
	}

	data, err := json.MarshalIndent(resources, "", "  ")
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to marshal result: %w", err)), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, nil, nil
}
