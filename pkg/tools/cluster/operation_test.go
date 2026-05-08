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
	"testing"
)

func TestListOperationsArgs_Fields(t *testing.T) {
	args := listOperationsArgs{}
	args.ProjectID = "test-project"
	args.Location = "us-central1"

	if args.ProjectID != "test-project" {
		t.Errorf("ProjectID = %s, want test-project", args.ProjectID)
	}
	if args.Location != "us-central1" {
		t.Errorf("Location = %s, want us-central1", args.Location)
	}
}

func TestGetOperationArgs_Fields(t *testing.T) {
	args := getOperationArgs{}
	args.ProjectID = "test-project"
	args.Location = "us-central1"
	args.OperationID = "my-op"

	if args.ProjectID != "test-project" {
		t.Errorf("ProjectID = %s, want test-project", args.ProjectID)
	}
	if args.Location != "us-central1" {
		t.Errorf("Location = %s, want us-central1", args.Location)
	}
	if args.OperationID != "my-op" {
		t.Errorf("OperationID = %s, want my-op", args.OperationID)
	}
}

func TestCancelOperationArgs_Fields(t *testing.T) {
	args := cancelOperationArgs{}
	args.ProjectID = "test-project"
	args.Location = "us-central1"
	args.OperationID = "my-op"

	if args.ProjectID != "test-project" {
		t.Errorf("ProjectID = %s, want test-project", args.ProjectID)
	}
	if args.Location != "us-central1" {
		t.Errorf("Location = %s, want us-central1", args.Location)
	}
	if args.OperationID != "my-op" {
		t.Errorf("OperationID = %s, want my-op", args.OperationID)
	}
}

func TestListOperations_Handler(t *testing.T) {
	h := &handlers{}
	_, _, err := h.listOperations(context.Background(), nil, &listOperationsArgs{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if err.Error() != "client not initialized" {
		t.Errorf("Expected 'client not initialized', got %v", err)
	}
}

func TestGetOperation_Handler(t *testing.T) {
	h := &handlers{}
	_, _, err := h.getOperation(context.Background(), nil, &getOperationArgs{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if err.Error() != "client not initialized" {
		t.Errorf("Expected 'client not initialized', got %v", err)
	}
}

func TestCancelOperation_Handler(t *testing.T) {
	h := &handlers{}
	_, _, err := h.cancelOperation(context.Background(), nil, &cancelOperationArgs{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if err.Error() != "client not initialized" {
		t.Errorf("Expected 'client not initialized', got %v", err)
	}
}
