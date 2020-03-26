
---
title: "discovery.zephyr.solo.iogithub.com/solo-io/mesh-projects/api/discovery/v1alpha1/cluster.proto"
---

## Package : `discovery.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for github.com/solo-io/mesh-projects/api/discovery/v1alpha1/cluster.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## github.com/solo-io/mesh-projects/api/discovery/v1alpha1/cluster.proto


## Table of Contents
  - [KubernetesClusterSpec](#discovery.zephyr.solo.io.KubernetesClusterSpec)







<a name="discovery.zephyr.solo.io.KubernetesClusterSpec"></a>

### KubernetesClusterSpec
specifies a method by which to select pods within a mesh for the application of rules and policies


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| secretRef | [core.zephyr.solo.io.ResourceRef](#core.zephyr.solo.io.ResourceRef) |  | pointer to secret which contains the kubeconfig with information to connect to the remote cluster. |
| context | [string](#string) |  | context to use within the kubeconfig pointed to by the above reference |
| version | [string](#string) |  | version of kubernetes |
| cloud | [string](#string) |  | cloud provider, empty if unknown |
| writeNamespace | [string](#string) |  | namespace to use when writing Service Mesh Hub resources to this cluster |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

