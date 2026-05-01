package k8s

import (
	"context"
	"fmt"

	"bitbucket.org/creachadair/stringset"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
)

var (
	resourceVerbs = stringset.New("get", "list", "watch", "create", "update", "patch", "delete", "deletecollection", "use", "bind", "impersonate", "*", "approve", "sign", "escalate", "attest")
)

func checkK8sAuth(ctx context.Context, _ *mcp.CallToolRequest, args *checkK8sAuthArgs) (*mcp.CallToolResult, any, error) {
	client, res, err := getK8sClient(&args.Cluster)
	if res != nil || err != nil {
		return res, nil, err
	}

	text, err := checkK8sAuthImpl(ctx, args, client)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	return textResult("%s", text), nil, nil
}

func checkK8sAuthImpl(ctx context.Context, a *checkK8sAuthArgs, client *Clientset) (string, error) {
	authStatus := AuthStatus{}

	discClient, err := newDiscoveryClient(client.RestConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create discovery client: %w", err)
	}

	// Resolve the resource and validate the command.
	var resource *schema.GroupVersionResource
	if a.ResourceName != "" {
		var warning string
		var err error
		resource, warning, err = ResourceFor(ctx, discClient, a.ResourceName)
		if err != nil {
			return "", fmt.Errorf("failed to resolve resource: %w", err)
		}
		if warning != "" {
			authStatus.AddWarning(warning)
		}
	}
	validate(ctx, a, discClient, resource, &authStatus)

	// Create the self subject access review.
	ssar := createSSARUnstructured(a, resource)
	apiResource, err := ResolveAPIResourceByKind(ctx, discClient, "selfsubjectaccessreviews")
	if err != nil {
		return "", fmt.Errorf("failed to resolve kind to GVR: %w", err)
	}
	resourceInterface, err := apiResource.ResourceInterface(ctx, client.Dynamic, a.Namespace)
	if err != nil {
		return "", fmt.Errorf("failed to get resource interface: %w", err)
	}
	authRes, err := resourceInterface.Create(ctx, ssar, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create self subject access review: %w", err)
	}

	// Parse the self subject access review.
	statusMap, err := getSSARStatus(authRes)
	if err != nil {
		return "", fmt.Errorf("failed to get self subject access review status: %w", err)
	}
	allowed, err := getSSARAllowed(statusMap)
	if err != nil {
		return "", fmt.Errorf("failed to get self subject access review allowed: %w", err)
	}
	authStatus.SetAllowed(allowed)
	reason, err := getSSARStringField(statusMap, "reason")
	if err != nil {
		return "", fmt.Errorf("failed to get self subject access review reason: %w", err)
	}
	authStatus.SetReason(reason)
	evaluationError, err := getSSARStringField(statusMap, "evaluationError")
	if err != nil {
		return "", fmt.Errorf("failed to get self subject access review evaluation error: %w", err)
	}
	authStatus.SetEvaluationError(evaluationError)

	return fmt.Sprintf("%+v", authStatus), nil
}

func isKnownResourceVerb(s string) bool {
	return resourceVerbs.Contains(s)
}

func isNamespaced(ctx context.Context, client discovery.DiscoveryInterface, gvr *schema.GroupVersionResource) (bool, error) {
	if gvr.Resource == "*" {
		return true, nil
	}
	apiResourceList, err := client.ServerResourcesForGroupVersion(schema.GroupVersion{
		Group:   gvr.Group,
		Version: gvr.Version,
	}.String())
	if err != nil {
		return true, err
	}
	for _, resource := range apiResourceList.APIResources {
		if resource.Name == gvr.Resource {
			return resource.Namespaced, nil
		}
	}
	return false, fmt.Errorf("the server doesn't have a resource type '%s' in group '%s'", gvr.Resource, gvr.Group)
}

func validate(ctx context.Context, a *checkK8sAuthArgs, client discovery.DiscoveryInterface, gvr *schema.GroupVersionResource, authStatus *AuthStatus) {
	if gvr != nil && !gvr.Empty() && a.Namespace != "" {
		if namespaced, err := isNamespaced(ctx, client, gvr); err == nil && !namespaced {
			if len(gvr.Group) == 0 {
				authStatus.AddWarning(fmt.Sprintf("resource '%s' is not namespace scoped\n", gvr.Resource))
			} else {
				authStatus.AddWarning(fmt.Sprintf("resource '%s' is not namespace scoped in group '%s'\n", gvr.Resource, gvr.Group))
			}
		}
		if !isKnownResourceVerb(a.Verb) {
			authStatus.AddWarning(fmt.Sprintf("verb '%s' is not a known verb\n", a.Verb))
		}
	}
}

func createSSARUnstructured(a *checkK8sAuthArgs, resource *schema.GroupVersionResource) *unstructured.Unstructured {
	ssar := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "authorization.k8s.io/v1",
			"kind":       "SelfSubjectAccessReview",
		},
	}
	var group, res string
	if resource != nil {
		group = resource.Group
		res = resource.Resource
	}
	ssar.Object["spec"] = map[string]any{
		"resourceAttributes": map[string]any{
			"namespace":   a.Namespace,
			"verb":        a.Verb,
			"group":       group,
			"resource":    res,
			"subresource": a.Subresource,
			"name":        a.ResourceName,
		},
	}
	return ssar
}

func getSSARStatus(authRes *unstructured.Unstructured) (map[string]any, error) {
	status, statusFound := authRes.Object["status"]
	if !statusFound {
		return nil, fmt.Errorf("could not find status in self subject access review")
	}
	statusMap, ok := status.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("could not parse status in self subject access review")
	}
	return statusMap, nil
}

func getSSARAllowed(statusMap map[string]any) (bool, error) {
	allowed, allowedFound := statusMap["allowed"]
	if !allowedFound {
		return false, fmt.Errorf("could not find allowed in self subject access review")
	}
	allowedBool, ok := allowed.(bool)
	if !ok {
		return false, fmt.Errorf("could not parse allowed in self subject access review")
	}
	return allowedBool, nil
}

func getSSARStringField(statusMap map[string]any, key string) (string, error) {
	field, fieldFound := statusMap[key]
	if !fieldFound {
		return "", nil
	}
	fieldString, ok := field.(string)
	if !ok {
		return "", fmt.Errorf("could not parse %s in self subject access review", key)
	}
	return fieldString, nil
}
