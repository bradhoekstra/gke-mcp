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
	"time"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/params"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type getK8SLogsArgs struct {
	params.Cluster
	Name          string `json:"name" jsonschema:"Required. The name of the pod to retrieve logs from. Only 'pod' resource type is supported in this version."`
	Namespace     string `json:"namespace,omitempty" jsonschema:"Optional. The namespace of the resource. If not specified, \"default\" is used."`
	AllContainers bool   `json:"allContainers,omitempty" jsonschema:"Optional. If true, retrieve logs from all containers in the pod."`
	Container     string `json:"container,omitempty" jsonschema:"Optional. The name of the container to retrieve logs from. If not specified, logs from the first container are returned."`
	Previous      bool   `json:"previous,omitempty" jsonschema:"Optional. If true, retrieve logs from the previous instantiation of the container."`
	Timestamps    bool   `json:"timestamps,omitempty" jsonschema:"Optional. If true, include timestamps in the log output."`
	Since         string `json:"since,omitempty" jsonschema:"Optional. Retrieve logs since this duration ago (e.g. \"1h\", \"10m\")."`
	Tail          int64  `json:"tail,omitempty" jsonschema:"Optional. The number of lines from the end of the logs to show."`
}

func (h *handlers) getK8SLogs(ctx context.Context, _ *mcp.CallToolRequest, args *getK8SLogsArgs) (*mcp.CallToolResult, any, error) {
	clusterPath := args.ClusterPath()

	client, err := h.provider.KubernetesClient(ctx, clusterPath)
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to get kubernetes client: %w", err)), nil, nil
	}

	rt := "pod"
	name := args.Name
	if strings.Contains(name, "/") {
		parts := strings.Split(name, "/")
		if len(parts) != 2 || parts[1] == "" {
			return params.ErrorResult(fmt.Errorf("invalid resource name: %q, expected format is type/name", name)), nil, nil
		}
		rt, name = parts[0], parts[1]
	}

	if rt != "pod" {
		return params.ErrorResult(fmt.Errorf("only 'pod' resource type is supported for logs in this version, got %q", rt)), nil, nil
	}

	ns := args.Namespace
	if ns == "" {
		ns = "default"
	}

	opts := &corev1.PodLogOptions{
		Container:  args.Container,
		Previous:   args.Previous,
		Timestamps: args.Timestamps,
	}

	if args.Tail > 0 {
		opts.TailLines = &args.Tail
	}

	if args.Since != "" {
		d, err := time.ParseDuration(args.Since)
		if err != nil {
			return params.ErrorResult(fmt.Errorf("failed to parse since duration %q: %w", args.Since, err)), nil, nil
		}
		seconds := int64(d.Seconds())
		opts.SinceSeconds = &seconds
	}

	var containers []string
	if args.AllContainers || args.Container == "" {
		pod, err := client.CoreV1().Pods(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return params.ErrorResult(fmt.Errorf("failed to get pod: %w", err)), nil, nil
		}

		if args.AllContainers {
			for _, c := range pod.Spec.Containers {
				containers = append(containers, c.Name)
			}
			for _, c := range pod.Spec.InitContainers {
				containers = append(containers, c.Name)
			}
		} else if args.Container == "" {
			if len(pod.Spec.Containers) > 0 {
				containers = []string{pod.Spec.Containers[0].Name}
			}
		}
	} else {
		containers = []string{args.Container}
	}

	var allLogs []string
	for _, cName := range containers {
		cOpts := opts.DeepCopy()
		cOpts.Container = cName

		if len(containers) > 1 {
			allLogs = append(allLogs, fmt.Sprintf("===== Container: %s =====", cName))
		}

		req := client.CoreV1().Pods(ns).GetLogs(name, cOpts)
		podLogs, err := req.Stream(ctx)
		if err != nil {
			allLogs = append(allLogs, fmt.Sprintf("Error: failed to stream logs: %v", err))
			continue
		}

		const maxLogSize = 1024 * 1024 // 1MB safety limit
		buf := new(strings.Builder)
		n, err := io.Copy(buf, io.LimitReader(podLogs, maxLogSize))
		_ = podLogs.Close()

		if err != nil {
			allLogs = append(allLogs, fmt.Sprintf("Error: failed to read logs: %v", err))
			continue
		}

		logText := buf.String()
		if n >= maxLogSize {
			logText += "\n... (logs truncated due to size limit)"
		}
		if logText == "" {
			logText = "(no logs found)"
		}
		allLogs = append(allLogs, logText)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: strings.Join(allLogs, "\n")},
		},
	}, nil, nil
}
