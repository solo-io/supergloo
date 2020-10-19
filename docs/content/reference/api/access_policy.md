
---
title: "access_policy.proto"
---

## Package : `networking.smh.solo.io`



<a name="top"></a>

<a name="API Reference for access_policy.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## access_policy.proto


## Table of Contents
  - [AccessPolicySpec](#networking.smh.solo.io.AccessPolicySpec)
  - [AccessPolicyStatus](#networking.smh.solo.io.AccessPolicyStatus)
  - [AccessPolicyStatus.TrafficTargetsEntry](#networking.smh.solo.io.AccessPolicyStatus.TrafficTargetsEntry)







<a name="networking.smh.solo.io.AccessPolicySpec"></a>

### AccessPolicySpec
Access control policies apply ALLOW policies to communication in a mesh. Access control policies specify the following: ALLOW those requests that: originate from from **source workload**, target the **destination target**, and match the indicated request criteria (allowed_paths, allowed_methods, allowed_ports). Enforcement of access control is determined by the [VirtualMesh's GlobalAccessPolicy]({{% versioned_link_path fromRoot="/reference/api/virtual_mesh/#networking.smh.solo.io.VirtualMeshSpec.GlobalAccessPolicy" %}})


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sourceSelector | []networking.smh.solo.io.IdentitySelector | repeated | Requests originating from these pods will have the rule applied. Leave empty to have all pods in the mesh apply these policies.<br>Note that access control policies are mapped to source pods by their service account. If other pods share the same service account, this access control rule will apply to those pods as well.<br>For fine-grained access control policies, ensure that your service accounts properly reflect the desired boundary for your access control policies. |
| destinationSelector | []networking.smh.solo.io.TrafficTargetSelector | repeated | Requests destined for these pods will have the rule applied. Leave empty to apply to all destination pods in the mesh. |
| allowedPaths | []string | repeated | Optional. A list of HTTP paths or gRPC methods to allow. gRPC methods must be presented as fully-qualified name in the form of "/packageName.serviceName/methodName" and are case sensitive. Exact match, prefix match, and suffix match are supported for paths. For example, the path "/books/review" matches "/books/review" (exact match), "*books/" (suffix match), or "/books*" (prefix match).<br>If not specified, allow any path. |
| allowedMethods | []networking.smh.solo.io.HttpMethodValue | repeated | Optional. A list of HTTP methods to allow (e.g., "GET", "POST"). It is ignored in gRPC case because the value is always "POST". If not specified, allows any method. |
| allowedPorts | []uint32 | repeated | Optional. A list of ports which to allow. If not set any port is allowed. |






<a name="networking.smh.solo.io.AccessPolicyStatus"></a>

### AccessPolicyStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the AccessPolicy metadata. If the observedGeneration does not match generation, the controller has not received the most recent version of this resource. |
| state | networking.smh.solo.io.ApprovalState |  | The state of the overall resource. It will only show accepted if it has been successfully applied to all target meshes. |
| trafficTargets | []networking.smh.solo.io.AccessPolicyStatus.TrafficTargetsEntry | repeated | The status of the AccessPolicy for each TrafficTarget to which it has been applied. An AccessPolicy may be Accepted for some TrafficTargets and rejected for others. |
| workloads | []string | repeated | The list of Workloads to which this policy has been applied. |
| errors | []string | repeated | Any errors found while processing this generation of the resource. |






<a name="networking.smh.solo.io.AccessPolicyStatus.TrafficTargetsEntry"></a>

### AccessPolicyStatus.TrafficTargetsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
| value | networking.smh.solo.io.ApprovalStatus |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

