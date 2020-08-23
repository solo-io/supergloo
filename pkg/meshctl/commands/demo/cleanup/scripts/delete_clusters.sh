#!/bin/bash -ex

mgmtCluster=$0
remoteCluster=$1

kind delete cluster --name $mgmtCluster

if [ "$remoteCluster" != "" ]; then
  kind delete cluster --name $remoteCluster
fi