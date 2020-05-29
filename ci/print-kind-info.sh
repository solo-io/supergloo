#!/bin/bash

# generate names: $1 allows to make several envs in parallel 
managementPlane=management-plane-$1
remoteCluster=target-cluster-$1

kubectl --context kind-$managementPlane get pod -A
kubectl --context kind-$remoteCluster get pod -A
kubectl --context kind-$managementPlane get meshworkloads -A
kubectl --context kind-$managementPlane get meshservices -A

# and process and disk info to debug out of disk space issues in CI
ps -auxf
df -h