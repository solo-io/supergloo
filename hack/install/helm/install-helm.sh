#!/usr/bin/env bash

set -ex

# Expected to be run from a machine with kubectl and helm installed
# This script will initialize helm on kubernetes

kubectl apply -f helm-service-account.yaml

# Required for helm 2, this installs tiller on kubernetes in the "kube-system" namespace
helm init --service-account tiller --upgrade