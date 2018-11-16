#!/usr/bin/env bash

set -ex

VM_DRIVER=${1:-virtualbox}

minishift start --vm-driver=$VM_DRIVER --memory=8192MB
oc login -u system:admin
