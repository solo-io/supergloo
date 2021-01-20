
---

---

## Package : `envoy.type.matcher.v3`



<a name="top"></a>

<a name="API Reference for value.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## value.proto


## Table of Contents
  - [ListMatcher](#envoy.type.matcher.v3.ListMatcher)
  - [ValueMatcher](#envoy.type.matcher.v3.ValueMatcher)
  - [ValueMatcher.NullMatch](#envoy.type.matcher.v3.ValueMatcher.NullMatch)







<a name="envoy.type.matcher.v3.ListMatcher"></a>

### ListMatcher



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| oneOf | [envoy.type.matcher.v3.ValueMatcher]({{< versioned_link_path fromRoot="/reference/api/envoy.type.matcher.v3.value#envoy.type.matcher.v3.ValueMatcher" >}}) |  | If specified, at least one of the values in the list must match the value specified. |
  





<a name="envoy.type.matcher.v3.ValueMatcher"></a>

### ValueMatcher



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nullMatch | [envoy.type.matcher.v3.ValueMatcher.NullMatch]({{< versioned_link_path fromRoot="/reference/api/envoy.type.matcher.v3.value#envoy.type.matcher.v3.ValueMatcher.NullMatch" >}}) |  | If specified, a match occurs if and only if the target value is a NullValue. |
  | doubleMatch | [envoy.type.matcher.v3.DoubleMatcher]({{< versioned_link_path fromRoot="/reference/api/envoy.type.matcher.v3.number#envoy.type.matcher.v3.DoubleMatcher" >}}) |  | If specified, a match occurs if and only if the target value is a double value and is matched to this field. |
  | stringMatch | [envoy.type.matcher.v3.StringMatcher]({{< versioned_link_path fromRoot="/reference/api/envoy.type.matcher.v3.string#envoy.type.matcher.v3.StringMatcher" >}}) |  | If specified, a match occurs if and only if the target value is a string value and is matched to this field. |
  | boolMatch | bool |  | If specified, a match occurs if and only if the target value is a bool value and is equal to this field. |
  | presentMatch | bool |  | If specified, value match will be performed based on whether the path is referring to a valid primitive value in the metadata. If the path is referring to a non-primitive value, the result is always not matched. |
  | listMatch | [envoy.type.matcher.v3.ListMatcher]({{< versioned_link_path fromRoot="/reference/api/envoy.type.matcher.v3.value#envoy.type.matcher.v3.ListMatcher" >}}) |  | If specified, a match occurs if and only if the target value is a list value and is matched to this field. |
  





<a name="envoy.type.matcher.v3.ValueMatcher.NullMatch"></a>

### ValueMatcher.NullMatch






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

