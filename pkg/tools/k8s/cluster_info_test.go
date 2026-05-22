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
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func TestGetK8SClusterInfo(t *testing.T) {
	ctx := context.Background()

	fakeClientset := fake.NewSimpleClientset(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-dns",
			Namespace: "kube-system",
			Labels: map[string]string{
				"kubernetes.io/cluster-service": "true",
				"kubernetes.io/name":            "CoreDNS",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: "dns", Port: 53},
			},
		},
	})

	mockProvider := &mockClientProvider{
		kubernetesClient: fakeClientset,
		restConfig:       &rest.Config{Host: "https://10.0.0.1"},
	}

	h := &handlers{
		c:        &config.Config{},
		provider: mockProvider,
	}

	args := &getK8SClusterInfoArgs{}
	args.ProjectID = "p"
	args.Location = "l"
	args.ClusterName = "c"

	result, _, err := h.getK8SClusterInfo(ctx, &mcp.CallToolRequest{}, args)
	if err != nil {
		t.Fatalf("getK8SClusterInfo failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("getK8SClusterInfo returned error result: %v", result.Content[0])
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("result.Content[0] is not TextContent")
	}

	if !strings.Contains(textContent.Text, "Kubernetes control plane is running at https://10.0.0.1") {
		t.Errorf("output does not contain control plane info")
	}

	if !strings.Contains(textContent.Text, "CoreDNS is running at https://10.0.0.1/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy") {
		t.Errorf("output does not contain CoreDNS info")
	}
}
