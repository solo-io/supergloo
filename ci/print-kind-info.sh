#!/bin/bash

set -x

# generate names: $1 allows to make several envs in parallel 
managementPlane=management-plane-$1
remoteCluster=target-cluster-$1

kubectl --context kind-$managementPlane get pod -A
kubectl --context kind-$remoteCluster get pod -A
kubectl --context kind-$managementPlane get mesh -A
kubectl --context kind-$managementPlane get meshworkloads -A
kubectl --context kind-$managementPlane get meshservices -A
kubectl --context kind-$managementPlane get trafficpolicies -A -o yaml
kubectl --context kind-$managementPlane get settings -A -o yaml
kubectl --context kind-$managementPlane get virtualmesh -A -o yaml

kubectl --context kind-$managementPlane get virtualservices.networking.istio.io -A  -o yaml
kubectl --context kind-$managementPlane get destinationrules.networking.istio.io -A -o yaml

kubectl --context kind-$managementPlane -n service-mesh-hub logs deployment/mesh-discovery
kubectl --context kind-$managementPlane -n service-mesh-hub logs deployment/mesh-networking

# and process and disk info to debug out of disk space issues in CI
# this is too verbose: ps -auxf
df -h
