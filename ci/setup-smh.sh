#!/bin/bash -ex

#####################################
#
# Set up service mesh hub in the target kind cluster.
#
#####################################

cluster=$1
  
if [ "${cluster}" == "" ]; then
  cluster=master-cluster
fi

K="kubectl --context kind-${cluster}"

echo "deploying smh to ${cluster} from local images..."

# register our CRDs
${K} apply -f ../service-mesh-hub/install/helm/service-mesh-hub/crds

# namespace
${K} create ns service-mesh-hub || echo exists

PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )/.."

INSTALL_DIR="${PROJECT_ROOT}/install/"
DEFAULT_MANIFEST="${INSTALL_DIR}/service-mesh-hub-default.yaml"

## build
MAKE="make -C $PROJECT_ROOT"
eval "${MAKE} manifest-gen service-mesh-hub-image -B"

## install to kube

grep "image:" "${DEFAULT_MANIFEST}" \
  | awk '{print $3}' \
  | while read -r image; do
  kind load docker-image --name "${cluster}" "${image}"
done

${K} apply -n service-mesh-hub -f "${DEFAULT_MANIFEST}"

${K} -n service-mesh-hub rollout status deployment networking
${K} -n service-mesh-hub rollout status deployment discovery

echo setup successfully set up service-mesh-hub
