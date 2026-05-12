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

func TestFormatCustomColumns(t *testing.T) {
	items := []unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "pod-1",
				},
				"status": map[string]interface{}{
					"phase": "Running",
				},
			},
		},
		{
			Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "pod-2",
				},
				"status": map[string]interface{}{
					"phase": "Pending",
				},
			},
		},
	}

	customColumns := "NAME:.metadata.name,STATUS:.status.phase"
	out, err := FormatCustomColumns(items, customColumns)
	if err != nil {
		t.Fatalf("FormatCustomColumns failed: %v", err)
	}

	if !strings.Contains(out, "NAME") || !strings.Contains(out, "STATUS") {
		t.Errorf("Output missing headers: %s", out)
	}
	if !strings.Contains(out, "pod-1") || !strings.Contains(out, "Running") {
		t.Errorf("Output missing data for pod-1: %s", out)
	}
	if !strings.Contains(out, "pod-2") || !strings.Contains(out, "Pending") {
		t.Errorf("Output missing data for pod-2: %s", out)
	}
}

func TestFormatCustomColumns_Invalid(t *testing.T) {
	items := []unstructured.Unstructured{{Object: map[string]interface{}{}}}

	_, err := FormatCustomColumns(items, "INVALID")
	if err == nil {
		t.Error("Expected error for invalid column format, got nil")
	}

	_, err = FormatCustomColumns(items, "NAME:..metadata.name")
	if err == nil {
		t.Error("Expected error for recursive descent, got nil")
	}
}
