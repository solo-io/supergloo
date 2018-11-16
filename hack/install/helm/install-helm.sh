#!/usr/bin/env bash

set -ex

# Expected to be run from a machine with kubectl and helm 2 installed
# This script will initialize helm on kubernetes

kubectl apply -f helm-service-account.yaml

# This installs tiller on kubernetes in the "kube-system" namespace
helm init --service-account tiller --upgrade