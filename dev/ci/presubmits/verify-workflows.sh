#!/usr/bin/env bash
# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

WORKFLOW_DIR=".github/workflows"

if [ ! -d "$WORKFLOW_DIR" ]; then
    echo "No workflows directory found."
    exit 0
fi

FAILED=0

for file in "$WORKFLOW_DIR"/*.yml "$WORKFLOW_DIR"/*.yaml; do
    [ -e "$file" ] || continue
    
    # Check if file uses pull_request_target
    if grep -q "pull_request_target" "$file"; then
        echo "Checking $file for insecure pull_request_target usage..."
        
        # Check if it checks out untrusted code
        if grep -q "ref:.*github\.event\.pull_request\.head" "$file"; then
            echo "Error: $file uses pull_request_target and checks out untrusted code (pull_request.head)."
            FAILED=1
        fi
        
        if grep -q "ref:.*github\.head_ref" "$file"; then
            echo "Error: $file uses pull_request_target and checks out untrusted code (head_ref)."
            FAILED=1
        fi

        if grep -q "repository:.*github\.event\.pull_request\.head" "$file"; then
            echo "Error: $file uses pull_request_target and checks out an untrusted repository."
            FAILED=1
        fi
    fi
done

if [ $FAILED -ne 0 ]; then
    echo "Workflow security check failed."
    exit 1
fi

echo "Workflow security check passed."
exit 0
