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

// Package main provides the entrypoint for the gke-agent server.
package main

import (
	"context"
	"log"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/agents/gkeagent"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	ctx := context.Background()
	// Version is typically set at build time, using a placeholder here.
	cfg := config.New("0.1.0")

	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "GKE Agent Server",
			Version: "0.1.0",
		},
		&mcp.ServerOptions{
			Capabilities: &mcp.ServerCapabilities{
				Tools: &mcp.ToolCapabilities{ListChanged: true},
			},
		},
	)

	if err := gkeagent.Install(ctx, s, cfg); err != nil {
		log.Fatalf("Failed to install gkeagent tool: %v", err)
	}

	log.Println("Starting GKE Agent Server in stdio mode")
	tr := &mcp.LoggingTransport{Transport: &mcp.StdioTransport{}, Writer: log.Writer()}
	if err := s.Run(ctx, tr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
