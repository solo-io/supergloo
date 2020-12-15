#!/bin/bash -ex

#####################################
#
# Set up gloo mesh in the target kind cluster.
#
#####################################
set -o xtrace

cluster=$1
apiServerAddress=$2

PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )/.."
source ${PROJECT_ROOT}/ci/setup-funcs.sh

if [ "${cluster}" == "" ]; then
  cluster=mgmt-cluster
fi

K="kubectl --context kind-${cluster}"

echo "deploying gloo-mesh to ${cluster} from local images..."

## build and load GlooMesh docker images
MAKE="make -C $PROJECT_ROOT"
eval "${MAKE} clean-helm manifest-gen package-helm build-all-images -B"

setChartVariables

agentChart=${AGENT_CHART}
agentImage=${AGENT_IMAGE}
gloomeshChart=${GLOOMESH_CHART}

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
  --chart-file "${gloomeshChart}" \
  --namespace gloo-mesh \
  --cluster-name "${cluster}" \
  --verbose  \
  --api-server-address "${apiServerAddress}" \
  --cert-agent-chart-file "${agentChart}"


${K} -n gloo-mesh rollout status deployment networking
${K} -n gloo-mesh rollout status deployment discovery

echo setup successfully set up gloo-mesh
