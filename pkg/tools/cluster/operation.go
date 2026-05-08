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

package cluster

import (
	"context"
	"fmt"

	containerpb "cloud.google.com/go/container/apiv1/containerpb"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/params"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/encoding/protojson"
)

type listOperationsArgs struct {
	params.LocationRequired
}

type getOperationArgs struct {
	params.Operation
}

type cancelOperationArgs struct {
	params.Operation
}

func (h *handlers) listOperations(ctx context.Context, _ *mcp.CallToolRequest, args *listOperationsArgs) (*mcp.CallToolResult, any, error) {
	if h.cmClient == nil {
		return nil, nil, fmt.Errorf("client not initialized")
	}
	req := &containerpb.ListOperationsRequest{
		Parent: args.LocationPath(),
	}
	resp, err := h.cmClient.ListOperations(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: protojson.Format(resp)},
		},
	}, nil, nil
}

func (h *handlers) getOperation(ctx context.Context, _ *mcp.CallToolRequest, args *getOperationArgs) (*mcp.CallToolResult, any, error) {
	if h.cmClient == nil {
		return nil, nil, fmt.Errorf("client not initialized")
	}
	req := &containerpb.GetOperationRequest{
		Name: args.OperationPath(),
	}
	resp, err := h.cmClient.GetOperation(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: protojson.Format(resp)},
		},
	}, nil, nil
}

func (h *handlers) cancelOperation(ctx context.Context, _ *mcp.CallToolRequest, args *cancelOperationArgs) (*mcp.CallToolResult, any, error) {
	if h.cmClient == nil {
		return nil, nil, fmt.Errorf("client not initialized")
	}
	req := &containerpb.CancelOperationRequest{
		Name: args.OperationPath(),
	}
	err := h.cmClient.CancelOperation(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Operation %s cancelled successfully.", args.OperationPath())},
		},
	}, nil, nil
}
