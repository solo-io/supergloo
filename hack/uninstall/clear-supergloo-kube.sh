#!/usr/bin/env bash

set -x

kubectl delete install -n supergloo-system --all
kubectl delete mesh -n supergloo-system --all

