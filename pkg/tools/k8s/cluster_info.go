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
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/params"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type getK8SClusterInfoArgs struct {
	params.Cluster
}

func (h *handlers) getK8SClusterInfo(ctx context.Context, _ *mcp.CallToolRequest, args *getK8SClusterInfoArgs) (*mcp.CallToolResult, any, error) {
	clusterPath := args.ClusterPath()

	config, err := h.provider.RESTConfig(ctx, clusterPath)
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to get REST config: %w", err)), nil, nil
	}

	client, err := h.provider.KubernetesClient(ctx, clusterPath)
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to get kubernetes client: %w", err)), nil, nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Kubernetes control plane is at %s\n", config.Host))

	// Try to find CoreDNS or kube-dns
	_, err = client.CoreV1().Services("kube-system").Get(ctx, "kube-dns", metav1.GetOptions{})
	if err == nil {
		result.WriteString(fmt.Sprintf("CoreDNS is running at %s/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy\n", config.Host))
	} else {
		// try coredns name
		_, err = client.CoreV1().Services("kube-system").Get(ctx, "coredns", metav1.GetOptions{})
		if err == nil {
			result.WriteString(fmt.Sprintf("CoreDNS is running at %s/api/v1/namespaces/kube-system/services/coredns:dns/proxy\n", config.Host))
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.String()},
		},
	}, nil, nil
}
