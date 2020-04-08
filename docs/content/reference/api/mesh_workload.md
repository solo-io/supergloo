
---
title: "mesh_workload.proto"
---

## Package : `discovery.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for mesh_workload.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## mesh_workload.proto


## Table of Contents
  - [MeshWorkloadSpec](#discovery.zephyr.solo.io.MeshWorkloadSpec)
  - [MeshWorkloadSpec.KubeController](#discovery.zephyr.solo.io.MeshWorkloadSpec.KubeController)
  - [MeshWorkloadSpec.KubeController.LabelsEntry](#discovery.zephyr.solo.io.MeshWorkloadSpec.KubeController.LabelsEntry)
  - [MeshWorkloadStatus](#discovery.zephyr.solo.io.MeshWorkloadStatus)







<a name="discovery.zephyr.solo.io.MeshWorkloadSpec"></a>

### MeshWorkloadSpec
The MeshWorkload is an abstraction for a workload/client which we have discovered to be part of a given mesh. The Mesh object has references to the MeshWorkloads which belong to it.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeController | [MeshWorkloadSpec.KubeController](#discovery.zephyr.solo.io.MeshWorkloadSpec.KubeController) |  | The controller (e.g. deployment) that owns this workload |
| mesh | [core.zephyr.solo.io.ResourceRef](#core.zephyr.solo.io.ResourceRef) |  | The mesh with which this workload is associated |






<a name="discovery.zephyr.solo.io.MeshWorkloadSpec.KubeController"></a>

### MeshWorkloadSpec.KubeController



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeControllerRef | [core.zephyr.solo.io.ResourceRef](#core.zephyr.solo.io.ResourceRef) |  | Resource ref to the underlying kubernetes controller which is managing the pods associated with the workloads. It has the generic name kube_controller as it can represent either a deployment or a daemonset. Or potentially any other kubernetes object which creates injected pods. |
| labels | [][MeshWorkloadSpec.KubeController.LabelsEntry](#discovery.zephyr.solo.io.MeshWorkloadSpec.KubeController.LabelsEntry) | repeated | these are the labels directly from the pods that this controller owns NB: these are NEITHER the matchLabels nor the labels on the controller itself. we need these to determine which services are backed by this workload, and the service backing is determined by the pod labels. |
| serviceAccountName | [string](#string) |  | Service account attached to the pods owned by this controller |






<a name="discovery.zephyr.solo.io.MeshWorkloadSpec.KubeController.LabelsEntry"></a>

### MeshWorkloadSpec.KubeController.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="discovery.zephyr.solo.io.MeshWorkloadStatus"></a>

### MeshWorkloadStatus






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

