#!/bin/bash

# deploy the bookinfo app
kubectl create ns bookstore
kubectl apply -n bookstore -f bookinfo.yaml

# show the productpage
xdg-open http://localhost:9080
kubectl port-forward -n bookstore deployment/productpage-v1 9080

# install prometheus & grafana
kubectl apply \
  --filename https://raw.githubusercontent.com/giantswarm/kubernetes-prometheus/master/manifests-all.yaml

# install it all
supergloo init
# if running sg locally
kubectl delete deployment -n supergloo-system supergloo
# install meshes and ingress
supergloo install istio --name istio
supergloo install linkerd --name linkerd
supergloo install gloo --name gloo --target-meshes supergloo-system.istio

# create a route to grafana & show the ui
glooctl add route \
    --name vs \
    --dest-namespace supergloo-system \
    --dest-name monitoring-grafana-3000 \
    --path-prefix /
xdg-open $(glooctl proxy url)

kubectl label namespace bookstore istio-injection=enabled
kubectl delete pods -n bookstore --all


# tear everything down and start fresh
APPS="istio linkerd supergloo gloo bookinfo monitoring"
RESOURCES="clusterrole clusterrolebinding crd mutatingwebhookconfigurations.admissionregistration.k8s.io"
for app in ${APPS}; do for resource in ${RESOURCES}; do kubectl delete $(kubectl get $resource -oname | grep $app); done; done
