
---
title: "struct.proto"
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
`ListValue` is a wrapper around a repeated field of values.<br>The JSON representation for `ListValue` is JSON array.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| values | [][google.protobuf.Value]({{< ref "struct.md#google.protobuf.Value" >}}) | repeated | Repeated field of dynamically typed values. |
  





<a name="google.protobuf.Struct"></a>

### Struct
`Struct` represents a structured data value, consisting of fields which map to dynamically typed values. In some languages, `Struct` might be supported by a native representation. For example, in scripting languages like JS a struct is represented as an object. The details of that representation are described together with the proto support for the language.<br>The JSON representation for `Struct` is JSON object.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fields | [][google.protobuf.Struct.FieldsEntry]({{< ref "struct.md#google.protobuf.Struct.FieldsEntry" >}}) | repeated | Unordered map of dynamically typed values. |
  





<a name="google.protobuf.Struct.FieldsEntry"></a>

### Struct.FieldsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [google.protobuf.Value]({{< ref "struct.md#google.protobuf.Value" >}}) |  |  |
  





<a name="google.protobuf.Value"></a>

### Value
`Value` represents a dynamically typed value which can be either null, a number, a string, a boolean, a recursive struct value, or a list of values. A producer of value is expected to set one of that variants, absence of any variant indicates an error.<br>The JSON representation for `Value` is JSON value.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nullValue | [google.protobuf.NullValue]({{< ref "struct.md#google.protobuf.NullValue" >}}) |  | Represents a null value. |
  | numberValue | double |  | Represents a double value. |
  | stringValue | string |  | Represents a string value. |
  | boolValue | bool |  | Represents a boolean value. |
  | structValue | [google.protobuf.Struct]({{< ref "struct.md#google.protobuf.Struct" >}}) |  | Represents a structured value. |
  | listValue | [google.protobuf.ListValue]({{< ref "struct.md#google.protobuf.ListValue" >}}) |  | Represents a repeated `Value`. |
  




 <!-- end messages -->


<a name="google.protobuf.NullValue"></a>

### NullValue
`NullValue` is a singleton enumeration to represent the null value for the `Value` type union.<br> The JSON representation for `NullValue` is JSON `null`.

| Name | Number | Description |
| ---- | ------ | ----------- |
| NULL_VALUE | 0 | Null value. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

