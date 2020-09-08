---
title: "Guides Overview"
menuTitle: Guides
description: Guides for getting started using Service Mesh Hub
weight: 40
---

There are several guides available. To follow the guides included in this section, you will need at least two Kubernetes clusters with Service Mesh Hub installed on one. You can easily set up two such clusters [using Kind](#using-kind) as detailed below. Some of the guides also make use of the Bookinfo sample application. You can install the application by following the steps in the [Bookinfo deployment section](#bookinfo-deployment).

To become familiar with Service Mesh Hub, we recommend following the guides in the order that they appear here, starting from the top and going down.

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
3. Have multiple Kubernetes clusters ready to use, accessible in different `kubeconfig` contexts. If you don't have access to multiple Kubernetes clusters, see the section below, [Using Kind](#using-kind) to use Kubernetes in Docker (Kind) to spin up two clusters in containers.


## Assumptions made in the guides

We will assume in this guide that you have access to two clusters and the following two contexts are available in your `kubeconfig` file. 

Your actual context names will likely be different.

* `mgmt-cluster-context`
    - kubeconfig context pointing to a cluster where we will install and operate Service Mesh Hub
* `remote-cluster-context`
    - kubeconfig context pointing to a cluster where we will install and manage a service mesh using Service Mesh Hub 

We assume you've [installed Service Mesh Hub]({{% versioned_link_path fromRoot="/setup/#installing-with-meshctl" %}}) into the cluster represented by the context `mgmt-cluster-context`.


#### Two registered clusters
We also assume you've [registered]({{% versioned_link_path fromRoot="/setup/#register-a-cluster" %}}) both of those clusters with Service Mesh Hub:


```shell
meshctl cluster register \
  --remote-cluster-name mgmt-cluster \
  --remote-context mgmt-cluster-context
```

```shell
meshctl cluster register \
  --remote-cluster-name remote-cluster \
  --remote-context remote-cluster-context
```

At this point we have two clusters, `mgmt-cluster` and `remote-cluster` both registered with Service Mesh Hub which happens to be installed on the `mgmt-cluster` cluster.

## Using Kind

Kubernetes in Docker makes it easy to stand up Kubernetes clusters on your local desktop for testing and development. To assist with learning about Service Mesh Hub, you can use Kind to deploy two Kubernetes clusters. You can install Kind by following the [installation steps in their Quick Start guide](https://kind.sigs.k8s.io/docs/user/quick-start/).

Once you have Kind installed, you can create the two clusters in question by running the following:

```bash
meshctl demo init
```

The command will do the following:

* Create two kind clusters: mgmt-cluster and remote-cluster
* Install Service Mesh Hub on the management plane cluster
* Install Istio on both clusters
* Register both clusters with Service Mesh Hub

You should now be ready to start going through the guides.

#### Bookinfo deployment

For some parts of the guide, you'll want to have the [Bookinfo](https://istio.io/docs/examples/bookinfo/) demo deployed to two clusters. 

{{% notice note %}}
You'll want to first have [Istio installed for multi-cluster]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}) **before** installing the Bookinfo demo. 
{{% /notice %}}

The core components, including reviews-v1 and reviews-v2, are deployed to the management plane cluster, while `reviews-v3` is deployed on the remote cluster.

Deploy part of the bookinfo application to the `mgmt-cluster` cluster:

{{% notice note %}}
Be sure to switch the `kubeconfig` context to the `mgmt-cluster`
{{% /notice %}}

```shell
# Set the proper context value for each cluster
MGMT_CONTEXT=your_management_plane_context
REMOTE_CONTEXT=your_remote_context

kubectl config use-context $MGMT_CONTEXT

kubectl create ns bookinfo
kubectl label namespace bookinfo istio-injection=enabled
​
# we deploy everything except reviews-v3 to the management plane cluster
kubectl apply -n bookinfo -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'app,version notin (v3)'
kubectl apply -n bookinfo -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'account'
```

Now deploy only reviews-v3 to your `remote-cluster-context` cluster:

```shell
kubectl config use-context $REMOTE_CONTEXT

kubectl create ns bookinfo
kubectl label namespace bookinfo istio-injection=enabled
​
kubectl apply -n bookinfo -f  https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'app,version in (v3)' 
kubectl apply -n bookinfo -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'service=reviews' 
kubectl apply -n bookinfo -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'account=reviews' 
kubectl apply -n bookinfo -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'app=ratings' 
kubectl apply -n bookinfo -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'account=ratings' 
```

Now you have Bookinfo demo set up for the rest of the guides. From here you can [federate]({{% versioned_link_path fromRoot="/guides/federate_identity" %}}) your two clusters, or start configuring [multi-cluster access]({{% versioned_link_path fromRoot="/guides/access_control_intro" %}}) and [traffic policy]({{% versioned_link_path fromRoot="/guides/multicluster_communication" %}}).
