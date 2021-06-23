
---

title: "rate_limit.proto"

---

## Package : `ratelimit.networking.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for rate_limit.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## rate_limit.proto


## Table of Contents
  - [Action](#ratelimit.networking.mesh.gloo.solo.io.Action)
  - [Action.DestinationCluster](#ratelimit.networking.mesh.gloo.solo.io.Action.DestinationCluster)
  - [Action.GenericKey](#ratelimit.networking.mesh.gloo.solo.io.Action.GenericKey)
  - [Action.HeaderValueMatch](#ratelimit.networking.mesh.gloo.solo.io.Action.HeaderValueMatch)
  - [Action.HeaderValueMatch.HeaderMatcher](#ratelimit.networking.mesh.gloo.solo.io.Action.HeaderValueMatch.HeaderMatcher)
  - [Action.HeaderValueMatch.HeaderMatcher.Int64Range](#ratelimit.networking.mesh.gloo.solo.io.Action.HeaderValueMatch.HeaderMatcher.Int64Range)
  - [Action.Metadata](#ratelimit.networking.mesh.gloo.solo.io.Action.Metadata)
  - [Action.Metadata.MetadataKey](#ratelimit.networking.mesh.gloo.solo.io.Action.Metadata.MetadataKey)
  - [Action.Metadata.MetadataKey.PathSegment](#ratelimit.networking.mesh.gloo.solo.io.Action.Metadata.MetadataKey.PathSegment)
  - [Action.RemoteAddress](#ratelimit.networking.mesh.gloo.solo.io.Action.RemoteAddress)
  - [Action.RequestHeaders](#ratelimit.networking.mesh.gloo.solo.io.Action.RequestHeaders)
  - [Action.SourceCluster](#ratelimit.networking.mesh.gloo.solo.io.Action.SourceCluster)
  - [GatewayRateLimit](#ratelimit.networking.mesh.gloo.solo.io.GatewayRateLimit)
  - [RouteRateLimit](#ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit)
  - [RouteRateLimit.AdvancedRateLimit](#ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.AdvancedRateLimit)
  - [RouteRateLimit.AdvancedRateLimit.RateLimitActions](#ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.AdvancedRateLimit.RateLimitActions)
  - [RouteRateLimit.BasicRateLimit](#ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.BasicRateLimit)
  - [RouteRateLimit.BasicRateLimit.RateLimitRatio](#ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.BasicRateLimit.RateLimitRatio)

  - [Action.Metadata.Source](#ratelimit.networking.mesh.gloo.solo.io.Action.Metadata.Source)
  - [RouteRateLimit.BasicRateLimit.RateLimitRatio.Unit](#ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.BasicRateLimit.RateLimitRatio.Unit)






<a name="ratelimit.networking.mesh.gloo.solo.io.Action"></a>

### Action
Copied directly from envoy https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-msg-config-route-v3-ratelimit-action


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sourceCluster | [ratelimit.networking.mesh.gloo.solo.io.Action.SourceCluster]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.Action.SourceCluster" >}}) |  | Rate limit on source cluster. |
  | destinationCluster | [ratelimit.networking.mesh.gloo.solo.io.Action.DestinationCluster]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.Action.DestinationCluster" >}}) |  | Rate limit on destination cluster. |
  | requestHeaders | [ratelimit.networking.mesh.gloo.solo.io.Action.RequestHeaders]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.Action.RequestHeaders" >}}) |  | Rate limit on request headers. |
  | remoteAddress | [ratelimit.networking.mesh.gloo.solo.io.Action.RemoteAddress]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.Action.RemoteAddress" >}}) |  | Rate limit on remote address. |
  | genericKey | [ratelimit.networking.mesh.gloo.solo.io.Action.GenericKey]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.Action.GenericKey" >}}) |  | Rate limit on a generic key. |
  | headerValueMatch | [ratelimit.networking.mesh.gloo.solo.io.Action.HeaderValueMatch]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.Action.HeaderValueMatch" >}}) |  | Rate limit on the existence of request headers. |
  | metadata | [ratelimit.networking.mesh.gloo.solo.io.Action.Metadata]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.Action.Metadata" >}}) |  | Rate limit on metadata. |
  





<a name="ratelimit.networking.mesh.gloo.solo.io.Action.DestinationCluster"></a>

### Action.DestinationCluster
The following descriptor entry is appended to the descriptor:<br>```   ("destination_cluster", "<routed target cluster>") ```<br>Once a request matches against a route table rule, a routed cluster is determined by one of the following `route table configuration (envoy_api_msg_RouteConfiguration)` settings:<br>* `cluster (envoy_api_field_route.RouteAction.cluster)` indicates the upstream cluster   to route to. * `weighted_clusters (envoy_api_field_route.RouteAction.weighted_clusters)`   chooses a cluster randomly from a set of clusters with attributed weight. * `cluster_header (envoy_api_field_route.RouteAction.cluster_header)` indicates which   header in the request contains the target cluster.






<a name="ratelimit.networking.mesh.gloo.solo.io.Action.GenericKey"></a>

### Action.GenericKey
The following descriptor entry is appended to the descriptor:<br>```   ("generic_key", "<descriptor_value>") ```


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| descriptorValue | string |  | The value to use in the descriptor entry.<br>[(validate.rules).string.min_bytes = 1]; |
  





<a name="ratelimit.networking.mesh.gloo.solo.io.Action.HeaderValueMatch"></a>

### Action.HeaderValueMatch
The following descriptor entry is appended to the descriptor:<br>```   ("header_match", "<descriptor_value>") ```


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| descriptorValue | string |  | The value to use in the descriptor entry.<br>[(validate.rules).string.min_bytes = 1]; |
  | expectMatch | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | If set to true, the action will append a descriptor entry when the request matches the headers. If set to false, the action will append a descriptor entry when the request does not match the headers. The default value is true. |
  | headers | [][ratelimit.networking.mesh.gloo.solo.io.Action.HeaderValueMatch.HeaderMatcher]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.Action.HeaderValueMatch.HeaderMatcher" >}}) | repeated | Specifies a set of headers that the rate limit action should match on. The action will check the request’s headers against all the specified headers in the config. A match will happen if all the headers in the config are present in the request with the same values (or based on presence if the value field is not in the config).<br>[(validate.rules).repeated .min_items = 1]; |
  





<a name="ratelimit.networking.mesh.gloo.solo.io.Action.HeaderValueMatch.HeaderMatcher"></a>

### Action.HeaderValueMatch.HeaderMatcher



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Specifies the name of the header in the request.<br>[(validate.rules).string.min_bytes = 1]; |
  | exactMatch | string |  | If specified, header match will be performed based on the value of the header. |
  | regexMatch | string |  | If specified, this regex string is a regular expression rule which implies the entire request header value must match the regex. The rule will not match if only a subsequence of the request header value matches the regex. The regex grammar used in the value field is defined `(here)[https://en.cppreference.com/w/cpp/regex/ecmascript]`.<br>Examples:<br>* The regex *\d{3}* matches the value *123* * The regex *\d{3}* does not match the value *1234* * The regex *\d{3}* does not match the value *123.456*<br>[(validate.rules).string.max_bytes = 1024]; |
  | rangeMatch | [ratelimit.networking.mesh.gloo.solo.io.Action.HeaderValueMatch.HeaderMatcher.Int64Range]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.Action.HeaderValueMatch.HeaderMatcher.Int64Range" >}}) |  | If specified, header match will be performed based on range. The rule will match if the request header value is within this range. The entire request header value must represent an integer in base 10 notation: consisting of an optional plus or minus sign followed by a sequence of digits. The rule will not match if the header value does not represent an integer. Match will fail for empty values, floating point numbers or if only a subsequence of the header value is an integer.<br>Examples:<br>* For range [-10,0), route will match for header value -1, but not for 0, "somestring", 10.9,   "-1somestring" |
  | presentMatch | bool |  | If specified, header match will be performed based on whether the header is in the request. |
  | prefixMatch | string |  | If specified, header match will be performed based on the prefix of the header value. Note: empty prefix is not allowed, please use present_match instead.<br>Examples:<br>* The prefix *abcd* matches the value *abcdxyz*, but not for *abcxyz*.<br>[(validate.rules).string.min_bytes = 1]; |
  | suffixMatch | string |  | If specified, header match will be performed based on the suffix of the header value. Note: empty suffix is not allowed, please use present_match instead.<br>Examples:<br>* The suffix *abcd* matches the value *xyzabcd*, but not for *xyzbcd*.<br>[(validate.rules).string.min_bytes = 1]; |
  | invertMatch | bool |  | If specified, the match result will be inverted before checking. Defaults to false.<br>Examples:<br>* The regex *\d{3}* does not match the value *1234*, so it will match when inverted. * The range [-10,0) will match the value -1, so it will not match when inverted. |
  





<a name="ratelimit.networking.mesh.gloo.solo.io.Action.HeaderValueMatch.HeaderMatcher.Int64Range"></a>

### Action.HeaderValueMatch.HeaderMatcher.Int64Range
Specifies the int64 start and end of the range using half-open interval semantics [start, end).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| start | int64 |  | start of the range (inclusive) |
  | end | int64 |  | end of the range (exclusive) |
  





<a name="ratelimit.networking.mesh.gloo.solo.io.Action.Metadata"></a>

### Action.Metadata
The following descriptor entry is appended when the metadata contains a key value:   ("<descriptor_key>", "<value_queried_from_metadata>")


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| descriptorKey | string |  | Required. The key to use in the descriptor entry.<br>[(validate.rules).string = {min_len: 1}]; |
  | metadataKey | [ratelimit.networking.mesh.gloo.solo.io.Action.Metadata.MetadataKey]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.Action.Metadata.MetadataKey" >}}) |  | Required. Metadata struct that defines the key and path to retrieve the string value. A match will only happen if the value in the metadata is of type string.<br>[(validate.rules).message = {required: true}]; |
  | defaultValue | string |  | An optional value to use if *metadata_key* is empty. If not set and no value is present under the metadata_key then no descriptor is generated. |
  | source | [ratelimit.networking.mesh.gloo.solo.io.Action.Metadata.Source]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.Action.Metadata.Source" >}}) |  | Source of metadata<br>[(validate.rules).enum = {defined_only: true}]; |
  





<a name="ratelimit.networking.mesh.gloo.solo.io.Action.Metadata.MetadataKey"></a>

### Action.Metadata.MetadataKey
MetadataKey provides a general interface using `key` and `path` to retrieve value from [`Metadata`](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/base.proto#envoy-v3-api-msg-config-core-v3-metadata).<br>For example, for the following Metadata:<br>```yaml filter_metadata:   envoy.xxx:     prop:       foo: bar       xyz:         hello: envoy ```<br>The following MetadataKey will retrieve a string value "bar" from the Metadata.<br>```yaml key: envoy.xxx path: - key: prop - key: foo ```


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  | Required. The key name of Metadata to retrieve the Struct from the metadata. Typically, it represents a builtin subsystem or custom extension.<br>[(validate.rules).string = {min_len: 1}]; |
  | path | [][ratelimit.networking.mesh.gloo.solo.io.Action.Metadata.MetadataKey.PathSegment]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.Action.Metadata.MetadataKey.PathSegment" >}}) | repeated | Must have at least one element. The path to retrieve the Value from the Struct. It can be a prefix or a full path, e.g. ``[prop, xyz]`` for a struct or ``[prop, foo]`` for a string in the example, which depends on the particular scenario.<br>Note: Due to that only the key type segment is supported, the path can not specify a list unless the list is the last segment.<br>[(validate.rules).repeated = {min_items: 1}]; |
  





<a name="ratelimit.networking.mesh.gloo.solo.io.Action.Metadata.MetadataKey.PathSegment"></a>

### Action.Metadata.MetadataKey.PathSegment
Specifies the segment in a path to retrieve value from Metadata. Currently it is only supported to specify the key, i.e. field name, as one segment of a path.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  | Required. If specified, use the key to retrieve the value in a Struct.<br>[(validate.rules).string = {min_len: 1}]; |
  





<a name="ratelimit.networking.mesh.gloo.solo.io.Action.RemoteAddress"></a>

### Action.RemoteAddress
The following descriptor entry is appended to the descriptor and is populated using the trusted address from `x-forwarded-for (config_http_conn_man_headers_x-forwarded-for)`:<br>```   ("remote_address", "<trusted address from x-forwarded-for>") ```






<a name="ratelimit.networking.mesh.gloo.solo.io.Action.RequestHeaders"></a>

### Action.RequestHeaders
The following descriptor entry is appended when a header contains a key that matches the *header_name*:<br>```   ("<descriptor_key>", "<header_value_queried_from_header>") ```


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headerName | string |  | The header name to be queried from the request headers. The header’s value is used to populate the value of the descriptor entry for the descriptor_key.<br>[(validate.rules).string.min_bytes = 1]; |
  | descriptorKey | string |  | The key to use in the descriptor entry.<br>[(validate.rules).string.min_bytes = 1]; |
  





<a name="ratelimit.networking.mesh.gloo.solo.io.Action.SourceCluster"></a>

### Action.SourceCluster
The following descriptor entry is appended to the descriptor:<br>```   ("source_cluster", "<local service cluster>") ```<br><local service cluster> is derived from the :option:`--service-cluster` option.






<a name="ratelimit.networking.mesh.gloo.solo.io.GatewayRateLimit"></a>

### GatewayRateLimit
Configure the Rate-Limit Filter on a Gateway


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ratelimitServerRef | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  |  |
  | requestTimeout | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  |  |
  | denyOnFail | bool |  |  |
  | rateLimitBeforeAuth | bool |  | Set this is set to true if you would like to rate limit traffic before applying external auth to it. *Note*: When this is true, you will lose some features like being able to rate limit a request based on its auth state |
  





<a name="ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit"></a>

### RouteRateLimit
Rate limit configuration for a Route or TrafficPolicy. Configures rate limits for individual HTTP routes


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| basic | [ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.BasicRateLimit]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.BasicRateLimit" >}}) |  | Config for rate-limiting using simplified (gloo-specific) API |
  | advanced | [ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.AdvancedRateLimit]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.AdvancedRateLimit" >}}) |  | Partial config for GlooE rate-limiting based on Envoy's rate-limit service; supports Envoy's rate-limit service API. (reference here: https://github.com/lyft/ratelimit#configuration) Configure rate-limit *actions* here, which define how request characteristics get translated into descriptors used by the rate-limit service for rate-limiting. Configure rate-limit *descriptors* and their associated limits on the Gloo settings. Only one of `ratelimit` or `rate_limit_configs` can be set. |
  | configRefs | [common.mesh.gloo.solo.io.ObjectRefList]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.refs#common.mesh.gloo.solo.io.ObjectRefList" >}}) |  | References to RateLimitConfig resources. This is used to configure the GlooE rate limit server. Only one of `ratelimit` or `rate_limit_configs` can be set. |
  





<a name="ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.AdvancedRateLimit"></a>

### RouteRateLimit.AdvancedRateLimit
Use this field if you want to inline the Envoy rate limits for this VirtualHost. Note that this does not configure the rate limit server. If you are running Gloo Enterprise, you need to specify the server configuration via the appropriate field in the Gloo `Settings` resource. If you are running a custom rate limit server you need to configure it yourself.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| actions | [][ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.AdvancedRateLimit.RateLimitActions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.AdvancedRateLimit.RateLimitActions" >}}) | repeated | Define individual rate limits here. Each rate limit will be evaluated, if any rate limit would be throttled, the entire request returns a 429 (gets throttled) |
  





<a name="ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.AdvancedRateLimit.RateLimitActions"></a>

### RouteRateLimit.AdvancedRateLimit.RateLimitActions
Each action and setAction in the lists maps part of the request (or its context) to a descriptor. The tuple or set of descriptors generated by the provided actions is sent to the rate limit server and matched against rate limit rules. Order matters on provided actions but not on setActions, e.g. the following actions: - actions:   - requestHeaders:      descriptorKey: account_id      headerName: x-account-id   - requestHeaders:      descriptorKey: plan      headerName: x-plan define an ordered descriptor tuple like so: ('account_id', '<x-account-id value>'), ('plan', '<x-plan value>')<br>While the current form matches, the same tuple in reverse order would not match the following descriptor:<br>descriptors: - key: account_id   descriptors:   - key: plan     value: BASIC     rateLimit:       requestsPerUnit: 1       unit: MINUTE  - key: plan    value: PLUS    rateLimit:      requestsPerUnit: 20      unit: MINUTE<br>Similarly, the following setActions: - setActions:   - requestHeaders:      descriptorKey: account_id      headerName: x-account-id   - requestHeaders:      descriptorKey: plan      headerName: x-plan define an unordered descriptor set like so: {('account_id', '<x-account-id value>'), ('plan', '<x-plan value>')}<br>This set would match the following setDescriptor:<br>setDescriptors: - simpleDescriptors:   - key: plan     value: BASIC   - key: account_id  rateLimit:    requestsPerUnit: 20    unit: MINUTE<br>It would also match the following setDescriptor which includes only a subset of the setActions enumerated:<br>setDescriptors: - simpleDescriptors:   - key: account_id  rateLimit:    requestsPerUnit: 20    unit: MINUTE<br>It would even match the following setDescriptor. Any setActions list would match this setDescriptor which has simpleDescriptors omitted entirely:<br>setDescriptors: - rateLimit:    requestsPerUnit: 20    unit: MINUTE


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| actions | [][ratelimit.networking.mesh.gloo.solo.io.Action]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.Action" >}}) | repeated | TODO: come up with descriptive names and comments for these fields |
  | setActions | [][ratelimit.networking.mesh.gloo.solo.io.Action]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.Action" >}}) | repeated |  |
  





<a name="ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.BasicRateLimit"></a>

### RouteRateLimit.BasicRateLimit
Basic rate-limiting API


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| authorizedLimits | [ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.BasicRateLimit.RateLimitRatio]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.BasicRateLimit.RateLimitRatio" >}}) |  | limits for authorized users |
  | anonymousLimits | [ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.BasicRateLimit.RateLimitRatio]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.BasicRateLimit.RateLimitRatio" >}}) |  | limits for unauthorized users |
  





<a name="ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.BasicRateLimit.RateLimitRatio"></a>

### RouteRateLimit.BasicRateLimit.RateLimitRatio
A `RateLimitRatio` specifies the actual ratio of requests that will be permitted when there is a match.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| unit | [ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.BasicRateLimit.RateLimitRatio.Unit]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.BasicRateLimit.RateLimitRatio.Unit" >}}) |  |  |
  | requestsPerUnit | uint32 |  |  |
  




 <!-- end messages -->


<a name="ratelimit.networking.mesh.gloo.solo.io.Action.Metadata.Source"></a>

### Action.Metadata.Source


| Name | Number | Description |
| ---- | ------ | ----------- |
| DYNAMIC | 0 | Query [dynamic metadata](https://www.envoyproxy.io/docs/envoy/latest/configuration/advanced/well_known_dynamic_metadata#well-known-dynamic-metadata). |
| ROUTE_ENTRY | 1 | Query [route entry metadata](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-route-metadata). |



<a name="ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.BasicRateLimit.RateLimitRatio.Unit"></a>

### RouteRateLimit.BasicRateLimit.RateLimitRatio.Unit


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN | 0 |  |
| SECOND | 1 |  |
| MINUTE | 2 |  |
| HOUR | 3 |  |
| DAY | 4 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

