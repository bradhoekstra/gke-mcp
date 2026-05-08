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

func TestListK8SEventsArgs_Fields(t *testing.T) {
	args := listK8SEventsArgs{
		ResourceType:  "pod",
		Name:          "my-pod",
		Namespace:     "default",
		AllNamespaces: true,
		Limit:         100,
	}
	args.ProjectID = "p"
	args.Location = "l"
	args.ClusterName = "c"

	if args.ResourceType != "pod" {
		t.Errorf("ResourceType = %s, want pod", args.ResourceType)
	}
	if args.Name != "my-pod" {
		t.Errorf("Name = %s, want my-pod", args.Name)
	}
	if args.Namespace != "default" {
		t.Errorf("Namespace = %s, want default", args.Namespace)
	}
	if !args.AllNamespaces {
		t.Errorf("AllNamespaces = %v, want true", args.AllNamespaces)
	}
	if args.Limit != 100 {
		t.Errorf("Limit = %d, want 100", args.Limit)
	}
	if args.ClusterPath() != "projects/p/locations/l/clusters/c" {
		t.Errorf("ClusterPath = %s, want projects/p/locations/l/clusters/c", args.ClusterPath())
	}
}
