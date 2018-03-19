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

# Print commands
set -x

function process_result() {
    if [[ $1 -eq 0 ]]; then
        SUMMARY+="\nPASSED: $2 "
    else
        SUMMARY+="\nFAILED: $2 "
        ((FAILURE_COUNT++))
    fi
}

echo ${GOPATH}

# Build mixer binary
make mixs
MIXER_BINARY=$(make where-is-out)/mixs
ENVOY_BINARY=$(make where-is-out)/envoy

# Download Proxy Repo
cd ..
ls proxy || git clone https://github.com/istio/proxy
cd proxy
git pull

# A default value for ISTIO_ENVOY_VERSION is set by init.sh
git reset ${ISTIO_ENVOY_VERSION} --hard
#ENVOY_BINARY=$(pwd)/usr/local/bin/envoy
START_ENVOY=$(pwd)/src/envoy/http/mixer/start_envoy
cd ../istio

# Install Fortio
( cd vendor/istio.io/fortio ; go install . )

# Run Tests
TESTSPATH='tests/integration/example/tests'
TOTAL_FAILURE=0
SUMMARY='Tests Summary'

printf "Envoy date:"
ls -l ${ENVOY_BINARY}

printf "Mixer date:"
ls -l ${MIXER_BINARY}

printf "Envoy hash:"
md5sum ${ENVOY_BINARY}

TESTARG=(-envoy_binary ${ENVOY_BINARY} -envoy_start_script ${START_ENVOY} -mixer_binary ${MIXER_BINARY} -fortio_binary fortio)

go test -v ./tests/integration/example/tests/sample1 ${TESTARG[@]} $@
process_result $? sample1

go test -v ./tests/integration/example/tests/sample2 ${TESTARG[@]} $@
process_result $? sample2

printf "${SUMMARY}\n"
exit ${FAILURE_COUNT}
