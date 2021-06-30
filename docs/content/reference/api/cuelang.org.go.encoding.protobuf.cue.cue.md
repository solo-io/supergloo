
---

---

## Package : `cue`



<a name="top"></a>

<a name="API Reference for cue.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cue.proto


## Table of Contents
  - [FieldOptions](#cue.FieldOptions)


  - [File-level Extensions](#cue.proto-extensions)
  - [File-level Extensions](#cue.proto-extensions)





<a name="cue.FieldOptions"></a>

### FieldOptions



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| required | bool |  |  |
  | disableOpenapiValidation | bool |  | Enable this option to treat this field use an unstructured object in the OpenAPI schema for this field. This is currently required to disable infinite recursion when expanding references with CUE on recursive types. |
  




 <!-- end messages -->

 <!-- end enums -->


<a name="cue.proto-extensions"></a>

### File-level Extensions
| Extension | Type | Base | Number | Description |
| --------- | ---- | ---- | ------ | ----------- |
| opt | FieldOptions | .google.protobuf.FieldOptions | 1069 |  |
| val | string | .google.protobuf.FieldOptions | 123456 |  |

 <!-- end HasExtensions -->

 <!-- end services -->

