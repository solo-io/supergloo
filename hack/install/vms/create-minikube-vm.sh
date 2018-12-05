#!/usr/bin/env bash

set -ex

VM_DRIVER=${1:-virtualbox}

minikube start --vm-driver=$VM_DRIVER --memory=8192 --cpus=4 --kubernetes-version=v1.10.0