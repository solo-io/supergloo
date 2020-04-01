
---
title: "discovery.zephyr.solo.iogithub.com/solo-io/mesh-projects/api/discovery/v1alpha1/mesh_service.proto"
---

## Package : `discovery.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for github.com/solo-io/mesh-projects/api/discovery/v1alpha1/mesh_service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## github.com/solo-io/mesh-projects/api/discovery/v1alpha1/mesh_service.proto


## Table of Contents
  - [MeshServiceSpec](#discovery.zephyr.solo.io.MeshServiceSpec)
  - [MeshServiceSpec.Federation](#discovery.zephyr.solo.io.MeshServiceSpec.Federation)
  - [MeshServiceSpec.KubeService](#discovery.zephyr.solo.io.MeshServiceSpec.KubeService)
  - [MeshServiceSpec.KubeService.KubeServicePort](#discovery.zephyr.solo.io.MeshServiceSpec.KubeService.KubeServicePort)
  - [MeshServiceSpec.KubeService.LabelsEntry](#discovery.zephyr.solo.io.MeshServiceSpec.KubeService.LabelsEntry)
  - [MeshServiceSpec.KubeService.WorkloadSelectorLabelsEntry](#discovery.zephyr.solo.io.MeshServiceSpec.KubeService.WorkloadSelectorLabelsEntry)
  - [MeshServiceSpec.Subset](#discovery.zephyr.solo.io.MeshServiceSpec.Subset)
  - [MeshServiceSpec.SubsetsEntry](#discovery.zephyr.solo.io.MeshServiceSpec.SubsetsEntry)
  - [MeshServiceStatus](#discovery.zephyr.solo.io.MeshServiceStatus)







<a name="discovery.zephyr.solo.io.MeshServiceSpec"></a>

### MeshServiceSpec
The MeshService is an abstraction for a service which we have discovered to be part of a given mesh. The Mesh object has references to the MeshServices which belong to it.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeService | [MeshServiceSpec.KubeService](#discovery.zephyr.solo.io.MeshServiceSpec.KubeService) |  | Metadata about the kube-native service backing this MeshService. |
| mesh | [core.zephyr.solo.io.ResourceRef](#core.zephyr.solo.io.ResourceRef) |  | The mesh with which this service is associated. |
| subsets | [][MeshServiceSpec.SubsetsEntry](#discovery.zephyr.solo.io.MeshServiceSpec.SubsetsEntry) | repeated | Subsets for routing, based on labels. |
| federation | [MeshServiceSpec.Federation](#discovery.zephyr.solo.io.MeshServiceSpec.Federation) |  | Metadata about the decisions that Service Mesh Hub has made about what workloads this service is federated to. |






<a name="discovery.zephyr.solo.io.MeshServiceSpec.Federation"></a>

### MeshServiceSpec.Federation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| multiclusterDnsName | [string](#string) |  | For any workload that this service has federated to (i.e., any MeshWorkload whose ref appears in `federated_to_workloads`), a client in that workload will be able to reach this service at this DNS name. This includes workloads on clusters other than the one hosting this service. |
| federatedToWorkloads | [][core.zephyr.solo.io.ResourceRef](#core.zephyr.solo.io.ResourceRef) | repeated | The list of MeshWorkloads which are able to resolve this service's `multicluster_dns_name`. |






<a name="discovery.zephyr.solo.io.MeshServiceSpec.KubeService"></a>

### MeshServiceSpec.KubeService



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.zephyr.solo.io.ResourceRef](#core.zephyr.solo.io.ResourceRef) |  | A reference to the kube-native service that this MeshService represents. |
| workloadSelectorLabels | [][MeshServiceSpec.KubeService.WorkloadSelectorLabelsEntry](#discovery.zephyr.solo.io.MeshServiceSpec.KubeService.WorkloadSelectorLabelsEntry) | repeated | Selectors for the set of pods targeted by the k8s Service. |
| labels | [][MeshServiceSpec.KubeService.LabelsEntry](#discovery.zephyr.solo.io.MeshServiceSpec.KubeService.LabelsEntry) | repeated | Labels on the underyling k8s Service itself. |
| ports | [][MeshServiceSpec.KubeService.KubeServicePort](#discovery.zephyr.solo.io.MeshServiceSpec.KubeService.KubeServicePort) | repeated | The ports exposed by the underlying service. |






<a name="discovery.zephyr.solo.io.MeshServiceSpec.KubeService.KubeServicePort"></a>

### MeshServiceSpec.KubeService.KubeServicePort



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | [uint32](#uint32) |  | external-facing port for this service (i.e., NOT the service's target port on the backing pods) |
| name | [string](#string) |  |  |
| protocol | [string](#string) |  |  |






<a name="discovery.zephyr.solo.io.MeshServiceSpec.KubeService.LabelsEntry"></a>

### MeshServiceSpec.KubeService.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="discovery.zephyr.solo.io.MeshServiceSpec.KubeService.WorkloadSelectorLabelsEntry"></a>

### MeshServiceSpec.KubeService.WorkloadSelectorLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






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
| federationStatus | [core.zephyr.solo.io.Status](#core.zephyr.solo.io.Status) |  | The status of federation artifacts being written to remote clusters as a result of the federation metadata on this object's Spec. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

