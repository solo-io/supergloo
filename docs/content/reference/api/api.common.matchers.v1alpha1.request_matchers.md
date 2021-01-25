
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

  - [StatusCodeMatcher.Comparator](#common.matchers.mesh.gloo.solo.io.StatusCodeMatcher.Comparator)






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
Describes a matcher against HTTP response status codes.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | uint32 |  | the status code value to match against |
  | comparator | [common.matchers.mesh.gloo.solo.io.StatusCodeMatcher.Comparator]({{< versioned_link_path fromRoot="/reference/api/api.common.matchers.v1alpha1.request_matchers#common.matchers.mesh.gloo.solo.io.StatusCodeMatcher.Comparator" >}}) |  | the comparison type used for matching |
  




 <!-- end messages -->


<a name="common.matchers.mesh.gloo.solo.io.StatusCodeMatcher.Comparator"></a>

### StatusCodeMatcher.Comparator


| Name | Number | Description |
| ---- | ------ | ----------- |
| EQ | 0 | default, strict equality |
| GE | 1 | greater than or equal to |
| LE | 2 | less than or equal to |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

