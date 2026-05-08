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

package params

import (
	"testing"
)

func TestParent_Parse(t *testing.T) {
	tests := []struct {
		name        string
		parent      string
		wantProject string
		wantLoc     string
		wantCluster string
		wantErr     bool
	}{
		{
			name:        "valid parent",
			parent:      "projects/my-project/locations/us-central1/clusters/my-cluster",
			wantProject: "my-project",
			wantLoc:     "us-central1",
			wantCluster: "my-cluster",
			wantErr:     false,
		},
		{
			name:    "invalid format - too short",
			parent:  "projects/my-project/locations/us-central1",
			wantErr: true,
		},
		{
			name:    "invalid format - wrong prefix",
			parent:  "orgs/my-project/locations/us-central1/clusters/my-cluster",
			wantErr: true,
		},
		{
			name:    "invalid format - missing clusters",
			parent:  "projects/my-project/locations/us-central1/nodes/my-node",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parent{Parent: tt.parent}
			project, loc, cluster, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parent.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if project != tt.wantProject {
					t.Errorf("Parent.Parse() project = %v, want %v", project, tt.wantProject)
				}
				if loc != tt.wantLoc {
					t.Errorf("Parent.Parse() loc = %v, want %v", loc, tt.wantLoc)
				}
				if cluster != tt.wantCluster {
					t.Errorf("Parent.Parse() cluster = %v, want %v", cluster, tt.wantCluster)
				}
			}
		})
	}
}
