#!/bin/bash 

#####################################
#
# Set up two kind clusters:
#   1. a management plane which will have a service-mesh-hub namespace, and
#   2. a target cluster
#
# The management plane will have the appropriate secret for communicating with the target cluster
# Your kube context will be left pointing to the management plane cluster
# The target cluster will have Istio set up in the istio-system namespace in its demo profile
#
# To clean up **ALL** of your kind clusters, run this script as: `bash ci/setup-kind.sh cleanup`
# I had some trouble with the docker VM running out of disk space- run this cleanup step often if you can
#
#####################################

set -e

if [ "$1" == "cleanup" ]; then
  kind get clusters | while read -r r; do kind delete cluster --name "$r"; done
  exit 0
fi

make clean

# generate 16-character random suffix on these names
managementPlane=management-plane-$(xxd -l16 -ps /dev/urandom)
remoteCluster=target-cluster-$(xxd -l16 -ps /dev/urandom)

# set up each cluster
kind create cluster --name $managementPlane
kind create cluster --name $remoteCluster

printf "\n\n---\n"
echo "Finished setting up cluster $managementPlane"
echo "Finished setting up cluster $remoteCluster"

# set up kubectl to be pointing to the proper cluster
kubectl config use-context kind-$managementPlane

# ensure service-mesh-hub ns exists
kubectl create ns service-mesh-hub

# register all our CRDs in the management plane
ls install/helm/charts/custom-resource-definitions/crds | while read f; do kubectl apply -f install/helm/charts/custom-resource-definitions/crds/$f; done

# register all the CRDs in the target cluster too
ls install/helm/charts/custom-resource-definitions/crds | while read f; do kubectl --context kind-$remoteCluster apply -f install/helm/charts/custom-resource-definitions/crds/$f; done

# make all the docker images
# write the output to a temp file so that we can grab the image names out of it
# also ensure we clean up the file once we're done
tempFile=/tmp/images
make docker -B | tee $tempFile

function cleanup {
  rm $tempFile
}

trap cleanup EXIT

# grab the image names out of the `make docker` output
# the kind cluster name is just $managementPlane, not kind-$managementPlane; the latter is how kubectl is aware of it
sed -nE 's|Successfully tagged (.*$)|\1|p' $tempFile | while read f; do kind load docker-image --name $managementPlane $f; kind load docker-image --name $remoteCluster $f; done

# package up Helm
make package-index-app-helm -B

# install the app
# the helm version needs to strip the leading v out of the git describe output
helmVersion=$(git describe --tags --dirty | sed -E 's|^v(.*$)|\1|')
helm \
  -n service-mesh-hub \
  install service-mesh-hub \
  ./_output/helm/charts/management-plane/service-mesh-hub-$helmVersion.tgz

# generate the meshctl binary, register the remote cluster, and install Istio onto the remote cluster
make meshctl -B
./_output/meshctl cluster register \
  --remote-context kind-$managementPlane \
  --remote-cluster-name management-plane-cluster \
  --local-cluster-domain-override host.docker.internal \
  --dev-csr-agent-chart

./_output/meshctl cluster register \
  --remote-context kind-$remoteCluster \
  --remote-cluster-name target-cluster \
  --local-cluster-domain-override host.docker.internal \
  --dev-csr-agent-chart

./_output/meshctl istio install --profile=default --context kind-$remoteCluster
./_output/meshctl istio install --profile=minimal --context kind-$managementPlane
