#!/bin/bash -ex

masterCluster=$0

# set current context to master cluster
kubectl config use-context kind-${masterCluster}
