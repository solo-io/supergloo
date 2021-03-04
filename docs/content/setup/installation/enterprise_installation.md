---
title: "Enterprise"
menuTitle: Enterprise
description: Install Gloo Mesh Enterprise
weight: 100
---

{{% notice note %}} Gloo Mesh Enterprise is required for this feature. {{% /notice %}}

This document describes how to install Gloo Mesh Enterprise.
A conceptual overview of the Gloo Mesh Enterprise architecture can be found [here]({{% versioned_link_path fromRoot="/concepts/relay" %}}).

## Helm

The source for the Gloo Mesh Enterprise Helm chart is available on [GitHub](https://github.com/solo-io/gloo-mesh-enterprise-helm).

1. Add the Helm repo

```shell
helm repo add gloo-mesh-enterprise https://storage.googleapis.com/gloo-mesh-enterprise/gloo-mesh-enterprise
```

2. (optional) View available versions

```shell
helm search repo enterprise-networking
```

3. (optional) View Helm values

```shell
helm show values enterprise-networking/enterprise-networking
```

4. Install

{{% notice note %}} If you are running Gloo Mesh Enterprise's management plane on a cluster you intend to register (i.e. also run a service mesh), set the `enterprise-networking.cluster` value to the cluster name you intend to set for the management cluster at registration time. {{% /notice %}}

```shell
helm install enterprise-networking enterprise-networking/enterprise-networking --namespace gloo-mesh \
  --set licenseKey=${GLOO_MESH_ENTERRISE_LICENSE_KEY}
```

## meshctl

Instructions for installing Gloo Mesh Enterprise via meshctl are coming soon.