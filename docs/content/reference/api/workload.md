
---
title: "workload.proto"
---

## Package : `discovery.smh.solo.io`



<a name="top"></a>

<a name="API Reference for workload.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## workload.proto


## Table of Contents
  - [WorkloadSpec](#discovery.smh.solo.io.WorkloadSpec)
  - [WorkloadSpec.AppMesh](#discovery.smh.solo.io.WorkloadSpec.AppMesh)
  - [WorkloadSpec.AppMesh.ContainerPort](#discovery.smh.solo.io.WorkloadSpec.AppMesh.ContainerPort)
  - [WorkloadSpec.KubernetesWorkload](#discovery.smh.solo.io.WorkloadSpec.KubernetesWorkload)
  - [WorkloadSpec.KubernetesWorkload.PodLabelsEntry](#discovery.smh.solo.io.WorkloadSpec.KubernetesWorkload.PodLabelsEntry)
  - [WorkloadStatus](#discovery.smh.solo.io.WorkloadStatus)







<a name="discovery.smh.solo.io.WorkloadSpec"></a>

### WorkloadSpec
The Workload is an abstraction for a workload/client which mesh-discovery has discovered to be part of a given mesh (i.e. its traffic is managed by an in-mesh sidecar).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubernetes | [WorkloadSpec.KubernetesWorkload](#discovery.smh.solo.io.WorkloadSpec.KubernetesWorkload) |  | Information describing workloads backed by Kubernetes Pods. |
| mesh | [core.skv2.solo.io.ObjectRef](#core.skv2.solo.io.ObjectRef) |  | The mesh with which this workload is associated. |
| appMesh | [WorkloadSpec.AppMesh](#discovery.smh.solo.io.WorkloadSpec.AppMesh) |  | Appmesh specific metadata. |






<a name="discovery.smh.solo.io.WorkloadSpec.AppMesh"></a>

### WorkloadSpec.AppMesh
Information relevant to AppMesh-injected workloads.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| virtualNodeName | [string](#string) |  | The value of the env var APPMESH_VIRTUAL_NODE_NAME on the Appmesh envoy proxy container. |
| ports | [][WorkloadSpec.AppMesh.ContainerPort](#discovery.smh.solo.io.WorkloadSpec.AppMesh.ContainerPort) | repeated | Needed for declaring Appmesh VirtualNode listeners. |






<a name="discovery.smh.solo.io.WorkloadSpec.AppMesh.ContainerPort"></a>

### WorkloadSpec.AppMesh.ContainerPort
k8s application container ports.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | [uint32](#uint32) |  |  |
| protocol | [string](#string) |  |  |






<a name="discovery.smh.solo.io.WorkloadSpec.KubernetesWorkload"></a>

### WorkloadSpec.KubernetesWorkload
Information describing a Kubernetes-based workload (e.g. a Deployment or DaemonSet).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| controller | [core.skv2.solo.io.ClusterObjectRef](#core.skv2.solo.io.ClusterObjectRef) |  | Resource ref to the underlying kubernetes controller which is managing the pods associated with the workloads. It has the generic name controller as it can represent a deployment, daemonset, or statefulset. |
| podLabels | [][WorkloadSpec.KubernetesWorkload.PodLabelsEntry](#discovery.smh.solo.io.WorkloadSpec.KubernetesWorkload.PodLabelsEntry) | repeated | These are the labels directly from the pods that this controller owns. NB: these labels are read directly from the pod template metadata.labels defined in the workload spec. We need these to determine which services are backed by this workload. |
| serviceAccountName | [string](#string) |  | Service account attached to the pods owned by this controller. |






<a name="discovery.smh.solo.io.WorkloadSpec.KubernetesWorkload.PodLabelsEntry"></a>

### WorkloadSpec.KubernetesWorkload.PodLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="discovery.smh.solo.io.WorkloadStatus"></a>

### WorkloadStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | [int64](#int64) |  | The observed generation of the Workload. When this matches the Workload's metadata.generation it indicates that mesh-networking has reconciled the latest version of the Workload. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

