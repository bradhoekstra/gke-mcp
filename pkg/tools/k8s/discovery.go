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

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/restmapper"
)

// ResolveGVR resolves a resource kind or name (e.g., "pods", "deployments") to its GroupVersionResource.
func ResolveGVR(ctx context.Context, discoveryClient discovery.DiscoveryInterface, resource string) (schema.GroupVersionResource, bool, error) {
	groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return schema.GroupVersionResource{}, false, fmt.Errorf("failed to get API group resources: %w", err)
	}

	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	// Try to resolve as a direct resource name (e.g., "pods")
	gvr, err := mapper.ResourceFor(schema.GroupVersionResource{Resource: resource})
	if err == nil {
		gvk, err := mapper.KindFor(gvr)
		if err != nil {
			return gvr, false, err
		}
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return gvr, false, err
		}
		return gvr, mapping.Scope.Name() == "namespace", nil
	}

	// Then try to resolve as a kind (e.g., "Pod")
	gvk, err := mapper.KindFor(schema.GroupVersionResource{Resource: resource})
	if err == nil {
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return schema.GroupVersionResource{}, false, err
		}
		return mapping.Resource, mapping.Scope.Name() == "namespace", nil
	}

	return schema.GroupVersionResource{}, false, fmt.Errorf("failed to resolve resource %q: %w", resource, err)
}
