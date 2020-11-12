#!/bin/bash -ex

#####################################
#
# Set up gloo mesh in the target kind cluster.
#
#####################################

cluster=$1
apiServerAddress=$2

PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )/.."

if [ "${cluster}" == "" ]; then
  cluster=mgmt-cluster
fi

K="kubectl --context kind-${cluster}"

echo "deploying gloo-mesh to ${cluster} from local images..."

## build and load GlooMesh docker images
MAKE="make -C $PROJECT_ROOT"
eval "${MAKE} clean-helm manifest-gen package-helm build-all-images -B"

INSTALL_DIR="${PROJECT_ROOT}/install/"
DEFAULT_MANIFEST="${INSTALL_DIR}/gloo-mesh-default.yaml"

AGENT_VALUES=${INSTALL_DIR}/helm/cert-agent/values.yaml
AGENT_IMAGE_REGISTRY=$(cat ${AGENT_VALUES} | grep "registry: " | awk '{print $2}')
AGENT_IMAGE_REPOSITORY=$(cat ${AGENT_VALUES} | grep "repository: " | awk '{print $2}')
AGENT_IMAGE_TAG=$(cat ${AGENT_VALUES} | grep "tag: " | awk '{print $2}' | sed 's/"//g')

agentChart=${INSTALL_DIR}/helm/_output/charts/cert-agent/cert-agent-${AGENT_IMAGE_TAG}.tgz
agentImage=${AGENT_IMAGE_REGISTRY}/${AGENT_IMAGE_REPOSITORY}:${AGENT_IMAGE_TAG}

GLOOMESH_VALUES=${INSTALL_DIR}/helm/gloo-mesh/values.yaml
GLOOMESH_IMAGE_TAG=$(cat ${GLOOMESH_VALUES} | grep -m 1 "tag: " | awk '{print $2}' | sed 's/"//g')
glooMeshChart=${INSTALL_DIR}/helm/_output/charts/gloo-mesh/gloo-mesh-${GLOOMESH_IMAGE_TAG}.tgz

# load GlooMesh discovery and networking images
grep "image:" "${DEFAULT_MANIFEST}" \
  | awk '{print $3}' \
  | while read -r image; do
  kind load docker-image --name "${cluster}" "${image}"
done
# load cert-agent image
kind load docker-image --name "${cluster}" "${agentImage}"

## install to kube

go run "${PROJECT_ROOT}/cmd/meshctl/main.go" install \
  --kubecontext kind-"${cluster}" \
  --chart-file "${glooMeshChart}" \
  --namespace gloo-mesh \
  --register \
  --cluster-name "${cluster}" \
  --verbose  \
  --api-server-address "${apiServerAddress}" \
  --cert-agent-chart-file "${agentChart}"


${K} -n gloo-mesh rollout status deployment networking
${K} -n gloo-mesh rollout status deployment discovery

echo setup successfully set up gloo-mesh
