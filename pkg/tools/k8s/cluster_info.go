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
	"strconv"
	"strings"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/params"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	corev1 "k8s.io/api/core/v1"
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

	services, err := client.CoreV1().Services("kube-system").List(ctx, metav1.ListOptions{
		LabelSelector: "kubernetes.io/cluster-service=true",
	})
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to list services: %w", err)), nil, nil
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Kubernetes control plane is running at %s", config.Host))

	for _, s := range services.Items {
		lines = append(lines, fmt.Sprintf("%s is running at %s", serviceName(s), serviceLink(s, config.Host)))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: strings.Join(lines, "\n")},
		},
	}, nil, nil
}

func serviceName(s corev1.Service) string {
	if name, found := s.GetLabels()["kubernetes.io/name"]; found {
		return name
	}
	return s.GetName()
}

func serviceLink(s corev1.Service, host string) string {
	if len(s.Status.LoadBalancer.Ingress) > 0 {
		ingress := s.Status.LoadBalancer.Ingress[0]
		ip := ingress.IP
		if ip == "" {
			ip = ingress.Hostname
		}
		var link strings.Builder
		for _, port := range s.Spec.Ports {
			link.WriteString("http://" + ip + ":" + strconv.Itoa(int(port.Port)))
		}
		return link.String()
	}
	return host + "/api/v1/namespaces/" + s.Namespace + "/services/" + serviceProxyResourceName(s) + "/proxy"
}

func serviceProxyResourceName(service corev1.Service) string {
	serviceName := service.Name
	if len(service.Spec.Ports) > 0 {
		port := service.Spec.Ports[0]
		if scheme := guessScheme(port); len(scheme) > 0 {
			return scheme + ":" + serviceName + ":" + port.Name
		}
		if len(port.Name) > 0 {
			return serviceName + ":" + port.Name
		}
		return serviceName
	}
	return service.Name
}

func guessScheme(port corev1.ServicePort) string {
	if port.Name == "https" || port.Port == 443 {
		return "https"
	}
	return ""
}
