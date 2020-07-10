#!/bin/bash -ex

#####################################
#
# Set up one kind cluster with Istio (Ingress + Istiod only) and the Petstore application.
#
# To clean up **ALL** of your kind clusters, run this script as: `bash ci/setup-kind.sh cleanup`
#
#####################################

ARG=$1

if [ "$ARG" == "cleanup" ]; then
  kubectl delete ns dev-portal
  exit 0
fi

ARG=${ARG:=default}

## register crds

# register all our CRDs in the management plane
kubectl --context kind-$managementPlane apply -f ../service-mesh-hub/install/helm/charts/custom-resource-definitions/crds
kubectl --context kind-$managementPlane apply -f ../skv2/crds/multicluster.solo.io_v1alpha1_crds.yaml
# register all the CRDs in the target cluster too
##TODO:uncomment kubectl --context kind-$remoteCluster apply -f ../service-mesh-hub/install/helm/charts/custom-resource-definitions/crds


# namespace
kubectl create ns service-mesh-hub

# register cluster (TODO: add second cluster)
go run cmd/cli/main.go cluster register --cluster-name management-cluster-1

return # TODO: complete this

PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )/.."

INSTALL_DIR="${PROJECT_ROOT}/install/"
DEFAULT_MANIFEST="${INSTALL_DIR}/service-mesh-hub-default.yaml"

## build
MAKE="make -C $PROJECT_ROOT"
eval "${MAKE} manifest-gen build-images -B"


## install to kube

grep "image:" "${DEFAULT_MANIFEST}" \
  | grep -v "ext-auth-service" \
  | grep -v "rate-limiter" \
  | grep -v "redis" \
  | awk '{print $3}' \
  | while read -r image; do
  kind load docker-image --name $cluster $image
done

kubectl apply -n service-mesh-hub -f $DEFAULT_MANIFEST

kubectl -n service-mesh-hub rollout status deployment mesh-networking
kubectl -n service-mesh-hub rollout status deployment mesh-discovery

echo setup successfully set up service-mesh-hub
