#!/usr/bin/env bash

# Copyright 2022 The Kubernetes Authors.
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

set -o errexit -o nounset -o pipefail

set -x

# cd to the repo root
REPO_ROOT=$(git rev-parse --show-toplevel)
cd "${REPO_ROOT}"

TAG="${TAG:-"$(date +v%Y%m%d)-$(git describe --always --dirty)"}"
SERVICE_BASENAME="${SERVICE_BASENAME:-k8s-infra-porche-sandbox}"
IMAGE_REPO="${IMAGE_REPO:-gcr.io/k8s-staging-infra-tools/redirectserver}"
PROJECT="${PROJECT:-k8s-infra-porche-sandbox}"

REGIONS=(
    asia-east1
    asia-northeast1
    asia-northeast2
    asia-south1
    australia-southeast1
    europe-north1
    europe-southwest1
    europe-west1
    europe-west2
    europe-west4
    europe-west8
    europe-west9
    southamerica-west1
    us-central1
    us-east1
    us-east4
    us-east5
    us-south1
    us-west1
    us-west2
)

FALLBACK_LOCATION=https://d1qobdt0ghfgcv.cloudfront.net/

FEATURE_FLAGS=AllowRegionToBeSpecified

BUCKET_PREFIX=test-artifacts-k8s-io

for REGION in "${REGIONS[@]}"; do
    gcloud --project="${PROJECT}" \
        run services update "${SERVICE_BASENAME}-${REGION}" \
        --image "${IMAGE_REPO}:${TAG}" \
        --region "${REGION}" \
        --concurrency 5 \
        --max-instances 3 \
        `# NOTE: should match number of cores configured` \
        --update-env-vars GOMAXPROCS=1,FALLBACK_LOCATION=${FALLBACK_LOCATION},BUCKET_PREFIX=${BUCKET_PREFIX},FEATURE_FLAGS=${FEATURE_FLAGS} \
        `# TODO: if we use this to deploy prod, we need to handle this differently` \
        --args=-v=3
done
