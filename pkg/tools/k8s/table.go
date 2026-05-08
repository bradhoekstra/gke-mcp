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
	"bytes"
	"encoding/json"
	"fmt"
	"text/tabwriter"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// FormatTable formats a Kubernetes Table object (stored in Unstructured) into a human-readable string.
func FormatTable(obj *unstructured.Unstructured) (string, error) {
	// Convert Unstructured to metav1.Table
	data, err := obj.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("failed to marshal unstructured object: %w", err)
	}

	var table metav1.Table
	if err := json.Unmarshal(data, &table); err != nil {
		return "", fmt.Errorf("failed to unmarshal Table: %w", err)
	}

	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 8, 2, ' ', 0)

	// Write headers
	for i, col := range table.ColumnDefinitions {
		_, _ = fmt.Fprint(w, col.Name)
		if i < len(table.ColumnDefinitions)-1 {
			_, _ = fmt.Fprint(w, "\t")
		}
	}
	_, _ = fmt.Fprintln(w)

	// Write rows
	for _, row := range table.Rows {
		for i, cell := range row.Cells {
			_, _ = fmt.Fprintf(w, "%v", cell)
			if i < len(row.Cells)-1 {
				_, _ = fmt.Fprint(w, "\t")
			}
		}
		_, _ = fmt.Fprintln(w)
	}

	if err := w.Flush(); err != nil {
		return "", err
	}

	return buf.String(), nil
}
