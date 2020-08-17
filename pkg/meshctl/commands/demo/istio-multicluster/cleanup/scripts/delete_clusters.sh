#!/bin/bash -ex

masterCluster=$0
remoteCluster=$1

kind delete cluster --name $masterCluster
kind delete cluster --name $remoteCluster
