
---

title: "access_policy.proto"

---

## Package : `networking.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for access_policy.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## access_policy.proto


## Table of Contents
  - [AccessPolicySpec](#networking.mesh.gloo.solo.io.AccessPolicySpec)
  - [AccessPolicyStatus](#networking.mesh.gloo.solo.io.AccessPolicyStatus)
  - [AccessPolicyStatus.TrafficTargetsEntry](#networking.mesh.gloo.solo.io.AccessPolicyStatus.TrafficTargetsEntry)







<a name="networking.mesh.gloo.solo.io.AccessPolicySpec"></a>

### AccessPolicySpec
Grants communication permission between selected identities (i.e. traffic sources) and TrafficTargets (i.e. traffic targets). Explicitly granted access permission is required if a [VirtualMesh's GlobalAccessPolicy]({{% versioned_link_path fromRoot="/reference/api/virtual_mesh/#networking.mesh.gloo.solo.io.VirtualMeshSpec.GlobalAccessPolicy" %}}) is set to `ENABLED`.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sourceSelector | [][common.mesh.gloo.solo.io.IdentitySelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1alpha2.selectors#common.mesh.gloo.solo.io.IdentitySelector" >}}) | repeated | Specify the identities of Workloads (i.e. traffic sources) for which to apply this AccessPolicy. Leave empty to apply the AccessPolicy to all Workloads colocated in the destination's Mesh. |
  | destinationSelector | [][common.mesh.gloo.solo.io.TrafficTargetSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1alpha2.selectors#common.mesh.gloo.solo.io.TrafficTargetSelector" >}}) | repeated | Specify the TrafficTargets for which to apply this AccessPolicy. Leave empty to apply the AccessPolicy to all TrafficTargets. |
  | allowedPaths | []string | repeated | Optional.* A list of HTTP paths or gRPC methods to allow. gRPC methods must be presented as fully-qualified name in the form of "/packageName.serviceName/methodName" and are case sensitive. Exact match, prefix match, and suffix match are supported for paths. For example, the path "/books/review" matches "/books/review" (exact match), "*books/" (suffix match), or "/books*" (prefix match).<br>If not specified, allow any path. |
  | allowedMethods | [][networking.mesh.gloo.solo.io.HttpMethodValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.http#networking.mesh.gloo.solo.io.HttpMethodValue" >}}) | repeated | Optional.* A list of HTTP methods to allow (e.g., "GET", "POST"). It is ignored in gRPC case because the value is always "POST". If not specified, allows any method. |
  | allowedPorts | []uint32 | repeated | Optional.* A list of ports which to allow. If not set any port is allowed. |
  





<a name="networking.mesh.gloo.solo.io.AccessPolicyStatus"></a>

### AccessPolicyStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the AccessPolicy metadata. If the `observedGeneration` does not match `metadata.generation`, Gloo Mesh has not processed the most recent version of this resource. |
  | state | [networking.mesh.gloo.solo.io.ApprovalState]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.validation_state#networking.mesh.gloo.solo.io.ApprovalState" >}}) |  | The state of the overall resource. It will only show accepted if it has been successfully applied to selected TrafficTargets. |
  | trafficTargets | [][networking.mesh.gloo.solo.io.AccessPolicyStatus.TrafficTargetsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.access_policy#networking.mesh.gloo.solo.io.AccessPolicyStatus.TrafficTargetsEntry" >}}) | repeated | The status of the AccessPolicy for each TrafficTarget to which it has been applied. An AccessPolicy may be accepted for some TrafficTargets and rejected for others. |
  | workloads | []string | repeated | The list of Workloads to which this policy has been applied. |
  | errors | []string | repeated | Any errors found while processing this generation of the resource. |
  





<a name="networking.mesh.gloo.solo.io.AccessPolicyStatus.TrafficTargetsEntry"></a>

### AccessPolicyStatus.TrafficTargetsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [networking.mesh.gloo.solo.io.ApprovalStatus]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.validation_state#networking.mesh.gloo.solo.io.ApprovalStatus" >}}) |  |  |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

