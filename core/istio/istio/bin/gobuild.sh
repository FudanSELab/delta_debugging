#!/bin/bash
#
# Copyright 2017 Istio Authors. All Rights Reserved.
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
# This script builds and link stamps the output

VERBOSE=${VERBOSE:-"0"}
V=""
if [[ "${VERBOSE}" == "1" ]];then
    V="-x"
    set -x
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

OUT=${1:?"output path"}
VERSION_PACKAGE=${2:?"version go package"} # istio.io/istio/pkg/version
BUILDPATH=${3:?"path to build"}

set -e

GOOS=${GOOS:-linux}
GOARCH=${GOARCH:-amd64}
GOBINARY=${GOBINARY:-go}
GOPKG="$GOPATH/pkg"
BUILDINFO=${BUILDINFO:-""}
STATIC=${STATIC:-1}
LDFLAGS="-extldflags -static"
GOBUILDFLAGS=${GOBUILDFLAGS:-""}
GCFLAGS=${GCFLAGS:-}
export CGO_ENABLED=0

if [[ "${STATIC}" !=  "1" ]];then
    LDFLAGS=""
fi

# gather buildinfo if not already provided
# For a release build BUILDINFO should be produced
# at the beginning of the build and used throughout
if [[ -z ${BUILDINFO} ]];then
    BUILDINFO=$(mktemp)
    ${ROOT}/bin/get_workspace_status > ${BUILDINFO}
fi

# BUILD LD_VERSIONFLAGS
LD_VERSIONFLAGS=""
while read line; do
    read SYMBOL VALUE < <(echo $line)
    LD_VERSIONFLAGS=${LD_VERSIONFLAGS}" -X ${VERSION_PACKAGE}.${SYMBOL}=${VALUE}"
done < "${BUILDINFO}"

# forgoing -i (incremental build) because it will be deprecated by tool chain. 
time GOOS=${GOOS} GOARCH=${GOARCH} ${GOBINARY} build ${V} ${GOBUILDFLAGS} ${GCFLAGS:+-gcflags "${GCFLAGS}"} -o ${OUT} \
       -pkgdir=${GOPKG}/${GOOS}_${GOARCH} -ldflags "${LDFLAGS} ${LD_VERSIONFLAGS}" "${BUILDPATH}"
