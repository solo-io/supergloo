#!/usr/bin/env bash

kubectl delete ns --all
kubectl delete mutatingwebhookconfigurations.admissionregistration.k8s.io istio-sidecar-injector

kubectl delete ns bookinfo consul supergloo-system gloo-system istio-system
kubectl delete mutatingwebhookconfigurations.admissionregistration.k8s.io --all
kubectl delete pod static-client static-server
helm del --purge test-consul-mesh
helm del --purge test-istio-mesh
kubectl delete $(kubectl get crd -o name|grep istio)
kubectl delete $(kubectl get clusterroles.rbac.authorization.k8s.io -o name|grep istio)
kubectl delete $(kubectl get clusterrolebindings.rbac.authorization.k8s.io -o name|grep istio)
kubectl delete configmaps istio-galley-configuration