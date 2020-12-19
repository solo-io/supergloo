
---
title: "type.proto"
---

## Package : `google.protobuf`



<a name="top"></a>

<a name="API Reference for type.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## type.proto


## Table of Contents
  - [Enum](#google.protobuf.Enum)
  - [EnumValue](#google.protobuf.EnumValue)
  - [Field](#google.protobuf.Field)
  - [Option](#google.protobuf.Option)
  - [Type](#google.protobuf.Type)

  - [Field.Cardinality](#google.protobuf.Field.Cardinality)
  - [Field.Kind](#google.protobuf.Field.Kind)
  - [Syntax](#google.protobuf.Syntax)






<a name="google.protobuf.Enum"></a>

### Enum
Enum type definition.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Enum type name. |
  | enumvalue | [][google.protobuf.EnumValue]({{< ref "type.md#google.protobuf.EnumValue" >}}) | repeated | Enum value definitions. |
  | options | [][google.protobuf.Option]({{< ref "type.md#google.protobuf.Option" >}}) | repeated | Protocol buffer options. |
  | sourceContext | [google.protobuf.SourceContext]({{< ref "source_context.md#google.protobuf.SourceContext" >}}) |  | The source context. |
  | syntax | [google.protobuf.Syntax]({{< ref "type.md#google.protobuf.Syntax" >}}) |  | The source syntax. |
  





<a name="google.protobuf.EnumValue"></a>

### EnumValue
Enum value definition.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Enum value name. |
  | number | int32 |  | Enum value number. |
  | options | [][google.protobuf.Option]({{< ref "type.md#google.protobuf.Option" >}}) | repeated | Protocol buffer options. |
  





<a name="google.protobuf.Field"></a>

### Field
A single field of a message type.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [google.protobuf.Field.Kind]({{< ref "type.md#google.protobuf.Field.Kind" >}}) |  | The field type. |
  | cardinality | [google.protobuf.Field.Cardinality]({{< ref "type.md#google.protobuf.Field.Cardinality" >}}) |  | The field cardinality. |
  | number | int32 |  | The field number. |
  | name | string |  | The field name. |
  | typeUrl | string |  | The field type URL, without the scheme, for message or enumeration types. Example: `"type.googleapis.com/google.protobuf.Timestamp"`. |
  | oneofIndex | int32 |  | The index of the field type in `Type.oneofs`, for message or enumeration types. The first type has index 1; zero means the type is not in the list. |
  | packed | bool |  | Whether to use alternative packed wire representation. |
  | options | [][google.protobuf.Option]({{< ref "type.md#google.protobuf.Option" >}}) | repeated | The protocol buffer options. |
  | jsonName | string |  | The field JSON name. |
  | defaultValue | string |  | The string value of the default value of this field. Proto2 syntax only. |
  





<a name="google.protobuf.Option"></a>

### Option
A protocol buffer option, which can be attached to a message, field, enumeration, etc.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The option's name. For protobuf built-in options (options defined in descriptor.proto), this is the short name. For example, `"map_entry"`. For custom options, it should be the fully-qualified name. For example, `"google.api.http"`. |
  | value | [google.protobuf.Any]({{< ref "any.md#google.protobuf.Any" >}}) |  | The option's value packed in an Any message. If the value is a primitive, the corresponding wrapper type defined in google/protobuf/wrappers.proto should be used. If the value is an enum, it should be stored as an int32 value using the google.protobuf.Int32Value type. |
  





<a name="google.protobuf.Type"></a>

### Type
A protocol buffer message type.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The fully qualified message name. |
  | fields | [][google.protobuf.Field]({{< ref "type.md#google.protobuf.Field" >}}) | repeated | The list of fields. |
  | oneofs | []string | repeated | The list of types appearing in `oneof` definitions in this type. |
  | options | [][google.protobuf.Option]({{< ref "type.md#google.protobuf.Option" >}}) | repeated | The protocol buffer options. |
  | sourceContext | [google.protobuf.SourceContext]({{< ref "source_context.md#google.protobuf.SourceContext" >}}) |  | The source context. |
  | syntax | [google.protobuf.Syntax]({{< ref "type.md#google.protobuf.Syntax" >}}) |  | The source syntax. |
  




 <!-- end messages -->


<a name="google.protobuf.Field.Cardinality"></a>

### Field.Cardinality
Whether a field is optional, required, or repeated.

| Name | Number | Description |
| ---- | ------ | ----------- |
| CARDINALITY_UNKNOWN | 0 | For fields with unknown cardinality. |
| CARDINALITY_OPTIONAL | 1 | For optional fields. |
| CARDINALITY_REQUIRED | 2 | For required fields. Proto2 syntax only. |
| CARDINALITY_REPEATED | 3 | For repeated fields. |



<a name="google.protobuf.Field.Kind"></a>

### Field.Kind
Basic field types.

| Name | Number | Description |
| ---- | ------ | ----------- |
| TYPE_UNKNOWN | 0 | Field type unknown. |
| TYPE_DOUBLE | 1 | Field type double. |
| TYPE_FLOAT | 2 | Field type float. |
| TYPE_INT64 | 3 | Field type int64. |
| TYPE_UINT64 | 4 | Field type uint64. |
| TYPE_INT32 | 5 | Field type int32. |
| TYPE_FIXED64 | 6 | Field type fixed64. |
| TYPE_FIXED32 | 7 | Field type fixed32. |
| TYPE_BOOL | 8 | Field type bool. |
| TYPE_STRING | 9 | Field type string. |
| TYPE_GROUP | 10 | Field type group. Proto2 syntax only, and deprecated. |
| TYPE_MESSAGE | 11 | Field type message. |
| TYPE_BYTES | 12 | Field type bytes. |
| TYPE_UINT32 | 13 | Field type uint32. |
| TYPE_ENUM | 14 | Field type enum. |
| TYPE_SFIXED32 | 15 | Field type sfixed32. |
| TYPE_SFIXED64 | 16 | Field type sfixed64. |
| TYPE_SINT32 | 17 | Field type sint32. |
| TYPE_SINT64 | 18 | Field type sint64. |



<a name="google.protobuf.Syntax"></a>

### Syntax
The syntax in which a protocol buffer element is defined.

| Name | Number | Description |
| ---- | ------ | ----------- |
| SYNTAX_PROTO2 | 0 | Syntax `proto2`. |
| SYNTAX_PROTO3 | 1 | Syntax `proto3`. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

