#!/bin/bash
#
# Copyright 2017 Istio Authors
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

set -o errexit

if [ "$#" -ne 1 ]; then
    echo Missing version parameter
    echo Usage: build-services.sh \<version\>
    exit 1
fi

VERSION=$1
SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

pushd $SCRIPTDIR/productpage
  docker build -t istio/examples-bookinfo-productpage-v1:${VERSION} .
popd

pushd $SCRIPTDIR/details
  #plain build -- no calling external book service to fetch topics
  docker build -t istio/examples-bookinfo-details-v1:${VERSION} --build-arg service_version=v1 .
  #with calling external book service to fetch topic for the book
  docker build -t istio/examples-bookinfo-details-v2:${VERSION} --build-arg service_version=v2 \
	 --build-arg enable_external_book_service=true .
popd

pushd $SCRIPTDIR/reviews
  #java build the app.
  docker run --rm -v `pwd`:/usr/bin/app:rw niaquinto/gradle clean build
  pushd reviews-wlpcfg
    #plain build -- no ratings
    docker build -t istio/examples-bookinfo-reviews-v1:${VERSION} --build-arg service_version=v1 .
    #with ratings black stars
    docker build -t istio/examples-bookinfo-reviews-v2:${VERSION} --build-arg service_version=v2 \
	   --build-arg enable_ratings=true .
    #with ratings red stars
    docker build -t istio/examples-bookinfo-reviews-v3:${VERSION} --build-arg service_version=v3 \
	   --build-arg enable_ratings=true --build-arg star_color=red .
  popd
popd

pushd $SCRIPTDIR/ratings
  docker build -t istio/examples-bookinfo-ratings-v1:${VERSION} --build-arg service_version=v1 .
  docker build -t istio/examples-bookinfo-ratings-v2:${VERSION} --build-arg service_version=v2 .
popd

pushd $SCRIPTDIR/mysql
  docker build -t istio/examples-bookinfo-mysqldb:${VERSION} .
popd

pushd $SCRIPTDIR/mongodb
  docker build -t istio/examples-bookinfo-mongodb:${VERSION} .
popd
