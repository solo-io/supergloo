
---
title: "traffic_target.proto"
---

## Package : `discovery.smh.solo.io`



<a name="top"></a>

<a name="API Reference for traffic_target.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## traffic_target.proto


## Table of Contents
  - [TrafficTargetSpec](#discovery.smh.solo.io.TrafficTargetSpec)
  - [TrafficTargetSpec.KubeService](#discovery.smh.solo.io.TrafficTargetSpec.KubeService)
  - [TrafficTargetSpec.KubeService.KubeServicePort](#discovery.smh.solo.io.TrafficTargetSpec.KubeService.KubeServicePort)
  - [TrafficTargetSpec.KubeService.LabelsEntry](#discovery.smh.solo.io.TrafficTargetSpec.KubeService.LabelsEntry)
  - [TrafficTargetSpec.KubeService.Subset](#discovery.smh.solo.io.TrafficTargetSpec.KubeService.Subset)
  - [TrafficTargetSpec.KubeService.SubsetsEntry](#discovery.smh.solo.io.TrafficTargetSpec.KubeService.SubsetsEntry)
  - [TrafficTargetSpec.KubeService.WorkloadSelectorLabelsEntry](#discovery.smh.solo.io.TrafficTargetSpec.KubeService.WorkloadSelectorLabelsEntry)
  - [TrafficTargetStatus](#discovery.smh.solo.io.TrafficTargetStatus)
  - [TrafficTargetStatus.AppliedAccessPolicy](#discovery.smh.solo.io.TrafficTargetStatus.AppliedAccessPolicy)
  - [TrafficTargetStatus.AppliedFederation](#discovery.smh.solo.io.TrafficTargetStatus.AppliedFederation)
  - [TrafficTargetStatus.AppliedTrafficPolicy](#discovery.smh.solo.io.TrafficTargetStatus.AppliedTrafficPolicy)







<a name="discovery.smh.solo.io.TrafficTargetSpec"></a>

### TrafficTargetSpec
The TrafficTarget is an abstraction for a traffic target which we have discovered to be part of a given mesh.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeService | discovery.smh.solo.io.TrafficTargetSpec.KubeService |  | Metadata about the kube-native traffic target backing this TrafficTarget. |
| mesh | core.skv2.solo.io.ObjectRef |  | The mesh with which this traffic target is associated. |






<a name="discovery.smh.solo.io.TrafficTargetSpec.KubeService"></a>

### TrafficTargetSpec.KubeService



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | core.skv2.solo.io.ClusterObjectRef |  | A reference to the kube-native traffic target that this TrafficTarget represents. |
| workloadSelectorLabels | []discovery.smh.solo.io.TrafficTargetSpec.KubeService.WorkloadSelectorLabelsEntry | repeated | Selectors for the set of pods targeted by the k8s Service. |
| labels | []discovery.smh.solo.io.TrafficTargetSpec.KubeService.LabelsEntry | repeated | Labels on the underlying k8s Service itself. |
| ports | []discovery.smh.solo.io.TrafficTargetSpec.KubeService.KubeServicePort | repeated | The ports exposed by the underlying service. |
| subsets | []discovery.smh.solo.io.TrafficTargetSpec.KubeService.SubsetsEntry | repeated | Subsets for routing, based on labels. |






<a name="discovery.smh.solo.io.TrafficTargetSpec.KubeService.KubeServicePort"></a>

### TrafficTargetSpec.KubeService.KubeServicePort



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | uint32 |  | External-facing port for this k8s service (NOT the service's target port on the backing pods). |
| name | string |  |  |
| protocol | string |  |  |






<a name="discovery.smh.solo.io.TrafficTargetSpec.KubeService.LabelsEntry"></a>

### TrafficTargetSpec.KubeService.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
| value | string |  |  |






<a name="discovery.smh.solo.io.TrafficTargetSpec.KubeService.Subset"></a>

### TrafficTargetSpec.KubeService.Subset
Subsets for routing, based on labels.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| values | []string | repeated |  |






<a name="discovery.smh.solo.io.TrafficTargetSpec.KubeService.SubsetsEntry"></a>

### TrafficTargetSpec.KubeService.SubsetsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
| value | discovery.smh.solo.io.TrafficTargetSpec.KubeService.Subset |  |  |






<a name="discovery.smh.solo.io.TrafficTargetSpec.KubeService.WorkloadSelectorLabelsEntry"></a>

### TrafficTargetSpec.KubeService.WorkloadSelectorLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
| value | string |  |  |






<a name="discovery.smh.solo.io.TrafficTargetStatus"></a>

### TrafficTargetStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the TrafficPolicy metadata. if the observedGeneration does not match generation, the controller has not received the most recent version of this resource. |
| appliedTrafficPolicies | []discovery.smh.solo.io.TrafficTargetStatus.AppliedTrafficPolicy | repeated | The set of Traffic Policies that have been applied to this TrafficTarget |
| appliedAccessPolicies | []discovery.smh.solo.io.TrafficTargetStatus.AppliedAccessPolicy | repeated | The set of Access Policies that have been applied to this TrafficTarget |






<a name="discovery.smh.solo.io.TrafficTargetStatus.AppliedAccessPolicy"></a>

### TrafficTargetStatus.AppliedAccessPolicy
AppliedAccessPolicy represents a access policy that has been applied to the TrafficTarget. if an existing Access Policy becomes invalid, the last applied policy will be used


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | core.skv2.solo.io.ObjectRef |  | reference to the access policy |
| observedGeneration | int64 |  | the observed generation of the accepted access policy |
| spec | networking.smh.solo.io.AccessPolicySpec |  | the last known valid spec of the access policy |






<a name="discovery.smh.solo.io.TrafficTargetStatus.AppliedFederation"></a>

### TrafficTargetStatus.AppliedFederation
Federation policy applied to this TrafficTarget, allowing access to the traffic target from other meshes/clusters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| multiclusterDnsName | string |  | For any workload that this traffic target has federated to (i.e., any Workload whose ref appears in `federated_to_workloads`), a client in that workload will be able to reach this traffic target at this DNS name. This includes workloads on clusters other than the one hosting this service. |
| federatedToMeshes | []core.skv2.solo.io.ObjectRef | repeated | The list of Meshes which are able to resolve this service's `multicluster_dns_name`. |






<a name="discovery.smh.solo.io.TrafficTargetStatus.AppliedTrafficPolicy"></a>

### TrafficTargetStatus.AppliedTrafficPolicy
AppliedTrafficPolicy represents a traffic policy that has been applied to the TrafficTarget. if an existing Traffic Policy becomes invalid, the last applied policy will be used


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | core.skv2.solo.io.ObjectRef |  | reference to the traffic policy |
| observedGeneration | int64 |  | the observed generation of the accepted traffic policy |
| spec | networking.smh.solo.io.TrafficPolicySpec |  | the last known valid spec of the traffic policy |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

