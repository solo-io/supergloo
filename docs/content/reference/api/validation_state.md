
---
title: "validation_state.proto"
---

## Package : `networking.smh.solo.io`



<a name="top"></a>

<a name="API Reference for validation_state.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## validation_state.proto


## Table of Contents
  - [ApprovalStatus](#networking.smh.solo.io.ApprovalStatus)

  - [ApprovalState](#networking.smh.solo.io.ApprovalState)






<a name="networking.smh.solo.io.ApprovalStatus"></a>

### ApprovalStatus
The approval status of a policy that has been applied to a discovery resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| acceptanceOrder | uint32 |  | AcceptanceOrder represents the order in which the Policy was accepted and applied to a discovery resource. The first accepted policy will have an acceptance_order of 0, the second 1, etc. When conflicts are detected in the system, the Policy with the lowest acceptance_order will be chosen (and all other conflicting policies will be rejected). |
| state | networking.smh.solo.io.ApprovalState |  | The result of attempting to apply the policy to the discovery resource, reported by the Policy controller (mesh-networking). |
| errors | []string | repeated | Any errors observed which prevented the resource from being Accepted. |





 <!-- end messages -->


<a name="networking.smh.solo.io.ApprovalState"></a>

### ApprovalState
State of a Policy resource reflected in the status by Service Mesh Hub while processing a resource.

| Name | Number | Description |
| ---- | ------ | ----------- |
| PENDING | 0 | Resources are in a Pending state before they have been processed by Service Mesh Hub. |
| ACCEPTED | 1 | Resources are in a Accepted state when they are valid and have been applied successfully to the Service Mesh Hub configuration. |
| INVALID | 2 | Resources are in an Invalid state when they contain incorrect configuration parameters, such as missing required values or invalid resource references. An invalid state can also result when a resource's configuration is valid but conflicts with another resource which was accepted in an earlier point in time. |
| FAILED | 3 | Resources are in a Failed state when they contain correct configuration parameters, but the server encountered an error trying to synchronize the system to the desired state. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

