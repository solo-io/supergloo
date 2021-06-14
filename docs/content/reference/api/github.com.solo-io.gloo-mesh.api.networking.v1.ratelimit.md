
---

title: "ratelimit.proto"

---

## Package : `networking.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for ratelimit.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## ratelimit.proto


## Table of Contents
  - [Action](#networking.mesh.gloo.solo.io.Action)
  - [Action.DestinationCluster](#networking.mesh.gloo.solo.io.Action.DestinationCluster)
  - [Action.GenericKey](#networking.mesh.gloo.solo.io.Action.GenericKey)
  - [Action.HeaderValueMatch](#networking.mesh.gloo.solo.io.Action.HeaderValueMatch)
  - [Action.HeaderValueMatch.HeaderMatcher](#networking.mesh.gloo.solo.io.Action.HeaderValueMatch.HeaderMatcher)
  - [Action.HeaderValueMatch.HeaderMatcher.Int64Range](#networking.mesh.gloo.solo.io.Action.HeaderValueMatch.HeaderMatcher.Int64Range)
  - [Action.MetaData](#networking.mesh.gloo.solo.io.Action.MetaData)
  - [Action.MetaData.MetadataKey](#networking.mesh.gloo.solo.io.Action.MetaData.MetadataKey)
  - [Action.MetaData.MetadataKey.PathSegment](#networking.mesh.gloo.solo.io.Action.MetaData.MetadataKey.PathSegment)
  - [Action.RemoteAddress](#networking.mesh.gloo.solo.io.Action.RemoteAddress)
  - [Action.RequestHeaders](#networking.mesh.gloo.solo.io.Action.RequestHeaders)
  - [Action.SourceCluster](#networking.mesh.gloo.solo.io.Action.SourceCluster)
  - [Descriptor](#networking.mesh.gloo.solo.io.Descriptor)
  - [IngressRateLimit](#networking.mesh.gloo.solo.io.IngressRateLimit)
  - [RateLimit](#networking.mesh.gloo.solo.io.RateLimit)
  - [RateLimitActions](#networking.mesh.gloo.solo.io.RateLimitActions)
  - [RateLimitConfigRef](#networking.mesh.gloo.solo.io.RateLimitConfigRef)
  - [RateLimitConfigRefs](#networking.mesh.gloo.solo.io.RateLimitConfigRefs)
  - [RateLimitConfigSpec](#networking.mesh.gloo.solo.io.RateLimitConfigSpec)
  - [RateLimitConfigSpec.Raw](#networking.mesh.gloo.solo.io.RateLimitConfigSpec.Raw)
  - [RateLimitConfigStatus](#networking.mesh.gloo.solo.io.RateLimitConfigStatus)
  - [RateLimitRouteExtension](#networking.mesh.gloo.solo.io.RateLimitRouteExtension)
  - [RateLimitVhostExtension](#networking.mesh.gloo.solo.io.RateLimitVhostExtension)
  - [Ratelimit](#networking.mesh.gloo.solo.io.Ratelimit)
  - [RatelimitConfig](#networking.mesh.gloo.solo.io.RatelimitConfig)
  - [RatelimitSettings](#networking.mesh.gloo.solo.io.RatelimitSettings)
  - [ServiceSettings](#networking.mesh.gloo.solo.io.ServiceSettings)
  - [SetDescriptor](#networking.mesh.gloo.solo.io.SetDescriptor)
  - [Settings](#networking.mesh.gloo.solo.io.Settings)
  - [SimpleDescriptor](#networking.mesh.gloo.solo.io.SimpleDescriptor)

  - [Action.MetaData.Source](#networking.mesh.gloo.solo.io.Action.MetaData.Source)
  - [RateLimit.Unit](#networking.mesh.gloo.solo.io.RateLimit.Unit)
  - [RateLimitConfigStatus.State](#networking.mesh.gloo.solo.io.RateLimitConfigStatus.State)






<a name="networking.mesh.gloo.solo.io.Action"></a>

### Action
Copied directly from envoy https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-msg-config-route-v3-ratelimit-action


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sourceCluster | [networking.mesh.gloo.solo.io.Action.SourceCluster]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.Action.SourceCluster" >}}) |  | Rate limit on source cluster. |
  | destinationCluster | [networking.mesh.gloo.solo.io.Action.DestinationCluster]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.Action.DestinationCluster" >}}) |  | Rate limit on destination cluster. |
  | requestHeaders | [networking.mesh.gloo.solo.io.Action.RequestHeaders]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.Action.RequestHeaders" >}}) |  | Rate limit on request headers. |
  | remoteAddress | [networking.mesh.gloo.solo.io.Action.RemoteAddress]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.Action.RemoteAddress" >}}) |  | Rate limit on remote address. |
  | genericKey | [networking.mesh.gloo.solo.io.Action.GenericKey]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.Action.GenericKey" >}}) |  | Rate limit on a generic key. |
  | headerValueMatch | [networking.mesh.gloo.solo.io.Action.HeaderValueMatch]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.Action.HeaderValueMatch" >}}) |  | Rate limit on the existence of request headers. |
  | metadata | [networking.mesh.gloo.solo.io.Action.MetaData]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.Action.MetaData" >}}) |  | Rate limit on metadata. |
  





<a name="networking.mesh.gloo.solo.io.Action.DestinationCluster"></a>

### Action.DestinationCluster
The following descriptor entry is appended to the descriptor:<br>```   ("destination_cluster", "<routed target cluster>") ```<br>Once a request matches against a route table rule, a routed cluster is determined by one of the following `route table configuration (envoy_api_msg_RouteConfiguration)` settings:<br>* `cluster (envoy_api_field_route.RouteAction.cluster)` indicates the upstream cluster   to route to. * `weighted_clusters (envoy_api_field_route.RouteAction.weighted_clusters)`   chooses a cluster randomly from a set of clusters with attributed weight. * `cluster_header (envoy_api_field_route.RouteAction.cluster_header)` indicates which   header in the request contains the target cluster.






<a name="networking.mesh.gloo.solo.io.Action.GenericKey"></a>

### Action.GenericKey
The following descriptor entry is appended to the descriptor:<br>```   ("generic_key", "<descriptor_value>") ```


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| descriptorValue | string |  | The value to use in the descriptor entry.<br>[(validate.rules).string.min_bytes = 1]; |
  





<a name="networking.mesh.gloo.solo.io.Action.HeaderValueMatch"></a>

### Action.HeaderValueMatch
The following descriptor entry is appended to the descriptor:<br>```   ("header_match", "<descriptor_value>") ```


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| descriptorValue | string |  | The value to use in the descriptor entry.<br>[(validate.rules).string.min_bytes = 1]; |
  | expectMatch | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | If set to true, the action will append a descriptor entry when the request matches the headers. If set to false, the action will append a descriptor entry when the request does not match the headers. The default value is true. |
  | headers | [][networking.mesh.gloo.solo.io.Action.HeaderValueMatch.HeaderMatcher]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.Action.HeaderValueMatch.HeaderMatcher" >}}) | repeated | Specifies a set of headers that the rate limit action should match on. The action will check the request’s headers against all the specified headers in the config. A match will happen if all the headers in the config are present in the request with the same values (or based on presence if the value field is not in the config).<br>[(validate.rules).repeated .min_items = 1]; |
  





<a name="networking.mesh.gloo.solo.io.Action.HeaderValueMatch.HeaderMatcher"></a>

### Action.HeaderValueMatch.HeaderMatcher



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Specifies the name of the header in the request.<br>[(validate.rules).string.min_bytes = 1]; |
  | exactMatch | string |  | If specified, header match will be performed based on the value of the header. |
  | regexMatch | string |  | If specified, this regex string is a regular expression rule which implies the entire request header value must match the regex. The rule will not match if only a subsequence of the request header value matches the regex. The regex grammar used in the value field is defined `(here)[https://en.cppreference.com/w/cpp/regex/ecmascript]`.<br>Examples:<br>* The regex *\d{3}* matches the value *123* * The regex *\d{3}* does not match the value *1234* * The regex *\d{3}* does not match the value *123.456*<br>[(validate.rules).string.max_bytes = 1024]; |
  | rangeMatch | [networking.mesh.gloo.solo.io.Action.HeaderValueMatch.HeaderMatcher.Int64Range]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.Action.HeaderValueMatch.HeaderMatcher.Int64Range" >}}) |  | If specified, header match will be performed based on range. The rule will match if the request header value is within this range. The entire request header value must represent an integer in base 10 notation: consisting of an optional plus or minus sign followed by a sequence of digits. The rule will not match if the header value does not represent an integer. Match will fail for empty values, floating point numbers or if only a subsequence of the header value is an integer.<br>Examples:<br>* For range [-10,0), route will match for header value -1, but not for 0, "somestring", 10.9,   "-1somestring" |
  | presentMatch | bool |  | If specified, header match will be performed based on whether the header is in the request. |
  | prefixMatch | string |  | If specified, header match will be performed based on the prefix of the header value. Note: empty prefix is not allowed, please use present_match instead.<br>Examples:<br>* The prefix *abcd* matches the value *abcdxyz*, but not for *abcxyz*.<br>[(validate.rules).string.min_bytes = 1]; |
  | suffixMatch | string |  | If specified, header match will be performed based on the suffix of the header value. Note: empty suffix is not allowed, please use present_match instead.<br>Examples:<br>* The suffix *abcd* matches the value *xyzabcd*, but not for *xyzbcd*.<br>[(validate.rules).string.min_bytes = 1]; |
  | invertMatch | bool |  | If specified, the match result will be inverted before checking. Defaults to false.<br>Examples:<br>* The regex *\d{3}* does not match the value *1234*, so it will match when inverted. * The range [-10,0) will match the value -1, so it will not match when inverted. |
  





<a name="networking.mesh.gloo.solo.io.Action.HeaderValueMatch.HeaderMatcher.Int64Range"></a>

### Action.HeaderValueMatch.HeaderMatcher.Int64Range
Specifies the int64 start and end of the range using half-open interval semantics [start, end).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| start | int64 |  | start of the range (inclusive) |
  | end | int64 |  | end of the range (exclusive) |
  





<a name="networking.mesh.gloo.solo.io.Action.MetaData"></a>

### Action.MetaData
The following descriptor entry is appended when the metadata contains a key value:   ("<descriptor_key>", "<value_queried_from_metadata>")


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| descriptorKey | string |  | Required. The key to use in the descriptor entry.<br>[(validate.rules).string = {min_len: 1}]; |
  | metadataKey | [networking.mesh.gloo.solo.io.Action.MetaData.MetadataKey]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.Action.MetaData.MetadataKey" >}}) |  | Required. Metadata struct that defines the key and path to retrieve the string value. A match will only happen if the value in the metadata is of type string.<br>[(validate.rules).message = {required: true}]; |
  | defaultValue | string |  | An optional value to use if *metadata_key* is empty. If not set and no value is present under the metadata_key then no descriptor is generated. |
  | source | [networking.mesh.gloo.solo.io.Action.MetaData.Source]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.Action.MetaData.Source" >}}) |  | Source of metadata<br>[(validate.rules).enum = {defined_only: true}]; |
  





<a name="networking.mesh.gloo.solo.io.Action.MetaData.MetadataKey"></a>

### Action.MetaData.MetadataKey
MetadataKey provides a general interface using `key` and `path` to retrieve value from [`Metadata`](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/base.proto#envoy-v3-api-msg-config-core-v3-metadata).<br>For example, for the following Metadata:<br>```yaml filter_metadata:   envoy.xxx:     prop:       foo: bar       xyz:         hello: envoy ```<br>The following MetadataKey will retrieve a string value "bar" from the Metadata.<br>```yaml key: envoy.xxx path: - key: prop - key: foo ```


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  | Required. The key name of Metadata to retrieve the Struct from the metadata. Typically, it represents a builtin subsystem or custom extension.<br>[(validate.rules).string = {min_len: 1}]; |
  | path | [][networking.mesh.gloo.solo.io.Action.MetaData.MetadataKey.PathSegment]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.Action.MetaData.MetadataKey.PathSegment" >}}) | repeated | Must have at least one element. The path to retrieve the Value from the Struct. It can be a prefix or a full path, e.g. ``[prop, xyz]`` for a struct or ``[prop, foo]`` for a string in the example, which depends on the particular scenario.<br>Note: Due to that only the key type segment is supported, the path can not specify a list unless the list is the last segment.<br>[(validate.rules).repeated = {min_items: 1}]; |
  





<a name="networking.mesh.gloo.solo.io.Action.MetaData.MetadataKey.PathSegment"></a>

### Action.MetaData.MetadataKey.PathSegment
Specifies the segment in a path to retrieve value from Metadata. Currently it is only supported to specify the key, i.e. field name, as one segment of a path.<br>// TODO: cue doesn't like oneofs        oneof segment {          // option (validate.required) = true;<br>         // Required. If specified, use the key to retrieve the value in a Struct.          string key = 1; // [(validate.rules).string = {min_len: 1}];        }






<a name="networking.mesh.gloo.solo.io.Action.RemoteAddress"></a>

### Action.RemoteAddress
The following descriptor entry is appended to the descriptor and is populated using the trusted address from `x-forwarded-for (config_http_conn_man_headers_x-forwarded-for)`:<br>```   ("remote_address", "<trusted address from x-forwarded-for>") ```






<a name="networking.mesh.gloo.solo.io.Action.RequestHeaders"></a>

### Action.RequestHeaders
The following descriptor entry is appended when a header contains a key that matches the *header_name*:<br>```   ("<descriptor_key>", "<header_value_queried_from_header>") ```


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headerName | string |  | The header name to be queried from the request headers. The header’s value is used to populate the value of the descriptor entry for the descriptor_key.<br>[(validate.rules).string.min_bytes = 1]; |
  | descriptorKey | string |  | The key to use in the descriptor entry.<br>[(validate.rules).string.min_bytes = 1]; |
  





<a name="networking.mesh.gloo.solo.io.Action.SourceCluster"></a>

### Action.SourceCluster
The following descriptor entry is appended to the descriptor:<br>```   ("source_cluster", "<local service cluster>") ```<br><local service cluster> is derived from the :option:`--service-cluster` option.






<a name="networking.mesh.gloo.solo.io.Descriptor"></a>

### Descriptor
A descriptor is a list of key/value pairs that the rate limit server uses to select the correct rate limit to use when limiting. Descriptors are case-sensitive.<br>Each configuration contains a top level descriptor list and potentially multiple nested lists beneath that. The format is:<br>``` descriptors:   - key: <rule key: required>     value: <rule value: optional>     rate_limit: (optional block)       unit: <see below: required>       requests_per_unit: <see below: required>     descriptors: (optional block)       - ... (nested repetition of above) ```<br>Each descriptor in a descriptor list must have a key. It can also optionally have a value to enable a more specific match. The `rate_limit` block is optional and, if present, sets up an actual rate limit rule. If the rate limit is not present and there are no nested descriptors, then the descriptor is effectively whitelisted. Otherwise, nested descriptors allow more complex matching and rate limiting scenarios.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  | The key of the descriptor. This field is required. |
  | value | string |  | Optional value for the descriptor. If omitted, the server will create a rate limit for each value that is provided for this descriptor in rate limit requests. |
  | weight | uint32 |  | Each top-level Descriptor defines a new Rate Limit "rule". When a request comes in, rate limit actions are applied to the request to generate descriptor tuples that are sent to the rate limit server. If any rule is triggered then the entire request returns HTTP 429 Too Many Requests.<br>Typically, rule priority is signalled by nesting descriptors, as the most specific rule match for the descriptor tuple generated by the rate limit actions is used. In rare cases this is too restrictive; instead you can set rule priority by setting weights on your descriptors.<br>All rules with the highest weight are processed, if any of these rules trigger rate limiting then the entire request will return a 429. Rules that are not considered for rate limiting are ignored in the rate limit server, and their request count is not incremented in the rate limit server cache.<br>Defaults to 0; thus all rules are evaluated by default. |
  | alwaysApply | bool |  | A boolean override for rule priority via weighted rules. Any rule with `alwaysApply` set to `true` will always be considered for rate limiting, regardless of the rule's weight. The rule with the highest weight will still be considered. (this can be a rule that also has `alwaysApply` set to `true`)<br>Defaults to false. |
  





<a name="networking.mesh.gloo.solo.io.IngressRateLimit"></a>

### IngressRateLimit
Basic rate-limiting API


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| authorizedLimits | [networking.mesh.gloo.solo.io.RateLimit]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.RateLimit" >}}) |  |  |
  | anonymousLimits | [networking.mesh.gloo.solo.io.RateLimit]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.RateLimit" >}}) |  |  |
  





<a name="networking.mesh.gloo.solo.io.RateLimit"></a>

### RateLimit
A `RateLimit` specifies the actual rate limit that will be used when there is a match.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| unit | [networking.mesh.gloo.solo.io.RateLimit.Unit]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.RateLimit.Unit" >}}) |  |  |
  | requestsPerUnit | uint32 |  |  |
  





<a name="networking.mesh.gloo.solo.io.RateLimitActions"></a>

### RateLimitActions
Each action and setAction in the lists maps part of the request (or its context) to a descriptor. The tuple or set of descriptors generated by the provided actions is sent to the rate limit server and matched against rate limit rules. Order matters on provided actions but not on setActions, e.g. the following actions: - actions:   - requestHeaders:      descriptorKey: account_id      headerName: x-account-id   - requestHeaders:      descriptorKey: plan      headerName: x-plan define an ordered descriptor tuple like so: ('account_id', '<x-account-id value>'), ('plan', '<x-plan value>')<br>While the current form matches, the same tuple in reverse order would not match the following descriptor:<br>descriptors: - key: account_id   descriptors:   - key: plan     value: BASIC     rateLimit:       requestsPerUnit: 1       unit: MINUTE  - key: plan    value: PLUS    rateLimit:      requestsPerUnit: 20      unit: MINUTE<br>Similarly, the following setActions: - setActions:   - requestHeaders:      descriptorKey: account_id      headerName: x-account-id   - requestHeaders:      descriptorKey: plan      headerName: x-plan define an unordered descriptor set like so: {('account_id', '<x-account-id value>'), ('plan', '<x-plan value>')}<br>This set would match the following setDescriptor:<br>setDescriptors: - simpleDescriptors:   - key: plan     value: BASIC   - key: account_id  rateLimit:    requestsPerUnit: 20    unit: MINUTE<br>It would also match the following setDescriptor which includes only a subset of the setActions enumerated:<br>setDescriptors: - simpleDescriptors:   - key: account_id  rateLimit:    requestsPerUnit: 20    unit: MINUTE<br>It would even match the following setDescriptor. Any setActions list would match this setDescriptor which has simpleDescriptors omitted entirely:<br>setDescriptors: - rateLimit:    requestsPerUnit: 20    unit: MINUTE


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| actions | [][networking.mesh.gloo.solo.io.Action]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.Action" >}}) | repeated |  |
  | setActions | [][networking.mesh.gloo.solo.io.Action]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.Action" >}}) | repeated |  |
  





<a name="networking.mesh.gloo.solo.io.RateLimitConfigRef"></a>

### RateLimitConfigRef
A reference to a `RateLimitConfig` resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  |  |
  | namespace | string |  |  |
  





<a name="networking.mesh.gloo.solo.io.RateLimitConfigRefs"></a>

### RateLimitConfigRefs
A list of references to `RateLimitConfig` resources. Each resource represents a rate limit policy that will be independently enforced.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| refs | [][networking.mesh.gloo.solo.io.RateLimitConfigRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.RateLimitConfigRef" >}}) | repeated |  |
  





<a name="networking.mesh.gloo.solo.io.RateLimitConfigSpec"></a>

### RateLimitConfigSpec
A `RateLimitConfig` describes a rate limit policy.






<a name="networking.mesh.gloo.solo.io.RateLimitConfigSpec.Raw"></a>

### RateLimitConfigSpec.Raw
This object allows users to specify rate limit policies using the raw configuration formats used by the server and the client (Envoy). When using this configuration type, it is up to the user to ensure that server and client configurations match to implement the desired behavior. The server (and the client libraries that are shipped with it) will ensure that there are no collisions between raw configurations defined on separate `RateLimitConfig` resources.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| descriptors | [][networking.mesh.gloo.solo.io.Descriptor]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.Descriptor" >}}) | repeated | The descriptors that will be applied to the server. |
  | rateLimits | [][networking.mesh.gloo.solo.io.RateLimitActions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.RateLimitActions" >}}) | repeated | Actions specify how the client (Envoy) will compose the descriptors that will be sent to the server to make a rate limiting decision. |
  | setDescriptors | [][networking.mesh.gloo.solo.io.SetDescriptor]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.SetDescriptor" >}}) | repeated | The set descriptors that will be applied to the server. |
  





<a name="networking.mesh.gloo.solo.io.RateLimitConfigStatus"></a>

### RateLimitConfigStatus
The current status of the `RateLimitConfig`.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| state | [networking.mesh.gloo.solo.io.RateLimitConfigStatus.State]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.RateLimitConfigStatus.State" >}}) |  | The current state of the `RateLimitConfig`. |
  | message | string |  | A human-readable string explaining the status. |
  | observedGeneration | int64 |  | The observed generation of the resource. When this matches the metadata.generation of the resource, it indicates the status is up-to-date. |
  





<a name="networking.mesh.gloo.solo.io.RateLimitRouteExtension"></a>

### RateLimitRouteExtension
Use this field if you want to inline the Envoy rate limits for this Route. Note that this does not configure the rate limit server. If you are running Gloo Enterprise, you need to specify the server configuration via the appropriate field in the Gloo `Settings` resource. If you are running a custom rate limit server you need to configure it yourself.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| includeVhRateLimits | bool |  | Whether or not to include rate limits as defined on the VirtualHost in addition to rate limits on the Route. |
  | rateLimits | [][networking.mesh.gloo.solo.io.RateLimitActions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.RateLimitActions" >}}) | repeated | Define individual rate limits here. Each rate limit will be evaluated, if any rate limit would be throttled, the entire request returns a 429 (gets throttled) |
  





<a name="networking.mesh.gloo.solo.io.RateLimitVhostExtension"></a>

### RateLimitVhostExtension
Use this field if you want to inline the Envoy rate limits for this VirtualHost. Note that this does not configure the rate limit server. If you are running Gloo Enterprise, you need to specify the server configuration via the appropriate field in the Gloo `Settings` resource. If you are running a custom rate limit server you need to configure it yourself.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rateLimits | [][networking.mesh.gloo.solo.io.RateLimitActions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.RateLimitActions" >}}) | repeated | Define individual rate limits here. Each rate limit will be evaluated, if any rate limit would be throttled, the entire request returns a 429 (gets throttled) |
  





<a name="networking.mesh.gloo.solo.io.Ratelimit"></a>

### Ratelimit



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ratelimitBasic | [networking.mesh.gloo.solo.io.IngressRateLimit]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.IngressRateLimit" >}}) |  | Config for rate-limiting using simplified (gloo-specific) API |
  | rateLimitConfigs | [networking.mesh.gloo.solo.io.RateLimitConfigRefs]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.RateLimitConfigRefs" >}}) |  | References to RateLimitConfig resources. This is used to configure the GlooE rate limit server. Only one of `ratelimit` or `rate_limit_configs` can be set. |
  





<a name="networking.mesh.gloo.solo.io.RatelimitConfig"></a>

### RatelimitConfig
Ratelimit filter config.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| domain | string |  | Ratelimit domain |
  | descriptors | [networking.mesh.gloo.solo.io.Descriptor]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.Descriptor" >}}) |  |  |
  | setDescriptors | [networking.mesh.gloo.solo.io.SetDescriptor]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.SetDescriptor" >}}) |  |  |
  





<a name="networking.mesh.gloo.solo.io.RatelimitSettings"></a>

### RatelimitSettings
TODO: fix ratelimit settings


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ratelimitConfig | [networking.mesh.gloo.solo.io.Ratelimit]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.Ratelimit" >}}) |  |  |
  | ratelimitSettings | [networking.mesh.gloo.solo.io.ServiceSettings]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.ServiceSettings" >}}) |  | Enterprise-only: Partial config for GlooE's rate-limiting service, based on Envoy's rate-limit service; supports Envoy's rate-limit service API. (reference here: https://github.com/lyft/ratelimit#configuration) Configure rate-limit *descriptors* here, which define the limits for requests based on their descriptors. Configure rate-limits (composed of *actions*, which define how request characteristics get translated into descriptors) on the VirtualHost or its routes |
  | ratelimitServer | [networking.mesh.gloo.solo.io.Settings]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.Settings" >}}) |  | Enterprise-only: Settings for the rate limiting server itself |
  





<a name="networking.mesh.gloo.solo.io.ServiceSettings"></a>

### ServiceSettings
API based on Envoy's rate-limit service API. (reference here: https://github.com/lyft/ratelimit#configuration) Sample configuration below:<br>descriptors: - key: account_id  descriptors:  - key: plan    value: BASIC    rateLimit:      requestsPerUnit: 1      unit: MINUTE  - key: plan    value: PLUS    rateLimit:      requestsPerUnit: 20      unit: MINUTE


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| descriptors | [][networking.mesh.gloo.solo.io.Descriptor]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.Descriptor" >}}) | repeated |  |
  | setDescriptors | [][networking.mesh.gloo.solo.io.SetDescriptor]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.SetDescriptor" >}}) | repeated |  |
  





<a name="networking.mesh.gloo.solo.io.SetDescriptor"></a>

### SetDescriptor
A setDescriptor is a list of key/value pairs that the rate limit server uses to select the correct rate limit to use when limiting with the set style. Descriptors are case-sensitive.<br>Each configuration contains a simpleDescriptor list and a rateLimit. The format is:<br>``` set_descriptors:  - simple_descriptors: (optional block)      - key: <rule key: required>        value: <rule value: optional>      - ... (repetition of above)    rate_limit:      requests_per_unit: <see below: required>      unit: <see below: required>    always_apply: <bool value: optional>  - ... (repetition of above) ```<br>Each SetDescriptor defines a new Rate Limit "rule". When a request comes in, rate limit actions are applied to the request to generate descriptor tuples that are sent to the rate limit server. If any rule is triggered then the entire request returns HTTP 429 Too Many Requests.<br>The `rate_limit` block sets up an actual rate limit rule.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| simpleDescriptors | [][networking.mesh.gloo.solo.io.SimpleDescriptor]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.SimpleDescriptor" >}}) | repeated | Simple descriptor key/value pairs. |
  | rateLimit | [networking.mesh.gloo.solo.io.RateLimit]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.RateLimit" >}}) |  | Rate limit rule for the descriptor. |
  | alwaysApply | bool |  | Typically, rule priority is signalled by rule ordering, as the first rule match for the descriptor tuple generated by the rate limit actions is used.<br>In some cases this is too restrictive; A boolean override can be specified. Any rule with `alwaysApply` set to `true` will always be considered for rate limiting, regardless of the rule's place in the ordered list of rules. The first rule to match will still be considered. (This can be a rule that also has `alwaysApply` set to `true`.)<br>If any of these rules trigger rate limiting then the entire request will return a 429. Rules that are not considered for rate limiting are ignored in the rate limit server, and their request count is not incremented in the rate limit server cache.<br>Defaults to false. |
  





<a name="networking.mesh.gloo.solo.io.Settings"></a>

### Settings



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ratelimitServerRef | [networking.mesh.gloo.solo.io.RateLimitConfigRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit#networking.mesh.gloo.solo.io.RateLimitConfigRef" >}}) |  |  |
  | requestTimeout | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  |  |
  | denyOnFail | bool |  |  |
  | rateLimitBeforeAuth | bool |  | Set this is set to true if you would like to rate limit traffic before applying external auth to it. *Note*: When this is true, you will lose some features like being able to rate limit a request based on its auth state |
  





<a name="networking.mesh.gloo.solo.io.SimpleDescriptor"></a>

### SimpleDescriptor
A simpleDescriptor is a list of key/value pairs that the rate limit server uses to select the correct rate limit to use when limiting with the set style. Descriptors are case-sensitive.<br>The format is:<br>```  simple_descriptors:    - key: <rule key: required>      value: <rule value: optional>    - ... (repetition of above) ```<br>Each simpleDescriptor in a simpleDescriptor list must have a key. It can also optionally have a value to enable a more specific match.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  | The key of the descriptor. This field is required. |
  | value | string |  | Optional value for the descriptor. If omitted, the server will create a rate limit for each value that is provided for this descriptor in rate limit requests. |
  




 <!-- end messages -->


<a name="networking.mesh.gloo.solo.io.Action.MetaData.Source"></a>

### Action.MetaData.Source


| Name | Number | Description |
| ---- | ------ | ----------- |
| DYNAMIC | 0 | Query [dynamic metadata](https://www.envoyproxy.io/docs/envoy/latest/configuration/advanced/well_known_dynamic_metadata#well-known-dynamic-metadata). |
| ROUTE_ENTRY | 1 | Query [route entry metadata](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-route-metadata). |



<a name="networking.mesh.gloo.solo.io.RateLimit.Unit"></a>

### RateLimit.Unit


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN | 0 |  |
| SECOND | 1 |  |
| MINUTE | 2 |  |
| HOUR | 3 |  |
| DAY | 4 |  |



<a name="networking.mesh.gloo.solo.io.RateLimitConfigStatus.State"></a>

### RateLimitConfigStatus.State


| Name | Number | Description |
| ---- | ------ | ----------- |
| PENDING | 0 |  |
| ACCEPTED | 1 |  |
| REJECTED | 2 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

