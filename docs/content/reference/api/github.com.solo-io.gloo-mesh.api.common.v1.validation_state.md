
---

title: "validation_state.proto"

---

## Package : `common.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for validation_state.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## validation_state.proto


## Table of Contents

  - [ApprovalState](#common.mesh.gloo.solo.io.ApprovalState)





 <!-- end messages -->


<a name="common.mesh.gloo.solo.io.ApprovalState"></a>

### ApprovalState
State of a Policy resource reflected in the status by Gloo Mesh while processing a resource.

| Name | Number | Description |
| ---- | ------ | ----------- |
| PENDING | 0 | Resources are in a Pending state before they have been processed by Gloo Mesh. |
| ACCEPTED | 1 | Resources are in a Accepted state when they are valid and have been applied successfully to the Gloo Mesh configuration. |
| INVALID | 2 | Resources are in an Invalid state when they contain incorrect configuration parameters, such as missing required values or invalid resource references. An invalid state can also result when a resource's configuration is valid but conflicts with another resource which was accepted in an earlier point in time. |
| FAILED | 3 | Resources are in a Failed state when they contain correct configuration parameters, but the server encountered an error trying to synchronize the system to the desired state. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

