#!/bin/bash -ex

cluster=$0

mgmtIngressAddress=$(kubectl --context kind-mgmt-cluster get node -ojson | jq -r ".items[0].status.addresses[0].address")
mgmtIngressPort=$(kubectl --context kind-mgmt-cluster -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="http2")].nodePort}')

printf ${mgmtIngressAddress}:${mgmtIngressPort}
