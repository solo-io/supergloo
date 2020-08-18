
---
title: "mesh_workload.proto"
---

## Package : `discovery.smh.solo.io`



<a name="top"></a>

<a name="API Reference for mesh_workload.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## mesh_workload.proto


## Table of Contents
  - [MeshWorkloadSpec](#discovery.smh.solo.io.MeshWorkloadSpec)
  - [MeshWorkloadSpec.AppMesh](#discovery.smh.solo.io.MeshWorkloadSpec.AppMesh)
  - [MeshWorkloadSpec.AppMesh.ContainerPort](#discovery.smh.solo.io.MeshWorkloadSpec.AppMesh.ContainerPort)
  - [MeshWorkloadSpec.KubernertesWorkload](#discovery.smh.solo.io.MeshWorkloadSpec.KubernertesWorkload)
  - [MeshWorkloadSpec.KubernertesWorkload.PodLabelsEntry](#discovery.smh.solo.io.MeshWorkloadSpec.KubernertesWorkload.PodLabelsEntry)
  - [MeshWorkloadStatus](#discovery.smh.solo.io.MeshWorkloadStatus)







<a name="discovery.smh.solo.io.MeshWorkloadSpec"></a>

### MeshWorkloadSpec
The MeshWorkload is an abstraction for a workload/client which mesh-discovery has discovered to be part of a given mesh (i.e. its traffic is managed by an in-mesh sidecar). The Mesh object has references to the MeshWorkloads which belong to it.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubernetes | [MeshWorkloadSpec.KubernertesWorkload](#discovery.smh.solo.io.MeshWorkloadSpec.KubernertesWorkload) |  | information describing workloads backed by Kubernetes Pods. |
| mesh | [core.skv2.solo.io.ObjectRef](#core.skv2.solo.io.ObjectRef) |  | The mesh with which this workload is associated |
| appMesh | [MeshWorkloadSpec.AppMesh](#discovery.smh.solo.io.MeshWorkloadSpec.AppMesh) |  | Appmesh specific metadata |






<a name="discovery.smh.solo.io.MeshWorkloadSpec.AppMesh"></a>

### MeshWorkloadSpec.AppMesh
information relevant to AppMesh-injected workloads


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| virtualNodeName | [string](#string) |  | The value of the env var APPMESH_VIRTUAL_NODE_NAME on the Appmesh envoy proxy container |
| ports | [][MeshWorkloadSpec.AppMesh.ContainerPort](#discovery.smh.solo.io.MeshWorkloadSpec.AppMesh.ContainerPort) | repeated | Needed for declaring Appmesh VirtualNode listeners |






<a name="discovery.smh.solo.io.MeshWorkloadSpec.AppMesh.ContainerPort"></a>

### MeshWorkloadSpec.AppMesh.ContainerPort
k8s application container ports


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | [uint32](#uint32) |  |  |
| protocol | [string](#string) |  |  |






<a name="discovery.smh.solo.io.MeshWorkloadSpec.KubernertesWorkload"></a>

### MeshWorkloadSpec.KubernertesWorkload
information describing a Kubernetes-based workload (e.g. a Deployment or DaemonSet)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| controller | [core.skv2.solo.io.ClusterObjectRef](#core.skv2.solo.io.ClusterObjectRef) |  | Resource ref to the underlying kubernetes controller which is managing the pods associated with the workloads. It has the generic name controller as it can represent a deployment, daemonset, or statefulset. |
| podLabels | [][MeshWorkloadSpec.KubernertesWorkload.PodLabelsEntry](#discovery.smh.solo.io.MeshWorkloadSpec.KubernertesWorkload.PodLabelsEntry) | repeated | these are the labels directly from the pods that this controller owns NB: these labels are read directly from the pod template metadata.labels defined in the workload spec. We need these to determine which services are backed by this workload, and the service backing is determined by the pod labels. |
| serviceAccountName | [string](#string) |  | Service account attached to the pods owned by this controller |






<a name="discovery.smh.solo.io.MeshWorkloadSpec.KubernertesWorkload.PodLabelsEntry"></a>

### MeshWorkloadSpec.KubernertesWorkload.PodLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="discovery.smh.solo.io.MeshWorkloadStatus"></a>

### MeshWorkloadStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | [int64](#int64) |  | The observed generation of the MeshWorkload. When this matches the MeshWorkload's metadata.generation, it indicates that mesh-networking has reconciled the latest version of the MeshWorkload. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

