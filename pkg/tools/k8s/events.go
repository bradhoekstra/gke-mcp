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
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/params"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
)

type listK8SEventsArgs struct {
	params.Cluster
	Name          string `json:"name,omitempty" jsonschema:"Optional. The name of the resource to retrieve events for."`
	Namespace     string `json:"namespace,omitempty" jsonschema:"Optional. The namespace of the resource. If not specified and allNamespaces is false, 'default' is used."`
	ResourceType  string `json:"resourceType,omitempty" jsonschema:"Optional. The type of the resource to retrieve events for."`
	AllNamespaces bool   `json:"allNamespaces,omitempty" jsonschema:"Optional. If true, retrieve events from all namespaces."`
	Limit         int64  `json:"limit,omitempty" jsonschema:"Optional. The maximum number of events to return. If not specified, 500 is used."`
}

func (h *handlers) listK8SEvents(ctx context.Context, _ *mcp.CallToolRequest, args *listK8SEventsArgs) (*mcp.CallToolResult, any, error) {
	clusterPath := args.ClusterPath()

	config, err := h.provider.RESTConfig(ctx, clusterPath)
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to get rest config: %w", err)), nil, nil
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to create clientset: %w", err)), nil, nil
	}

	apiVersion := ""
	kind := ""
	if args.ResourceType != "" {
		discoveryClient, err := h.provider.DiscoveryClient(ctx, clusterPath)
		if err != nil {
			return params.ErrorResult(fmt.Errorf("failed to get discovery client: %w", err)), nil, nil
		}

		groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
		if err != nil {
			return params.ErrorResult(fmt.Errorf("failed to get API group resources: %w", err)), nil, nil
		}
		mapper := restmapper.NewDiscoveryRESTMapper(groupResources)

		// Try to resolve as a direct resource name (e.g., "pods")
		gvr, err := mapper.ResourceFor(schema.GroupVersionResource{Resource: args.ResourceType})
		var gvk schema.GroupVersionKind
		if err == nil {
			gvk, err = mapper.KindFor(gvr)
			if err != nil {
				return params.ErrorResult(fmt.Errorf("failed to get kind for resource: %w", err)), nil, nil
			}
		} else {
			// Then try to resolve as a kind (e.g., "Pod")
			gvk, err = mapper.KindFor(schema.GroupVersionResource{Resource: args.ResourceType})
			if err != nil {
				return params.ErrorResult(fmt.Errorf("failed to resolve resource type %q: %w", args.ResourceType, err)), nil, nil
			}
		}

		kind = gvk.Kind
		apiVersion = gvk.GroupVersion().String()
	}

	namespace := args.Namespace
	if !args.AllNamespaces && namespace == "" {
		namespace = "default"
	}

	limit := args.Limit
	if limit <= 0 {
		limit = 500
	}

	selector := ""
	if !args.AllNamespaces {
		selector = fmt.Sprintf("involvedObject.namespace=%s,", namespace)
	}
	if kind != "" {
		selector = selector + fmt.Sprintf("involvedObject.kind=%s,involvedObject.name=%s", kind, args.Name)
		if apiVersion != "" {
			selector = selector + fmt.Sprintf(",involvedObject.apiVersion=%s", apiVersion)
		}
	}
	selector = strings.TrimSuffix(selector, ",")

	eventList, err := clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		Limit:         limit,
		FieldSelector: selector,
	})
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to list events: %w", err)), nil, nil
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', 0)

	if args.AllNamespaces {
		_, _ = fmt.Fprintf(w, "NAMESPACE\t")
	}
	_, _ = fmt.Fprintf(w, "LAST SEEN\tTYPE\tREASON\tOBJECT\tMESSAGE\n")

	for _, event := range eventList.Items {
		if args.AllNamespaces {
			_, _ = fmt.Fprintf(w, "%s\t", event.InvolvedObject.Namespace)
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s/%s\t%s\n",
			getInterval(event),
			event.Type,
			event.Reason,
			event.InvolvedObject.Kind,
			event.InvolvedObject.Name,
			event.Message,
		)
	}
	_ = w.Flush()

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: buf.String()},
		},
	}, nil, nil
}

func getInterval(e corev1.Event) string {
	var interval string
	firstTimestampSince := translateMicroTimestampSince(e.EventTime)
	if e.EventTime.IsZero() {
		firstTimestampSince = translateTimestampSince(e.FirstTimestamp)
	}
	if e.Series != nil {
		interval = fmt.Sprintf("%s (x%d over %s)", translateMicroTimestampSince(e.Series.LastObservedTime), e.Series.Count, firstTimestampSince)
	} else if e.Count > 1 {
		interval = fmt.Sprintf("%s (x%d over %s)", translateTimestampSince(e.LastTimestamp), e.Count, firstTimestampSince)
	} else {
		interval = firstTimestampSince
	}
	return interval
}

func translateMicroTimestampSince(timestamp metav1.MicroTime) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}
	return duration.HumanDuration(time.Since(timestamp.Time))
}

func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}
	return duration.HumanDuration(time.Since(timestamp.Time))
}
