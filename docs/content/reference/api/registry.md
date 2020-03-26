
---
title: "discovery.zephyr.solo.iogithub.com/solo-io/mesh-projects/api/discovery/v1alpha1/registry.proto"
---

## Package : `discovery.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for github.com/solo-io/mesh-projects/api/discovery/v1alpha1/registry.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## github.com/solo-io/mesh-projects/api/discovery/v1alpha1/registry.proto


## Table of Contents
  - [Federation](#discovery.zephyr.solo.io.Federation)
  - [KubePod](#discovery.zephyr.solo.io.KubePod)
  - [KubePod.LabelsEntry](#discovery.zephyr.solo.io.KubePod.LabelsEntry)
  - [KubeService](#discovery.zephyr.solo.io.KubeService)
  - [KubeService.LabelsEntry](#discovery.zephyr.solo.io.KubeService.LabelsEntry)
  - [KubeService.WorkloadSelectorLabelsEntry](#discovery.zephyr.solo.io.KubeService.WorkloadSelectorLabelsEntry)
  - [KubeServicePort](#discovery.zephyr.solo.io.KubeServicePort)
  - [MeshServiceSpec](#discovery.zephyr.solo.io.MeshServiceSpec)
  - [MeshServiceSpec.Subset](#discovery.zephyr.solo.io.MeshServiceSpec.Subset)
  - [MeshServiceSpec.SubsetsEntry](#discovery.zephyr.solo.io.MeshServiceSpec.SubsetsEntry)
  - [MeshServiceStatus](#discovery.zephyr.solo.io.MeshServiceStatus)
  - [MeshWorkloadSpec](#discovery.zephyr.solo.io.MeshWorkloadSpec)
  - [MeshWorkloadStatus](#discovery.zephyr.solo.io.MeshWorkloadStatus)







<a name="discovery.zephyr.solo.io.Federation"></a>

### Federation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| multiclusterDnsName | [string](#string) |  |  |
| federatedToWorkloads | [][core.zephyr.solo.io.ResourceRef](#core.zephyr.solo.io.ResourceRef) | repeated |  |






<a name="discovery.zephyr.solo.io.KubePod"></a>

### KubePod



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| labels | [][KubePod.LabelsEntry](#discovery.zephyr.solo.io.KubePod.LabelsEntry) | repeated | these are the labels directly from the pods that this controller owns NB: these are NEITHER the matchLabels nor the labels on the controller itself. we need these to determine which services are backed by this workload, and the service backing is determined by the pod labels |
| serviceAccountName | [string](#string) |  |  |






<a name="discovery.zephyr.solo.io.KubePod.LabelsEntry"></a>

### KubePod.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="discovery.zephyr.solo.io.KubeService"></a>

### KubeService



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.zephyr.solo.io.ResourceRef](#core.zephyr.solo.io.ResourceRef) |  |  |
| workloadSelectorLabels | [][KubeService.WorkloadSelectorLabelsEntry](#discovery.zephyr.solo.io.KubeService.WorkloadSelectorLabelsEntry) | repeated | selectors for the set of pods targeted by the k8s Service |
| labels | [][KubeService.LabelsEntry](#discovery.zephyr.solo.io.KubeService.LabelsEntry) | repeated | labels on the underyling k8s Service itself |
| ports | [][KubeServicePort](#discovery.zephyr.solo.io.KubeServicePort) | repeated |  |






<a name="discovery.zephyr.solo.io.KubeService.LabelsEntry"></a>

### KubeService.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="discovery.zephyr.solo.io.KubeService.WorkloadSelectorLabelsEntry"></a>

### KubeService.WorkloadSelectorLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="discovery.zephyr.solo.io.KubeServicePort"></a>

### KubeServicePort



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | [uint32](#uint32) |  | external-facing port for this service (i.e., NOT the service's target port on the backing pods) |
| name | [string](#string) |  |  |
| protocol | [string](#string) |  |  |






<a name="discovery.zephyr.solo.io.MeshServiceSpec"></a>

### MeshServiceSpec
The MeshService is an abstraction for a service which we have discovered to be, or are told, is part of a given mesh. The Mesh object has references to the MeshServices which belong to it.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeService | [KubeService](#discovery.zephyr.solo.io.KubeService) |  | Resource ref to the underlying resource which this service is representing. It can potentially represent multiple different resource types, but currently supports only kubernetes services<br>The type is specified on the ResourceRef.APIGroup and ResourceRef.Kind fields |
| mesh | [core.zephyr.solo.io.ResourceRef](#core.zephyr.solo.io.ResourceRef) |  | The mesh with which this service is associated |
| subsets | [][MeshServiceSpec.SubsetsEntry](#discovery.zephyr.solo.io.MeshServiceSpec.SubsetsEntry) | repeated |  |
| federation | [Federation](#discovery.zephyr.solo.io.Federation) |  |  |






<a name="discovery.zephyr.solo.io.MeshServiceSpec.Subset"></a>

### MeshServiceSpec.Subset



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| values | [][string](#string) | repeated |  |






<a name="discovery.zephyr.solo.io.MeshServiceSpec.SubsetsEntry"></a>

### MeshServiceSpec.SubsetsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [MeshServiceSpec.Subset](#discovery.zephyr.solo.io.MeshServiceSpec.Subset) |  |  |






<a name="discovery.zephyr.solo.io.MeshServiceStatus"></a>

### MeshServiceStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| federationStatus | [core.zephyr.solo.io.ComputedStatus](#core.zephyr.solo.io.ComputedStatus) |  |  |






<a name="discovery.zephyr.solo.io.MeshWorkloadSpec"></a>

### MeshWorkloadSpec
The MeshWorkload is an abstraction for a workload/client which we have discovered to be, or are told, is part of a given mesh. The Mesh object has references to the MeshWorkloads which belong to it.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeControllerRef | [core.zephyr.solo.io.ResourceRef](#core.zephyr.solo.io.ResourceRef) |  | Resource ref to the underlying kubernetes controller which is managing the pods associated with the workloads. It has the generic name kube_controller as it can represent either a deployment or a daemonset. Or potentially any other kubernetes object which creates injected pods.<br>The type is specified on the ResourceRef.APIGroup and ResourceRef.Kind fields |
| kubePod | [KubePod](#discovery.zephyr.solo.io.KubePod) |  |  |
| mesh | [core.zephyr.solo.io.ResourceRef](#core.zephyr.solo.io.ResourceRef) |  | The mesh with which this workload is associated |






<a name="discovery.zephyr.solo.io.MeshWorkloadStatus"></a>

### MeshWorkloadStatus






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

