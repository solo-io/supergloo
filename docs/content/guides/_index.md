---
title: "Guides Overview"
menuTitle: Guides
description: Guides for getting started using Service Mesh Hub
weight: 40
---

There are several guides available. The complete list can be seen below. A reasonable
introduction to Service Mesh Hub can be obtained by following the guides in the order
that they appear here, starting from the top and going down.

* [Installing Multi-cluster Istio]({{% versioned_link_path fromRoot="/guides/installing_istio" %}})
* [Intro to Discovery]({{% versioned_link_path fromRoot="/guides/discovery_intro" %}})
* [Intro to Access Control]({{% versioned_link_path fromRoot="/guides/access_control_intro" %}})
* [Multicluster Communication]({{% versioned_link_path fromRoot="/guides/multicluster_communication" %}})

## Pre-requisites

There are three pre-requisites to following these guides:

1. Install `kubectl`
    - https://kubernetes.io/docs/tasks/tools/install-kubectl/
2. Install `meshctl`
    - https://github.com/solo-io/service-mesh-hub/releases
3. Have multiple Kubernetes clusters ready to use, accessible in different `kubeconfig` contexts. If you don't have access to multiple Kubernetes clusters, see the [Getting Started Guide]({{% versioned_link_path fromRoot="/getting_started/" %}}) to use Kubernetes in Docker (Kind) to spin up two clusters in containers.


## Assumptions made in the guides

We will assume in this guide that we have access to two clusters and the following two contexts available in our `kubeconfig` file. 

Your actual context names will likely be different.

* `management-plane-context`
    - kubeconfig context pointing to a cluster where we will install and operate Service Mesh Hub
* `remote-cluster-context`
    - kubeconfig context pointing to a cluster where we will install and manage a service mesh using Service Mesh Hub 

We assume you've [installed Service Mesh Hub]({{% versioned_link_path fromRoot="/setup/#installing-with-meshctl" %}}) into the cluster represented by `management-plane-context` the context.


#### Two registered clusters
We also assume you've [registered]({{% versioned_link_path fromRoot="/setup/#register-a-cluster" %}}) both of those clusters with Service Mesh Hub:


```shell
meshctl cluster register \
  --remote-cluster-name management-plane \
  --remote-context management-plane-context
```

```shell
meshctl cluster register \
  --remote-cluster-name new-remote-cluster \
  --remote-context remote-cluster-context
```

At this point we have two clusters, `management-plane-context` and `remote-cluster-context` both registered with Service Mesh Hub which happens to be installed on the `management-plane-context` cluster.

#### Bookinfo deployed on two clusters

We deploy parts of the [Bookinfo]() demo to two clusters. The core components, including reviews-v1 and reviews-v2, are deployed to `management-plane-cluster`, while `reviews-v3` is deployed on the `remote-cluster-context` cluster.

Deploy part of the bookinfo application to the `management-plane-context` cluster:

{{% notice note %}}
Be sure to switch the `kubeconfig` context to the `management-plane-context`
{{% /notice %}}

```shell
kubectl config use-context management-plane-context # your management-plane context name may be different
kubectl label namespace default istio-injection=enabled
​
# we deploy everything except reviews-v3 to the management-plane cluster
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'app,version notin (v3)'
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'account'
```

Now deploy only reviews-v3 to your `remote-cluster-context` cluster:

{{% notice note %}}
Be sure to switch the `kubeconfig` context to the `remote-cluster-context`
{{% /notice %}}

```shell
kubectl config use-context remote-cluster-context # your remote cluster context name may be different

kubectl label namespace default istio-injection=enabled
​
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'app,version in (v3)' 
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'service=reviews' 
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'account=reviews' 
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'app=ratings' 
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'account=ratings' 
```