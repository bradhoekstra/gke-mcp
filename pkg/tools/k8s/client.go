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
	"net/http"
	"strings"

	containerpb "cloud.google.com/go/container/apiv1/containerpb"
	"github.com/googleapis/gax-go/v2"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

// clientProvider provides Kubernetes clients for a GKE cluster.
type clientProvider struct {
	cmClient cmClient
}

type cmClient interface {
	GetCluster(ctx context.Context, req *containerpb.GetClusterRequest, opts ...gax.CallOption) (*containerpb.Cluster, error)
}

// NewClientProvider creates a new clientProvider.
func NewClientProvider(cmClient cmClient) *clientProvider {
	return &clientProvider{
		cmClient: cmClient,
	}
}

// RESTConfig returns a rest.Config for the given cluster.
func (p *clientProvider) RESTConfig(ctx context.Context, clusterPath string) (*rest.Config, error) {
	req := &containerpb.GetClusterRequest{
		Name: clusterPath,
	}
	resp, err := p.cmClient.GetCluster(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster %s: %w", clusterPath, err)
	}

	clusterCaCertificate := resp.GetMasterAuth().GetClusterCaCertificate()
	endpoint := resp.GetEndpoint()

	if clusterCaCertificate == "" || endpoint == "" {
		return nil, fmt.Errorf("clusterCaCertificate or endpoint not found for cluster")
	}

	if !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}

	caData, err := base64.StdEncoding.DecodeString(clusterCaCertificate)
	if err != nil {
		return nil, fmt.Errorf("failed to decode clusterCaCertificate: %w", err)
	}

	config := &rest.Config{
		Host: endpoint,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: caData,
		},
		ExecProvider: &api.ExecConfig{
			APIVersion:         "client.authentication.k8s.io/v1beta1",
			Command:            "gke-gcloud-auth-plugin",
			ProvideClusterInfo: true,
		},
	}

	return config, nil
}

// DynamicClient returns a dynamic.Interface for the given cluster.
func (p *clientProvider) DynamicClient(ctx context.Context, clusterPath string) (dynamic.Interface, error) {
	config, err := p.RESTConfig(ctx, clusterPath)
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(config)
}

// DynamicClientWithHeaders returns a dynamic.Interface that adds specific headers to every request.
func (p *clientProvider) DynamicClientWithHeaders(ctx context.Context, clusterPath string, headerName, headerValue string) (dynamic.Interface, error) {
	config, err := p.RESTConfig(ctx, clusterPath)
	if err != nil {
		return nil, err
	}
	
	config.Wrap(func(rt http.RoundTripper) http.RoundTripper {
		return &HeaderRoundTripper{
			Wrapped:     rt,
			HeaderName:  headerName,
			HeaderValue: headerValue,
		}
	})
	
	return dynamic.NewForConfig(config)
}

// DiscoveryClient returns a discovery.DiscoveryInterface for the given cluster.
func (p *clientProvider) DiscoveryClient(ctx context.Context, clusterPath string) (discovery.DiscoveryInterface, error) {
	config, err := p.RESTConfig(ctx, clusterPath)
	if err != nil {
		return nil, err
	}
	return discovery.NewDiscoveryClientForConfig(config)
}

// HeaderRoundTripper is an http.RoundTripper that adds a specific header to each request.
type HeaderRoundTripper struct {
	Wrapped     http.RoundTripper
	HeaderName  string
	HeaderValue string
}

func (t *HeaderRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	reqCopy := req.Clone(req.Context())
	reqCopy.Header.Set(t.HeaderName, t.HeaderValue)
	return t.Wrapped.RoundTrip(reqCopy)
}
