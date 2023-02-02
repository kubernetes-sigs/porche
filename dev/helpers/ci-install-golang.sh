#!/usr/bin/env bash

# Copyright 2023 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# script to install go in CI
set -o errexit
set -o nounset
set -o pipefail

VERSION="1.19.5"
HASH="36519702ae2fd573c9869461990ae550c8c0d955cd28d2827a6b159fda81ff95"

REPO_ROOT=$(git rev-parse --show-toplevel)

cd /tmp/
echo "Installing go ${VERSION} from upstream to ensure CI version ..."
curl -L https://go.dev/dl/go${VERSION}.linux-amd64.tar.gz -o go.tar.gz

sha256sum --check - <<EOF
${HASH}  go.tar.gz
EOF

tar zxf go.tar.gz

export PATH="/tmp/go/bin:${PATH}"
cd "${REPO_ROOT}"
go version
