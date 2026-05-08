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
	"testing"
)

func TestGetK8SResourceArgs_Fields(t *testing.T) {
	args := getK8SResourceArgs{
		ResourceType:  "pod",
		Name:          "my-pod",
		Namespace:     "default",
		LabelSelector: "app=myapp",
		FieldSelector: "status.phase=Running",
		OutputFormat:  "yaml",
	}
	args.Parent.Parent = "projects/p/locations/l/clusters/c"

	if args.ResourceType != "pod" {
		t.Errorf("ResourceType = %s, want pod", args.ResourceType)
	}
	if args.Name != "my-pod" {
		t.Errorf("Name = %s, want my-pod", args.Name)
	}
	if args.Namespace != "default" {
		t.Errorf("Namespace = %s, want default", args.Namespace)
	}
	if args.LabelSelector != "app=myapp" {
		t.Errorf("LabelSelector = %s, want app=myapp", args.LabelSelector)
	}
	if args.FieldSelector != "status.phase=Running" {
		t.Errorf("FieldSelector = %s, want status.phase=Running", args.FieldSelector)
	}
	if args.OutputFormat != "yaml" {
		t.Errorf("OutputFormat = %s, want yaml", args.OutputFormat)
	}
	if args.Parent.Parent != "projects/p/locations/l/clusters/c" {
		t.Errorf("Parent = %s, want projects/p/locations/l/clusters/c", args.Parent.Parent)
	}
}
