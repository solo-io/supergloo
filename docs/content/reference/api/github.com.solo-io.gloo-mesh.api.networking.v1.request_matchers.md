
---

title: "request_matchers.proto"

---

## Package : `networking.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for request_matchers.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## request_matchers.proto


## Table of Contents
  - [HeaderMatcher](#networking.mesh.gloo.solo.io.HeaderMatcher)
  - [HttpMatcher](#networking.mesh.gloo.solo.io.HttpMatcher)
  - [HttpMatcher.QueryParameterMatcher](#networking.mesh.gloo.solo.io.HttpMatcher.QueryParameterMatcher)
  - [StatusCodeMatcher](#networking.mesh.gloo.solo.io.StatusCodeMatcher)

  - [StatusCodeMatcher.Comparator](#networking.mesh.gloo.solo.io.StatusCodeMatcher.Comparator)






<a name="networking.mesh.gloo.solo.io.HeaderMatcher"></a>

### HeaderMatcher
Describes a matcher against HTTP request headers.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Specify the name of the header in the request. |
  | value | string |  | Specify the value of the header. If the value is absent a request that has the name header will match, regardless of the header’s value. |
  | regex | bool |  | Specify whether the header value should be treated as regex. |
  | invertMatch | bool |  | If set to true, the result of the match will be inverted. Defaults to false.<br>Examples:<br>- name=foo, invert_match=true: matches if no header named `foo` is present - name=foo, value=bar, invert_match=true: matches if no header named `foo` with value `bar` is present - name=foo, value=``\d{3}``, regex=true, invert_match=true: matches if no header named `foo` with a value consisting of three integers is present. |
  





<a name="networking.mesh.gloo.solo.io.HttpMatcher"></a>

### HttpMatcher
Specify HTTP request level match criteria. All specified conditions must be satisfied for a match to occur.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The name assigned to a match. The match's name will be concatenated with the parent route's name and will be logged in the access logs for requests matching this route. |
  | uri | [networking.mesh.gloo.solo.io.StringMatch]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.string_match#networking.mesh.gloo.solo.io.StringMatch" >}}) |  | Specify match criteria against the targeted path. |
  | headers | [][networking.mesh.gloo.solo.io.HeaderMatcher]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.request_matchers#networking.mesh.gloo.solo.io.HeaderMatcher" >}}) | repeated | Specify a set of headers which requests must match in entirety (all headers must match). |
  | queryParameters | [][networking.mesh.gloo.solo.io.HttpMatcher.QueryParameterMatcher]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.request_matchers#networking.mesh.gloo.solo.io.HttpMatcher.QueryParameterMatcher" >}}) | repeated | Specify a set of URL query parameters which requests must match in entirety (all query params must match). |
  | method | string |  | Specify an HTTP method to match against. |
  





<a name="networking.mesh.gloo.solo.io.HttpMatcher.QueryParameterMatcher"></a>

### HttpMatcher.QueryParameterMatcher
Specify match criteria against the target URL's query parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Specify the name of a key that must be present in the requested path's query string. |
  | value | string |  | Specify the value of the query parameter keyed on `name`. |
  | regex | bool |  | If true, treat `value` as a regular expression. |
  





<a name="networking.mesh.gloo.solo.io.StatusCodeMatcher"></a>

### StatusCodeMatcher
Describes a matcher against HTTP response status codes.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | uint32 |  | The status code value to match against. |
  | comparator | [networking.mesh.gloo.solo.io.StatusCodeMatcher.Comparator]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.request_matchers#networking.mesh.gloo.solo.io.StatusCodeMatcher.Comparator" >}}) |  | The comparison type used for matching. |
  




 <!-- end messages -->


<a name="networking.mesh.gloo.solo.io.StatusCodeMatcher.Comparator"></a>

### StatusCodeMatcher.Comparator


| Name | Number | Description |
| ---- | ------ | ----------- |
| EQ | 0 | Strict equality. |
| GE | 1 | Greater than or equal to. |
| LE | 2 | Less than or equal to. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

