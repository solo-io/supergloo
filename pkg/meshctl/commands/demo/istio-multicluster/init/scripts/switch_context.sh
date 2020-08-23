#!/bin/bash -ex

mgmtCluster=$0

# set current context to management cluster
kubectl config use-context kind-${mgmtCluster}
