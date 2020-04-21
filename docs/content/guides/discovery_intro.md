---
title: Discovery Intro
menuTitle: Discovery Intro
weight: 2
---

## Pre-Guide Notes

In this guide we will learn about the four main discovery resources:

1. Kubernetes Clusters
    - representation of a cluster that Service Mesh Hub is aware of and is authorized to
talk to its Kubernetes API server
    - *note*: not currently discovered, but created by `meshctl` at cluster registration time
2. Meshes
    - representation of a service mesh control plane that has been discovered 
3. Mesh Workloads
    - representation of a pod that is a member of a service mesh; this is often determined by the presence of
an injected proxy sidecar
4. Mesh Services
    - representation of a kubernetes service that is backed by Mesh Workload pods, e.g.
pods that are a member of the service mesh

To illustrate these concepts, we will assume that:

* we have just installed Istio as described in the "Installing Istio" guide
* Service Mesh Hub is running
* the cluster running Istio has been registered as described in [top-level intro]({{% versioned_link_path fromRoot="/getting_started/guides" %}})
* after Istio was installed, you deployed the `bookinfo` app into Istio's cluster as described in the [top-level intro]({{% versioned_link_path fromRoot="/getting_started/guides" %}})

## Guide

Ensure that your kubeconfig has the correct context set as its `currentContext`:

```shell
kubectl config use-context management-plane-context
```

### Kubernetes Clusters

Check to see that we wrote config when we registered our remote cluster:

```shell
kubectl -n service-mesh-hub get kubernetesclusters
```

```shell
NAME                 AGE
new-remote-cluster   6m53s
```

### Meshes

Check to see that Istio has been discovered:

```shell
kubectl -n service-mesh-hub get meshes
```

```
NAME                                    AGE
istio-istio-system-new-remote-cluster   99s
```

We can print it in YAML form to see all the information we discovered:

```shell
kubectl -n service-mesh-hub get mesh istio-istio-system-new-remote-cluster -oyaml
```

(snipped for brevity)

```
apiVersion: discovery.zephyr.solo.io/v1alpha1
kind: Mesh
metadata:
  ...
  labels:
    discovered_by: istio-mesh-discovery
  name: istio-istio-system-new-remote-cluster
  namespace: service-mesh-hub
  ...
spec:
  cluster:
    name: management-plane-cluster
    namespace: service-mesh-hub
  istio:
    citadelInfo:
      citadelNamespace: istio-system
      citadelServiceAccount: istiod-service-account
      trustDomain: cluster.local
    installation:
      installationNamespace: istio-system
      version: 1.5.1

```

### Mesh Workloads

Check to see that the `bookinfo` pods have been correctly identified as Mesh Workloads:

```shell
kubectl -n service-mesh-hub get meshworkloads
```

```
NAME                                              AGE
istio-details-v1-default-new-remote-cluster       15h
istio-productpage-v1-default-new-remote-cluster   15h
istio-ratings-v1-default-new-remote-cluster       15h
istio-reviews-v1-default-new-remote-cluster       15h
istio-reviews-v2-default-new-remote-cluster       15h
```

### Mesh Services

Similarly for the `bookinfo` services:

```shell
kubectl -n service-mesh-hub get meshservices
```

```
NAME                                   AGE
details-default-new-remote-cluster       15h
productpage-default-new-remote-cluster   15h
ratings-default-new-remote-cluster       15h
reviews-default-new-remote-cluster       15h
```
