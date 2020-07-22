#!/bin/bash

set -x

# generate names: $1 allows to make several envs in parallel 
masterCluster=master-cluster
remoteCluster=remote-cluster

kubectl --context kind-$masterCluster get pod -A
kubectl --context kind-$remoteCluster get pod -A
kubectl --context kind-$masterCluster get mesh -A
kubectl --context kind-$masterCluster get meshworkloads -A
kubectl --context kind-$masterCluster get meshservices -A
kubectl --context kind-$masterCluster get trafficpolicies -A -o yaml
kubectl --context kind-$masterCluster get accesspolicies -A -o yaml
kubectl --context kind-$masterCluster get virtualmesh -A -o yaml

kubectl --context kind-$masterCluster -n service-mesh-hub logs deployment/discovery
kubectl --context kind-$masterCluster -n service-mesh-hub logs deployment/networking

# and process and disk info to debug out of disk space issues in CI
# this is too verbose: ps -auxf
df -h
