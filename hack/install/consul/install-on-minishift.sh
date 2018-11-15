#!/usr/bin/env bash

set -ex

oc new-project consul
kubectl apply -f make-cluster-admin.yaml
helm install -f helm-overrides.yaml https://github.com/hashicorp/consul-helm/archive/v0.3.0.tar.gz
kubectl get mutatingwebhookconfigurations.admissionregistration.k8s.io -o yaml | sed "s/name: .*consul-connect-injector-cfg/name: consul-connect-injector-cfg/g" | kubectl apply --validate=false -f -

set +x

echo "Install successful. May take up to a minute for all pods to be ready."