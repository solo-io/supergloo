---
title: Multicluster Communication
menuTitle: Multicluster Communication
weight: 75
---

## Pre-Guide Notes

To illustrate the concepts in this guide, we will assume that:

* we have just installed Istio as described in the ["Installing Istio" guide]({{% versioned_link_path fromRoot="/guides/installing_istio" %}})
* Service Mesh Hub is running
* two clusters running Istio have been registered as described in the [top-level intro]({{% versioned_link_path fromRoot="/guides" %}})
* the two meshes are grouped together in a *Virtual Mesh*

A relevant definition of a Service Mesh Hub concept that we'll need in this guide:

* *Virtual Mesh* - a Virtual Mesh is a like a higher-level mesh, and may contain one or more service meshes. 
Global Service Mesh Hub configuration like enforcement of access control is applied on the level of a Virtual Mesh, and 
services in meshes that are grouped together in a Virtual Mesh can communicate with each other as if the network was flat.

## Guide

Deploy part of the bookinfo application to the management plane cluster:

```shell
kubectl config use-context management-plane-context # your management-plane context name may be different
kubectl label namespace default istio-injection=enabled
​
# we deploy everything except reviews-v3 to the management-plane cluster
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'app,version notin (v3)'
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'account'
```

Now deploy only reviews-v3 to your remote cluster:

```shell
kubectl config use-context remote-cluster-context # your remote cluster context name may be different

kubectl label namespace default istio-injection=enabled
​
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'app,version in (v3)' --context $REMOTE_CTX
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'service=reviews' --context $REMOTE_CTX
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'account=reviews' --context $REMOTE_CTX
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'app=ratings' --context $REMOTE_CTX
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'account=ratings' --context $REMOTE_CTX
```

We will now perform a *multicluster traffic split*, splitting traffic from the productpage service in the management plane cluster
between local v1 and v2 instances of the reviews service and a remote v3 instance of the reviews service:

```shell
# make sure that your context is still pointing to the management plane cluster

kubectl apply -f - <<EOF
apiVersion: networking.zephyr.solo.io/v1alpha1
kind: TrafficPolicy
metadata:
  namespace: service-mesh-hub
  name: simple
spec:
  destinationSelector:
    serviceRefs:
      services:
        - cluster: management-plane
          name: reviews
          namespace: default
  trafficShift:
    destinations:
      - destination:
          cluster: remote-cluster
          name: reviews
          namespace: default
        weight: 75
      - destination:
          cluster: management-plane
          name: reviews
          namespace: default
        weight: 15
        subset:
          version: v1
      - destination:
          cluster: management-plane
          name: reviews
          namespace: default
        weight: 10
        subset:
          version: v2
EOF
```

You should occasionally see traffic being routed to the reviews-v3 service, which will produce red-colored stars on the product page.

To go into slightly more detail here: The above TrafficPolicy says that:

* any traffic destined for the *reviews service* in the *management plane cluster*
* should be split across several different destinations
* 25% of traffic gets split between the v1 and v2 instances of the reviews service in the management plane cluster
* 75% of traffic gets sent to the v3 instance of the reviews service on the remote cluster

We have successfully set up multicluster traffic across our application! Note that this has been done transparently to the application;
the application can continue talking to what it believes is the local instance of the service, but, behind the scenes, we have
instead routed that traffic to an entirely different cluster. Note that this is interesting in its own right, that we have easily
achieved multicluster communication, but also in other scenarios like in fast disaster recovery: We can quickly route traffic 
to healthy instances of a service in an entirely different datacenter.
