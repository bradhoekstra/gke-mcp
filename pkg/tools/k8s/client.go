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
	"net/http"
	"strings"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ClientProvider provides Kubernetes clients for a GKE cluster.
type ClientProvider struct {
}

// NewClientProvider creates a new ClientProvider.
func NewClientProvider() *ClientProvider {
	return &ClientProvider{}
}

// RESTConfig returns a rest.Config for the given cluster.
func (p *ClientProvider) RESTConfig(_ context.Context, clusterPath string) (*rest.Config, error) {
	// Extract context name from clusterPath
	// clusterPath format: projects/PROJECT/locations/LOCATION/clusters/CLUSTER
	parts := strings.Split(clusterPath, "/")
	if len(parts) != 6 {
		return nil, fmt.Errorf("invalid cluster path: %s", clusterPath)
	}
	project := parts[1]
	location := parts[3]
	cluster := parts[5]
	contextName := fmt.Sprintf("gke_%s_%s_%s", project, location, cluster)

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{CurrentContext: contextName}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	return kubeConfig.ClientConfig()
}

// DynamicClient returns a dynamic.Interface for the given cluster.
func (p *ClientProvider) DynamicClient(ctx context.Context, clusterPath string) (dynamic.Interface, error) {
	config, err := p.RESTConfig(ctx, clusterPath)
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(config)
}

// DynamicClientWithHeaders returns a dynamic.Interface that adds specific headers to every request.
func (p *ClientProvider) DynamicClientWithHeaders(ctx context.Context, clusterPath string, headerName, headerValue string) (dynamic.Interface, error) {
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
func (p *ClientProvider) DiscoveryClient(ctx context.Context, clusterPath string) (discovery.DiscoveryInterface, error) {
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

// RoundTrip implements the http.RoundTripper interface.
func (t *HeaderRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	reqCopy := req.Clone(req.Context())
	reqCopy.Header.Set(t.HeaderName, t.HeaderValue)
	return t.Wrapped.RoundTrip(reqCopy)
}
