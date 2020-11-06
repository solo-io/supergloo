#!/bin/bash

set -x

# generate names: $1 allows to make several envs in parallel 
mgmtCluster=mgmt-cluster
remoteCluster=remote-cluster

kubectl --context kind-$mgmtCluster get pod -A
kubectl --context kind-$remoteCluster get pod -A
kubectl --context kind-$mgmtCluster describe pod -A
kubectl --context kind-$remoteCluster describe pod -A
kubectl --context kind-$mgmtCluster get mesh -A
kubectl --context kind-$mgmtCluster get workloads -A
kubectl --context kind-$mgmtCluster get traffictargets -A
kubectl --context kind-$mgmtCluster get trafficpolicies -A -o yaml
kubectl --context kind-$mgmtCluster get accesspolicies -A -o yaml
kubectl --context kind-$mgmtCluster get virtualmesh -A -o yaml

kubectl --context kind-$mgmtCluster -n gloo-mesh logs deployment/discovery
kubectl --context kind-$mgmtCluster -n gloo-mesh logs deployment/networking

kubectl --context kind-$mgmtCluster -n gloo-mesh port-forward deployment/discovery 9091& sleep 2; echo INPUTS:; curl -v localhost:9091/snapshots/input; echo OUTPUTS:; curl -v localhost:9091/snapshots/input; killall kubectl

kubectl --context kind-$mgmtCluster -n gloo-mesh port-forward deployment/networking 9091& sleep 2; echo INPUTS:; curl -v localhost:9091/snapshots/input; echo OUTPUTS:; curl -v localhost:9091/snapshots/input; killall kubectl

# and process and disk info to debug out of disk space issues in CI
# this is too verbose: ps -auxf
df -h
