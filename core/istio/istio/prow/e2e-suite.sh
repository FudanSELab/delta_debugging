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


#######################################################
# e2e-suite triggered after istio/presubmit succeeded #
#######################################################

WD=$(dirname $0)
WD=$(cd $WD; pwd)
ROOT=$(dirname $WD)

# Exit immediately for non zero status
set -e
# Check unset variables
set -u
# Print commands
set -x

TESTS_TARGETS="e2e_simple e2e_mixer e2e_bookinfo e2e_upgrade"
SINGLE_MODE=false
E2E_ARGS=()

source ${ROOT}/prow/lib.sh
setup_and_export_git_sha

if [ "${CI:-}" == 'bootstrap' ]; then
  # bootsrap upload all artifacts in _artifacts to the log bucket.
  ARTIFACTS_DIR=${ARTIFACTS_DIR:-"${GOPATH}/src/istio.io/istio/_artifacts"}
  E2E_ARGS+=(--test_logs_path="${ARTIFACTS_DIR}")
fi

ISTIO_GO=$(cd $(dirname $0)/..; pwd)

export HUB=${HUB:-"gcr.io/istio-testing"}
export TAG="${GIT_SHA}"

make init

# getopts only handles single character flags
for ((i=1; i<=$#; i++)); do
    case ${!i} in
        # -s/--single_test to specify only one test to run.
        # e.g. "-s e2e_mixer" will only trigger e2e mixer_test
        -s|--single_test) SINGLE_MODE=true; ((i++)); SINGLE_TEST=${!i}
        continue
        ;;
    esac
    E2E_ARGS+=( ${!i} )
done

echo 'Running ISTIO E2E Test(s)'
if ${SINGLE_MODE}; then
    echo "Executing single e2e test"

    # Check if it's a valid test file
    VALID_TEST=false
    for T in ${TESTS_TARGETS[@]}; do
        if [ "${T}" == "${SINGLE_TEST}" ]; then
            VALID_TEST=true
            time ISTIO_DOCKER_HUB=$HUB E2E_ARGS="${E2E_ARGS[@]}" make "${SINGLE_TEST}"
        fi
    done
    if [ "${VALID_TEST}" == "false" ]; then
      echo "Invalid e2e test target, must be one of ${TESTSPATH}"
      # Fail if it's not a valid test file
      process_result 1 'Invalid test target'
    fi

else
    echo "Executing e2e test suite"
    time ISTIO_DOCKER_HUB=$HUB E2E_ARGS="${E2E_ARGS[@]}" make e2e_all
fi

if [ "${CI:-}" == 'bootstrap' ] && [ -f junit.xml ]; then
  # allow bootsrap to upload junit results
  mv junit.xml ${ARTIFACTS_DIR}
fi
