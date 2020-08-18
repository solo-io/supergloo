
---
title: "kubernetes_cluster.proto"
---

## Package : `discovery.smh.solo.io`



<a name="top"></a>

<a name="API Reference for kubernetes_cluster.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kubernetes_cluster.proto


## Table of Contents
  - [KubernetesClusterSpec](#discovery.smh.solo.io.KubernetesClusterSpec)







<a name="discovery.smh.solo.io.KubernetesClusterSpec"></a>

### KubernetesClusterSpec
Representation of a Kubernetes cluster that has been registered in Service Mesh Hub.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| secretRef | [core.smh.solo.io.ResourceRef](#core.smh.solo.io.ResourceRef) |  | pointer to secret which contains the kubeconfig with information to connect to the remote cluster. |
| context | [string](#string) |  | context to use within the kubeconfig pointed to by the above reference |
| version | [string](#string) |  | version of kubernetes |
| cloud | [string](#string) |  | cloud provider, empty if unknown |
| writeNamespace | [string](#string) |  | namespace to use when writing Service Mesh Hub resources to this cluster |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

