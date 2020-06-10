
---
title: "mesh_service.proto"
---

## Package : `discovery.smh.solo.io`



<a name="top"></a>

<a name="API Reference for mesh_service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## mesh_service.proto


## Table of Contents
  - [MeshServiceSpec](#discovery.smh.solo.io.MeshServiceSpec)
  - [MeshServiceSpec.Federation](#discovery.smh.solo.io.MeshServiceSpec.Federation)
  - [MeshServiceSpec.KubeService](#discovery.smh.solo.io.MeshServiceSpec.KubeService)
  - [MeshServiceSpec.KubeService.KubeServicePort](#discovery.smh.solo.io.MeshServiceSpec.KubeService.KubeServicePort)
  - [MeshServiceSpec.KubeService.LabelsEntry](#discovery.smh.solo.io.MeshServiceSpec.KubeService.LabelsEntry)
  - [MeshServiceSpec.KubeService.WorkloadSelectorLabelsEntry](#discovery.smh.solo.io.MeshServiceSpec.KubeService.WorkloadSelectorLabelsEntry)
  - [MeshServiceSpec.Subset](#discovery.smh.solo.io.MeshServiceSpec.Subset)
  - [MeshServiceSpec.SubsetsEntry](#discovery.smh.solo.io.MeshServiceSpec.SubsetsEntry)
  - [MeshServiceStatus](#discovery.smh.solo.io.MeshServiceStatus)
  - [MeshServiceStatus.ValidatedTrafficPolicy](#discovery.smh.solo.io.MeshServiceStatus.ValidatedTrafficPolicy)







<a name="discovery.smh.solo.io.MeshServiceSpec"></a>

### MeshServiceSpec
The MeshService is an abstraction for a service which we have discovered to be part of a given mesh. The Mesh object has references to the MeshServices which belong to it.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeService | [MeshServiceSpec.KubeService](#discovery.smh.solo.io.MeshServiceSpec.KubeService) |  | Metadata about the kube-native service backing this MeshService. |
| mesh | [core.smh.solo.io.ResourceRef](#core.smh.solo.io.ResourceRef) |  | The mesh with which this service is associated. |
| subsets | [][MeshServiceSpec.SubsetsEntry](#discovery.smh.solo.io.MeshServiceSpec.SubsetsEntry) | repeated | Subsets for routing, based on labels. |
| federation | [MeshServiceSpec.Federation](#discovery.smh.solo.io.MeshServiceSpec.Federation) |  | Metadata about the decisions that Service Mesh Hub has made about what workloads this service is federated to. |






<a name="discovery.smh.solo.io.MeshServiceSpec.Federation"></a>

### MeshServiceSpec.Federation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| multiclusterDnsName | [string](#string) |  | For any workload that this service has federated to (i.e., any MeshWorkload whose ref appears in `federated_to_workloads`), a client in that workload will be able to reach this service at this DNS name. This includes workloads on clusters other than the one hosting this service. |
| federatedToWorkloads | [][core.smh.solo.io.ResourceRef](#core.smh.solo.io.ResourceRef) | repeated | The list of MeshWorkloads which are able to resolve this service's `multicluster_dns_name`. |






<a name="discovery.smh.solo.io.MeshServiceSpec.KubeService"></a>

### MeshServiceSpec.KubeService



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.smh.solo.io.ResourceRef](#core.smh.solo.io.ResourceRef) |  | A reference to the kube-native service that this MeshService represents. |
| workloadSelectorLabels | [][MeshServiceSpec.KubeService.WorkloadSelectorLabelsEntry](#discovery.smh.solo.io.MeshServiceSpec.KubeService.WorkloadSelectorLabelsEntry) | repeated | Selectors for the set of pods targeted by the k8s Service. |
| labels | [][MeshServiceSpec.KubeService.LabelsEntry](#discovery.smh.solo.io.MeshServiceSpec.KubeService.LabelsEntry) | repeated | Labels on the underlying k8s Service itself. |
| ports | [][MeshServiceSpec.KubeService.KubeServicePort](#discovery.smh.solo.io.MeshServiceSpec.KubeService.KubeServicePort) | repeated | The ports exposed by the underlying service. |






<a name="discovery.smh.solo.io.MeshServiceSpec.KubeService.KubeServicePort"></a>

### MeshServiceSpec.KubeService.KubeServicePort



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | [uint32](#uint32) |  | external-facing port for this service (i.e., NOT the service's target port on the backing pods) |
| name | [string](#string) |  |  |
| protocol | [string](#string) |  |  |






<a name="discovery.smh.solo.io.MeshServiceSpec.KubeService.LabelsEntry"></a>

### MeshServiceSpec.KubeService.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="discovery.smh.solo.io.MeshServiceSpec.KubeService.WorkloadSelectorLabelsEntry"></a>

### MeshServiceSpec.KubeService.WorkloadSelectorLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="discovery.smh.solo.io.MeshServiceSpec.Subset"></a>

### MeshServiceSpec.Subset



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| values | [][string](#string) | repeated |  |






<a name="discovery.smh.solo.io.MeshServiceSpec.SubsetsEntry"></a>

### MeshServiceSpec.SubsetsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [MeshServiceSpec.Subset](#discovery.smh.solo.io.MeshServiceSpec.Subset) |  |  |






<a name="discovery.smh.solo.io.MeshServiceStatus"></a>

### MeshServiceStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| federationStatus | [core.smh.solo.io.Status](#core.smh.solo.io.Status) |  | The status of federation artifacts being written to remote clusters as a result of the federation metadata on this object's Spec. |
| validatedTrafficPolicies | [][MeshServiceStatus.ValidatedTrafficPolicy](#discovery.smh.solo.io.MeshServiceStatus.ValidatedTrafficPolicy) | repeated |  |






<a name="discovery.smh.solo.io.MeshServiceStatus.ValidatedTrafficPolicy"></a>

### MeshServiceStatus.ValidatedTrafficPolicy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.smh.solo.io.ResourceRef](#core.smh.solo.io.ResourceRef) |  |  |
| trafficPolicySpec | [networking.smh.solo.io.TrafficPolicySpec](#networking.smh.solo.io.TrafficPolicySpec) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

