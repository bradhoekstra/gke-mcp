#!/usr/bin/env bash
# Copyright 2025 Google LLC
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

# Colors
BLUE='\033[0;34m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "${REPO_ROOT}"

echo -e "${BLUE}Running build...${NC}"
if npm run build; then
    echo -e "${GREEN}✓ Build passed.${NC}"
else
    echo -e "${RED}✗ Build failed.${NC}"
    exit 1
fi

echo -e "${BLUE}Running tests...${NC}"
if npm run test; then
    echo -e "${GREEN}✓ Tests passed.${NC}"
else
    echo -e "${RED}✗ Tests failed.${NC}"
    exit 1
fi

echo -e "${GREEN}Local presubmit checks complete.${NC}"
