
---

---

## Package : `validate`



<a name="top"></a>

<a name="API Reference for validate.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## validate.proto


## Table of Contents
  - [AnyRules](#validate.AnyRules)
  - [BoolRules](#validate.BoolRules)
  - [BytesRules](#validate.BytesRules)
  - [DoubleRules](#validate.DoubleRules)
  - [DurationRules](#validate.DurationRules)
  - [EnumRules](#validate.EnumRules)
  - [FieldRules](#validate.FieldRules)
  - [Fixed32Rules](#validate.Fixed32Rules)
  - [Fixed64Rules](#validate.Fixed64Rules)
  - [FloatRules](#validate.FloatRules)
  - [Int32Rules](#validate.Int32Rules)
  - [Int64Rules](#validate.Int64Rules)
  - [MapRules](#validate.MapRules)
  - [MessageRules](#validate.MessageRules)
  - [RepeatedRules](#validate.RepeatedRules)
  - [SFixed32Rules](#validate.SFixed32Rules)
  - [SFixed64Rules](#validate.SFixed64Rules)
  - [SInt32Rules](#validate.SInt32Rules)
  - [SInt64Rules](#validate.SInt64Rules)
  - [StringRules](#validate.StringRules)
  - [TimestampRules](#validate.TimestampRules)
  - [UInt32Rules](#validate.UInt32Rules)
  - [UInt64Rules](#validate.UInt64Rules)

  - [KnownRegex](#validate.KnownRegex)

  - [File-level Extensions](#validate.proto-extensions)
  - [File-level Extensions](#validate.proto-extensions)
  - [File-level Extensions](#validate.proto-extensions)





<a name="validate.AnyRules"></a>

### AnyRules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| required | bool | optional | Required specifies that this field must be set |
  | in | []string | repeated | In specifies that this field's `type_url` must be equal to one of the specified values. |
  | notIn | []string | repeated | NotIn specifies that this field's `type_url` must not be equal to any of the specified values. |
  





<a name="validate.BoolRules"></a>

### BoolRules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | bool | optional | Const specifies that this field must be exactly the specified value |
  





<a name="validate.BytesRules"></a>

### BytesRules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | bytes | optional | Const specifies that this field must be exactly the specified value |
  | len | uint64 | optional | Len specifies that this field must be the specified number of bytes |
  | minLen | uint64 | optional | MinLen specifies that this field must be the specified number of bytes at a minimum |
  | maxLen | uint64 | optional | MaxLen specifies that this field must be the specified number of bytes at a maximum |
  | pattern | string | optional | Pattern specifes that this field must match against the specified regular expression (RE2 syntax). The included expression should elide any delimiters. |
  | prefix | bytes | optional | Prefix specifies that this field must have the specified bytes at the beginning of the string. |
  | suffix | bytes | optional | Suffix specifies that this field must have the specified bytes at the end of the string. |
  | contains | bytes | optional | Contains specifies that this field must have the specified bytes anywhere in the string. |
  | in | []bytes | repeated | In specifies that this field must be equal to one of the specified values |
  | notIn | []bytes | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
  | ip | bool | optional | Ip specifies that the field must be a valid IP (v4 or v6) address in byte format |
  | ipv4 | bool | optional | Ipv4 specifies that the field must be a valid IPv4 address in byte format |
  | ipv6 | bool | optional | Ipv6 specifies that the field must be a valid IPv6 address in byte format |
  





<a name="validate.DoubleRules"></a>

### DoubleRules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | double | optional | Const specifies that this field must be exactly the specified value |
  | lt | double | optional | Lt specifies that this field must be less than the specified value, exclusive |
  | lte | double | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
  | gt | double | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
  | gte | double | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
  | in | []double | repeated | In specifies that this field must be equal to one of the specified values |
  | notIn | []double | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
  





<a name="validate.DurationRules"></a>

### DurationRules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| required | bool | optional | Required specifies that this field must be set |
  | const | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) | optional | Const specifies that this field must be exactly the specified value |
  | lt | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) | optional | Lt specifies that this field must be less than the specified value, exclusive |
  | lte | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) | optional | Lt specifies that this field must be less than the specified value, inclusive |
  | gt | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) | optional | Gt specifies that this field must be greater than the specified value, exclusive |
  | gte | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) | optional | Gte specifies that this field must be greater than the specified value, inclusive |
  | in | [][google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) | repeated | In specifies that this field must be equal to one of the specified values |
  | notIn | [][google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
  





<a name="validate.EnumRules"></a>

### EnumRules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | int32 | optional | Const specifies that this field must be exactly the specified value |
  | definedOnly | bool | optional | DefinedOnly specifies that this field must be only one of the defined values for this enum, failing on any undefined value. |
  | in | []int32 | repeated | In specifies that this field must be equal to one of the specified values |
  | notIn | []int32 | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
  





<a name="validate.FieldRules"></a>

### FieldRules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [validate.MessageRules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.MessageRules" >}}) | optional |  |
  | float | [validate.FloatRules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.FloatRules" >}}) | optional | Scalar Field Types |
  | double | [validate.DoubleRules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.DoubleRules" >}}) | optional |  |
  | int32 | [validate.Int32Rules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.Int32Rules" >}}) | optional |  |
  | int64 | [validate.Int64Rules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.Int64Rules" >}}) | optional |  |
  | uint32 | [validate.UInt32Rules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.UInt32Rules" >}}) | optional |  |
  | uint64 | [validate.UInt64Rules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.UInt64Rules" >}}) | optional |  |
  | sint32 | [validate.SInt32Rules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.SInt32Rules" >}}) | optional |  |
  | sint64 | [validate.SInt64Rules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.SInt64Rules" >}}) | optional |  |
  | fixed32 | [validate.Fixed32Rules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.Fixed32Rules" >}}) | optional |  |
  | fixed64 | [validate.Fixed64Rules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.Fixed64Rules" >}}) | optional |  |
  | sfixed32 | [validate.SFixed32Rules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.SFixed32Rules" >}}) | optional |  |
  | sfixed64 | [validate.SFixed64Rules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.SFixed64Rules" >}}) | optional |  |
  | bool | [validate.BoolRules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.BoolRules" >}}) | optional |  |
  | string | [validate.StringRules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.StringRules" >}}) | optional |  |
  | bytes | [validate.BytesRules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.BytesRules" >}}) | optional |  |
  | enum | [validate.EnumRules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.EnumRules" >}}) | optional | Complex Field Types |
  | repeated | [validate.RepeatedRules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.RepeatedRules" >}}) | optional |  |
  | map | [validate.MapRules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.MapRules" >}}) | optional |  |
  | any | [validate.AnyRules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.AnyRules" >}}) | optional | Well-Known Field Types |
  | duration | [validate.DurationRules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.DurationRules" >}}) | optional |  |
  | timestamp | [validate.TimestampRules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.TimestampRules" >}}) | optional |  |
  





<a name="validate.Fixed32Rules"></a>

### Fixed32Rules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | fixed32 | optional | Const specifies that this field must be exactly the specified value |
  | lt | fixed32 | optional | Lt specifies that this field must be less than the specified value, exclusive |
  | lte | fixed32 | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
  | gt | fixed32 | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
  | gte | fixed32 | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
  | in | []fixed32 | repeated | In specifies that this field must be equal to one of the specified values |
  | notIn | []fixed32 | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
  





<a name="validate.Fixed64Rules"></a>

### Fixed64Rules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | fixed64 | optional | Const specifies that this field must be exactly the specified value |
  | lt | fixed64 | optional | Lt specifies that this field must be less than the specified value, exclusive |
  | lte | fixed64 | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
  | gt | fixed64 | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
  | gte | fixed64 | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
  | in | []fixed64 | repeated | In specifies that this field must be equal to one of the specified values |
  | notIn | []fixed64 | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
  





<a name="validate.FloatRules"></a>

### FloatRules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | float | optional | Const specifies that this field must be exactly the specified value |
  | lt | float | optional | Lt specifies that this field must be less than the specified value, exclusive |
  | lte | float | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
  | gt | float | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
  | gte | float | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
  | in | []float | repeated | In specifies that this field must be equal to one of the specified values |
  | notIn | []float | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
  





<a name="validate.Int32Rules"></a>

### Int32Rules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | int32 | optional | Const specifies that this field must be exactly the specified value |
  | lt | int32 | optional | Lt specifies that this field must be less than the specified value, exclusive |
  | lte | int32 | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
  | gt | int32 | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
  | gte | int32 | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
  | in | []int32 | repeated | In specifies that this field must be equal to one of the specified values |
  | notIn | []int32 | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
  





<a name="validate.Int64Rules"></a>

### Int64Rules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | int64 | optional | Const specifies that this field must be exactly the specified value |
  | lt | int64 | optional | Lt specifies that this field must be less than the specified value, exclusive |
  | lte | int64 | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
  | gt | int64 | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
  | gte | int64 | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
  | in | []int64 | repeated | In specifies that this field must be equal to one of the specified values |
  | notIn | []int64 | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
  





<a name="validate.MapRules"></a>

### MapRules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| minPairs | uint64 | optional | MinPairs specifies that this field must have the specified number of KVs at a minimum |
  | maxPairs | uint64 | optional | MaxPairs specifies that this field must have the specified number of KVs at a maximum |
  | noSparse | bool | optional | NoSparse specifies values in this field cannot be unset. This only applies to map's with message value types. |
  | keys | [validate.FieldRules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.FieldRules" >}}) | optional | Keys specifies the constraints to be applied to each key in the field. |
  | values | [validate.FieldRules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.FieldRules" >}}) | optional | Values specifies the constraints to be applied to the value of each key in the field. Message values will still have their validations evaluated unless skip is specified here. |
  





<a name="validate.MessageRules"></a>

### MessageRules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| skip | bool | optional | Skip specifies that the validation rules of this field should not be evaluated |
  | required | bool | optional | Required specifies that this field must be set |
  





<a name="validate.RepeatedRules"></a>

### RepeatedRules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| minItems | uint64 | optional | MinItems specifies that this field must have the specified number of items at a minimum |
  | maxItems | uint64 | optional | MaxItems specifies that this field must have the specified number of items at a maximum |
  | unique | bool | optional | Unique specifies that all elements in this field must be unique. This contraint is only applicable to scalar and enum types (messages are not supported). |
  | items | [validate.FieldRules]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.FieldRules" >}}) | optional | Items specifies the contraints to be applied to each item in the field. Repeated message fields will still execute validation against each item unless skip is specified here. |
  





<a name="validate.SFixed32Rules"></a>

### SFixed32Rules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | sfixed32 | optional | Const specifies that this field must be exactly the specified value |
  | lt | sfixed32 | optional | Lt specifies that this field must be less than the specified value, exclusive |
  | lte | sfixed32 | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
  | gt | sfixed32 | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
  | gte | sfixed32 | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
  | in | []sfixed32 | repeated | In specifies that this field must be equal to one of the specified values |
  | notIn | []sfixed32 | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
  





<a name="validate.SFixed64Rules"></a>

### SFixed64Rules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | sfixed64 | optional | Const specifies that this field must be exactly the specified value |
  | lt | sfixed64 | optional | Lt specifies that this field must be less than the specified value, exclusive |
  | lte | sfixed64 | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
  | gt | sfixed64 | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
  | gte | sfixed64 | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
  | in | []sfixed64 | repeated | In specifies that this field must be equal to one of the specified values |
  | notIn | []sfixed64 | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
  





<a name="validate.SInt32Rules"></a>

### SInt32Rules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | sint32 | optional | Const specifies that this field must be exactly the specified value |
  | lt | sint32 | optional | Lt specifies that this field must be less than the specified value, exclusive |
  | lte | sint32 | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
  | gt | sint32 | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
  | gte | sint32 | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
  | in | []sint32 | repeated | In specifies that this field must be equal to one of the specified values |
  | notIn | []sint32 | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
  





<a name="validate.SInt64Rules"></a>

### SInt64Rules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | sint64 | optional | Const specifies that this field must be exactly the specified value |
  | lt | sint64 | optional | Lt specifies that this field must be less than the specified value, exclusive |
  | lte | sint64 | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
  | gt | sint64 | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
  | gte | sint64 | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
  | in | []sint64 | repeated | In specifies that this field must be equal to one of the specified values |
  | notIn | []sint64 | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
  





<a name="validate.StringRules"></a>

### StringRules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | string | optional | Const specifies that this field must be exactly the specified value |
  | len | uint64 | optional | Len specifies that this field must be the specified number of characters (Unicode code points). Note that the number of characters may differ from the number of bytes in the string. |
  | minLen | uint64 | optional | MinLen specifies that this field must be the specified number of characters (Unicode code points) at a minimum. Note that the number of characters may differ from the number of bytes in the string. |
  | maxLen | uint64 | optional | MaxLen specifies that this field must be the specified number of characters (Unicode code points) at a maximum. Note that the number of characters may differ from the number of bytes in the string. |
  | lenBytes | uint64 | optional | LenBytes specifies that this field must be the specified number of bytes at a minimum |
  | minBytes | uint64 | optional | MinBytes specifies that this field must be the specified number of bytes at a minimum |
  | maxBytes | uint64 | optional | MaxBytes specifies that this field must be the specified number of bytes at a maximum |
  | pattern | string | optional | Pattern specifes that this field must match against the specified regular expression (RE2 syntax). The included expression should elide any delimiters. |
  | prefix | string | optional | Prefix specifies that this field must have the specified substring at the beginning of the string. |
  | suffix | string | optional | Suffix specifies that this field must have the specified substring at the end of the string. |
  | contains | string | optional | Contains specifies that this field must have the specified substring anywhere in the string. |
  | notContains | string | optional | NotContains specifies that this field cannot have the specified substring anywhere in the string. |
  | in | []string | repeated | In specifies that this field must be equal to one of the specified values |
  | notIn | []string | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
  | email | bool | optional | Email specifies that the field must be a valid email address as defined by RFC 5322 |
  | hostname | bool | optional | Hostname specifies that the field must be a valid hostname as defined by RFC 1034. This constraint does not support internationalized domain names (IDNs). |
  | ip | bool | optional | Ip specifies that the field must be a valid IP (v4 or v6) address. Valid IPv6 addresses should not include surrounding square brackets. |
  | ipv4 | bool | optional | Ipv4 specifies that the field must be a valid IPv4 address. |
  | ipv6 | bool | optional | Ipv6 specifies that the field must be a valid IPv6 address. Valid IPv6 addresses should not include surrounding square brackets. |
  | uri | bool | optional | Uri specifies that the field must be a valid, absolute URI as defined by RFC 3986 |
  | uriRef | bool | optional | UriRef specifies that the field must be a valid URI as defined by RFC 3986 and may be relative or absolute. |
  | address | bool | optional | Address specifies that the field must be either a valid hostname as defined by RFC 1034 (which does not support internationalized domain names or IDNs), or it can be a valid IP (v4 or v6). |
  | uuid | bool | optional | Uuid specifies that the field must be a valid UUID as defined by RFC 4122 |
  | wellKnownRegex | [validate.KnownRegex]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.protoc-gen-validate.validate.validate#validate.KnownRegex" >}}) | optional | WellKnownRegex specifies a common well known pattern defined as a regex. |
  | strict | bool | optional | This applies to regexes HTTP_HEADER_NAME and HTTP_HEADER_VALUE to enable strict header validation. By default, this is true, and HTTP header validations are RFC-compliant. Setting to false will enable a looser validations that only disallows \r\n\0 characters, which can be used to bypass header matching rules. Default: true |
  





<a name="validate.TimestampRules"></a>

### TimestampRules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| required | bool | optional | Required specifies that this field must be set |
  | const | [google.protobuf.Timestamp]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.timestamp#google.protobuf.Timestamp" >}}) | optional | Const specifies that this field must be exactly the specified value |
  | lt | [google.protobuf.Timestamp]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.timestamp#google.protobuf.Timestamp" >}}) | optional | Lt specifies that this field must be less than the specified value, exclusive |
  | lte | [google.protobuf.Timestamp]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.timestamp#google.protobuf.Timestamp" >}}) | optional | Lte specifies that this field must be less than the specified value, inclusive |
  | gt | [google.protobuf.Timestamp]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.timestamp#google.protobuf.Timestamp" >}}) | optional | Gt specifies that this field must be greater than the specified value, exclusive |
  | gte | [google.protobuf.Timestamp]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.timestamp#google.protobuf.Timestamp" >}}) | optional | Gte specifies that this field must be greater than the specified value, inclusive |
  | ltNow | bool | optional | LtNow specifies that this must be less than the current time. LtNow can only be used with the Within rule. |
  | gtNow | bool | optional | GtNow specifies that this must be greater than the current time. GtNow can only be used with the Within rule. |
  | within | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) | optional | Within specifies that this field must be within this duration of the current time. This constraint can be used alone or with the LtNow and GtNow rules. |
  





<a name="validate.UInt32Rules"></a>

### UInt32Rules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | uint32 | optional | Const specifies that this field must be exactly the specified value |
  | lt | uint32 | optional | Lt specifies that this field must be less than the specified value, exclusive |
  | lte | uint32 | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
  | gt | uint32 | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
  | gte | uint32 | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
  | in | []uint32 | repeated | In specifies that this field must be equal to one of the specified values |
  | notIn | []uint32 | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
  





<a name="validate.UInt64Rules"></a>

### UInt64Rules



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | uint64 | optional | Const specifies that this field must be exactly the specified value |
  | lt | uint64 | optional | Lt specifies that this field must be less than the specified value, exclusive |
  | lte | uint64 | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
  | gt | uint64 | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
  | gte | uint64 | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
  | in | []uint64 | repeated | In specifies that this field must be equal to one of the specified values |
  | notIn | []uint64 | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
  




 <!-- end messages -->


<a name="validate.KnownRegex"></a>

### KnownRegex


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN | 0 |  |
| HTTP_HEADER_NAME | 1 | HTTP header name as defined by RFC 7230. |
| HTTP_HEADER_VALUE | 2 | HTTP header value as defined by RFC 7230. |


 <!-- end enums -->


<a name="validate.proto-extensions"></a>

### File-level Extensions
| Extension | Type | Base | Number | Description |
| --------- | ---- | ---- | ------ | ----------- |
| rules | FieldRules | .google.protobuf.FieldOptions | 1071 |  |
| disabled | bool | .google.protobuf.MessageOptions | 1071 |  |
| required | bool | .google.protobuf.OneofOptions | 1071 |  |

 <!-- end HasExtensions -->

 <!-- end services -->

