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
	"fmt"
	"strings"
	"text/tabwriter"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/util/jsonpath"
)

// FormatCustomColumns formats a list of unstructured items using the provided custom columns string.
// The format is "HEADER:JSONPATH,HEADER:JSONPATH".
func FormatCustomColumns(items []unstructured.Unstructured, customColumns string) (string, error) {
	if strings.Contains(customColumns, "..") || strings.Contains(customColumns, "?(") {
		return "", fmt.Errorf("invalid custom column format: recursive descent '..' and filter '?()' expressions are not supported")
	}

	columns := strings.Split(customColumns, ",")
	var headers []string
	var jsonPaths []*jsonpath.JSONPath

	for _, col := range columns {
		parts := strings.SplitN(col, ":", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid custom column format: %s. Expected HEADER:JSONPATH", col)
		}
		headers = append(headers, parts[0])

		jp := jsonpath.New("custom")
		// kubectl expects jsonpath without the surrounding {}
		if err := jp.Parse(fmt.Sprintf("{%s}", parts[1])); err != nil {
			return "", fmt.Errorf("failed to parse jsonpath for column %q: %w", col, err)
		}
		jsonPaths = append(jsonPaths, jp)
	}

	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 8, 2, ' ', 0)

	// Write headers
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// Write rows
	for _, item := range items {
		var row []string
		for _, jp := range jsonPaths {
			results, err := jp.FindResults(item.Object)
			if err != nil {
				row = append(row, "<error>")
				continue
			}
			if len(results) > 0 && len(results[0]) > 0 {
				row = append(row, fmt.Sprintf("%v", results[0][0].Interface()))
			} else {
				row = append(row, "<none>")
			}
		}
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	if err := w.Flush(); err != nil {
		return "", err
	}

	return buf.String(), nil
}
