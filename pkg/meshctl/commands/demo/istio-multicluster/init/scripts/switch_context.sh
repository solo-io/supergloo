#!/bin/bash -ex

managementCluster=$0

# set current context to management cluster
kubectl config use-context kind-${managementCluster}
