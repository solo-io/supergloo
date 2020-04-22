---
title: Access Control Intro
menuTitle: Access Control Intro
weight: 30
---

## Pre-Guide Notes

To illustrate the concepts in this guide, we will assume that:

* we have just installed Istio as described in the ["Installing Istio" guide]({{% versioned_link_path fromRoot="/guides/installing_istio" %}})
* Service Mesh Hub is running
* the cluster running Istio has been registered as described in the [top-level intro]({{% versioned_link_path fromRoot="/guides" %}})
* after Istio was installed, you deployed the `bookinfo` app into Istio's cluster as described in the [top-level intro]({{% versioned_link_path fromRoot="/guides" %}})

A relevant definition of a Service Mesh Hub concept that we'll need in this guide:

* *Virtual Mesh* - a Virtual Mesh is a like a higher-level mesh, and may contain one or more service meshes. 
Global Service Mesh Hub configuration like enforcement of access control is applied on the level of a Virtual Mesh, and 
services in meshes that are grouped together in a Virtual Mesh can communicate with each other as if the network was flat.

## Guide

Ensure that your kubeconfig has the correct context set as its `currentContext`:

```shell
kubectl config use-context remote-cluster-context
```

In another shell, start a port-forward to the bookinfo demo:

```shell
kubectl -n default port-forward deployment/productpage-v1 9080:9080
```

In a browser, visit `localhost:9080` (potentially selecting "normal user" if this is your first time using the app)
and verify that both the book details and the reviews are loading correctly. 1/3 of the time, the reviews should have
no stars, 1/3 of the time it will have black stars, and 1/3 of the time it should have red stars.

Now we're going to enable strict access control, which will prevent communication between services by default. Switch your
kubeconfig context back to the management plane so we can configure Service Mesh Hub:

```shell
kubectl config use-context management-plane-context
```

Let's review the discovered name of the Istio mesh we are working with:

```shell
kubectl -n service-mesh-hub get meshes
```

```
NAME                                    AGE
istio-istio-system-new-remote-cluster   99s
```

Now we're going to put that discovered mesh in a Virtual Mesh so that we can apply Virtual-Mesh-level configuration to it:

```shell
kubectl apply -f - <<EOF
apiVersion: networking.zephyr.solo.io/v1alpha1
kind: VirtualMesh
metadata:
  name: demo-virtual-mesh
  namespace: service-mesh-hub
spec:
  enforceAccessControl: true
  meshes:
  - name: istio-istio-system-new-remote-cluster
    namespace: service-mesh-hub
EOF
```

Note the setting `enforceAccessControl: true`. Since we have no `ALLOW` policies set up, all traffic should be denied by default.
We can reload `localhost:9080` and observe that the page should be broken. 

To allow traffic to continue as before, we apply the following two pieces of config. Respectively these
configs:

1. Allow the productpage service to talk to anything in the default namespace
2. Allow the reviews service to talk ony to the ratings service

```shell
kubectl apply --context $MGMT_PLANE_CTX -f - <<EOF
apiVersion: networking.zephyr.solo.io/v1alpha1
kind: AccessControlPolicy
metadata:
  namespace: service-mesh-hub
  name: productpage
spec:
  sourceSelector:
    serviceAccountRefs:
      serviceAccounts:
        - name: bookinfo-productpage
          namespace: default
          cluster: remote-cluster
  destinationSelector:
    matcher:
      namespaces:
        - default
---
apiVersion: networking.zephyr.solo.io/v1alpha1
kind: AccessControlPolicy
metadata:
  namespace: service-mesh-hub
  name: reviews
spec:
  sourceSelector:
    serviceAccountRefs:
      serviceAccounts:
        - name: bookinfo-reviews
          namespace: default
          cluster: remote-cluster
  destinationSelector:
    matcher:
      namespaces:
        - default
      labels:
        service: ratings
EOF
```

Observe that "Book Details" and "Book Reviews" (with the ratings stars) are working again because we have enabled traffic to the backing services.
