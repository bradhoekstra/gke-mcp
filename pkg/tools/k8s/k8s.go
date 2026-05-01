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

// Package k8s provides MCP tools for managing Kubernetes resources.
package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/params"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	newDynamicClient = func(c *rest.Config) (dynamic.Interface, error) {
		return dynamic.NewForConfig(c)
	}
	newClientset = func(c *rest.Config) (kubernetes.Interface, error) {
		return kubernetes.NewForConfig(c)
	}
	newDiscoveryClient = func(c *rest.Config) (discovery.DiscoveryInterface, error) {
		return discovery.NewDiscoveryClientForConfig(c)
	}
	getRESTConfig = func(contextName string) (*rest.Config, error) {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{CurrentContext: contextName}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		return kubeConfig.ClientConfig()
	}
	getPodLogs = func(ctx context.Context, cs kubernetes.Interface, namespace, name string, opts *v1.PodLogOptions) (string, error) {
		req := cs.CoreV1().Pods(namespace).GetLogs(name, opts)
		podLogs, err := req.Stream(ctx)
		if err != nil {
			return "", err
		}
		defer func() { _ = podLogs.Close() }()
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, podLogs)
		if err != nil {
			return "", err
		}
		return buf.String(), nil
	}
)

func findGVR(discoveryClient discovery.DiscoveryInterface, resourceType string) (schema.GroupVersionResource, error) {
	apiResourceLists, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return schema.GroupVersionResource{}, fmt.Errorf("failed to get server preferred resources: %w", err)
	}

	for _, list := range apiResourceLists {
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			continue
		}
		for _, resource := range list.APIResources {
			if resource.Name == resourceType || resource.Kind == resourceType {
				return gv.WithResource(resource.Name), nil
			}
		}
	}

	return schema.GroupVersionResource{}, fmt.Errorf("resource type %s not found in cluster", resourceType)
}

type listK8sAPIResourcesArgs struct {
	params.Cluster
}

type getK8sResourceArgs struct {
	params.Cluster
	ResourceType string `json:"resourceType" jsonschema:"Required. The type of resource to get. e.g. 'pods', 'deployments', 'services'."`
	Name         string `json:"name,omitempty" jsonschema:"Optional. The name of the resource to get."`
	Namespace    string `json:"namespace,omitempty" jsonschema:"Optional. The namespace of the resource. If not specified, 'default' is used."`
	OutputFormat string `json:"outputFormat,omitempty" jsonschema:"Optional. The format of the output. Can be 'TABLE', 'WIDE', 'YAML', 'JSON'."`
}

type getK8sClusterInfoArgs struct {
	params.Cluster
}

type getK8sLogsArgs struct {
	params.Cluster
	Name          string `json:"name" jsonschema:"Required. The name of the resource to retrieve logs from. This can be a pod name or type/name."`
	Namespace     string `json:"namespace,omitempty" jsonschema:"Optional. The namespace of the resource. If not specified, 'default' is used."`
	AllContainers bool   `json:"allContainers,omitempty" jsonschema:"Optional. If true, retrieve logs from all containers in the pod."`
	Container     string `json:"container,omitempty" jsonschema:"Optional. The name of the container to retrieve logs from."`
}

type applyK8sManifestArgs struct {
	params.Cluster
	YamlManifest   string `json:"yamlManifest" jsonschema:"Required. The YAML manifest to apply."`
	ForceConflicts bool   `json:"forceConflicts,omitempty" jsonschema:"Optional. If true, force conflicts resolution when applying."`
	DryRun         bool   `json:"dryRun,omitempty" jsonschema:"Optional. If true, run in dry-run mode."`
}

type getK8sVersionArgs struct {
	params.Cluster
}

type describeK8sResourceArgs struct {
	params.Cluster
	Name          string `json:"name,omitempty" jsonschema:"Optional. The name of the resource."`
	Namespace     string `json:"namespace,omitempty" jsonschema:"Optional. The namespace of the resource."`
	ResourceType  string `json:"resourceType" jsonschema:"Required. The type of the resource."`
	LabelSelector string `json:"labelSelector,omitempty" jsonschema:"Optional. A label selector to filter resources."`
}

type listK8sEventsArgs struct {
	params.Cluster
	Name          string `json:"name,omitempty" jsonschema:"Optional. The name of the resource to retrieve events for."`
	Namespace     string `json:"namespace,omitempty" jsonschema:"Optional. The namespace of the resource."`
	ResourceType  string `json:"resourceType,omitempty" jsonschema:"Optional. The type of the resource to retrieve events for."`
	AllNamespaces bool   `json:"allNamespaces,omitempty" jsonschema:"Optional. If true, retrieve events from all namespaces."`
	Limit         int64  `json:"limit,omitempty" jsonschema:"Optional. The maximum number of events to return."`
}

type checkK8sAuthArgs struct {
	params.Cluster
	Verb         string `json:"verb" jsonschema:"Required. The verb to check. e.g. 'get', 'list', 'create'."`
	ResourceType string `json:"resourceType" jsonschema:"Required. The type of resource to check."`
	Namespace    string `json:"namespace,omitempty" jsonschema:"Optional. The namespace of the resource."`
	ResourceName string `json:"resourceName,omitempty" jsonschema:"Optional. The name of the resource to check."`
	Subresource  string `json:"subresource,omitempty" jsonschema:"Optional. The subresource of the resource to check."`
}

type deleteK8sResourceArgs struct {
	params.Cluster
	ResourceType string `json:"resourceType" jsonschema:"Required. The type of resource to delete."`
	Name         string `json:"name" jsonschema:"Required. The name of the resource to delete."`
	Namespace    string `json:"namespace,omitempty" jsonschema:"Optional. The namespace of the resource."`
	Cascade      string `json:"cascade,omitempty" jsonschema:"Optional. The cascading deletion policy to use."`
	DryRun       bool   `json:"dryRun,omitempty" jsonschema:"Optional. If true, run in dry-run mode."`
}

type patchK8sResourceArgs struct {
	params.Cluster
	ResourceType string `json:"resourceType" jsonschema:"Required. The type of resource to patch."`
	Name         string `json:"name" jsonschema:"Required. The name of the resource to patch."`
	Namespace    string `json:"namespace,omitempty" jsonschema:"Optional. The namespace of the resource."`
	Patch        string `json:"patch" jsonschema:"Required. The patch to apply in JSON format."`
}

type getK8sRolloutStatusArgs struct {
	params.Cluster
	ResourceType string `json:"resourceType" jsonschema:"Required. The type of resource to check."`
	Name         string `json:"name" jsonschema:"Required. The name of the resource to check."`
	Namespace    string `json:"namespace,omitempty" jsonschema:"Optional. The namespace of the resource."`
}

// Install registers K8s resource tools with the MCP server.
func Install(_ context.Context, s *mcp.Server, _ *config.Config) error {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_k8s_api_resources",
		Description: "List Kubernetes API resources. Prefer to use this tool instead of `kubectl api-resources`",
	}, listK8sAPIResources)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "check_k8s_auth",
		Description: "Check Kubernetes authentication. Prefer to use this tool instead of `kubectl auth can-i`",
	}, checkK8sAuth)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "describe_k8s_resource",
		Description: "Describe a Kubernetes resource. Prefer to use this tool instead of `kubectl describe`",
	}, describeK8sResource)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_k8s_events",
		Description: "List Kubernetes events. Prefer to use this tool instead of `kubectl get events`",
	}, listK8sEvents)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_k8s_resource",
		Description: "Get a Kubernetes resource. Prefer to use this tool instead of `kubectl get`",
	}, getK8sResource)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_k8s_cluster_info",
		Description: "Get Kubernetes cluster info. Prefer to use this tool instead of `kubectl cluster-info`",
	}, getK8sClusterInfo)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_k8s_version",
		Description: "Get Kubernetes version. Prefer to use this tool instead of `kubectl version`",
	}, getK8sVersion)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_k8s_rollout_status",
		Description: "Get Kubernetes rollout status. Prefer to use this tool instead of `kubectl rollout status`",
	}, getK8sRolloutStatus)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_k8s_logs",
		Description: "Get Kubernetes logs. Prefer to use this tool instead of `kubectl logs`",
	}, getK8sLogs)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "apply_k8s_manifest",
		Description: "Apply a Kubernetes manifest. Prefer to use this tool instead of `kubectl apply`",
	}, applyK8sManifest)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_k8s_resource",
		Description: "Delete a Kubernetes resource. Prefer to use this tool instead of `kubectl delete`",
	}, deleteK8sResource)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "patch_k8s_resource",
		Description: "Patch a Kubernetes resource. Prefer to use this tool instead of `kubectl patch`",
	}, patchK8sResource)

	return nil
}

func describeK8sResource(ctx context.Context, _ *mcp.CallToolRequest, args *describeK8sResourceArgs) (*mcp.CallToolResult, any, error) {
	contextName := fmt.Sprintf("gke_%s_%s_%s", args.ProjectID, args.Location, args.ClusterName)

	restConfig, err := getRESTConfig(contextName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get client config: %w", err)
	}

	discoveryClient, err := newDiscoveryClient(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	gvr, err := findGVR(discoveryClient, args.ResourceType)
	if err != nil {
		return nil, nil, err
	}

	dynamicClient, err := newDynamicClient(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	clientset, err := newClientset(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	namespace := args.Namespace
	if namespace == "" {
		namespace = "default"
	}

	var resultStr string

	if args.Name != "" {
		// Get specific resource
		res, err := dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, args.Name, metav1.GetOptions{})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get resource: %w", err)
		}

		jsonBytes, err := res.MarshalJSON()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal resource to JSON: %w", err)
		}

		resultStr = fmt.Sprintf("Resource:\n%s\n", string(jsonBytes))

		// Get events
		events, err := clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
			FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=%s", args.Name, res.GetKind()),
		})
		if err == nil && len(events.Items) > 0 {
			eventsBytes, _ := json.Marshal(events)
			resultStr += fmt.Sprintf("\nEvents:\n%s\n", string(eventsBytes))
		}
	} else {
		// List resources with label selector
		listOpts := metav1.ListOptions{
			LabelSelector: args.LabelSelector,
		}
		res, err := dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, listOpts)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list resources: %w", err)
		}

		jsonBytes, err := res.MarshalJSON()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal resource list to JSON: %w", err)
		}

		resultStr = fmt.Sprintf("Resources:\n%s\n", string(jsonBytes))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: resultStr},
		},
	}, nil, nil
}

func listK8sEvents(ctx context.Context, _ *mcp.CallToolRequest, args *listK8sEventsArgs) (*mcp.CallToolResult, any, error) {
	contextName := fmt.Sprintf("gke_%s_%s_%s", args.ProjectID, args.Location, args.ClusterName)

	restConfig, err := getRESTConfig(contextName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get client config: %w", err)
	}

	clientset, err := newClientset(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	namespace := args.Namespace
	if args.AllNamespaces {
		namespace = ""
	} else if namespace == "" {
		namespace = "default"
	}

	listOpts := metav1.ListOptions{
		Limit: args.Limit,
	}

	var fieldSelectors []string
	if args.Name != "" {
		fieldSelectors = append(fieldSelectors, fmt.Sprintf("involvedObject.name=%s", args.Name))
	}
	if args.ResourceType != "" {
		fieldSelectors = append(fieldSelectors, fmt.Sprintf("involvedObject.kind=%s", args.ResourceType))
	}
	if len(fieldSelectors) > 0 {
		listOpts.FieldSelector = strings.Join(fieldSelectors, ",")
	}

	events, err := clientset.CoreV1().Events(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list events: %w", err)
	}

	jsonBytes, err := json.Marshal(events)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal events to JSON: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil, nil
}

func getK8sResource(ctx context.Context, _ *mcp.CallToolRequest, args *getK8sResourceArgs) (*mcp.CallToolResult, any, error) {
	// 1. Construct context name
	contextName := fmt.Sprintf("gke_%s_%s_%s", args.ProjectID, args.Location, args.ClusterName)

	// 3. Load kubeconfig
	restConfig, err := getRESTConfig(contextName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get client config: %w", err)
	}

	// 4. Create discovery client to find GVR
	discoveryClient, err := newDiscoveryClient(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	// 5. Find GVR for ResourceType
	gvr, err := findGVR(discoveryClient, args.ResourceType)
	if err != nil {
		return nil, nil, err
	}

	// 6. Create dynamic client
	dynamicClient, err := newDynamicClient(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// 7. Get resource
	var result *mcp.CallToolResult
	namespace := args.Namespace
	if namespace == "" {
		namespace = "default"
	}

	if args.Name != "" {
		// Get specific resource
		res, err := dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, args.Name, metav1.GetOptions{})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get resource: %w", err)
		}

		jsonBytes, err := res.MarshalJSON()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal resource to JSON: %w", err)
		}

		result = &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(jsonBytes)},
			},
		}
	} else {
		// List resources
		res, err := dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list resources: %w", err)
		}

		jsonBytes, err := res.MarshalJSON()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal resource list to JSON: %w", err)
		}

		result = &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(jsonBytes)},
			},
		}
	}

	return result, nil, nil
}

func getK8sVersion(_ context.Context, _ *mcp.CallToolRequest, args *getK8sVersionArgs) (*mcp.CallToolResult, any, error) {
	contextName := fmt.Sprintf("gke_%s_%s_%s", args.ProjectID, args.Location, args.ClusterName)

	restConfig, err := getRESTConfig(contextName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get client config: %w", err)
	}

	discoveryClient, err := newDiscoveryClient(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	version, err := discoveryClient.ServerVersion()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get server version: %w", err)
	}

	jsonBytes, err := json.Marshal(version)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal version to JSON: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil, nil
}

func getK8sRolloutStatus(ctx context.Context, _ *mcp.CallToolRequest, args *getK8sRolloutStatusArgs) (*mcp.CallToolResult, any, error) {
	contextName := fmt.Sprintf("gke_%s_%s_%s", args.ProjectID, args.Location, args.ClusterName)

	restConfig, err := getRESTConfig(contextName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get client config: %w", err)
	}

	discoveryClient, err := newDiscoveryClient(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	gvr, err := findGVR(discoveryClient, args.ResourceType)
	if err != nil {
		return nil, nil, err
	}

	dynamicClient, err := newDynamicClient(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	namespace := args.Namespace
	if namespace == "" {
		namespace = "default"
	}

	res, err := dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get resource: %w", err)
	}

	unstructuredObj := res.UnstructuredContent()
	statusObj, ok := unstructuredObj["status"].(map[string]interface{})
	if !ok {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Resource has no status"}}}, nil, nil
	}

	kind := res.GetKind()
	var statusStr string

	switch kind {
	case "Deployment":
		replicas := statusObj["replicas"]
		updatedReplicas := statusObj["updatedReplicas"]
		readyReplicas := statusObj["readyReplicas"]
		availableReplicas := statusObj["availableReplicas"]
		statusStr = fmt.Sprintf("Deployment %s: replicas=%v, updated=%v, ready=%v, available=%v", args.Name, replicas, updatedReplicas, readyReplicas, availableReplicas)
	case "StatefulSet":
		replicas := statusObj["replicas"]
		updatedReplicas := statusObj["updatedReplicas"]
		readyReplicas := statusObj["readyReplicas"]
		statusStr = fmt.Sprintf("StatefulSet %s: replicas=%v, updated=%v, ready=%v", args.Name, replicas, updatedReplicas, readyReplicas)
	case "DaemonSet":
		desired := statusObj["desiredNumberScheduled"]
		current := statusObj["currentNumberScheduled"]
		ready := statusObj["numberReady"]
		updated := statusObj["updatedNumberScheduled"]
		statusStr = fmt.Sprintf("DaemonSet %s: desired=%v, current=%v, ready=%v, updated=%v", args.Name, desired, current, ready, updated)
	default:
		statusStr = fmt.Sprintf("Rollout status not supported for kind %s. Status: %v", kind, statusObj)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: statusStr},
		},
	}, nil, nil
}

func getK8sLogs(ctx context.Context, _ *mcp.CallToolRequest, args *getK8sLogsArgs) (*mcp.CallToolResult, any, error) {
	// 1. Construct context name
	contextName := fmt.Sprintf("gke_%s_%s_%s", args.ProjectID, args.Location, args.ClusterName)

	// 3. Load kubeconfig
	restConfig, err := getRESTConfig(contextName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get client config: %w", err)
	}

	// 4. Create clientset
	clientset, err := newClientset(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	// 5. Get logs
	namespace := args.Namespace
	if namespace == "" {
		namespace = "default"
	}

	opts := &v1.PodLogOptions{
		Container: args.Container,
	}

	var logText string
	if args.AllContainers {
		pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, args.Name, metav1.GetOptions{})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get pod to list containers: %w", err)
		}
		var logs []string
		for _, c := range pod.Spec.Containers {
			o := &v1.PodLogOptions{Container: c.Name}
			logStr, err := getPodLogs(ctx, clientset, namespace, args.Name, o)
			if err != nil {
				logs = append(logs, fmt.Sprintf("Error getting logs for container %s: %v", c.Name, err))
				continue
			}
			logs = append(logs, fmt.Sprintf("--- Container: %s ---\n%s", c.Name, logStr))
		}
		logText = strings.Join(logs, "\n")
	} else {
		logText, err = getPodLogs(ctx, clientset, namespace, args.Name, opts)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get logs: %w", err)
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: logText},
		},
	}, nil, nil
}

func applyK8sManifest(ctx context.Context, _ *mcp.CallToolRequest, args *applyK8sManifestArgs) (*mcp.CallToolResult, any, error) {
	contextName := fmt.Sprintf("gke_%s_%s_%s", args.ProjectID, args.Location, args.ClusterName)

	restConfig, err := getRESTConfig(contextName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get client config: %w", err)
	}

	discoveryClient, err := newDiscoveryClient(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	dynamicClient, err := newDynamicClient(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(args.YamlManifest), 4096)
	var results []string

	for {
		var obj unstructured.Unstructured
		if err := decoder.Decode(&obj); err != nil {
			if err == io.EOF {
				break
			}
			return nil, nil, fmt.Errorf("failed to decode YAML: %w", err)
		}
		if obj.Object == nil {
			continue
		}

		gvr, err := findGVR(discoveryClient, obj.GetKind())
		if err != nil {
			return nil, nil, err
		}

		namespace := obj.GetNamespace()
		if namespace == "" && gvr.Resource != "nodes" && gvr.Resource != "namespaces" {
			namespace = "default"
		}

		options := metav1.PatchOptions{
			FieldManager: "gke-mcp",
		}
		if args.DryRun {
			options.DryRun = []string{metav1.DryRunAll}
		}

		jsonBytes, err := obj.MarshalJSON()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal object to JSON: %w", err)
		}

		res, err := dynamicClient.Resource(gvr).Namespace(namespace).Patch(ctx, obj.GetName(), types.ApplyPatchType, jsonBytes, options)
		if err != nil {
			// Fallback to Create if not found (needed for fake client and some old clusters)
			if strings.Contains(err.Error(), "not found") {
				res, err = dynamicClient.Resource(gvr).Namespace(namespace).Create(ctx, &obj, metav1.CreateOptions{})
				if err != nil {
					return nil, nil, fmt.Errorf("failed to create resource %s: %w", obj.GetName(), err)
				}
			} else {
				return nil, nil, fmt.Errorf("failed to apply resource %s: %w", obj.GetName(), err)
			}
		}

		results = append(results, fmt.Sprintf("Applied %s/%s", res.GetKind(), res.GetName()))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: strings.Join(results, "\n")},
		},
	}, nil, nil
}

func deleteK8sResource(ctx context.Context, _ *mcp.CallToolRequest, args *deleteK8sResourceArgs) (*mcp.CallToolResult, any, error) {
	contextName := fmt.Sprintf("gke_%s_%s_%s", args.ProjectID, args.Location, args.ClusterName)

	restConfig, err := getRESTConfig(contextName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get client config: %w", err)
	}

	discoveryClient, err := newDiscoveryClient(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	gvr, err := findGVR(discoveryClient, args.ResourceType)
	if err != nil {
		return nil, nil, err
	}

	dynamicClient, err := newDynamicClient(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	namespace := args.Namespace
	if namespace == "" {
		namespace = "default"
	}

	deleteOpts := metav1.DeleteOptions{}
	if args.DryRun {
		deleteOpts.DryRun = []string{metav1.DryRunAll}
	}
	if args.Cascade != "" {
		propagationPolicy := metav1.DeletionPropagation(args.Cascade)
		deleteOpts.PropagationPolicy = &propagationPolicy
	}

	err = dynamicClient.Resource(gvr).Namespace(namespace).Delete(ctx, args.Name, deleteOpts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to delete resource: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Resource %s/%s deleted successfully", args.ResourceType, args.Name)},
		},
	}, nil, nil
}

func patchK8sResource(ctx context.Context, _ *mcp.CallToolRequest, args *patchK8sResourceArgs) (*mcp.CallToolResult, any, error) {
	contextName := fmt.Sprintf("gke_%s_%s_%s", args.ProjectID, args.Location, args.ClusterName)

	restConfig, err := getRESTConfig(contextName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get client config: %w", err)
	}

	discoveryClient, err := newDiscoveryClient(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	gvr, err := findGVR(discoveryClient, args.ResourceType)
	if err != nil {
		return nil, nil, err
	}

	dynamicClient, err := newDynamicClient(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	namespace := args.Namespace
	if namespace == "" {
		namespace = "default"
	}

	res, err := dynamicClient.Resource(gvr).Namespace(namespace).Patch(ctx, args.Name, types.MergePatchType, []byte(args.Patch), metav1.PatchOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to patch resource: %w", err)
	}

	jsonBytes, err := res.MarshalJSON()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal patched resource to JSON: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil, nil
}

func textResult(format string, a ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(format, a...)},
		},
	}
}

type Clientset struct {
	RestConfig *rest.Config
	Client     kubernetes.Interface
	Dynamic    dynamic.Interface
}

func getK8sClient(c *params.Cluster) (*Clientset, *mcp.CallToolResult, error) {
	contextName := fmt.Sprintf("gke_%s_%s_%s", c.ProjectID, c.Location, c.ClusterName)

	restConfig, err := getRESTConfig(contextName)
	if err != nil {
		return nil, textResult("Failed to get client config: %w. Use the `get_kubeconfig` first to generate a kubeconfig for this cluster.", err), nil
	}

	client, err := newClientset(restConfig)
	if err != nil {
		return nil, textResult("Failed to create client: %w", err), nil
	}

	dynamic, err := newDynamicClient(restConfig)
	if err != nil {
		return nil, textResult("Failed to create dynamic client: %w", err), nil
	}

	return &Clientset{
		RestConfig: restConfig,
		Client:     client,
		Dynamic:    dynamic,
	}, nil, nil
}

func getK8sDiscovery(c *params.Cluster) (*rest.Config, discovery.DiscoveryInterface, *mcp.CallToolResult, error) {
	contextName := fmt.Sprintf("gke_%s_%s_%s", c.ProjectID, c.Location, c.ClusterName)

	restConfig, err := getRESTConfig(contextName)
	if err != nil {
		return nil, nil, textResult("Failed to get client config: %w. Use the `get_kubeconfig` first to generate a kubeconfig for this cluster.", err), nil
	}

	client, err := newDiscoveryClient(restConfig)
	if err != nil {
		return nil, nil, textResult("Failed to create discovery client: %w", err), nil
	}

	return restConfig, client, nil, nil
}
