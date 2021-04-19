---
title: "Getting Started"
menuTitle: Getting Started
description: How to get started using Gloo Mesh
weight: 10
---

Welcome to Gloo Mesh, the open-source, multi-cluster, multi-mesh management plane. Gloo Mesh simplifies service-mesh operations and lets you manage multiple clusters of a service mesh from a centralized management plane. Gloo Mesh takes care of things like shared-trust/root CA federation, workload discovery, unified multi-cluster/global traffic policy, access policy, and more. 

## Getting `meshctl`

Gloo Mesh has a CLI tool called `meshctl` that helps bootstrap Gloo Mesh, register clusters, describe configured resources, and more. Get the latest `meshctl` from the [releases page on solo-io/gloo-mesh](https://github.com/solo-io/gloo-mesh/releases).

You can also quickly install like this:

```shell
curl -sL https://run.solo.io/meshctl/install | sh
```

Once you've downloaded the correct binary for your architecture, run the following to make sure it's working correctly:

```shell
meshctl version
```

You can add `meshctl` to your PATH ([Windows](https://helpdeskgeek.com/windows-10/add-windows-path-environment-variable/) [Mac](https://osxdaily.com/2014/08/14/add-new-path-to-path-command-line/) [Linux](https://linuxize.com/post/how-to-add-directory-to-path-in-linux/)) for global access on the command line.

## Deploying Gloo Mesh

In this section, we detail a few ways to get you up and running with Gloo Mesh and Gloo Mesh Enterprise. For detailed
information on each aspect of Gloo Mesh installation, check out the [setup guides]({{% versioned_link_path fromRoot="/setup/" %}}).

{{% children description="true" %}}
