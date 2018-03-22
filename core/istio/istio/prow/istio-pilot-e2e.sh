#!/bin/bash

# Copyright 2017 Istio Authors

#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at

#       http://www.apache.org/licenses/LICENSE-2.0

#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

WD=$(dirname $0)
WD=$(cd $WD; pwd)
ROOT=$(dirname $WD)

#######################################
# Presubmit script triggered by Prow. #
#######################################

# Exit immediately for non zero status
set -e
# Check unset variables
set -u
# Print commands
set -x

source ${ROOT}/prow/lib.sh
setup_and_export_git_sha

export NUM_NODES=4
source "${ROOT}/prow/cluster_lib.sh"

trap delete_cluster EXIT
create_cluster 'e2e-pilot'

HUB="gcr.io/istio-testing"

cd ${GOPATH}/src/istio.io/istio
make depend e2e_pilot HUB="${HUB}" TAG="${GIT_SHA}" TESTOPTS="-logtostderr -mixer=true -use-sidecar-injector=true -use-admission-webhook=false"
