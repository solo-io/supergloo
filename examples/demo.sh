#!/bin/bash

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

# deploy the bookinfo app
kubectl create ns bookstore
kubectl label namespace bookstore istio-injection=enabled
kubectl apply -n bookstore -f bookinfo.yaml

# show the productpage
xdg-open http://localhost:9080
kubectl port-forward -n bookstore deployment/productpage-v1 9080

glooctl add route \
    --name vs \
    --dest-namespace supergloo-system \
    --dest-name bookstore-productpage-9080 \
    --path-prefix /


# create a route to grafana & show the ui
glooctl add route \
    --name vs \
    --dest-namespace supergloo-system \
    --dest-name monitoring-grafana-3000 \
    --path-prefix /
xdg-open $(glooctl proxy url)

supergloo set mesh stats \
    --target-mesh supergloo-system.istio \
    --prometheus-configmap monitoring.prometheus-core


supergloo apply routingrule trafficshifting \
    --dest-upstreams supergloo-system.bookstore-reviews-9080 \
    --target-mesh supergloo-system.istio \
    --destination supergloo-system.bookstore-reviews-v3-9080:1 \
    --name reviews-blue-green

supergloo apply routingrule faultinjection abort http \
    --target-mesh supergloo-system.istio \
     -p 50 -s 404  --name rule1 \
    --dest-upstreams supergloo-system.bookstore-reviews-9080



supergloo apply routingrule trafficshifting \
    --dest-upstreams supergloo-system.bookstore-reviews-9080 \
    --target-mesh supergloo-system.istio \
    --destination supergloo-system.bookstore-reviews-v2-9080:1 \
    --name reviews-blue-green



# tear everything down and start fresh
APPS="istio linkerd supergloo gloo bookinfo monitoring"
RESOURCES="clusterrole clusterrolebinding crd mutatingwebhookconfigurations.admissionregistration.k8s.io"
for app in ${APPS}; do for resource in ${RESOURCES}; do kubectl delete $(kubectl get $resource -oname | grep $app)&; done; done
k delete ns bookstore istio-system linkerd supergloo-system gloo
