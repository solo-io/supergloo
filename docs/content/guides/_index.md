---
title: "Guides"
menuTitle: Guides
description: Guides for getting started using Service Mesh Hub
weight: 40
---

# Service Mesh Hub Guides

Over the course of these guides, we will explore the core
features of Service Mesh Hub, including:

* Discovery
* Routing
* Access Control
* Multicluster Federation

We will do that by managing and configuring Istio's `bookinfo` demo, which you can start up by running

```shell
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml
```

Through that application, we will see how we can use Service Mesh Hub to gain insights into and easily
control a complex multicluster application.

## Guides Overview

There are three pre-requisites to following these guides:

1. Install `kubectl`
    - https://kubernetes.io/docs/tasks/tools/install-kubectl/
2. Install `meshctl`
    - https://github.com/solo-io/service-mesh-hub/releases
3. Have multiple Kubernetes clusters ready to use, accessible in different kubeconfig contexts 

There are several guides available. The complete list can be seen below. A reasonable
introduction to Service Mesh Hub can be obtained by following the guides in the order
that they appear here, starting from the top and going down.

* [Installing Istio]({{% versioned_link_path fromRoot="/getting_started/guides/installing_istio" %}})
* [Intro to Discovery]({{% versioned_link_path fromRoot="/getting_started/guides/discovery_intro" %}})
* [Intro to Access Control]({{% versioned_link_path fromRoot="/getting_started/guides/access_control_intro" %}})
* [Multicluster Communication]({{% versioned_link_path fromRoot="/getting_started/guides/multicluster_communication" %}})
