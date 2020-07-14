#!/bin/bash

#####################################
#
# Set up a single Kind cluster with SMH installed
#
#####################################

if [ "$1" == "cleanup" ]; then
  kind delete cluster --name "management-plane"
  exit 0
fi

make clean

# generate 16-character random suffix on these names
managementPlane=management-plane

# set up each cluster
# This config is roughly based on: https://kind.sigs.k8s.io/docs/user/ingress/
cat <<EOF | kind create cluster --name $managementPlane --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
        authorization-mode: "AlwaysAllow"
  extraPortMappings:
  - containerPort: 32001
    hostPort: 32001
    protocol: TCP
EOF

printf "\n\n---\n"
echo "Finished setting up cluster $managementPlane"

# set up kubectl to be pointing to the proper cluster
kubectl config use-context kind-$managementPlane

# ensure service-mesh-hub ns exists
kubectl create ns --context kind-$managementPlane  service-mesh-hub

# leaving this in for the time being as there is a race with helm installing CRDs
# register all our CRDs in the management plane
ls install/helm/charts/custom-resource-definitions/crds | while read f; do kubectl --context kind-$managementPlane apply -f install/helm/charts/custom-resource-definitions/crds/$f; done


# make all the docker images
# write the output to a temp file so that we can grab the image names out of it
# also ensure we clean up the file once we're done
tempFile=/tmp/images

# don't proceed if the code doesn't compile
set -e
make docker -B | tee $tempFile

# allow failures again- this is how the waiting works below
set +e

function cleanup {
  rm $tempFile
}

trap cleanup EXIT

# grab the image names out of the `make docker` output
# the kind cluster name is just $managementPlane, not kind-$managementPlane; the latter is how kubectl is aware of it
sed -nE 's|Successfully tagged (.*$)|\1|p' $tempFile | while read f; do kind load docker-image --name $managementPlane $f; done

# create Helm packages
make -s package-index-mgmt-plane-helm -B
make -s package-index-csr-agent-helm -B

# generate the meshctl binary
make meshctl -B
# install the app
# the helm version needs to strip the leading v out of the git describe output
helmVersion=$(git describe --tags --dirty | sed -E 's|^v(.*$)|\1|')
./_output/meshctl install --file ./_output/helm/charts/management-plane/service-mesh-hub-$helmVersion.tgz

#register the management cluster
./_output/meshctl cluster register \
  --remote-context kind-$managementPlane \
  --remote-cluster-name management-plane-cluster \
  --local-cluster-domain-override host.docker.internal \
  --dev-csr-agent-chart
