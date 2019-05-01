#!/usr/bin/env bash

kubectl delete mutatingwebhookconfigurations.admissionregistration.k8s.io istio-sidecar-injector
kubectl delete ns istio-system
kubectl delete $(kubectl get crd -o name|grep istio)
kubectl delete $(kubectl get clusterroles.rbac.authorization.k8s.io -o name|grep istio)
kubectl delete $(kubectl get clusterrolebindings.rbac.authorization.k8s.io -o name|grep istio)

supergloo init --dryrun | kubectl delete -f -
kubectl delete ns supergloo-system
