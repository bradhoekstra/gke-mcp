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

// Package k8s provides MCP tools for interacting with Kubernetes resources.
package k8s

import (
	"context"
	"fmt"

	container "cloud.google.com/go/container/apiv1"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/api/option"
)

type handlers struct {
	c        *config.Config
	cmClient *container.ClusterManagerClient
	provider *clientProvider
}

// Install registers Kubernetes-related tools with the MCP server.
func Install(ctx context.Context, s *mcp.Server, c *config.Config) error {
	cmClient, err := container.NewClusterManagerClient(ctx, option.WithUserAgent(c.UserAgent()))
	if err != nil {
		return fmt.Errorf("failed to create cluster manager client: %w", err)
	}

	h := &handlers{
		c:        c,
		cmClient: cmClient,
		provider: NewClientProvider(cmClient),
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_k8s_resource",
		Description: "Gets one or more Kubernetes resources from a cluster. Resources can be filtered by type, name, namespace, and label selectors. Returns the resources in YAML format. This is similar to running `kubectl get`.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, h.getK8SResource)

	return nil
}
