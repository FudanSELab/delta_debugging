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


#######################################
#                                     #
#             e2e-suite               #
#                                     #
#######################################

# Exit immediately for non zero status
set -e
# Check unset variables
set -u
# Print commands
set -x

# Check https://github.com/istio/test-infra/blob/master/boskos/configs.yaml
# for exiting resources types
RESOURCE_TYPE='gke-e2e-test'
OWNER="$(basename "${BASH_SOURCE[0]}")"
INFO_PATH="$(mktemp)"
FILE_LOG="$(mktemp)"
ROOT=$(cd $(dirname $0)/..; pwd)

function cleanup() {
  mason_cleanup
  cat "${FILE_LOG}"
}

source "${ROOT}/prow/mason_lib.sh"
source "${ROOT}/prow/cluster_lib.sh"

trap mason_cleanup EXIT
get_resource "${RESOURCE_TYPE}" "${OWNER}" "${INFO_PATH}" "${FILE_LOG}"
setup_cluster

echo 'Running e2e with rbac, no auth Tests'
./prow/e2e-suite.sh "$@"
