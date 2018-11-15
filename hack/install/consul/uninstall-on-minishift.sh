#!/usr/bin/env bash

set -ex

kubectl get mutatingwebhookconfigurations.admissionregistration.k8s.io | grep consul | awk '{print $1}' | xargs kubectl delete mutatingwebhookconfigurations.admissionregistration.k8s.io
oc delete project consul

set +x

echo "Uninstall successful. May take up to a minute for namespace to be terminated."