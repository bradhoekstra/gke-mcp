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
	"io"
	"strings"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/params"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/yaml"
)

type applyK8SManifestArgs struct {
	params.Cluster
	YamlManifest   string `json:"yamlManifest" jsonschema:"Required. The YAML manifest to apply."`
	ForceConflicts bool   `json:"forceConflicts,omitempty" jsonschema:"Optional. If true, force conflicts resolution when applying."`
	DryRun         bool   `json:"dryRun,omitempty" jsonschema:"Optional. If true, run in dry-run mode."`
}

func (h *handlers) applyK8SManifest(ctx context.Context, _ *mcp.CallToolRequest, args *applyK8SManifestArgs) (*mcp.CallToolResult, any, error) {
	clusterPath := args.ClusterPath()

	discoveryClient, err := h.provider.DiscoveryClient(ctx, clusterPath)
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to get discovery client: %w", err)), nil, nil
	}

	dynamicClient, err := h.provider.DynamicClient(ctx, clusterPath)
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to get dynamic client: %w", err)), nil, nil
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient))

	unstructuredObjects, err := yamlToUnstructured(strings.NewReader(args.YamlManifest))
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to parse YAML manifest: %w", err)), nil, nil
	}

	if len(unstructuredObjects) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "No resources found to apply."},
			},
		}, nil, nil
	}

	var resultBuilder strings.Builder
	var errors []string

	for i, obj := range unstructuredObjects {
		gvk := obj.GroupVersionKind()
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			errors = append(errors, fmt.Sprintf("document %d with kind %s get REST mapping: %v", i+1, gvk.Kind, err))
			continue
		}
		gvr := mapping.Resource
		name := obj.GetName()

		applyOptions := metav1.ApplyOptions{
			FieldManager: "gke-mcp-agent",
			Force:        args.ForceConflicts,
		}
		if args.DryRun {
			applyOptions.DryRun = []string{"All"}
		}

		var appliedObj *unstructured.Unstructured
		if mapping.Scope.Name() == "namespace" {
			ns := obj.GetNamespace()
			if ns == "" {
				errors = append(errors, fmt.Sprintf("document %d with kind %s: namespace is required but not specified", i+1, gvk.Kind))
				continue
			}
			appliedObj, err = dynamicClient.Resource(gvr).Namespace(ns).Apply(ctx, name, obj, applyOptions)
		} else {
			appliedObj, err = dynamicClient.Resource(gvr).Apply(ctx, name, obj, applyOptions)
		}

		if err != nil {
			errors = append(errors, fmt.Sprintf("apply resource %s/%s (kind: %s): %v", obj.GetNamespace(), name, gvk.Kind, err))
			continue
		}

		data, err := yaml.Marshal(appliedObj)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to marshal applied resource %s/%s to YAML: %v", appliedObj.GetNamespace(), appliedObj.GetName(), err))
			continue
		}

		resultBuilder.WriteString("---\n")
		resultBuilder.WriteString(string(data))
	}

	result := resultBuilder.String()
	if len(errors) > 0 {
		result += "\nErrors:\n" + strings.Join(errors, "\n")
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
		IsError: len(errors) > 0,
	}, nil, nil
}

func yamlToUnstructured(input io.Reader) ([]*unstructured.Unstructured, error) {
	decoder := k8syaml.NewYAMLToJSONDecoder(input)
	var objects []*unstructured.Unstructured

	for {
		obj := &unstructured.Unstructured{}
		err := decoder.Decode(obj)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if obj.GetKind() != "" {
			objects = append(objects, obj)
		}
	}
	return objects, nil
}
