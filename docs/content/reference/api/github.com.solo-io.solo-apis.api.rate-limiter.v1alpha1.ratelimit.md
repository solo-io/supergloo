
---

---

## Package : `ratelimit.api.solo.io`



<a name="top"></a>

<a name="API Reference for ratelimit.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## ratelimit.proto


## Table of Contents
  - [Action](#ratelimit.api.solo.io.Action)
  - [Action.DestinationCluster](#ratelimit.api.solo.io.Action.DestinationCluster)
  - [Action.GenericKey](#ratelimit.api.solo.io.Action.GenericKey)
  - [Action.HeaderValueMatch](#ratelimit.api.solo.io.Action.HeaderValueMatch)
  - [Action.HeaderValueMatch.HeaderMatcher](#ratelimit.api.solo.io.Action.HeaderValueMatch.HeaderMatcher)
  - [Action.HeaderValueMatch.HeaderMatcher.Int64Range](#ratelimit.api.solo.io.Action.HeaderValueMatch.HeaderMatcher.Int64Range)
  - [Action.MetaData](#ratelimit.api.solo.io.Action.MetaData)
  - [Action.MetaData.MetadataKey](#ratelimit.api.solo.io.Action.MetaData.MetadataKey)
  - [Action.MetaData.MetadataKey.PathSegment](#ratelimit.api.solo.io.Action.MetaData.MetadataKey.PathSegment)
  - [Action.RemoteAddress](#ratelimit.api.solo.io.Action.RemoteAddress)
  - [Action.RequestHeaders](#ratelimit.api.solo.io.Action.RequestHeaders)
  - [Action.SourceCluster](#ratelimit.api.solo.io.Action.SourceCluster)
  - [Descriptor](#ratelimit.api.solo.io.Descriptor)
  - [RateLimit](#ratelimit.api.solo.io.RateLimit)
  - [RateLimitActions](#ratelimit.api.solo.io.RateLimitActions)
  - [RateLimitConfigSpec](#ratelimit.api.solo.io.RateLimitConfigSpec)
  - [RateLimitConfigSpec.Raw](#ratelimit.api.solo.io.RateLimitConfigSpec.Raw)
  - [RateLimitConfigStatus](#ratelimit.api.solo.io.RateLimitConfigStatus)
  - [SetDescriptor](#ratelimit.api.solo.io.SetDescriptor)
  - [SimpleDescriptor](#ratelimit.api.solo.io.SimpleDescriptor)

  - [Action.MetaData.Source](#ratelimit.api.solo.io.Action.MetaData.Source)
  - [RateLimit.Unit](#ratelimit.api.solo.io.RateLimit.Unit)
  - [RateLimitConfigStatus.State](#ratelimit.api.solo.io.RateLimitConfigStatus.State)






<a name="ratelimit.api.solo.io.Action"></a>

### Action



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sourceCluster | [ratelimit.api.solo.io.Action.SourceCluster]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.Action.SourceCluster" >}}) |  | Rate limit on source cluster. |
  | destinationCluster | [ratelimit.api.solo.io.Action.DestinationCluster]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.Action.DestinationCluster" >}}) |  | Rate limit on destination cluster. |
  | requestHeaders | [ratelimit.api.solo.io.Action.RequestHeaders]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.Action.RequestHeaders" >}}) |  | Rate limit on request headers. |
  | remoteAddress | [ratelimit.api.solo.io.Action.RemoteAddress]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.Action.RemoteAddress" >}}) |  | Rate limit on remote address. |
  | genericKey | [ratelimit.api.solo.io.Action.GenericKey]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.Action.GenericKey" >}}) |  | Rate limit on a generic key. |
  | headerValueMatch | [ratelimit.api.solo.io.Action.HeaderValueMatch]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.Action.HeaderValueMatch" >}}) |  | Rate limit on the existence of request headers. |
  | metadata | [ratelimit.api.solo.io.Action.MetaData]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.Action.MetaData" >}}) |  | Rate limit on metadata. |
  





<a name="ratelimit.api.solo.io.Action.DestinationCluster"></a>

### Action.DestinationCluster







<a name="ratelimit.api.solo.io.Action.GenericKey"></a>

### Action.GenericKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| descriptorValue | string |  | The value to use in the descriptor entry.<br>[(validate.rules).string.min_bytes = 1]; |
  





<a name="ratelimit.api.solo.io.Action.HeaderValueMatch"></a>

### Action.HeaderValueMatch



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| descriptorValue | string |  | The value to use in the descriptor entry.<br>[(validate.rules).string.min_bytes = 1]; |
  | expectMatch | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | If set to true, the action will append a descriptor entry when the request matches the headers. If set to false, the action will append a descriptor entry when the request does not match the headers. The default value is true. |
  | headers | [][ratelimit.api.solo.io.Action.HeaderValueMatch.HeaderMatcher]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.Action.HeaderValueMatch.HeaderMatcher" >}}) | repeated | Specifies a set of headers that the rate limit action should match on. The action will check the request’s headers against all the specified headers in the config. A match will happen if all the headers in the config are present in the request with the same values (or based on presence if the value field is not in the config).<br>[(validate.rules).repeated .min_items = 1]; |
  





<a name="ratelimit.api.solo.io.Action.HeaderValueMatch.HeaderMatcher"></a>

### Action.HeaderValueMatch.HeaderMatcher



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Specifies the name of the header in the request.<br>[(validate.rules).string.min_bytes = 1]; |
  | exactMatch | string |  | If specified, header match will be performed based on the value of the header. |
  | regexMatch | string |  | If specified, this regex string is a regular expression rule which implies the entire request header value must match the regex. The rule will not match if only a subsequence of the request header value matches the regex. The regex grammar used in the value field is defined `(here)[https://en.cppreference.com/w/cpp/regex/ecmascript]`.<br>Examples:<br>* The regex *\d{3}* matches the value *123* * The regex *\d{3}* does not match the value *1234* * The regex *\d{3}* does not match the value *123.456*<br>[(validate.rules).string.max_bytes = 1024]; |
  | rangeMatch | [ratelimit.api.solo.io.Action.HeaderValueMatch.HeaderMatcher.Int64Range]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.Action.HeaderValueMatch.HeaderMatcher.Int64Range" >}}) |  | If specified, header match will be performed based on range. The rule will match if the request header value is within this range. The entire request header value must represent an integer in base 10 notation: consisting of an optional plus or minus sign followed by a sequence of digits. The rule will not match if the header value does not represent an integer. Match will fail for empty values, floating point numbers or if only a subsequence of the header value is an integer.<br>Examples:<br>* For range [-10,0), route will match for header value -1, but not for 0, "somestring", 10.9,   "-1somestring" |
  | presentMatch | bool |  | If specified, header match will be performed based on whether the header is in the request. |
  | prefixMatch | string |  | If specified, header match will be performed based on the prefix of the header value. Note: empty prefix is not allowed, please use present_match instead.<br>Examples:<br>* The prefix *abcd* matches the value *abcdxyz*, but not for *abcxyz*.<br>[(validate.rules).string.min_bytes = 1]; |
  | suffixMatch | string |  | If specified, header match will be performed based on the suffix of the header value. Note: empty suffix is not allowed, please use present_match instead.<br>Examples:<br>* The suffix *abcd* matches the value *xyzabcd*, but not for *xyzbcd*.<br>[(validate.rules).string.min_bytes = 1]; |
  | invertMatch | bool |  | If specified, the match result will be inverted before checking. Defaults to false.<br>Examples:<br>* The regex *\d{3}* does not match the value *1234*, so it will match when inverted. * The range [-10,0) will match the value -1, so it will not match when inverted. |
  





<a name="ratelimit.api.solo.io.Action.HeaderValueMatch.HeaderMatcher.Int64Range"></a>

### Action.HeaderValueMatch.HeaderMatcher.Int64Range



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| start | int64 |  | start of the range (inclusive) |
  | end | int64 |  | end of the range (exclusive) |
  





<a name="ratelimit.api.solo.io.Action.MetaData"></a>

### Action.MetaData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| descriptorKey | string |  | Required. The key to use in the descriptor entry.<br>[(validate.rules).string = {min_len: 1}]; |
  | metadataKey | [ratelimit.api.solo.io.Action.MetaData.MetadataKey]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.Action.MetaData.MetadataKey" >}}) |  | Required. Metadata struct that defines the key and path to retrieve the string value. A match will only happen if the value in the metadata is of type string.<br>[(validate.rules).message = {required: true}]; |
  | defaultValue | string |  | An optional value to use if *metadata_key* is empty. If not set and no value is present under the metadata_key then no descriptor is generated. |
  | source | [ratelimit.api.solo.io.Action.MetaData.Source]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.Action.MetaData.Source" >}}) |  | Source of metadata<br>[(validate.rules).enum = {defined_only: true}]; |
  





<a name="ratelimit.api.solo.io.Action.MetaData.MetadataKey"></a>

### Action.MetaData.MetadataKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  | Required. The key name of Metadata to retrieve the Struct from the metadata. Typically, it represents a builtin subsystem or custom extension.<br>[(validate.rules).string = {min_len: 1}]; |
  | path | [][ratelimit.api.solo.io.Action.MetaData.MetadataKey.PathSegment]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.Action.MetaData.MetadataKey.PathSegment" >}}) | repeated | Must have at least one element. The path to retrieve the Value from the Struct. It can be a prefix or a full path, e.g. ``[prop, xyz]`` for a struct or ``[prop, foo]`` for a string in the example, which depends on the particular scenario.<br>Note: Due to that only the key type segment is supported, the path can not specify a list unless the list is the last segment.<br>[(validate.rules).repeated = {min_items: 1}]; |
  





<a name="ratelimit.api.solo.io.Action.MetaData.MetadataKey.PathSegment"></a>

### Action.MetaData.MetadataKey.PathSegment



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  | Required. If specified, use the key to retrieve the value in a Struct.<br>[(validate.rules).string = {min_len: 1}]; |
  





<a name="ratelimit.api.solo.io.Action.RemoteAddress"></a>

### Action.RemoteAddress







<a name="ratelimit.api.solo.io.Action.RequestHeaders"></a>

### Action.RequestHeaders



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headerName | string |  | The header name to be queried from the request headers. The header’s value is used to populate the value of the descriptor entry for the descriptor_key.<br>[(validate.rules).string.min_bytes = 1]; |
  | descriptorKey | string |  | The key to use in the descriptor entry.<br>[(validate.rules).string.min_bytes = 1]; |
  





<a name="ratelimit.api.solo.io.Action.SourceCluster"></a>

### Action.SourceCluster







<a name="ratelimit.api.solo.io.Descriptor"></a>

### Descriptor



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  | The key of the descriptor. This field is required. |
  | value | string |  | Optional value for the descriptor. If omitted, the server will create a rate limit for each value that is provided for this descriptor in rate limit requests. |
  | rateLimit | [ratelimit.api.solo.io.RateLimit]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.RateLimit" >}}) |  | Optional rate limit rule for the descriptor. |
  | descriptors | [][ratelimit.api.solo.io.Descriptor]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.Descriptor" >}}) | repeated | Nested descriptors. |
  | weight | uint32 |  | Each top-level Descriptor defines a new Rate Limit "rule". When a request comes in, rate limit actions are applied to the request to generate descriptor tuples that are sent to the rate limit server. If any rule is triggered then the entire request returns HTTP 429 Too Many Requests.<br>Typically, rule priority is signalled by nesting descriptors, as the most specific rule match for the descriptor tuple generated by the rate limit actions is used. In rare cases this is too restrictive; instead you can set rule priority by setting weights on your descriptors.<br>All rules with the highest weight are processed, if any of these rules trigger rate limiting then the entire request will return a 429. Rules that are not considered for rate limiting are ignored in the rate limit server, and their request count is not incremented in the rate limit server cache.<br>Defaults to 0; thus all rules are evaluated by default. |
  | alwaysApply | bool |  | A boolean override for rule priority via weighted rules. Any rule with `alwaysApply` set to `true` will always be considered for rate limiting, regardless of the rule's weight. The rule with the highest weight will still be considered. (this can be a rule that also has `alwaysApply` set to `true`)<br>Defaults to false. |
  





<a name="ratelimit.api.solo.io.RateLimit"></a>

### RateLimit



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| unit | [ratelimit.api.solo.io.RateLimit.Unit]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.RateLimit.Unit" >}}) |  |  |
  | requestsPerUnit | uint32 |  |  |
  





<a name="ratelimit.api.solo.io.RateLimitActions"></a>

### RateLimitActions



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| actions | [][ratelimit.api.solo.io.Action]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.Action" >}}) | repeated |  |
  | setActions | [][ratelimit.api.solo.io.Action]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.Action" >}}) | repeated |  |
  





<a name="ratelimit.api.solo.io.RateLimitConfigSpec"></a>

### RateLimitConfigSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| raw | [ratelimit.api.solo.io.RateLimitConfigSpec.Raw]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.RateLimitConfigSpec.Raw" >}}) |  | Define a policy using the raw configuration format used by the server and the client (Envoy). |
  





<a name="ratelimit.api.solo.io.RateLimitConfigSpec.Raw"></a>

### RateLimitConfigSpec.Raw



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| descriptors | [][ratelimit.api.solo.io.Descriptor]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.Descriptor" >}}) | repeated | The descriptors that will be applied to the server. |
  | rateLimits | [][ratelimit.api.solo.io.RateLimitActions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.RateLimitActions" >}}) | repeated | Actions specify how the client (Envoy) will compose the descriptors that will be sent to the server to make a rate limiting decision. |
  | setDescriptors | [][ratelimit.api.solo.io.SetDescriptor]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.SetDescriptor" >}}) | repeated | The set descriptors that will be applied to the server. |
  





<a name="ratelimit.api.solo.io.RateLimitConfigStatus"></a>

### RateLimitConfigStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| state | [ratelimit.api.solo.io.RateLimitConfigStatus.State]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.RateLimitConfigStatus.State" >}}) |  | The current state of the `RateLimitConfig`. |
  | message | string |  | A human-readable string explaining the status. |
  | observedGeneration | int64 |  | The observed generation of the resource. When this matches the metadata.generation of the resource, it indicates the status is up-to-date. |
  





<a name="ratelimit.api.solo.io.SetDescriptor"></a>

### SetDescriptor



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| simpleDescriptors | [][ratelimit.api.solo.io.SimpleDescriptor]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.SimpleDescriptor" >}}) | repeated | Simple descriptor key/value pairs. |
  | rateLimit | [ratelimit.api.solo.io.RateLimit]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.RateLimit" >}}) |  | Rate limit rule for the descriptor. |
  | alwaysApply | bool |  | Typically, rule priority is signalled by rule ordering, as the first rule match for the descriptor tuple generated by the rate limit actions is used.<br>In some cases this is too restrictive; A boolean override can be specified. Any rule with `alwaysApply` set to `true` will always be considered for rate limiting, regardless of the rule's place in the ordered list of rules. The first rule to match will still be considered. (This can be a rule that also has `alwaysApply` set to `true`.)<br>If any of these rules trigger rate limiting then the entire request will return a 429. Rules that are not considered for rate limiting are ignored in the rate limit server, and their request count is not incremented in the rate limit server cache.<br>Defaults to false. |
  





<a name="ratelimit.api.solo.io.SimpleDescriptor"></a>

### SimpleDescriptor



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  | The key of the descriptor. This field is required. |
  | value | string |  | Optional value for the descriptor. If omitted, the server will create a rate limit for each value that is provided for this descriptor in rate limit requests. |
  




 <!-- end messages -->


<a name="ratelimit.api.solo.io.Action.MetaData.Source"></a>

### Action.MetaData.Source


| Name | Number | Description |
| ---- | ------ | ----------- |
| DYNAMIC | 0 | Query [dynamic metadata](https://www.envoyproxy.io/docs/envoy/latest/configuration/advanced/well_known_dynamic_metadata#well-known-dynamic-metadata). |
| ROUTE_ENTRY | 1 | Query [route entry metadata](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-route-metadata). |



<a name="ratelimit.api.solo.io.RateLimit.Unit"></a>

### RateLimit.Unit


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN | 0 |  |
| SECOND | 1 |  |
| MINUTE | 2 |  |
| HOUR | 3 |  |
| DAY | 4 |  |



<a name="ratelimit.api.solo.io.RateLimitConfigStatus.State"></a>

### RateLimitConfigStatus.State


| Name | Number | Description |
| ---- | ------ | ----------- |
| PENDING | 0 |  |
| ACCEPTED | 1 |  |
| REJECTED | 2 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

