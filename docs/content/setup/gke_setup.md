---
title: "Gloo Mesh Setup in Production"
menuTitle: Using GKE
description: Deploying Gloo Mesh in Production
weight: 2
---

This document shows some of the Production configuration options that may be useful.
We will continue to add to this document and welcome users of Gloo Mesh Enterprise
to send PRs to this as well.


## Install Istio

[Deploy Istio]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}) or
[Gloo Mesh Istio]({{% versioned_link_path fromRoot="/setup/gloo_mesh_istio/" %}})
on both clusters.

## Install Gloo Mesh

##### OSS (community version)

1) Follow the [Gloo Mesh installation instructions]({{% versioned_link_path fromRoot="/setup/installation/community_installation" %}}).

2) [Register your clusters]({{% versioned_link_path fromRoot="/setup/cluster_registration/community_cluster_registration" %}}).

##### Enterprise

1) Go through the [Gloo Mesh Enterprise Prerequisites]({{% versioned_link_path fromRoot="/setup/prerequisites/enterprise_prerequisites" %}}).

2) [Register your enterprise clusters]({{% versioned_link_path fromRoot="/setup/cluster_registration/enterprise_cluster_registration" %}}).

## Follow our Guides

3) Install the [bookinfo example deployment]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployment" %}})

4) Check out some of our other [guides]({{% versioned_link_path fromRoot="/guides" %}})!
