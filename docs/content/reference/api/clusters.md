
---
title: "clusters.proto"
---

## Package : `rpc.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for clusters.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## clusters.proto


## Table of Contents
  - [KubernetesCluster](#rpc.zephyr.solo.io.KubernetesCluster)
  - [KubernetesCluster.LabelsEntry](#rpc.zephyr.solo.io.KubernetesCluster.LabelsEntry)
  - [ListKubernetesClustersRequest](#rpc.zephyr.solo.io.ListKubernetesClustersRequest)
  - [ListKubernetesClustersResponse](#rpc.zephyr.solo.io.ListKubernetesClustersResponse)



  - [KubernetesClusterApi](#rpc.zephyr.solo.io.KubernetesClusterApi)




<a name="rpc.zephyr.solo.io.KubernetesCluster"></a>

### KubernetesCluster



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| spec | [discovery.zephyr.solo.io.KubernetesClusterSpec](#discovery.zephyr.solo.io.KubernetesClusterSpec) |  |  |
| ref | [core.zephyr.solo.io.ResourceRef](#core.zephyr.solo.io.ResourceRef) |  |  |
| labels | [][KubernetesCluster.LabelsEntry](#rpc.zephyr.solo.io.KubernetesCluster.LabelsEntry) | repeated |  |






<a name="rpc.zephyr.solo.io.KubernetesCluster.LabelsEntry"></a>

### KubernetesCluster.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="rpc.zephyr.solo.io.ListKubernetesClustersRequest"></a>

### ListKubernetesClustersRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| clusters | [][KubernetesCluster](#rpc.zephyr.solo.io.KubernetesCluster) | repeated |  |






<a name="rpc.zephyr.solo.io.ListKubernetesClustersResponse"></a>

### ListKubernetesClustersResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="rpc.zephyr.solo.io.KubernetesClusterApi"></a>

### KubernetesClusterApi


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ListClusters | [ListKubernetesClustersRequest](#rpc.zephyr.solo.io.ListKubernetesClustersRequest) | [ListKubernetesClustersRequest](#rpc.zephyr.solo.io.ListKubernetesClustersRequest) |  |

 <!-- end services -->

