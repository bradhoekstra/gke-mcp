package k8s

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"bitbucket.org/creachadair/stringset"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

var (
	nonStandardResourceNames = stringset.New("users", "groups")
)

// APIResource represents a Kubernetes API resource.
type APIResource struct {
	restMapping *meta.RESTMapping
}

// GroupVersionKind returns the group version kind of the API resource.
func (r *APIResource) GroupVersionKind() schema.GroupVersionKind {
	return r.restMapping.GroupVersionKind
}

// Resource returns the plural resource name.
func (r *APIResource) Resource() string {
	return r.restMapping.Resource.Resource
}

// ResourceInterface returns the dynamic resource interface for the API resource.
// If the resource is not namespaced, the namespace will be ignored.
func (r *APIResource) ResourceInterface(ctx context.Context, di dynamic.Interface, namespace string) (dynamic.ResourceInterface, error) {
	if r.restMapping.Scope == meta.RESTScopeNamespace {
		return di.Resource(r.restMapping.Resource).Namespace(namespace), nil
	}
	return di.Resource(r.restMapping.Resource), nil
}

// IsNamespaced returns true if the resource is namespaced.
func (r *APIResource) IsNamespaced() bool {
	return r.restMapping.Scope.Name() == meta.RESTScopeNameNamespace
}

// ResolveAPIResourceByKind resolves the k8s API resource for the given kind.
// Kinds can be provided in both short, singular and plural forms and the matching
// process is case insensitive.
// In case of multiple matches, the preferred group and version are used.
// If a match is not found, a cmderror.ErrNotFound is returned.
func ResolveAPIResourceByKind(ctx context.Context, client discovery.DiscoveryInterface, kind string) (*APIResource, error) {
	log.Printf("ResolveAPIResourceByKind: %s", kind)
	start := time.Now()
	defer func() {
		log.Printf("ResolveAPIResourceByKind: %s took %v", kind, time.Since(start))
	}()
	_, apiResources, err := client.ServerGroupsAndResources()
	if err != nil && len(apiResources) == 0 {
		// Only fail if there are no resources. Note that discovery API can return a partial list of
		// resources. This can happen for example, if an aggregated API is not responsive.
		// If a partial list is returned, we still try to find the resource in the list. Otherwise,
		// a broken aggregated API can block the entire discovery call.
		return nil, fmt.Errorf("failed to get server resources: %w", err)
	}
	log.Printf("Found %d API resources. Partial results: %v", countAPIResources(apiResources), err != nil)
	restmapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(client))
	for _, apiList := range apiResources {
		for _, r := range apiList.APIResources {
			if kindMatches(kind, r) {
				// Note that even though we technically found the resource, we don't know what the
				// preferred group and version are.
				// Delegate into the restmapper to find the best match.
				gvr, err := restmapper.KindFor(schema.GroupVersionResource{Resource: r.Name})
				if err != nil {
					return nil, fmt.Errorf("failed to find best version for resource: %w", err)
				}
				mapping, err := restmapper.RESTMapping(gvr.GroupKind(), gvr.Version)
				if err != nil {
					return nil, fmt.Errorf("failed to find REST mapping for resource: %w", err)
				}
				return &APIResource{restMapping: mapping}, nil
			}
		}
	}
	return nil, fmt.Errorf("the server doesn't have a resource type %q", kind)
}

// ResourceFor returns the GroupVersionResource for the given resource argument.
// If the resource argument is "*", it returns a GroupVersionResource with the resource set to "*".
// If the resource argument is not found, it returns a GroupVersionResource with the resource set
// to the argument and a warning message.
func ResourceFor(ctx context.Context, client discovery.DiscoveryInterface, resourceArg string) (*schema.GroupVersionResource, string, error) {
	log.Printf("ResourceFor: %s", resourceArg)
	start := time.Now()
	defer func() {
		log.Printf("ResourceFor: %s took %v", resourceArg, time.Since(start))
	}()

	if resourceArg == "*" {
		return &schema.GroupVersionResource{Resource: resourceArg}, "", nil
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(client))

	fullySpecifiedGVR, groupResource := schema.ParseResourceArg(strings.ToLower(resourceArg))
	gvr := schema.GroupVersionResource{}
	if fullySpecifiedGVR != nil {
		gvr, _ = mapper.ResourceFor(*fullySpecifiedGVR)
	}
	if gvr.Empty() {
		var err error
		gvr, err = mapper.ResourceFor(groupResource.WithVersion(""))
		if err != nil {
			var warning string
			if !nonStandardResourceNames.Contains(groupResource.String()) {
				if len(groupResource.Group) == 0 {
					warning = fmt.Sprintf("the server doesn't have a resource type '%s'\n", groupResource.Resource)
				} else {
					warning = fmt.Sprintf("the server doesn't have a resource type '%s' in group '%s'\n", groupResource.Resource, groupResource.Group)
				}
			}
			return &schema.GroupVersionResource{Resource: resourceArg}, warning, nil
		}
	}
	return &gvr, "", nil
}

func countAPIResources(apiResources []*metav1.APIResourceList) (count int) {
	for _, apiList := range apiResources {
		count += len(apiList.APIResources)
	}
	return count
}

func kindMatches(kind string, r metav1.APIResource) bool {
	if strings.EqualFold(r.Name, kind) || strings.EqualFold(r.Kind, kind) {
		return true
	}
	if r.ShortNames != nil {
		for _, s := range r.ShortNames {
			if strings.EqualFold(s, kind) {
				return true
			}
		}
	}
	return false
}
