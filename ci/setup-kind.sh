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

if [ "$1" == "cleanup" ]; then
  kind get clusters | while read -r r; do kind delete cluster --name "$r"; done
  exit 0
fi

managementPlane=management-plane-$(xxd -l16 -ps /dev/urandom)
remoteCluster=target-cluster-$(xxd -l16 -ps /dev/urandom)

kind create cluster --name $managementPlane
kind create cluster --name $remoteCluster

printf "\n\n---\n"
echo "Finished setting up cluster $managementPlane"
echo "Finished setting up cluster $remoteCluster"

kubectl config use-context kind-$managementPlane
kubectl create ns service-mesh-hub
ls install/helm/charts/custom-resource-definitions/crds | while read f; do kubectl apply -f install/helm/charts/custom-resource-definitions/crds/$f; done

make meshctl -B

./_output/meshctl cluster register --remote-context kind-$remoteCluster --remote-cluster-name target-cluster

./_output/meshctl istio install --profile=demo --context kind-$remoteCluster
