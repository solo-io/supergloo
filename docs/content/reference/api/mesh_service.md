
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
  - [MeshServiceSpec.KubeService](#discovery.smh.solo.io.MeshServiceSpec.KubeService)
  - [MeshServiceSpec.KubeService.KubeServicePort](#discovery.smh.solo.io.MeshServiceSpec.KubeService.KubeServicePort)
  - [MeshServiceSpec.KubeService.LabelsEntry](#discovery.smh.solo.io.MeshServiceSpec.KubeService.LabelsEntry)
  - [MeshServiceSpec.KubeService.Subset](#discovery.smh.solo.io.MeshServiceSpec.KubeService.Subset)
  - [MeshServiceSpec.KubeService.SubsetsEntry](#discovery.smh.solo.io.MeshServiceSpec.KubeService.SubsetsEntry)
  - [MeshServiceSpec.KubeService.WorkloadSelectorLabelsEntry](#discovery.smh.solo.io.MeshServiceSpec.KubeService.WorkloadSelectorLabelsEntry)
  - [MeshServiceStatus](#discovery.smh.solo.io.MeshServiceStatus)
  - [MeshServiceStatus.AppliedAccessPolicy](#discovery.smh.solo.io.MeshServiceStatus.AppliedAccessPolicy)
  - [MeshServiceStatus.AppliedFederation](#discovery.smh.solo.io.MeshServiceStatus.AppliedFederation)
  - [MeshServiceStatus.AppliedTrafficPolicy](#discovery.smh.solo.io.MeshServiceStatus.AppliedTrafficPolicy)







<a name="discovery.smh.solo.io.MeshServiceSpec"></a>

### MeshServiceSpec
The MeshService is an abstraction for a service which we have discovered to be part of a given mesh. The Mesh object has references to the MeshServices which belong to it.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeService | [MeshServiceSpec.KubeService](#discovery.smh.solo.io.MeshServiceSpec.KubeService) |  | Metadata about the kube-native service backing this MeshService. |
| mesh | [core.skv2.solo.io.ObjectRef](#core.skv2.solo.io.ObjectRef) |  | The mesh with which this service is associated. |






<a name="discovery.smh.solo.io.MeshServiceSpec.KubeService"></a>

### MeshServiceSpec.KubeService



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.skv2.solo.io.ClusterObjectRef](#core.skv2.solo.io.ClusterObjectRef) |  | A reference to the kube-native service that this MeshService represents. |
| workloadSelectorLabels | [][MeshServiceSpec.KubeService.WorkloadSelectorLabelsEntry](#discovery.smh.solo.io.MeshServiceSpec.KubeService.WorkloadSelectorLabelsEntry) | repeated | Selectors for the set of pods targeted by the k8s Service. |
| labels | [][MeshServiceSpec.KubeService.LabelsEntry](#discovery.smh.solo.io.MeshServiceSpec.KubeService.LabelsEntry) | repeated | Labels on the underlying k8s Service itself. |
| ports | [][MeshServiceSpec.KubeService.KubeServicePort](#discovery.smh.solo.io.MeshServiceSpec.KubeService.KubeServicePort) | repeated | The ports exposed by the underlying service. |
| subsets | [][MeshServiceSpec.KubeService.SubsetsEntry](#discovery.smh.solo.io.MeshServiceSpec.KubeService.SubsetsEntry) | repeated | Subsets for routing, based on labels. |






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






<a name="discovery.smh.solo.io.MeshServiceSpec.KubeService.Subset"></a>

### MeshServiceSpec.KubeService.Subset



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| values | [][string](#string) | repeated |  |






<a name="discovery.smh.solo.io.MeshServiceSpec.KubeService.SubsetsEntry"></a>

### MeshServiceSpec.KubeService.SubsetsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [MeshServiceSpec.KubeService.Subset](#discovery.smh.solo.io.MeshServiceSpec.KubeService.Subset) |  |  |






<a name="discovery.smh.solo.io.MeshServiceSpec.KubeService.WorkloadSelectorLabelsEntry"></a>

### MeshServiceSpec.KubeService.WorkloadSelectorLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="discovery.smh.solo.io.MeshServiceStatus"></a>

### MeshServiceStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | [int64](#int64) |  | The most recent generation observed in the the TrafficPolicy metadata. if the observedGeneration does not match generation, the controller has not received the most recent version of this resource. |
| appliedTrafficPolicies | [][MeshServiceStatus.AppliedTrafficPolicy](#discovery.smh.solo.io.MeshServiceStatus.AppliedTrafficPolicy) | repeated | The set of Traffic Policies that have been applied to this MeshService |
| appliedAccessPolicies | [][MeshServiceStatus.AppliedAccessPolicy](#discovery.smh.solo.io.MeshServiceStatus.AppliedAccessPolicy) | repeated | The set of Access Policies that have been applied to this MeshService |






<a name="discovery.smh.solo.io.MeshServiceStatus.AppliedAccessPolicy"></a>

### MeshServiceStatus.AppliedAccessPolicy
AppliedAccessPolicy represents a access policy that has been applied to the MeshService. if an existing Access Policy becomes invalid, the last applied policy will be used


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.skv2.solo.io.ObjectRef](#core.skv2.solo.io.ObjectRef) |  | reference to the access policy |
| observedGeneration | [int64](#int64) |  | the observed generation of the accepted access policy |
| spec | [networking.smh.solo.io.AccessPolicySpec](#networking.smh.solo.io.AccessPolicySpec) |  | the last known valid spec of the access policy |






<a name="discovery.smh.solo.io.MeshServiceStatus.AppliedFederation"></a>

### MeshServiceStatus.AppliedFederation
Federation policy applied to this MeshService, allowing access to the service from other meshes/clusters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| multiclusterDnsName | [string](#string) |  | For any workload that this service has federated to (i.e., any MeshWorkload whose ref appears in `federated_to_workloads`), a client in that workload will be able to reach this service at this DNS name. This includes workloads on clusters other than the one hosting this service. |
| federatedToMeshes | [][core.skv2.solo.io.ObjectRef](#core.skv2.solo.io.ObjectRef) | repeated | The list of Meshes which are able to resolve this service's `multicluster_dns_name`. |






<a name="discovery.smh.solo.io.MeshServiceStatus.AppliedTrafficPolicy"></a>

### MeshServiceStatus.AppliedTrafficPolicy
AppliedTrafficPolicy represents a traffic policy that has been applied to the MeshService. if an existing Traffic Policy becomes invalid, the last applied policy will be used


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.skv2.solo.io.ObjectRef](#core.skv2.solo.io.ObjectRef) |  | reference to the traffic policy |
| observedGeneration | [int64](#int64) |  | the observed generation of the accepted traffic policy |
| spec | [networking.smh.solo.io.TrafficPolicySpec](#networking.smh.solo.io.TrafficPolicySpec) |  | the last known valid spec of the traffic policy |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

