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
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"

	containerpb "cloud.google.com/go/container/apiv1/containerpb"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/params"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"k8s.io/client-go/tools/clientcmd"
	k8sClientApi "k8s.io/client-go/tools/clientcmd/api"
)

type getK8SResourceArgs struct {
	params.Cluster
	ResourceType  string `json:"resourceType" jsonschema:"Required. The type of resource to retrieve. Kubernetes resource/kind name in singular form, lower case. e.g. \"pod\", \"deployment\", \"service\"."`
	Name          string `json:"name,omitempty" jsonschema:"Optional. The name of the resource to retrieve. If not specified, all resources of the given type are returned."`
	Namespace     string `json:"namespace,omitempty" jsonschema:"Optional. The namespace of the resource. If not specified, all namespaces are searched."`
	LabelSelector string `json:"labelSelector,omitempty" jsonschema:"Optional. A label selector to filter resources."`
	FieldSelector string `json:"fieldSelector,omitempty" jsonschema:"Optional. A field selector to filter resources."`
	// TODO: Support whenTime and customColumns if needed.
	OutputFormat string `json:"outputFormat,omitempty" jsonschema:"Optional. The output format. One of: (table, wide, yaml, json). If not specified, defaults to table."`
}

func (h *handlers) getK8SResource(ctx context.Context, _ *mcp.CallToolRequest, args *getK8SResourceArgs) (*mcp.CallToolResult, any, error) {
	// 1. Get cluster info from GKE API to setup kubeconfig
	req := &containerpb.GetClusterRequest{
		Name: args.ClusterPath(),
	}
	resp, err := h.cmClient.GetCluster(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get cluster %s: %w", args.ClusterPath(), err)
	}

	// 2. Setup temporary kubeconfig
	kubeconfigPath, cleanup, err := h.setupTempKubeconfig(resp, args.ProjectID, args.Location, args.ClusterName)
	if err != nil {
		return nil, nil, err
	}
	defer cleanup()

	// 3. Build kubectl command
	kubectlArgs := []string{"--kubeconfig", kubeconfigPath, "get", args.ResourceType}
	if args.Name != "" {
		kubectlArgs = append(kubectlArgs, args.Name)
	}
	if args.Namespace != "" {
		kubectlArgs = append(kubectlArgs, "--namespace", args.Namespace)
	} else {
		kubectlArgs = append(kubectlArgs, "--all-namespaces")
	}
	if args.LabelSelector != "" {
		kubectlArgs = append(kubectlArgs, "--selector", args.LabelSelector)
	}
	if args.FieldSelector != "" {
		kubectlArgs = append(kubectlArgs, "--field-selector", args.FieldSelector)
	}

	switch strings.ToLower(args.OutputFormat) {
	case "wide":
		kubectlArgs = append(kubectlArgs, "-o", "wide")
	case "yaml":
		kubectlArgs = append(kubectlArgs, "-o", "yaml")
	case "json":
		kubectlArgs = append(kubectlArgs, "-o", "json")
	case "table", "":
		// default
	default:
		return nil, nil, fmt.Errorf("unsupported output format: %s", args.OutputFormat)
	}

	// 4. Run kubectl
	cmd := exec.CommandContext(ctx, "kubectl", kubectlArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error running kubectl: %v\nOutput: %s", err, string(out))},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(out)},
		},
	}, nil, nil
}

func (h *handlers) setupTempKubeconfig(resp *containerpb.Cluster, projectID, location, clusterName string) (string, func(), error) {
	clusterCaCertificate := resp.GetMasterAuth().GetClusterCaCertificate()
	endpoint := resp.GetEndpoint()

	if clusterCaCertificate == "" || endpoint == "" {
		return "", nil, fmt.Errorf("clusterCaCertificate or endpoint not found for cluster")
	}

	if !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}

	caData, err := base64.StdEncoding.DecodeString(clusterCaCertificate)
	if err != nil {
		return "", nil, fmt.Errorf("failed to decode clusterCaCertificate: %w", err)
	}

	contextName := fmt.Sprintf("gke_%s_%s_%s", projectID, location, clusterName)

	config := k8sClientApi.Config{
		Clusters: map[string]*k8sClientApi.Cluster{
			contextName: {
				Server:                   endpoint,
				CertificateAuthorityData: caData,
			},
		},
		Contexts: map[string]*k8sClientApi.Context{
			contextName: {
				Cluster:  contextName,
				AuthInfo: contextName,
			},
		},
		AuthInfos: map[string]*k8sClientApi.AuthInfo{
			contextName: {
				Exec: &k8sClientApi.ExecConfig{
					APIVersion:         "client.authentication.k8s.io/v1beta1",
					Command:            "gke-gcloud-auth-plugin",
					ProvideClusterInfo: true,
				},
			},
		},
		CurrentContext: contextName,
	}

	tmpFile, err := os.CreateTemp("", "kubeconfig-")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp kubeconfig: %w", err)
	}
	tmpFile.Close()

	if err := clientcmd.WriteToFile(config, tmpFile.Name()); err != nil {
		os.Remove(tmpFile.Name())
		return "", nil, fmt.Errorf("failed to write kubeconfig to file: %w", err)
	}

	cleanup := func() {
		os.Remove(tmpFile.Name())
	}

	return tmpFile.Name(), cleanup, nil
}
