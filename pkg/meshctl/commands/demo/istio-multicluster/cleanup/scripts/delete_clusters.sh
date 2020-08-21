#!/bin/bash -ex

managementCluster=$0
remoteCluster=$1

kind delete cluster --name $managementCluster
kind delete cluster --name $remoteCluster
