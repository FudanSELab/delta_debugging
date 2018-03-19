#!/bin/bash
#
# Copyright 2017,2018 Istio Authors. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# Script entry into Istio smoketest run with a k8s cluster with Istio installed.
# This is intended as a smoketest for external callers that require a simple
# script entry point. $WORKSPACE is the root dir into which the Istio repo is
# expected to be cloned at the desired version.
#
# Usage: e2e_istio_preinstalled.sh all|bookinfo|mixer|simple

set -o errexit
set -o nounset
set -o pipefail

[[  $1 == "all" || $1 == "bookinfo" || $1 == "mixer" || $1 == "simple" ]] || { echo "$1 is not a legal test name"; exit 1; }

declare -a tests
[[  $1 == "all" ]] && tests=("bookinfo" "mixer" "simple") || tests=($1)

HUB=gcr.io/istio-testing

cd ${WORKSPACE}/github.com/istio/istio
TAG=`git rev-parse HEAD`

for t in ${tests[@]}; do
  go test -v -timeout 20m ./tests/e2e/tests/${t} -args \
  --skip_setup --namespace istio-system \
  --mixer_tag ${TAG} --pilot_tag ${TAG} --proxy_tag ${TAG} --ca_tag ${TAG} \
  --mixer_hub ${HUB} --pilot_hub ${HUB} --proxy_hub ${HUB} --ca_hub ${HUB} \
  --istioctl_url https://storage.googleapis.com/istio-artifacts/pilot/${TAG}/artifacts/istioctl
done
