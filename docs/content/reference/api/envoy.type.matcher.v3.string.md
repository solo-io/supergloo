
---

---

## Package : `envoy.type.matcher.v3`



<a name="top"></a>

<a name="API Reference for string.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## string.proto


## Table of Contents
  - [ListStringMatcher](#envoy.type.matcher.v3.ListStringMatcher)
  - [StringMatcher](#envoy.type.matcher.v3.StringMatcher)







<a name="envoy.type.matcher.v3.ListStringMatcher"></a>

### ListStringMatcher



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| patterns | [][envoy.type.matcher.v3.StringMatcher]({{< versioned_link_path fromRoot="/reference/api/envoy.type.matcher.v3.string#envoy.type.matcher.v3.StringMatcher" >}}) | repeated |  |
  





<a name="envoy.type.matcher.v3.StringMatcher"></a>

### StringMatcher



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| exact | string |  | The input string must match exactly the string specified here.<br>Examples:<br>* *abc* only matches the value *abc*. |
  | prefix | string |  | The input string must have the prefix specified here. Note: empty prefix is not allowed, please use regex instead.<br>Examples:<br>* *abc* matches the value *abc.xyz* |
  | suffix | string |  | The input string must have the suffix specified here. Note: empty prefix is not allowed, please use regex instead.<br>Examples:<br>* *abc* matches the value *xyz.abc* |
  | safeRegex | [envoy.type.matcher.v3.RegexMatcher]({{< versioned_link_path fromRoot="/reference/api/envoy.type.matcher.v3.regex#envoy.type.matcher.v3.RegexMatcher" >}}) |  | The input string must match the regular expression specified here. |
  | contains | string |  | The input string must have the substring specified here. Note: empty contains match is not allowed, please use regex instead.<br>Examples:<br>* *abc* matches the value *xyz.abc.def* |
  | ignoreCase | bool |  | If true, indicates the exact/prefix/suffix matching should be case insensitive. This has no effect for the safe_regex match. For example, the matcher *data* will match both input string *Data* and *data* if set to true. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

