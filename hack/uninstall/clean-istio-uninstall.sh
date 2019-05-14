#!/usr/bin/env bash

supergloo init --dry-run | kubectl delete -f -
kubectl delete ns supergloo-system
APPS="istio linkerd supergloo gloo bookinfo monitoring"
RESOURCES="clusterrole clusterrolebinding crd mutatingwebhookconfigurations.admissionregistration.k8s.io"
for app in ${APPS}; do for resource in ${RESOURCES}; do echo deleting $resource $app; kubectl delete $(kubectl get $resource -oname | grep $app)& done; done
