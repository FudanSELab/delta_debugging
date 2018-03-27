#!/bin/bash

# Copyright 2017 Istio Authors
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at

#       http://www.apache.org/licenses/LICENSE-2.0

#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/.."
VERSION_FILE="istio.VERSION"
SRC_VERSION_FILE="${ROOT}/${VERSION_FILE}"
TEMP_DIR="/tmp"
DEST_DIR=$ROOT
COMPONENT_FILES=false

set -o errexit
set -o pipefail
set -x

function usage() {
  [[ -n "${1}" ]] && echo "${1}"

  cat <<EOF
usage: ${BASH_SOURCE[0]} [options ...]"
  options:
    -i ... URL to download istioctl binary
    -p ... <hub>,<tag> for the pilot docker image
    -x ... <hub>,<tag> for the mixer docker image
    -c ... <hub>,<tag> for the istio-ca docker image
    -a ... <hub>,<tag> Specifies same hub and tag for pilot, mixer, proxy, and istio-ca containers
    -h ... <hub>,<tag> for the hyperkube docker image
    -r ... tag for proxy debian package
    -n ... <namespace> namespace in which to install Istio control plane components
    -A ... URL to download auth debian packages
    -P ... URL to download pilot debian packages
    -E ... URL to download proxy debian packages
    -d ... directory to store file (optional, defaults to source code tree)
    -D ... enable debug for proxy (optional, false or true, default is false)
    -m ... true|false Create the individual component files as well as the all-in-one
EOF
  exit 2
}

source "$SRC_VERSION_FILE" || error_exit "Could not source versions"

while getopts :i:n:p:x:c:a:h:r:P:d:D:m: arg; do
  case ${arg} in
    i) ISTIOCTL_URL="${OPTARG}";;
    n) ISTIO_NAMESPACE="${OPTARG}";;
    p) PILOT_HUB_TAG="${OPTARG}";; # Format: "<hub>,<tag>"
    x) MIXER_HUB_TAG="${OPTARG}";; # Format: "<hub>,<tag>"
    c) CA_HUB_TAG="${OPTARG}";; # Format: "<hub>,<tag>"
    a) ALL_HUB_TAG="${OPTARG}";; # Format: "<hub>,<tag>"
    h) HYPERKUBE_HUB_TAG="${OPTARG}";; # Format: "<hub>,<tag>"
    r) PROXY_TAG="${OPTARG}";;
    P) PILOT_DEBIAN_URL="${OPTARG}";;
    d) DEST_DIR="${OPTARG}";;
    D) PROXY_DEBUG="${OPTARG}";;
    m) COMPONENT_FILES=true;;
    *) usage;;
  esac
done

if [[ -n ${ALL_HUB_TAG} ]]; then
    PILOT_HUB="$(echo ${ALL_HUB_TAG}|cut -f1 -d,)"
    PILOT_TAG="$(echo ${ALL_HUB_TAG}|cut -f2 -d,)"
    PROXY_HUB="$(echo ${ALL_HUB_TAG}|cut -f1 -d,)"
    PROXY_TAG="$(echo ${ALL_HUB_TAG}|cut -f2 -d,)"
    MIXER_HUB="$(echo ${ALL_HUB_TAG}|cut -f1 -d,)"
    MIXER_TAG="$(echo ${ALL_HUB_TAG}|cut -f2 -d,)"
    CA_HUB="$(echo ${ALL_HUB_TAG}|cut -f1 -d,)"
    CA_TAG="$(echo ${ALL_HUB_TAG}|cut -f2 -d,)"
fi

if [[ -n ${PILOT_HUB_TAG} ]]; then
    PILOT_HUB="$(echo ${PILOT_HUB_TAG}|cut -f1 -d,)"
    PILOT_TAG="$(echo ${PILOT_HUB_TAG}|cut -f2 -d,)"
fi

if [[ -n ${MIXER_HUB_TAG} ]]; then
    MIXER_HUB="$(echo ${MIXER_HUB_TAG}|cut -f1 -d,)"
    MIXER_TAG="$(echo ${MIXER_HUB_TAG}|cut -f2 -d,)"
fi

if [[ -n ${CA_HUB_TAG} ]]; then
    CA_HUB="$(echo ${CA_HUB_TAG}|cut -f1 -d,)"
    CA_TAG="$(echo ${CA_HUB_TAG}|cut -f2 -d,)"
fi

if [[ -n ${HYPERKUBE_HUB_TAG} ]]; then
    HYPERKUBE_HUB="$(echo ${HYPERKUBE_HUB_TAG}|cut -f1 -d,)"
    HYPERKUBE_TAG="$(echo ${HYPERKUBE_HUB_TAG}|cut -f2 -d,)"
fi

# handle PROXY_DEBUG conversion to proxy_debug or proxy image
PROXY_IMAGE="proxy"
if [[ "${PROXY_DEBUG}" == "true" ]]; then
    echo "# Use proxy_debug image"
    PROXY_IMAGE="proxy_debug"
fi

function error_exit() {
  # ${BASH_SOURCE[1]} is the file name of the caller.
  echo "${BASH_SOURCE[1]}: line ${BASH_LINENO[0]}: ${1:-Unknown Error.} (exit ${2:-1})" 1>&2
  exit ${2:-1}
}

#
# In-place portable sed operation
# the sed -i operation is not defined by POSIX and hence is not portable
#
function execute_sed() {
  sed -e "${1}" $2 > $2.new
  mv -- $2.new $2
}

# Generated merge yaml files for easy installation
function merge_files() {
  SRC=$TEMP_DIR/templates
  DEST=$DEST_DIR/install/kubernetes

  # istio.yaml and istio-auth.yaml file contain cluster-wide installations
  ISTIO=$DEST/istio.yaml
  ISTIO_AUTH=$DEST/istio-auth.yaml
  ISTIO_ONE_NAMESPACE=$DEST/istio-one-namespace.yaml
  ISTIO_ONE_NAMESPACE_AUTH=$DEST/istio-one-namespace-auth.yaml
  ISTIO_SIDECAR_INJECTOR=$DEST/istio-sidecar-injector.yaml
  ISTIO_SIDECAR_INJECTOR_CONFIGMAP_DEBUG=$DEST/istio-sidecar-injector-configmap-debug.yaml
  ISTIO_SIDECAR_INJECTOR_CONFIGMAP_RELEASE=$DEST/istio-sidecar-injector-configmap-release.yaml
  ISTIO_CA_PLUGIN_CERTS=$DEST/istio-ca-plugin-certs.yaml
  ISTIO_CA_HEALTH_CHECK=$DEST/istio-ca-with-health-check.yaml


  if [ "$COMPONENT_FILES" = true ]; then
    echo "generating component files"
      COMPONENT_DIR=$DEST/components
      if [ ! -d "$COMPONENT_DIR" ]; then
        mkdir -p $COMPONENT_DIR
      fi
      cat $SRC/istio-ns.yaml.tmpl >> $COMPONENT_DIR/istio-ns.yaml
      cat $SRC/istio-rbac-beta.yaml.tmpl >> $COMPONENT_DIR/istio-rbac-beta.yaml
      cat $SRC/istio-mixer.yaml.tmpl >> $COMPONENT_DIR/istio-mixer.yaml
      cat $SRC/istio-config.yaml.tmpl >> $COMPONENT_DIR/istio-config.yaml
      cat $SRC/istio-pilot.yaml.tmpl >> $COMPONENT_DIR/istio-pilot.yaml
      cat $SRC/istio-ingress.yaml.tmpl >> $COMPONENT_DIR/istio-ingress.yaml
  fi

  echo "# GENERATED FILE. Use with Kubernetes 1.7+" > $ISTIO
  echo "# TO UPDATE, modify files in install/kubernetes/templates and run install/updateVersion.sh" >> $ISTIO
  cat $SRC/istio-ns.yaml.tmpl >> $ISTIO
  cat $SRC/istio-rbac-beta.yaml.tmpl >> $ISTIO
  cat $SRC/istio-mixer.yaml.tmpl >> $ISTIO
  cat $SRC/istio-config.yaml.tmpl >> $ISTIO
  cat $SRC/istio-pilot.yaml.tmpl >> $ISTIO
  cat $SRC/istio-ingress.yaml.tmpl >> $ISTIO

  cp $ISTIO $ISTIO_ONE_NAMESPACE
  cat $SRC/istio-ca.yaml.tmpl >> $ISTIO

  cp $ISTIO $ISTIO_AUTH
  execute_sed "s/# authPolicy: MUTUAL_TLS/authPolicy: MUTUAL_TLS/" $ISTIO_AUTH
  execute_sed "s/# controlPlaneAuthPolicy: MUTUAL_TLS/controlPlaneAuthPolicy: MUTUAL_TLS/" $ISTIO_AUTH
  execute_sed "s/NONE #--controlPlaneAuthPolicy/MUTUAL_TLS/" $ISTIO_AUTH
  execute_sed "s/envoy_mixer.json/envoy_mixer_auth.json/" $ISTIO_AUTH
  execute_sed "s/envoy_pilot.json/envoy_pilot_auth.json/" $ISTIO_AUTH

  # restrict pilot controllers to a single namespace in the test file
  execute_sed "s|args: \[\"discovery\", \"-v\", \"2\"|args: \[\"discovery\", \"-v\", \"2\", \"-a\", \"${ISTIO_NAMESPACE}\"|" $ISTIO_ONE_NAMESPACE
  cat $SRC/istio-ca-one-namespace.yaml.tmpl >> $ISTIO_ONE_NAMESPACE

  cp $ISTIO_ONE_NAMESPACE $ISTIO_ONE_NAMESPACE_AUTH
  execute_sed "s/# authPolicy: MUTUAL_TLS/authPolicy: MUTUAL_TLS/" $ISTIO_ONE_NAMESPACE_AUTH
  execute_sed "s/# controlPlaneAuthPolicy: MUTUAL_TLS/controlPlaneAuthPolicy: MUTUAL_TLS/" $ISTIO_ONE_NAMESPACE_AUTH
  execute_sed "s/NONE #--controlPlaneAuthPolicy/MUTUAL_TLS/" $ISTIO_ONE_NAMESPACE_AUTH
  execute_sed "s/envoy_mixer.json/envoy_mixer_auth.json/" $ISTIO_ONE_NAMESPACE_AUTH
  execute_sed "s/envoy_pilot.json/envoy_pilot_auth.json/" $ISTIO_ONE_NAMESPACE_AUTH

  echo "# GENERATED FILE. Use with Kubernetes 1.9+" > $ISTIO_SIDECAR_INJECTOR
  echo "# TO UPDATE, modify files in install/kubernetes/templates and run install/updateVersion.sh" >> $ISTIO_SIDECAR_INJECTOR
  cat ${SRC}/istio-sidecar-injector.yaml.tmpl >> $ISTIO_SIDECAR_INJECTOR

  echo "# GENERATED FILE. Use with Kubernetes 1.9+" > $ISTIO_SIDECAR_INJECTOR_CONFIGMAP_DEBUG
  echo "# TO UPDATE, modify files in install/kubernetes/templates and run install/updateVersion.sh" >> $ISTIO_SIDECAR_INJECTOR_CONFIGMAP_DEBUG
  cat ${SRC}/istio-sidecar-injector-configmap-debug.yaml.tmpl >> $ISTIO_SIDECAR_INJECTOR_CONFIGMAP_DEBUG

  echo "# GENERATED FILE. Use with Kubernetes 1.9+" > $ISTIO_SIDECAR_INJECTOR_CONFIGMAP_RELEASE
  echo "# TO UPDATE, modify files in install/kubernetes/templates and run install/updateVersion.sh" >> $ISTIO_SIDECAR_INJECTOR_CONFIGMAP_RELEASE
  cat ${SRC}/istio-sidecar-injector-configmap-release.yaml.tmpl >> $ISTIO_SIDECAR_INJECTOR_CONFIGMAP_RELEASE

  echo "# GENERATED FILE. Use with Kubernetes 1.7+" > $ISTIO_CA_PLUGIN_CERTS
  echo "# TO UPDATE, modify files in install/kubernetes/templates and run install/updateVersion.sh" >> $ISTIO_CA_PLUGIN_CERTS
  cat $SRC/istio-ca-plugin-certs.yaml.tmpl >> $ISTIO_CA_PLUGIN_CERTS

  echo "# GENERATED FILE. Use with Kubernetes 1.7+" > $ISTIO_CA_HEALTH_CHECK
  echo "# TO UPDATE, modify files in install/kubernetes/templates and run install/updateVersion.sh" >> $ISTIO_CA_HEALTH_CHECK
  cat $SRC/istio-ca-with-health-check.yaml.tmpl >> $ISTIO_CA_HEALTH_CHECK
}

function update_version_file() {
  cat <<EOF > "${DEST_DIR}/${VERSION_FILE}"
# DO NOT EDIT THIS FILE MANUALLY instead use
# install/updateVersion.sh (see install/README.md)
export CA_HUB="${CA_HUB}"
export CA_TAG="${CA_TAG}"
export MIXER_HUB="${MIXER_HUB}"
export MIXER_TAG="${MIXER_TAG}"
export PILOT_HUB="${PILOT_HUB}"
export PILOT_TAG="${PILOT_TAG}"
export ISTIOCTL_URL="${ISTIOCTL_URL}"
export PROXY_TAG="${PROXY_TAG}"
export PROXY_DEBUG="${PROXY_DEBUG}"
export ISTIO_NAMESPACE="${ISTIO_NAMESPACE}"
export PILOT_DEBIAN_URL="${PILOT_DEBIAN_URL}"
export FORTIO_HUB="${FORTIO_HUB}"
export FORTIO_TAG="${FORTIO_TAG}"
export HYPERKUBE_HUB="${HYPERKUBE_HUB}"
export HYPERKUBE_TAG="${HYPERKUBE_TAG}"
EOF
}

#
# Updating helm's values.yaml for the current versions
#
function update_helm_version() {
  pushd $TEMP_DIR/templates/helm/istio
  local HELM_FILE=$DEST_DIR/install/kubernetes/helm/istio/values.yaml

  execute_sed "s|{CA_HUB}|${CA_HUB}|"       values.yaml.tmpl
  execute_sed "s|{CA_TAG}|${CA_TAG}|"       values.yaml.tmpl
  execute_sed "s|{PROXY_HUB}|${PILOT_HUB}|" values.yaml.tmpl
  execute_sed "s|{PROXY_IMAGE}|${PROXY_IMAGE}|" values.yaml.tmpl
  execute_sed "s|{PROXY_TAG}|${PILOT_TAG}|" values.yaml.tmpl
  execute_sed "s|{PROXY_DEBUG}|${PROXY_DEBUG}|" values.yaml.tmpl
  execute_sed "s|{PILOT_HUB}|${PILOT_HUB}|" values.yaml.tmpl
  execute_sed "s|{PILOT_TAG}|${PILOT_TAG}|" values.yaml.tmpl
  execute_sed "s|{MIXER_HUB}|${MIXER_HUB}|" values.yaml.tmpl
  execute_sed "s|{MIXER_TAG}|${MIXER_TAG}|" values.yaml.tmpl
  execute_sed "s|{HYPERKUBE_HUB}|${HYPERKUBE_HUB}|" values.yaml.tmpl
  execute_sed "s|{HYPERKUBE_TAG}|${HYPERKUBE_TAG}|" values.yaml.tmpl

  echo "# GENERATED FILE. Use with Kubernetes 1.7+" > $HELM_FILE
  echo "# TO UPDATE, modify files in install/kubernetes/templates/helm/istio and run install/updateVersion.sh" >> $HELM_FILE
  cat values.yaml.tmpl >> $HELM_FILE

  popd
}

function update_istio_install() {
  pushd $TEMP_DIR/templates
  execute_sed "s|{ISTIO_NAMESPACE}|${ISTIO_NAMESPACE}|" istio-ns.yaml.tmpl
  execute_sed "s|{ISTIO_NAMESPACE}|${ISTIO_NAMESPACE}|" istio-rbac-beta.yaml.tmpl
  execute_sed "s|{ISTIO_NAMESPACE}|${ISTIO_NAMESPACE}|" istio-config.yaml.tmpl
  execute_sed "s|{ISTIO_NAMESPACE}|${ISTIO_NAMESPACE}|" istio-pilot.yaml.tmpl
  execute_sed "s|{ISTIO_NAMESPACE}|${ISTIO_NAMESPACE}|" istio-ingress.yaml.tmpl
  execute_sed "s|{ISTIO_NAMESPACE}|${ISTIO_NAMESPACE}|" istio-mixer.yaml.tmpl
  execute_sed "s|{ISTIO_NAMESPACE}|${ISTIO_NAMESPACE}|" istio-ca.yaml.tmpl
  execute_sed "s|{ISTIO_NAMESPACE}|${ISTIO_NAMESPACE}|" istio-ca-one-namespace.yaml.tmpl
  execute_sed "s|{ISTIO_NAMESPACE}|${ISTIO_NAMESPACE}|" istio-ca-plugin-certs.yaml.tmpl
  execute_sed "s|{ISTIO_NAMESPACE}|${ISTIO_NAMESPACE}|" istio-ca-with-health-check.yaml.tmpl
  execute_sed "s|{ISTIO_NAMESPACE}|${ISTIO_NAMESPACE}|" istio-sidecar-injector.yaml.tmpl
  execute_sed "s|{ISTIO_NAMESPACE}|${ISTIO_NAMESPACE}|" istio-sidecar-injector-configmap-debug.yaml.tmpl
  execute_sed "s|{ISTIO_NAMESPACE}|${ISTIO_NAMESPACE}|" istio-sidecar-injector-configmap-release.yaml.tmpl

  execute_sed "s|image: {PILOT_HUB}/\(.*\):{PILOT_TAG}|image: ${PILOT_HUB}/\1:${PILOT_TAG}|" istio-pilot.yaml.tmpl
  execute_sed "s|image: {PROXY_HUB}/{PROXY_IMAGE}:{PROXY_TAG}|image: ${PILOT_HUB}/${PROXY_IMAGE}:${PILOT_TAG}|" istio-pilot.yaml.tmpl
  execute_sed "s|image: {MIXER_HUB}/\(.*\):{MIXER_TAG}|image: ${MIXER_HUB}/\1:${MIXER_TAG}|" istio-mixer.yaml.tmpl
  execute_sed "s|image: {PROXY_HUB}/{PROXY_IMAGE}:{PROXY_TAG}|image: ${PILOT_HUB}/${PROXY_IMAGE}:${PILOT_TAG}|" istio-mixer.yaml.tmpl
  execute_sed "s|image: {CA_HUB}/\(.*\):{CA_TAG}|image: ${CA_HUB}/\1:${CA_TAG}|" istio-ca.yaml.tmpl
  execute_sed "s|image: {CA_HUB}/\(.*\):{CA_TAG}|image: ${CA_HUB}/\1:${CA_TAG}|" istio-ca-one-namespace.yaml.tmpl
  execute_sed "s|image: {CA_HUB}/\(.*\):{CA_TAG}|image: ${CA_HUB}/\1:${CA_TAG}|" istio-ca-plugin-certs.yaml.tmpl
  execute_sed "s|image: {CA_HUB}/\(.*\):{CA_TAG}|image: ${CA_HUB}/\1:${CA_TAG}|" istio-ca-with-health-check.yaml.tmpl

  execute_sed "s|{PILOT_HUB}|${PILOT_HUB}|" istio-sidecar-injector.yaml.tmpl
  execute_sed "s|{PILOT_TAG}|${PILOT_TAG}|" istio-sidecar-injector.yaml.tmpl
  execute_sed "s|{PROXY_IMAGE}|${PROXY_IMAGE}|" istio-sidecar-injector.yaml.tmpl

  execute_sed "s|{PILOT_HUB}|${PILOT_HUB}|" istio-sidecar-injector-configmap-debug.yaml.tmpl
  execute_sed "s|{PILOT_TAG}|${PILOT_TAG}|" istio-sidecar-injector-configmap-debug.yaml.tmpl
  execute_sed "s|{PILOT_HUB}|${PILOT_HUB}|" istio-sidecar-injector-configmap-release.yaml.tmpl
  execute_sed "s|{PILOT_TAG}|${PILOT_TAG}|" istio-sidecar-injector-configmap-release.yaml.tmpl


  execute_sed "s|image: {PROXY_HUB}/{PROXY_IMAGE}:{PROXY_TAG}|image: ${PILOT_HUB}/${PROXY_IMAGE}:${PILOT_TAG}|" istio-ingress.yaml.tmpl
  popd
}

function update_istio_addons() {
  DEST=$DEST_DIR/install/kubernetes/addons
  mkdir -p $DEST
  pushd $TEMP_DIR/templates/addons
  execute_sed "s|image: {MIXER_HUB}/\(.*\):{MIXER_TAG}|image: ${MIXER_HUB}/\1:${MIXER_TAG}|" grafana.yaml.tmpl
  execute_sed "s|image: {MIXER_HUB}/\(.*\):{MIXER_TAG}|image: ${MIXER_HUB}/\1:${MIXER_TAG}|" servicegraph.yaml.tmpl
  sed "s|{ISTIO_NAMESPACE}|${ISTIO_NAMESPACE}|" grafana.yaml.tmpl  > $DEST/grafana.yaml
  sed "s|{ISTIO_NAMESPACE}|${ISTIO_NAMESPACE}|" prometheus.yaml.tmpl > $DEST/prometheus.yaml
  sed "s|{ISTIO_NAMESPACE}|${ISTIO_NAMESPACE}|" servicegraph.yaml.tmpl > $DEST/servicegraph.yaml
  sed "s|{ISTIO_NAMESPACE}|${ISTIO_NAMESPACE}|" zipkin.yaml.tmpl > $DEST/zipkin.yaml
  sed "s|{ISTIO_NAMESPACE}|${ISTIO_NAMESPACE}|" zipkin-to-stackdriver.yaml.tmpl > $DEST/zipkin-to-stackdriver.yaml
  popd
}

function update_istio_install_docker() {
  pushd $TEMP_DIR/templates
  execute_sed "s|image: {PILOT_HUB}/\(.*\):{PILOT_TAG}|image: ${PILOT_HUB}/\1:${PILOT_TAG}|" istio.yaml.tmpl
  execute_sed "s|image: {PILOT_HUB}/\(.*\):{PILOT_TAG}|image: ${PILOT_HUB}/\1:${PILOT_TAG}|" bookinfo.sidecars.yaml.tmpl
  execute_sed "s|image: {PILOT_HUB}/\(.*\):{PROXY_TAG}|image: ${PILOT_HUB}/\1:${PROXY_TAG}|" bookinfo.sidecars.yaml.tmpl
  popd
}

# Generated merge yaml files for easy installation
function merge_files_docker() {
  TYPE=$1
  SRC=$TEMP_DIR/templates

  # Merge istio.yaml install file
  INSTALL_DEST=$DEST_DIR/install/$TYPE
  ISTIO=${INSTALL_DEST}/istio.yaml

  mkdir -p $INSTALL_DEST
  echo "# GENERATED FILE. Use with Docker-Compose and ${TYPE}" > $ISTIO
  echo "# TO UPDATE, modify files in install/${TYPE}/templates and run install/updateVersion.sh" >> $ISTIO
  cat $SRC/istio.yaml.tmpl >> $ISTIO

  # Merge bookinfo.sidecars.yaml sample file
  SAMPLES_DEST=$DEST_DIR/samples/bookinfo/$TYPE
  BOOKINFO=${SAMPLES_DEST}/bookinfo.sidecars.yaml

  mkdir -p $SAMPLES_DEST
  echo "# GENERATED FILE. Use with Docker-Compose and ${TYPE}" > $BOOKINFO
  echo "# TO UPDATE, modify files in samples/bookinfo/${TYPE}/templates and run install/updateVersion.sh" >> $BOOKINFO
  cat $SRC/bookinfo.sidecars.yaml.tmpl >> $BOOKINFO
}

if [[ "$DEST_DIR" != "$ROOT" ]]; then
  if [ ! -d "$DEST_DIR" ]; then
    mkdir -p $DEST_DIR
  fi
  cp -R $ROOT/install $DEST_DIR/
  cp -R $ROOT/samples $DEST_DIR/
fi

mkdir -p $TEMP_DIR/templates
cp -R $ROOT/install/kubernetes/templates/* $TEMP_DIR/templates/
update_version_file
update_helm_version
update_istio_install
update_istio_addons
merge_files
rm -R $TEMP_DIR/templates

for platform in consul eureka
do
    cp -R $ROOT/install/$platform/templates $TEMP_DIR/templates
    cp -a $ROOT/samples/bookinfo/$platform/templates/. $TEMP_DIR/templates/
    update_istio_install_docker
    merge_files_docker $platform
    rm -R $TEMP_DIR/templates
done
