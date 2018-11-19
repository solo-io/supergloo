#!/usr/bin/env bash

set -ex

# Expected to be run from a machine with kubectl and helm 2 installed
# This script will initialize helm on kubernetes

if [ ! -f helm-service-account.yaml ]; then
    echo "File 'helm-service-account.yaml' found! Make sure you run out of the hack/install/helm directory."
    exit 1
fi

kubectl apply -f helm-service-account.yaml

# This installs tiller on kubernetes in the "kube-system" namespace
helm init --service-account tiller --upgrade
