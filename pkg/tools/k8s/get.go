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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/params"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

type getK8SResourceArgs struct {
	params.Cluster
	ResourceType  string `json:"resourceType" jsonschema:"Required. The type of resource to retrieve. Kubernetes resource/kind name in singular form, lower case. e.g. \"pod\", \"deployment\", \"service\"."`
	Name          string `json:"name,omitempty" jsonschema:"Optional. The name of the resource to retrieve. If not specified, all resources of the given type are returned."`
	Namespace     string `json:"namespace,omitempty" jsonschema:"Optional. The namespace of the resource. If not specified, all namespaces are searched."`
	LabelSelector string `json:"labelSelector,omitempty" jsonschema:"Optional. A label selector to filter resources."`
	FieldSelector string `json:"fieldSelector,omitempty" jsonschema:"Optional. A field selector to filter resources."`
	OutputFormat  string `json:"outputFormat,omitempty" jsonschema:"Optional. The output format. One of: (table, wide, yaml, json). If not specified, defaults to table."`
	CustomColumns string `json:"customColumns,omitempty" jsonschema:"Optional. The custom columns to output in the format HEADER:JSONPATH,HEADER:JSONPATH. e.g. 'NAME:.metadata.name,STATUS:.status.phase'. If specified, outputFormat is ignored."`
}

func (h *handlers) getK8SResource(ctx context.Context, _ *mcp.CallToolRequest, args *getK8SResourceArgs) (*mcp.CallToolResult, any, error) {
	clusterPath := args.ClusterPath()

	discoveryClient, err := h.provider.DiscoveryClient(ctx, clusterPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get discovery client: %w", err)
	}

	gvr, isNamespaced, err := ResolveGVR(ctx, discoveryClient, args.ResourceType)
	if err != nil {
		return nil, nil, err
	}

	useTable := (args.OutputFormat == "" || strings.ToLower(args.OutputFormat) == "table" || strings.ToLower(args.OutputFormat) == "wide") && args.CustomColumns == ""
	
	var dynamicClient dynamic.Interface
	if useTable {
		headerValue := "application/json;as=Table;v=v1;g=meta.k8s.io"
		if strings.ToLower(args.OutputFormat) == "wide" {
			headerValue += ",application/json;as=Table;v=v1beta1;g=meta.k8s.io;includeColumns=wide"
		}
		dynamicClient, err = h.provider.DynamicClientWithHeaders(ctx, clusterPath, "Accept", headerValue)
	} else {
		dynamicClient, err = h.provider.DynamicClient(ctx, clusterPath)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get dynamic client: %w", err)
	}

	var resourceInterface dynamic.ResourceInterface
	if isNamespaced && args.Namespace != "" {
		resourceInterface = dynamicClient.Resource(gvr).Namespace(args.Namespace)
	} else {
		resourceInterface = dynamicClient.Resource(gvr)
	}

	var result string
	if args.Name != "" {
		obj, err := resourceInterface.Get(ctx, args.Name, metav1.GetOptions{})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get resource: %w", err)
		}

		if args.CustomColumns != "" {
			result, err = FormatCustomColumns([]unstructured.Unstructured{*obj}, args.CustomColumns)
		} else {
			result, err = h.formatResource(obj, args.OutputFormat)
		}
		if err != nil {
			return nil, nil, err
		}
	} else {
		list, err := resourceInterface.List(ctx, metav1.ListOptions{
			LabelSelector: args.LabelSelector,
			FieldSelector: args.FieldSelector,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list resources: %w", err)
		}
		
		if args.CustomColumns != "" {
			result, err = FormatCustomColumns(list.Items, args.CustomColumns)
		} else if useTable {
			// In table mode, the list.Object itself is the Table
			result, err = FormatTable(&unstructured.Unstructured{Object: list.Object})
		} else {
			result, err = h.formatResourceList(list, args.OutputFormat)
		}
		if err != nil {
			return nil, nil, err
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
	}, nil, nil
}

func (h *handlers) formatResource(obj *unstructured.Unstructured, format string) (string, error) {
	switch strings.ToLower(format) {
	case "json":
		data, err := json.MarshalIndent(obj, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
	case "yaml":
		data, err := yaml.Marshal(obj)
		if err != nil {
			return "", err
		}
		return string(data), nil
	case "table", "wide", "":
		// If it's a Table object returned from server
		if obj.GetKind() == "Table" {
			return FormatTable(obj)
		}
		// Fallback to YAML if not a table
		data, err := yaml.Marshal(obj)
		if err != nil {
			return "", err
		}
		return string(data), nil
	default:
		return "", fmt.Errorf("unsupported output format: %s", format)
	}
}

func (h *handlers) formatResourceList(list *unstructured.UnstructuredList, format string) (string, error) {
	switch strings.ToLower(format) {
	case "json":
		data, err := json.MarshalIndent(list, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
	case "yaml":
		data, err := yaml.Marshal(list)
		if err != nil {
			return "", err
		}
		return string(data), nil
	default:
		return "", fmt.Errorf("unsupported output format: %s", format)
	}
}
