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
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestFormatTable(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind": "Table",
			"columnDefinitions": []interface{}{
				map[string]interface{}{"name": "NAME"},
				map[string]interface{}{"name": "STATUS"},
			},
			"rows": []interface{}{
				map[string]interface{}{
					"cells": []interface{}{"pod-1", "Running"},
				},
				map[string]interface{}{
					"cells": []interface{}{"pod-2", "Pending"},
				},
			},
		},
	}

	out, err := FormatTable(obj)
	if err != nil {
		t.Fatalf("FormatTable failed: %v", err)
	}

	if !strings.Contains(out, "NAME") || !strings.Contains(out, "STATUS") {
		t.Errorf("Output missing headers: %s", out)
	}
	if !strings.Contains(out, "pod-1") || !strings.Contains(out, "Running") {
		t.Errorf("Output missing data: %s", out)
	}
}
