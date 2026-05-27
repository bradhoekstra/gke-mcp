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

python3 - <<'EOF'
import os
import sys
import yaml

def check_workflow(file_path):
    with open(file_path, 'r') as f:
        try:
            wf = yaml.safe_load(f)
        except yaml.YAMLError as e:
            print(f"Error parsing {file_path}: {e}")
            return False

    if not wf:
        return True

    triggers = wf.get('on', {})
    has_pr_target = False
    if isinstance(triggers, str):
        has_pr_target = (triggers == 'pull_request_target')
    elif isinstance(triggers, list):
        has_pr_target = 'pull_request_target' in triggers
    elif isinstance(triggers, dict):
        has_pr_target = 'pull_request_target' in triggers

    if not has_pr_target:
        return True

    print(f"Checking {file_path} for insecure pull_request_target usage...")
    failed = False

    jobs = wf.get('jobs', {})
    for job_name, job in jobs.items():
        steps = job.get('steps', [])
        for step in steps:
            uses = step.get('uses', '')
            if 'actions/checkout' in uses:
                with_params = step.get('with', {})
                ref = str(with_params.get('ref', ''))
                repo = str(with_params.get('repository', ''))

                if 'github.event.pull_request.head' in ref:
                    print(f"Error: {file_path} (job: {job_name}) uses pull_request_target and checks out untrusted code (ref: {ref})")
                    failed = True
                if 'github.head_ref' in ref:
                    print(f"Error: {file_path} (job: {job_name}) uses pull_request_target and checks out untrusted code (ref: {ref})")
                    failed = True
                if 'github.event.pull_request.head' in repo:
                    print(f"Error: {file_path} (job: {job_name}) uses pull_request_target and checks out an untrusted repository (repository: {repo})")
                    failed = True

    return not failed

def main():
    workflow_dir = '.github/workflows'
    if not os.path.isdir(workflow_dir):
        print("No workflows directory found.")
        sys.exit(0)

    failed = False
    for filename in os.listdir(workflow_dir):
        if filename.endswith('.yml') or filename.endswith('.yaml'):
            file_path = os.path.join(workflow_dir, filename)
            if not check_workflow(file_path):
                failed = True

    if failed:
        print("Workflow security check failed.")
        sys.exit(1)

    print("Workflow security check passed.")
    sys.exit(0)

if __name__ == '__main__':
    main()
EOF
