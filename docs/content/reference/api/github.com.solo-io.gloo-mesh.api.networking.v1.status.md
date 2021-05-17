
---

title: "status.proto"

---

## Package : `networking.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for status.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## status.proto


## Table of Contents
  - [ApprovalStatus](#networking.mesh.gloo.solo.io.ApprovalStatus)







<a name="networking.mesh.gloo.solo.io.ApprovalStatus"></a>

### ApprovalStatus
The approval status of a policy that has been applied to a discovery resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| acceptanceOrder | uint32 |  | Represents the order in which the policy was accepted and applied to a discovery resource. The first accepted policy will have an acceptance_order of 0, the second 1, etc. When conflicts are detected in the system, the Policy with the lowest acceptance_order will be chosen and all other conflicting policies will be rejected. |
  | state | [common.mesh.gloo.solo.io.ApprovalState]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.validation_state#common.mesh.gloo.solo.io.ApprovalState" >}}) |  | The result of attempting to apply the policy to the discovery resource. |
  | errors | []string | repeated | Any errors observed which prevented the resource from being Accepted. |
  | warnings | []string | repeated | Any warnings observed while processing the resource. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

