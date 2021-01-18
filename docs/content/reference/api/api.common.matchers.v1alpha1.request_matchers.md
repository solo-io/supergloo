
---

title: "request_matchers.proto"

---

## Package : `common.matchers.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for request_matchers.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## request_matchers.proto


## Table of Contents
  - [HeaderMatcher](#common.matchers.mesh.gloo.solo.io.HeaderMatcher)
  - [StatusCodeMatcher](#common.matchers.mesh.gloo.solo.io.StatusCodeMatcher)
  - [StatusCodeMatcher.Range](#common.matchers.mesh.gloo.solo.io.StatusCodeMatcher.Range)







<a name="common.matchers.mesh.gloo.solo.io.HeaderMatcher"></a>

### HeaderMatcher
Describes a matcher against HTTP request headers.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Specifies the name of the header in the request. |
  | value | string |  | Specifies the value of the header. If the value is absent a request that has the name header will match, regardless of the header’s value. |
  | regex | bool |  | Specifies whether the header value should be treated as regex or not. |
  | invertMatch | bool |  | If set to true, the result of the match will be inverted. Defaults to false.<br>Examples: name=foo, invert_match=true: matches if no header named `foo` is present name=foo, value=bar, invert_match=true: matches if no header named `foo` with value `bar` is present name=foo, value=``\d{3}``, regex=true, invert_match=true: matches if no header named `foo` with a value consisting of three integers is present |
  





<a name="common.matchers.mesh.gloo.solo.io.StatusCodeMatcher"></a>

### StatusCodeMatcher
Describes a matchers against HTTP response status codes.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| exact | uint32 |  | Matches the status code exactly. |
  | range | [common.matchers.mesh.gloo.solo.io.StatusCodeMatcher.Range]({{< versioned_link_path fromRoot="/reference/api/api.common.matchers.v1alpha1.request_matchers#common.matchers.mesh.gloo.solo.io.StatusCodeMatcher.Range" >}}) |  |  |
  





<a name="common.matchers.mesh.gloo.solo.io.StatusCodeMatcher.Range"></a>

### StatusCodeMatcher.Range
Describes a range matcher against HTTP response status codes. Boundaries are inclusive.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | uint32 |  | The inclusive boundary value. |
  | isLte | bool |  | If true, treat the value as an inclusive upper bound. Otherwise, as an inclusive lower bound. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

