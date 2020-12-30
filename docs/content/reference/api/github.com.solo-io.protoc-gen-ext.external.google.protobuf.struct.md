
---

---

## Package : `google.protobuf`



<a name="top"></a>

<a name="API Reference for struct.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## struct.proto


## Table of Contents
  - [ListValue](#google.protobuf.ListValue)
  - [Struct](#google.protobuf.Struct)
  - [Struct.FieldsEntry](#google.protobuf.Struct.FieldsEntry)
  - [Value](#google.protobuf.Value)

  - [NullValue](#google.protobuf.NullValue)






<a name="google.protobuf.ListValue"></a>

### ListValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| values | [][google.protobuf.Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.struct#google.protobuf.Value" >}}) | repeated | Repeated field of dynamically typed values. |
  





<a name="google.protobuf.Struct"></a>

### Struct



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [][google.protobuf.Struct.FieldsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.struct#google.protobuf.Struct.FieldsEntry" >}}) | repeated | Unordered map of dynamically typed values. |
  





<a name="google.protobuf.Struct.FieldsEntry"></a>

### Struct.FieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [google.protobuf.Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.struct#google.protobuf.Value" >}}) |  |  |
  





<a name="google.protobuf.Value"></a>

### Value



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nullValue | [google.protobuf.NullValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.struct#google.protobuf.NullValue" >}}) |  | Represents a null value. |
  | numberValue | double |  | Represents a double value. |
  | stringValue | string |  | Represents a string value. |
  | boolValue | bool |  | Represents a boolean value. |
  | structValue | [google.protobuf.Struct]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.struct#google.protobuf.Struct" >}}) |  | Represents a structured value. |
  | listValue | [google.protobuf.ListValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.struct#google.protobuf.ListValue" >}}) |  | Represents a repeated `Value`. |
  




 <!-- end messages -->


<a name="google.protobuf.NullValue"></a>

### NullValue


| Name | Number | Description |
| ---- | ------ | ----------- |
| NULL_VALUE | 0 | Null value. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

