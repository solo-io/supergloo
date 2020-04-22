---
title: "Guides Overview"
menuTitle: Guides
description: Guides for getting started using Service Mesh Hub
weight: 40
---

We will assume in this guide that we have access to two clusters and the following two contexts available in our `kubeconfig` file. 

Your actual context names will likely be different.

* `management-plane-context`
    - kubeconfig context pointing to a cluster where we will install and operate Service Mesh Hub
* `remote-cluster-context`
    - kubeconfig context pointing to a cluster where we will install and manage a service mesh using Service Mesh Hub 

### Pre-requisites

There are three pre-requisites to following these guides:

1. Install `kubectl`
    - https://kubernetes.io/docs/tasks/tools/install-kubectl/
2. Install `meshctl`
    - https://github.com/solo-io/service-mesh-hub/releases
3. Have multiple Kubernetes clusters ready to use, accessible in different kubeconfig contexts. If you don't have access to multiple Kubernetes clusters, see the [Getting Started Guide]({{% versioned_link_path fromRoot="/getting_started/" %}}) to use Kubernetes in Docker (Kind) to spin up two clusters in containers.

There are several guides available. The complete list can be seen below. A reasonable
introduction to Service Mesh Hub can be obtained by following the guides in the order
that they appear here, starting from the top and going down.

* [Installing Multi-cluster Istio]({{% versioned_link_path fromRoot="/guides/installing_istio" %}})
* [Intro to Discovery]({{% versioned_link_path fromRoot="/guides/discovery_intro" %}})
* [Intro to Access Control]({{% versioned_link_path fromRoot="/guides/access_control_intro" %}})
* [Multicluster Communication]({{% versioned_link_path fromRoot="/guides/multicluster_communication" %}})
