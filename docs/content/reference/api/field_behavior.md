
---
title: "field_behavior.proto"
---

## Package : `google.api`



<a name="top"></a>

<a name="API Reference for field_behavior.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## field_behavior.proto


## Table of Contents

  - [FieldBehavior](#google.api.FieldBehavior)

  - [File-level Extensions](#field_behavior.proto-extensions)




 <!-- end messages -->


<a name="google.api.FieldBehavior"></a>

### FieldBehavior
An indicator of the behavior of a given field (for example, that a field is required in requests, or given as output but ignored as input). This **does not** change the behavior in protocol buffers itself; it only denotes the behavior and may affect how API tooling handles the field.<br>Note: This enum **may** receive new values in the future.

| Name | Number | Description |
| ---- | ------ | ----------- |
| FIELD_BEHAVIOR_UNSPECIFIED | 0 | Conventional default for enums. Do not use this. |
| OPTIONAL | 1 | Specifically denotes a field as optional. While all fields in protocol buffers are optional, this may be specified for emphasis if appropriate. |
| REQUIRED | 2 | Denotes a field as required. This indicates that the field **must** be provided as part of the request, and failure to do so will cause an error (usually `INVALID_ARGUMENT`). |
| OUTPUT_ONLY | 3 | Denotes a field as output only. This indicates that the field is provided in responses, but including the field in a request does nothing (the server *must* ignore it and *must not* throw an error as a result of the field's presence). |
| INPUT_ONLY | 4 | Denotes a field as input only. This indicates that the field is provided in requests, and the corresponding field is not included in output. |
| IMMUTABLE | 5 | Denotes a field as immutable. This indicates that the field may be set once in a request to create a resource, but may not be changed thereafter. |


 <!-- end enums -->


<a name="field_behavior.proto-extensions"></a>

### File-level Extensions
| Extension | Type | Base | Number | Description |
| --------- | ---- | ---- | ------ | ----------- |
| field_behavior | FieldBehavior | .google.protobuf.FieldOptions | 1052 | A designation of a specific field behavior (required, output only, etc.) in protobuf messages.<br>Examples:<br>  string name = 1 [(google.api.field_behavior) = REQUIRED];   State state = 1 [(google.api.field_behavior) = OUTPUT_ONLY];   google.protobuf.Duration ttl = 1     [(google.api.field_behavior) = INPUT_ONLY];   google.protobuf.Timestamp expire_time = 1     [(google.api.field_behavior) = OUTPUT_ONLY,      (google.api.field_behavior) = IMMUTABLE]; |

 <!-- end HasExtensions -->

 <!-- end services -->

