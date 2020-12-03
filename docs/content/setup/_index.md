---
title: "Setup"
menuTitle: Setup
description: Setting up Gloo Mesh
weight: 30
---

Gloo Mesh is deployed on a Kubernetes cluster and used to manage service meshes running on one or more Kubernetes clusters. The following documents detail how to use Kubernetes in Docker (Kind) to test Gloo Mesh, how to install Gloo Mesh or Gloo Mesh Enterprise, and how to register clusters containing a service mesh deployment with Gloo Mesh.

We use a Kubernetes cluster to host the management plane (Gloo Mesh) while each service mesh can run on its own independent cluster. If you don't have access to multiple clusters, see the [Getting Started Guide]({{% versioned_link_path fromRoot="/getting_started/" %}}) to get started with Kubernetes in Docker. 

![Gloo Mesh Architecture]({{% versioned_link_path fromRoot="/img/gloomesh-3clusters.png" %}})

You can install Gloo Mesh onto its own cluster and register remote clusters, or you can co-locate Gloo Mesh onto a cluster with a service mesh. The former (its own cluster) is the preferred deployment pattern, but for getting started, exploring, or to save resources, you can use the co-located deployment approach.

![Gloo Mesh Architecture]({{% versioned_link_path fromRoot="/img/gloomesh-2clusters.png" %}})

The following guides provide detailed instructions for getting Gloo Mesh deployed:

{{% children description="true" %}}