
---

---

## Package : `envoy.config.route.v3`



<a name="top"></a>

<a name="API Reference for route_components.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## route_components.proto


## Table of Contents
  - [CorsPolicy](#envoy.config.route.v3.CorsPolicy)
  - [Decorator](#envoy.config.route.v3.Decorator)
  - [DirectResponseAction](#envoy.config.route.v3.DirectResponseAction)
  - [FilterAction](#envoy.config.route.v3.FilterAction)
  - [HeaderMatcher](#envoy.config.route.v3.HeaderMatcher)
  - [HedgePolicy](#envoy.config.route.v3.HedgePolicy)
  - [InternalRedirectPolicy](#envoy.config.route.v3.InternalRedirectPolicy)
  - [QueryParameterMatcher](#envoy.config.route.v3.QueryParameterMatcher)
  - [RateLimit](#envoy.config.route.v3.RateLimit)
  - [RateLimit.Action](#envoy.config.route.v3.RateLimit.Action)
  - [RateLimit.Action.DestinationCluster](#envoy.config.route.v3.RateLimit.Action.DestinationCluster)
  - [RateLimit.Action.DynamicMetaData](#envoy.config.route.v3.RateLimit.Action.DynamicMetaData)
  - [RateLimit.Action.GenericKey](#envoy.config.route.v3.RateLimit.Action.GenericKey)
  - [RateLimit.Action.HeaderValueMatch](#envoy.config.route.v3.RateLimit.Action.HeaderValueMatch)
  - [RateLimit.Action.MetaData](#envoy.config.route.v3.RateLimit.Action.MetaData)
  - [RateLimit.Action.RemoteAddress](#envoy.config.route.v3.RateLimit.Action.RemoteAddress)
  - [RateLimit.Action.RequestHeaders](#envoy.config.route.v3.RateLimit.Action.RequestHeaders)
  - [RateLimit.Action.SourceCluster](#envoy.config.route.v3.RateLimit.Action.SourceCluster)
  - [RateLimit.Override](#envoy.config.route.v3.RateLimit.Override)
  - [RateLimit.Override.DynamicMetadata](#envoy.config.route.v3.RateLimit.Override.DynamicMetadata)
  - [RedirectAction](#envoy.config.route.v3.RedirectAction)
  - [RetryPolicy](#envoy.config.route.v3.RetryPolicy)
  - [RetryPolicy.RateLimitedRetryBackOff](#envoy.config.route.v3.RetryPolicy.RateLimitedRetryBackOff)
  - [RetryPolicy.ResetHeader](#envoy.config.route.v3.RetryPolicy.ResetHeader)
  - [RetryPolicy.RetryBackOff](#envoy.config.route.v3.RetryPolicy.RetryBackOff)
  - [RetryPolicy.RetryHostPredicate](#envoy.config.route.v3.RetryPolicy.RetryHostPredicate)
  - [RetryPolicy.RetryPriority](#envoy.config.route.v3.RetryPolicy.RetryPriority)
  - [Route](#envoy.config.route.v3.Route)
  - [Route.TypedPerFilterConfigEntry](#envoy.config.route.v3.Route.TypedPerFilterConfigEntry)
  - [RouteAction](#envoy.config.route.v3.RouteAction)
  - [RouteAction.HashPolicy](#envoy.config.route.v3.RouteAction.HashPolicy)
  - [RouteAction.HashPolicy.ConnectionProperties](#envoy.config.route.v3.RouteAction.HashPolicy.ConnectionProperties)
  - [RouteAction.HashPolicy.Cookie](#envoy.config.route.v3.RouteAction.HashPolicy.Cookie)
  - [RouteAction.HashPolicy.FilterState](#envoy.config.route.v3.RouteAction.HashPolicy.FilterState)
  - [RouteAction.HashPolicy.Header](#envoy.config.route.v3.RouteAction.HashPolicy.Header)
  - [RouteAction.HashPolicy.QueryParameter](#envoy.config.route.v3.RouteAction.HashPolicy.QueryParameter)
  - [RouteAction.MaxStreamDuration](#envoy.config.route.v3.RouteAction.MaxStreamDuration)
  - [RouteAction.RequestMirrorPolicy](#envoy.config.route.v3.RouteAction.RequestMirrorPolicy)
  - [RouteAction.UpgradeConfig](#envoy.config.route.v3.RouteAction.UpgradeConfig)
  - [RouteAction.UpgradeConfig.ConnectConfig](#envoy.config.route.v3.RouteAction.UpgradeConfig.ConnectConfig)
  - [RouteMatch](#envoy.config.route.v3.RouteMatch)
  - [RouteMatch.ConnectMatcher](#envoy.config.route.v3.RouteMatch.ConnectMatcher)
  - [RouteMatch.GrpcRouteMatchOptions](#envoy.config.route.v3.RouteMatch.GrpcRouteMatchOptions)
  - [RouteMatch.TlsContextMatchOptions](#envoy.config.route.v3.RouteMatch.TlsContextMatchOptions)
  - [Tracing](#envoy.config.route.v3.Tracing)
  - [VirtualCluster](#envoy.config.route.v3.VirtualCluster)
  - [VirtualHost](#envoy.config.route.v3.VirtualHost)
  - [VirtualHost.TypedPerFilterConfigEntry](#envoy.config.route.v3.VirtualHost.TypedPerFilterConfigEntry)
  - [WeightedCluster](#envoy.config.route.v3.WeightedCluster)
  - [WeightedCluster.ClusterWeight](#envoy.config.route.v3.WeightedCluster.ClusterWeight)
  - [WeightedCluster.ClusterWeight.TypedPerFilterConfigEntry](#envoy.config.route.v3.WeightedCluster.ClusterWeight.TypedPerFilterConfigEntry)

  - [RateLimit.Action.MetaData.Source](#envoy.config.route.v3.RateLimit.Action.MetaData.Source)
  - [RedirectAction.RedirectResponseCode](#envoy.config.route.v3.RedirectAction.RedirectResponseCode)
  - [RetryPolicy.ResetHeaderFormat](#envoy.config.route.v3.RetryPolicy.ResetHeaderFormat)
  - [RouteAction.ClusterNotFoundResponseCode](#envoy.config.route.v3.RouteAction.ClusterNotFoundResponseCode)
  - [RouteAction.InternalRedirectAction](#envoy.config.route.v3.RouteAction.InternalRedirectAction)
  - [VirtualHost.TlsRequirementType](#envoy.config.route.v3.VirtualHost.TlsRequirementType)






<a name="envoy.config.route.v3.CorsPolicy"></a>

### CorsPolicy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| allowOriginStringMatch | [][envoy.type.matcher.v3.StringMatcher]({{< versioned_link_path fromRoot="/reference/api/envoy.type.matcher.v3.string#envoy.type.matcher.v3.StringMatcher" >}}) | repeated | Specifies string patterns that match allowed origins. An origin is allowed if any of the string matchers match. |
  | allowMethods | string |  | Specifies the content for the *access-control-allow-methods* header. |
  | allowHeaders | string |  | Specifies the content for the *access-control-allow-headers* header. |
  | exposeHeaders | string |  | Specifies the content for the *access-control-expose-headers* header. |
  | maxAge | string |  | Specifies the content for the *access-control-max-age* header. |
  | allowCredentials | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | Specifies whether the resource allows credentials. |
  | filterEnabled | [envoy.config.core.v3.RuntimeFractionalPercent]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.RuntimeFractionalPercent" >}}) |  | Specifies the % of requests for which the CORS filter is enabled.<br>If neither ``enabled``, ``filter_enabled``, nor ``shadow_enabled`` are specified, the CORS filter will be enabled for 100% of the requests.<br>If :ref:`runtime_key <envoy_api_field_config.core.v3.RuntimeFractionalPercent.runtime_key>` is specified, Envoy will lookup the runtime key to get the percentage of requests to filter. |
  | shadowEnabled | [envoy.config.core.v3.RuntimeFractionalPercent]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.RuntimeFractionalPercent" >}}) |  | Specifies the % of requests for which the CORS policies will be evaluated and tracked, but not enforced.<br>This field is intended to be used when ``filter_enabled`` and ``enabled`` are off. One of those fields have to explicitly disable the filter in order for this setting to take effect.<br>If :ref:`runtime_key <envoy_api_field_config.core.v3.RuntimeFractionalPercent.runtime_key>` is specified, Envoy will lookup the runtime key to get the percentage of requests for which it will evaluate and track the request's *Origin* to determine if it's valid but will not enforce any policies. |
  





<a name="envoy.config.route.v3.Decorator"></a>

### Decorator



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| operation | string |  | The operation name associated with the request matched to this route. If tracing is enabled, this information will be used as the span name reported for this request.<br>.. note::<br>  For ingress (inbound) requests, or egress (outbound) responses, this value may be overridden   by the :ref:`x-envoy-decorator-operation   <config_http_filters_router_x-envoy-decorator-operation>` header. |
  | propagate | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | Whether the decorated details should be propagated to the other party. The default is true. |
  





<a name="envoy.config.route.v3.DirectResponseAction"></a>

### DirectResponseAction



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| status | uint32 |  | Specifies the HTTP response status to be returned. |
  | body | [envoy.config.core.v3.DataSource]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.DataSource" >}}) |  | Specifies the content of the response body. If this setting is omitted, no body is included in the generated response.<br>.. note::<br>  Headers can be specified using *response_headers_to_add* in the enclosing   :ref:`envoy_api_msg_config.route.v3.Route`, :ref:`envoy_api_msg_config.route.v3.RouteConfiguration` or   :ref:`envoy_api_msg_config.route.v3.VirtualHost`. |
  





<a name="envoy.config.route.v3.FilterAction"></a>

### FilterAction



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| action | [google.protobuf.Any]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.any#google.protobuf.Any" >}}) |  |  |
  





<a name="envoy.config.route.v3.HeaderMatcher"></a>

### HeaderMatcher



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Specifies the name of the header in the request. |
  | exactMatch | string |  | If specified, header match will be performed based on the value of the header. |
  | safeRegexMatch | [envoy.type.matcher.v3.RegexMatcher]({{< versioned_link_path fromRoot="/reference/api/envoy.type.matcher.v3.regex#envoy.type.matcher.v3.RegexMatcher" >}}) |  | If specified, this regex string is a regular expression rule which implies the entire request header value must match the regex. The rule will not match if only a subsequence of the request header value matches the regex. |
  | rangeMatch | [envoy.type.v3.Int64Range]({{< versioned_link_path fromRoot="/reference/api/envoy.type.v3.range#envoy.type.v3.Int64Range" >}}) |  | If specified, header match will be performed based on range. The rule will match if the request header value is within this range. The entire request header value must represent an integer in base 10 notation: consisting of an optional plus or minus sign followed by a sequence of digits. The rule will not match if the header value does not represent an integer. Match will fail for empty values, floating point numbers or if only a subsequence of the header value is an integer.<br>Examples:<br>* For range [-10,0), route will match for header value -1, but not for 0, "somestring", 10.9,   "-1somestring" |
  | presentMatch | bool |  | If specified, header match will be performed based on whether the header is in the request. |
  | prefixMatch | string |  | If specified, header match will be performed based on the prefix of the header value. Note: empty prefix is not allowed, please use present_match instead.<br>Examples:<br>* The prefix *abcd* matches the value *abcdxyz*, but not for *abcxyz*. |
  | suffixMatch | string |  | If specified, header match will be performed based on the suffix of the header value. Note: empty suffix is not allowed, please use present_match instead.<br>Examples:<br>* The suffix *abcd* matches the value *xyzabcd*, but not for *xyzbcd*. |
  | containsMatch | string |  | If specified, header match will be performed based on whether the header value contains the given value or not. Note: empty contains match is not allowed, please use present_match instead.<br>Examples:<br>* The value *abcd* matches the value *xyzabcdpqr*, but not for *xyzbcdpqr*. |
  | invertMatch | bool |  | If specified, the match result will be inverted before checking. Defaults to false.<br>Examples:<br>* The regex ``\d{3}`` does not match the value *1234*, so it will match when inverted. * The range [-10,0) will match the value -1, so it will not match when inverted. |
  





<a name="envoy.config.route.v3.HedgePolicy"></a>

### HedgePolicy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| initialRequests | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | Specifies the number of initial requests that should be sent upstream. Must be at least 1. Defaults to 1. [#not-implemented-hide:] |
  | additionalRequestChance | [envoy.type.v3.FractionalPercent]({{< versioned_link_path fromRoot="/reference/api/envoy.type.v3.percent#envoy.type.v3.FractionalPercent" >}}) |  | Specifies a probability that an additional upstream request should be sent on top of what is specified by initial_requests. Defaults to 0. [#not-implemented-hide:] |
  | hedgeOnPerTryTimeout | bool |  | Indicates that a hedged request should be sent when the per-try timeout is hit. This means that a retry will be issued without resetting the original request, leaving multiple upstream requests in flight. The first request to complete successfully will be the one returned to the caller.<br>* At any time, a successful response (i.e. not triggering any of the retry-on conditions) would be returned to the client. * Before per-try timeout, an error response (per retry-on conditions) would be retried immediately or returned ot the client   if there are no more retries left. * After per-try timeout, an error response would be discarded, as a retry in the form of a hedged request is already in progress.<br>Note: For this to have effect, you must have a :ref:`RetryPolicy <envoy_api_msg_config.route.v3.RetryPolicy>` that retries at least one error code and specifies a maximum number of retries.<br>Defaults to false. |
  





<a name="envoy.config.route.v3.InternalRedirectPolicy"></a>

### InternalRedirectPolicy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| maxInternalRedirects | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | An internal redirect is not handled, unless the number of previous internal redirects that a downstream request has encountered is lower than this value. In the case where a downstream request is bounced among multiple routes by internal redirect, the first route that hits this threshold, or does not set :ref:`internal_redirect_policy <envoy_api_field_config.route.v3.RouteAction.internal_redirect_policy>` will pass the redirect back to downstream.<br>If not specified, at most one redirect will be followed. |
  | redirectResponseCodes | []uint32 | repeated | Defines what upstream response codes are allowed to trigger internal redirect. If unspecified, only 302 will be treated as internal redirect. Only 301, 302, 303, 307 and 308 are valid values. Any other codes will be ignored. |
  | predicates | [][envoy.config.core.v3.TypedExtensionConfig]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.extension#envoy.config.core.v3.TypedExtensionConfig" >}}) | repeated | Specifies a list of predicates that are queried when an upstream response is deemed to trigger an internal redirect by all other criteria. Any predicate in the list can reject the redirect, causing the response to be proxied to downstream. |
  | allowCrossSchemeRedirect | bool |  | Allow internal redirect to follow a target URI with a different scheme than the value of x-forwarded-proto. The default is false. |
  





<a name="envoy.config.route.v3.QueryParameterMatcher"></a>

### QueryParameterMatcher



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Specifies the name of a key that must be present in the requested *path*'s query string. |
  | stringMatch | [envoy.type.matcher.v3.StringMatcher]({{< versioned_link_path fromRoot="/reference/api/envoy.type.matcher.v3.string#envoy.type.matcher.v3.StringMatcher" >}}) |  | Specifies whether a query parameter value should match against a string. |
  | presentMatch | bool |  | Specifies whether a query parameter should be present. |
  





<a name="envoy.config.route.v3.RateLimit"></a>

### RateLimit



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| stage | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | Refers to the stage set in the filter. The rate limit configuration only applies to filters with the same stage number. The default stage number is 0.<br>.. note::<br>  The filter supports a range of 0 - 10 inclusively for stage numbers. |
  | disableKey | string |  | The key to be set in runtime to disable this rate limit configuration. |
  | actions | [][envoy.config.route.v3.RateLimit.Action]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RateLimit.Action" >}}) | repeated | A list of actions that are to be applied for this rate limit configuration. Order matters as the actions are processed sequentially and the descriptor is composed by appending descriptor entries in that sequence. If an action cannot append a descriptor entry, no descriptor is generated for the configuration. See :ref:`composing actions <config_http_filters_rate_limit_composing_actions>` for additional documentation. |
  | limit | [envoy.config.route.v3.RateLimit.Override]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RateLimit.Override" >}}) |  | An optional limit override to be appended to the descriptor produced by this rate limit configuration. If the override value is invalid or cannot be resolved from metadata, no override is provided. See :ref:`rate limit override <config_http_filters_rate_limit_rate_limit_override>` for more information. |
  





<a name="envoy.config.route.v3.RateLimit.Action"></a>

### RateLimit.Action



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sourceCluster | [envoy.config.route.v3.RateLimit.Action.SourceCluster]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RateLimit.Action.SourceCluster" >}}) |  | Rate limit on source cluster. |
  | destinationCluster | [envoy.config.route.v3.RateLimit.Action.DestinationCluster]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RateLimit.Action.DestinationCluster" >}}) |  | Rate limit on destination cluster. |
  | requestHeaders | [envoy.config.route.v3.RateLimit.Action.RequestHeaders]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RateLimit.Action.RequestHeaders" >}}) |  | Rate limit on request headers. |
  | remoteAddress | [envoy.config.route.v3.RateLimit.Action.RemoteAddress]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RateLimit.Action.RemoteAddress" >}}) |  | Rate limit on remote address. |
  | genericKey | [envoy.config.route.v3.RateLimit.Action.GenericKey]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RateLimit.Action.GenericKey" >}}) |  | Rate limit on a generic key. |
  | headerValueMatch | [envoy.config.route.v3.RateLimit.Action.HeaderValueMatch]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RateLimit.Action.HeaderValueMatch" >}}) |  | Rate limit on the existence of request headers. |
  | dynamicMetadata | [envoy.config.route.v3.RateLimit.Action.DynamicMetaData]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RateLimit.Action.DynamicMetaData" >}}) |  | Rate limit on dynamic metadata.<br>.. attention::   This field has been deprecated in favor of the :ref:`metadata <envoy_api_field_config.route.v3.RateLimit.Action.metadata>` field |
  | metadata | [envoy.config.route.v3.RateLimit.Action.MetaData]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RateLimit.Action.MetaData" >}}) |  | Rate limit on metadata. |
  | extension | [envoy.config.core.v3.TypedExtensionConfig]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.extension#envoy.config.core.v3.TypedExtensionConfig" >}}) |  | Rate limit descriptor extension. See the rate limit descriptor extensions documentation. |
  





<a name="envoy.config.route.v3.RateLimit.Action.DestinationCluster"></a>

### RateLimit.Action.DestinationCluster







<a name="envoy.config.route.v3.RateLimit.Action.DynamicMetaData"></a>

### RateLimit.Action.DynamicMetaData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| descriptorKey | string |  | The key to use in the descriptor entry. |
  | metadataKey | [envoy.type.metadata.v3.MetadataKey]({{< versioned_link_path fromRoot="/reference/api/envoy.type.metadata.v3.metadata#envoy.type.metadata.v3.MetadataKey" >}}) |  | Metadata struct that defines the key and path to retrieve the string value. A match will only happen if the value in the dynamic metadata is of type string. |
  | defaultValue | string |  | An optional value to use if *metadata_key* is empty. If not set and no value is present under the metadata_key then no descriptor is generated. |
  





<a name="envoy.config.route.v3.RateLimit.Action.GenericKey"></a>

### RateLimit.Action.GenericKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| descriptorValue | string |  | The value to use in the descriptor entry. |
  | descriptorKey | string |  | An optional key to use in the descriptor entry. If not set it defaults to 'generic_key' as the descriptor key. |
  





<a name="envoy.config.route.v3.RateLimit.Action.HeaderValueMatch"></a>

### RateLimit.Action.HeaderValueMatch



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| descriptorValue | string |  | The value to use in the descriptor entry. |
  | expectMatch | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | If set to true, the action will append a descriptor entry when the request matches the headers. If set to false, the action will append a descriptor entry when the request does not match the headers. The default value is true. |
  | headers | [][envoy.config.route.v3.HeaderMatcher]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.HeaderMatcher" >}}) | repeated | Specifies a set of headers that the rate limit action should match on. The action will check the request’s headers against all the specified headers in the config. A match will happen if all the headers in the config are present in the request with the same values (or based on presence if the value field is not in the config). |
  





<a name="envoy.config.route.v3.RateLimit.Action.MetaData"></a>

### RateLimit.Action.MetaData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| descriptorKey | string |  | The key to use in the descriptor entry. |
  | metadataKey | [envoy.type.metadata.v3.MetadataKey]({{< versioned_link_path fromRoot="/reference/api/envoy.type.metadata.v3.metadata#envoy.type.metadata.v3.MetadataKey" >}}) |  | Metadata struct that defines the key and path to retrieve the string value. A match will only happen if the value in the metadata is of type string. |
  | defaultValue | string |  | An optional value to use if *metadata_key* is empty. If not set and no value is present under the metadata_key then no descriptor is generated. |
  | source | [envoy.config.route.v3.RateLimit.Action.MetaData.Source]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RateLimit.Action.MetaData.Source" >}}) |  | Source of metadata |
  





<a name="envoy.config.route.v3.RateLimit.Action.RemoteAddress"></a>

### RateLimit.Action.RemoteAddress







<a name="envoy.config.route.v3.RateLimit.Action.RequestHeaders"></a>

### RateLimit.Action.RequestHeaders



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headerName | string |  | The header name to be queried from the request headers. The header’s value is used to populate the value of the descriptor entry for the descriptor_key. |
  | descriptorKey | string |  | The key to use in the descriptor entry. |
  | skipIfAbsent | bool |  | If set to true, Envoy skips the descriptor while calling rate limiting service when header is not present in the request. By default it skips calling the rate limiting service if this header is not present in the request. |
  





<a name="envoy.config.route.v3.RateLimit.Action.SourceCluster"></a>

### RateLimit.Action.SourceCluster







<a name="envoy.config.route.v3.RateLimit.Override"></a>

### RateLimit.Override



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| dynamicMetadata | [envoy.config.route.v3.RateLimit.Override.DynamicMetadata]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RateLimit.Override.DynamicMetadata" >}}) |  | Limit override from dynamic metadata. |
  





<a name="envoy.config.route.v3.RateLimit.Override.DynamicMetadata"></a>

### RateLimit.Override.DynamicMetadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadataKey | [envoy.type.metadata.v3.MetadataKey]({{< versioned_link_path fromRoot="/reference/api/envoy.type.metadata.v3.metadata#envoy.type.metadata.v3.MetadataKey" >}}) |  | Metadata struct that defines the key and path to retrieve the struct value. The value must be a struct containing an integer "requests_per_unit" property and a "unit" property with a value parseable to :ref:`RateLimitUnit enum <envoy_api_enum_type.v3.RateLimitUnit>` |
  





<a name="envoy.config.route.v3.RedirectAction"></a>

### RedirectAction



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| httpsRedirect | bool |  | The scheme portion of the URL will be swapped with "https". |
  | schemeRedirect | string |  | The scheme portion of the URL will be swapped with this value. |
  | hostRedirect | string |  | The host portion of the URL will be swapped with this value. |
  | portRedirect | uint32 |  | The port value of the URL will be swapped with this value. |
  | pathRedirect | string |  | The path portion of the URL will be swapped with this value. Please note that query string in path_redirect will override the request's query string and will not be stripped.<br>For example, let's say we have the following routes:<br>- match: { path: "/old-path-1" }   redirect: { path_redirect: "/new-path-1" } - match: { path: "/old-path-2" }   redirect: { path_redirect: "/new-path-2", strip-query: "true" } - match: { path: "/old-path-3" }   redirect: { path_redirect: "/new-path-3?foo=1", strip_query: "true" }<br>1. if request uri is "/old-path-1?bar=1", users will be redirected to "/new-path-1?bar=1" 2. if request uri is "/old-path-2?bar=1", users will be redirected to "/new-path-2" 3. if request uri is "/old-path-3?bar=1", users will be redirected to "/new-path-3?foo=1" |
  | prefixRewrite | string |  | Indicates that during redirection, the matched prefix (or path) should be swapped with this value. This option allows redirect URLs be dynamically created based on the request.<br>.. attention::<br>  Pay attention to the use of trailing slashes as mentioned in   :ref:`RouteAction's prefix_rewrite <envoy_api_field_config.route.v3.RouteAction.prefix_rewrite>`. |
  | regexRewrite | [envoy.type.matcher.v3.RegexMatchAndSubstitute]({{< versioned_link_path fromRoot="/reference/api/envoy.type.matcher.v3.regex#envoy.type.matcher.v3.RegexMatchAndSubstitute" >}}) |  | Indicates that during redirect, portions of the path that match the pattern should be rewritten, even allowing the substitution of capture groups from the pattern into the new path as specified by the rewrite substitution string. This is useful to allow application paths to be rewritten in a way that is aware of segments with variable content like identifiers.<br>Examples using Google's `RE2 <https://github.com/google/re2>`_ engine:<br>* The path pattern ``^/service/([^/]+)(/.*)$`` paired with a substitution   string of ``\2/instance/\1`` would transform ``/service/foo/v1/api``   into ``/v1/api/instance/foo``.<br>* The pattern ``one`` paired with a substitution string of ``two`` would   transform ``/xxx/one/yyy/one/zzz`` into ``/xxx/two/yyy/two/zzz``.<br>* The pattern ``^(.*?)one(.*)$`` paired with a substitution string of   ``\1two\2`` would replace only the first occurrence of ``one``,   transforming path ``/xxx/one/yyy/one/zzz`` into ``/xxx/two/yyy/one/zzz``.<br>* The pattern ``(?i)/xxx/`` paired with a substitution string of ``/yyy/``   would do a case-insensitive match and transform path ``/aaa/XxX/bbb`` to   ``/aaa/yyy/bbb``. |
  | responseCode | [envoy.config.route.v3.RedirectAction.RedirectResponseCode]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RedirectAction.RedirectResponseCode" >}}) |  | The HTTP status code to use in the redirect response. The default response code is MOVED_PERMANENTLY (301). |
  | stripQuery | bool |  | Indicates that during redirection, the query portion of the URL will be removed. Default value is false. |
  





<a name="envoy.config.route.v3.RetryPolicy"></a>

### RetryPolicy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| retryOn | string |  | Specifies the conditions under which retry takes place. These are the same conditions documented for :ref:`config_http_filters_router_x-envoy-retry-on` and :ref:`config_http_filters_router_x-envoy-retry-grpc-on`. |
  | numRetries | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | Specifies the allowed number of retries. This parameter is optional and defaults to 1. These are the same conditions documented for :ref:`config_http_filters_router_x-envoy-max-retries`. |
  | perTryTimeout | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Specifies a non-zero upstream timeout per retry attempt. This parameter is optional. The same conditions documented for :ref:`config_http_filters_router_x-envoy-upstream-rq-per-try-timeout-ms` apply.<br>.. note::<br>  If left unspecified, Envoy will use the global   :ref:`route timeout <envoy_api_field_config.route.v3.RouteAction.timeout>` for the request.   Consequently, when using a :ref:`5xx <config_http_filters_router_x-envoy-retry-on>` based   retry policy, a request that times out will not be retried as the total timeout budget   would have been exhausted. |
  | retryPriority | [envoy.config.route.v3.RetryPolicy.RetryPriority]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RetryPolicy.RetryPriority" >}}) |  | Specifies an implementation of a RetryPriority which is used to determine the distribution of load across priorities used for retries. Refer to :ref:`retry plugin configuration <arch_overview_http_retry_plugins>` for more details. |
  | retryHostPredicate | [][envoy.config.route.v3.RetryPolicy.RetryHostPredicate]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RetryPolicy.RetryHostPredicate" >}}) | repeated | Specifies a collection of RetryHostPredicates that will be consulted when selecting a host for retries. If any of the predicates reject the host, host selection will be reattempted. Refer to :ref:`retry plugin configuration <arch_overview_http_retry_plugins>` for more details. |
  | hostSelectionRetryMaxAttempts | int64 |  | The maximum number of times host selection will be reattempted before giving up, at which point the host that was last selected will be routed to. If unspecified, this will default to retrying once. |
  | retriableStatusCodes | []uint32 | repeated | HTTP status codes that should trigger a retry in addition to those specified by retry_on. |
  | retryBackOff | [envoy.config.route.v3.RetryPolicy.RetryBackOff]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RetryPolicy.RetryBackOff" >}}) |  | Specifies parameters that control exponential retry back off. This parameter is optional, in which case the default base interval is 25 milliseconds or, if set, the current value of the `upstream.base_retry_backoff_ms` runtime parameter. The default maximum interval is 10 times the base interval. The documentation for :ref:`config_http_filters_router_x-envoy-max-retries` describes Envoy's back-off algorithm. |
  | rateLimitedRetryBackOff | [envoy.config.route.v3.RetryPolicy.RateLimitedRetryBackOff]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RetryPolicy.RateLimitedRetryBackOff" >}}) |  | Specifies parameters that control a retry back-off strategy that is used when the request is rate limited by the upstream server. The server may return a response header like ``Retry-After`` or ``X-RateLimit-Reset`` to provide feedback to the client on how long to wait before retrying. If configured, this back-off strategy will be used instead of the default exponential back off strategy (configured using `retry_back_off`) whenever a response includes the matching headers. |
  | retriableHeaders | [][envoy.config.route.v3.HeaderMatcher]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.HeaderMatcher" >}}) | repeated | HTTP response headers that trigger a retry if present in the response. A retry will be triggered if any of the header matches match the upstream response headers. The field is only consulted if 'retriable-headers' retry policy is active. |
  | retriableRequestHeaders | [][envoy.config.route.v3.HeaderMatcher]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.HeaderMatcher" >}}) | repeated | HTTP headers which must be present in the request for retries to be attempted. |
  





<a name="envoy.config.route.v3.RetryPolicy.RateLimitedRetryBackOff"></a>

### RetryPolicy.RateLimitedRetryBackOff



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resetHeaders | [][envoy.config.route.v3.RetryPolicy.ResetHeader]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RetryPolicy.ResetHeader" >}}) | repeated | Specifies the reset headers (like ``Retry-After`` or ``X-RateLimit-Reset``) to match against the response. Headers are tried in order, and matched case insensitive. The first header to be parsed successfully is used. If no headers match the default exponential back-off is used instead. |
  | maxInterval | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Specifies the maximum back off interval that Envoy will allow. If a reset header contains an interval longer than this then it will be discarded and the next header will be tried. Defaults to 300 seconds. |
  





<a name="envoy.config.route.v3.RetryPolicy.ResetHeader"></a>

### RetryPolicy.ResetHeader



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The name of the reset header.<br>.. note::<br>  If the header appears multiple times only the first value is used. |
  | format | [envoy.config.route.v3.RetryPolicy.ResetHeaderFormat]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RetryPolicy.ResetHeaderFormat" >}}) |  | The format of the reset header. |
  





<a name="envoy.config.route.v3.RetryPolicy.RetryBackOff"></a>

### RetryPolicy.RetryBackOff



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| baseInterval | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Specifies the base interval between retries. This parameter is required and must be greater than zero. Values less than 1 ms are rounded up to 1 ms. See :ref:`config_http_filters_router_x-envoy-max-retries` for a discussion of Envoy's back-off algorithm. |
  | maxInterval | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Specifies the maximum interval between retries. This parameter is optional, but must be greater than or equal to the `base_interval` if set. The default is 10 times the `base_interval`. See :ref:`config_http_filters_router_x-envoy-max-retries` for a discussion of Envoy's back-off algorithm. |
  





<a name="envoy.config.route.v3.RetryPolicy.RetryHostPredicate"></a>

### RetryPolicy.RetryHostPredicate



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  |  |
  | typedConfig | [google.protobuf.Any]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.any#google.protobuf.Any" >}}) |  |  |
  





<a name="envoy.config.route.v3.RetryPolicy.RetryPriority"></a>

### RetryPolicy.RetryPriority



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  |  |
  | typedConfig | [google.protobuf.Any]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.any#google.protobuf.Any" >}}) |  |  |
  





<a name="envoy.config.route.v3.Route"></a>

### Route



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Name for the route. |
  | match | [envoy.config.route.v3.RouteMatch]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RouteMatch" >}}) |  | Route matching parameters. |
  | route | [envoy.config.route.v3.RouteAction]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RouteAction" >}}) |  | Route request to some upstream cluster. |
  | redirect | [envoy.config.route.v3.RedirectAction]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RedirectAction" >}}) |  | Return a redirect. |
  | directResponse | [envoy.config.route.v3.DirectResponseAction]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.DirectResponseAction" >}}) |  | Return an arbitrary HTTP response directly, without proxying. |
  | filterAction | [envoy.config.route.v3.FilterAction]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.FilterAction" >}}) |  | [#not-implemented-hide:] If true, a filter will define the action (e.g., it could dynamically generate the RouteAction). [#comment: TODO(samflattery): Remove cleanup in route_fuzz_test.cc when implemented] |
  | metadata | [envoy.config.core.v3.Metadata]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.Metadata" >}}) |  | The Metadata field can be used to provide additional information about the route. It can be used for configuration, stats, and logging. The metadata should go under the filter namespace that will need it. For instance, if the metadata is intended for the Router filter, the filter name should be specified as *envoy.filters.http.router*. |
  | decorator | [envoy.config.route.v3.Decorator]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.Decorator" >}}) |  | Decorator for the matched route. |
  | typedPerFilterConfig | [][envoy.config.route.v3.Route.TypedPerFilterConfigEntry]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.Route.TypedPerFilterConfigEntry" >}}) | repeated | The typed_per_filter_config field can be used to provide route-specific configurations for filters. The key should match the filter name, such as *envoy.filters.http.buffer* for the HTTP buffer filter. Use of this field is filter specific; see the :ref:`HTTP filter documentation <config_http_filters>` for if and how it is utilized. |
  | requestHeadersToAdd | [][envoy.config.core.v3.HeaderValueOption]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.HeaderValueOption" >}}) | repeated | Specifies a set of headers that will be added to requests matching this route. Headers specified at this level are applied before headers from the enclosing :ref:`envoy_api_msg_config.route.v3.VirtualHost` and :ref:`envoy_api_msg_config.route.v3.RouteConfiguration`. For more information, including details on header value syntax, see the documentation on :ref:`custom request headers <config_http_conn_man_headers_custom_request_headers>`. |
  | requestHeadersToRemove | []string | repeated | Specifies a list of HTTP headers that should be removed from each request matching this route. |
  | responseHeadersToAdd | [][envoy.config.core.v3.HeaderValueOption]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.HeaderValueOption" >}}) | repeated | Specifies a set of headers that will be added to responses to requests matching this route. Headers specified at this level are applied before headers from the enclosing :ref:`envoy_api_msg_config.route.v3.VirtualHost` and :ref:`envoy_api_msg_config.route.v3.RouteConfiguration`. For more information, including details on header value syntax, see the documentation on :ref:`custom request headers <config_http_conn_man_headers_custom_request_headers>`. |
  | responseHeadersToRemove | []string | repeated | Specifies a list of HTTP headers that should be removed from each response to requests matching this route. |
  | tracing | [envoy.config.route.v3.Tracing]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.Tracing" >}}) |  | Presence of the object defines whether the connection manager's tracing configuration is overridden by this route specific instance. |
  | perRequestBufferLimitBytes | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | The maximum bytes which will be buffered for retries and shadowing. If set, the bytes actually buffered will be the minimum value of this and the listener per_connection_buffer_limit_bytes. |
  





<a name="envoy.config.route.v3.Route.TypedPerFilterConfigEntry"></a>

### Route.TypedPerFilterConfigEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [google.protobuf.Any]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.any#google.protobuf.Any" >}}) |  |  |
  





<a name="envoy.config.route.v3.RouteAction"></a>

### RouteAction



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cluster | string |  | Indicates the upstream cluster to which the request should be routed to. |
  | clusterHeader | string |  | Envoy will determine the cluster to route to by reading the value of the HTTP header named by cluster_header from the request headers. If the header is not found or the referenced cluster does not exist, Envoy will return a 404 response.<br>.. attention::<br>  Internally, Envoy always uses the HTTP/2 *:authority* header to represent the HTTP/1   *Host* header. Thus, if attempting to match on *Host*, match on *:authority* instead.<br>.. note::<br>  If the header appears multiple times only the first value is used. |
  | weightedClusters | [envoy.config.route.v3.WeightedCluster]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.WeightedCluster" >}}) |  | Multiple upstream clusters can be specified for a given route. The request is routed to one of the upstream clusters based on weights assigned to each cluster. See :ref:`traffic splitting <config_http_conn_man_route_table_traffic_splitting_split>` for additional documentation. |
  | clusterNotFoundResponseCode | [envoy.config.route.v3.RouteAction.ClusterNotFoundResponseCode]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RouteAction.ClusterNotFoundResponseCode" >}}) |  | The HTTP status code to use when configured cluster is not found. The default response code is 503 Service Unavailable. |
  | metadataMatch | [envoy.config.core.v3.Metadata]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.Metadata" >}}) |  | Optional endpoint metadata match criteria used by the subset load balancer. Only endpoints in the upstream cluster with metadata matching what's set in this field will be considered for load balancing. If using :ref:`weighted_clusters <envoy_api_field_config.route.v3.RouteAction.weighted_clusters>`, metadata will be merged, with values provided there taking precedence. The filter name should be specified as *envoy.lb*. |
  | prefixRewrite | string |  | Indicates that during forwarding, the matched prefix (or path) should be swapped with this value. This option allows application URLs to be rooted at a different path from those exposed at the reverse proxy layer. The router filter will place the original path before rewrite into the :ref:`x-envoy-original-path <config_http_filters_router_x-envoy-original-path>` header.<br>Only one of *prefix_rewrite* or :ref:`regex_rewrite <envoy_api_field_config.route.v3.RouteAction.regex_rewrite>` may be specified.<br>.. attention::<br>  Pay careful attention to the use of trailing slashes in the   :ref:`route's match <envoy_api_field_config.route.v3.Route.match>` prefix value.   Stripping a prefix from a path requires multiple Routes to handle all cases. For example,   rewriting */prefix* to */* and */prefix/etc* to */etc* cannot be done in a single   :ref:`Route <envoy_api_msg_config.route.v3.Route>`, as shown by the below config entries:<br>  .. code-block:: yaml<br>    - match:         prefix: "/prefix/"       route:         prefix_rewrite: "/"     - match:         prefix: "/prefix"       route:         prefix_rewrite: "/"<br>  Having above entries in the config, requests to */prefix* will be stripped to */*, while   requests to */prefix/etc* will be stripped to */etc*. |
  | regexRewrite | [envoy.type.matcher.v3.RegexMatchAndSubstitute]({{< versioned_link_path fromRoot="/reference/api/envoy.type.matcher.v3.regex#envoy.type.matcher.v3.RegexMatchAndSubstitute" >}}) |  | Indicates that during forwarding, portions of the path that match the pattern should be rewritten, even allowing the substitution of capture groups from the pattern into the new path as specified by the rewrite substitution string. This is useful to allow application paths to be rewritten in a way that is aware of segments with variable content like identifiers. The router filter will place the original path as it was before the rewrite into the :ref:`x-envoy-original-path <config_http_filters_router_x-envoy-original-path>` header.<br>Only one of :ref:`prefix_rewrite <envoy_api_field_config.route.v3.RouteAction.prefix_rewrite>` or *regex_rewrite* may be specified.<br>Examples using Google's `RE2 <https://github.com/google/re2>`_ engine:<br>* The path pattern ``^/service/([^/]+)(/.*)$`` paired with a substitution   string of ``\2/instance/\1`` would transform ``/service/foo/v1/api``   into ``/v1/api/instance/foo``.<br>* The pattern ``one`` paired with a substitution string of ``two`` would   transform ``/xxx/one/yyy/one/zzz`` into ``/xxx/two/yyy/two/zzz``.<br>* The pattern ``^(.*?)one(.*)$`` paired with a substitution string of   ``\1two\2`` would replace only the first occurrence of ``one``,   transforming path ``/xxx/one/yyy/one/zzz`` into ``/xxx/two/yyy/one/zzz``.<br>* The pattern ``(?i)/xxx/`` paired with a substitution string of ``/yyy/``   would do a case-insensitive match and transform path ``/aaa/XxX/bbb`` to   ``/aaa/yyy/bbb``. |
  | hostRewriteLiteral | string |  | Indicates that during forwarding, the host header will be swapped with this value. |
  | autoHostRewrite | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | Indicates that during forwarding, the host header will be swapped with the hostname of the upstream host chosen by the cluster manager. This option is applicable only when the destination cluster for a route is of type *strict_dns* or *logical_dns*. Setting this to true with other cluster types has no effect. |
  | hostRewriteHeader | string |  | Indicates that during forwarding, the host header will be swapped with the content of given downstream or :ref:`custom <config_http_conn_man_headers_custom_request_headers>` header. If header value is empty, host header is left intact.<br>.. attention::<br>  Pay attention to the potential security implications of using this option. Provided header   must come from trusted source.<br>.. note::<br>  If the header appears multiple times only the first value is used. |
  | hostRewritePathRegex | [envoy.type.matcher.v3.RegexMatchAndSubstitute]({{< versioned_link_path fromRoot="/reference/api/envoy.type.matcher.v3.regex#envoy.type.matcher.v3.RegexMatchAndSubstitute" >}}) |  | Indicates that during forwarding, the host header will be swapped with the result of the regex substitution executed on path value with query and fragment removed. This is useful for transitioning variable content between path segment and subdomain.<br>For example with the following config:<br>  .. code-block:: yaml<br>    host_rewrite_path_regex:       pattern:         google_re2: {}         regex: "^/(.+)/.+$"       substitution: \1<br>Would rewrite the host header to `envoyproxy.io` given the path `/envoyproxy.io/some/path`. |
  | timeout | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Specifies the upstream timeout for the route. If not specified, the default is 15s. This spans between the point at which the entire downstream request (i.e. end-of-stream) has been processed and when the upstream response has been completely processed. A value of 0 will disable the route's timeout.<br>.. note::<br>  This timeout includes all retries. See also   :ref:`config_http_filters_router_x-envoy-upstream-rq-timeout-ms`,   :ref:`config_http_filters_router_x-envoy-upstream-rq-per-try-timeout-ms`, and the   :ref:`retry overview <arch_overview_http_routing_retry>`. |
  | idleTimeout | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Specifies the idle timeout for the route. If not specified, there is no per-route idle timeout, although the connection manager wide :ref:`stream_idle_timeout <envoy_api_field_extensions.filters.network.http_connection_manager.v3.HttpConnectionManager.stream_idle_timeout>` will still apply. A value of 0 will completely disable the route's idle timeout, even if a connection manager stream idle timeout is configured.<br>The idle timeout is distinct to :ref:`timeout <envoy_api_field_config.route.v3.RouteAction.timeout>`, which provides an upper bound on the upstream response time; :ref:`idle_timeout <envoy_api_field_config.route.v3.RouteAction.idle_timeout>` instead bounds the amount of time the request's stream may be idle.<br>After header decoding, the idle timeout will apply on downstream and upstream request events. Each time an encode/decode event for headers or data is processed for the stream, the timer will be reset. If the timeout fires, the stream is terminated with a 408 Request Timeout error code if no upstream response header has been received, otherwise a stream reset occurs.<br>If the :ref:`overload action <config_overload_manager_overload_actions>` "envoy.overload_actions.reduce_timeouts" is configured, this timeout is scaled according to the value for :ref:`HTTP_DOWNSTREAM_STREAM_IDLE <envoy_api_enum_value_config.overload.v3.ScaleTimersOverloadActionConfig.TimerType.HTTP_DOWNSTREAM_STREAM_IDLE>`. |
  | retryPolicy | [envoy.config.route.v3.RetryPolicy]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RetryPolicy" >}}) |  | Indicates that the route has a retry policy. Note that if this is set, it'll take precedence over the virtual host level retry policy entirely (e.g.: policies are not merged, most internal one becomes the enforced policy). |
  | retryPolicyTypedConfig | [google.protobuf.Any]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.any#google.protobuf.Any" >}}) |  | [#not-implemented-hide:] Specifies the configuration for retry policy extension. Note that if this is set, it'll take precedence over the virtual host level retry policy entirely (e.g.: policies are not merged, most internal one becomes the enforced policy). :ref:`Retry policy <envoy_api_field_config.route.v3.VirtualHost.retry_policy>` should not be set if this field is used. |
  | requestMirrorPolicies | [][envoy.config.route.v3.RouteAction.RequestMirrorPolicy]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RouteAction.RequestMirrorPolicy" >}}) | repeated | Indicates that the route has request mirroring policies. |
  | priority | [envoy.config.core.v3.RoutingPriority]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.RoutingPriority" >}}) |  | Optionally specifies the :ref:`routing priority <arch_overview_http_routing_priority>`. |
  | rateLimits | [][envoy.config.route.v3.RateLimit]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RateLimit" >}}) | repeated | Specifies a set of rate limit configurations that could be applied to the route. |
  | includeVhRateLimits | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | Specifies if the rate limit filter should include the virtual host rate limits. By default, if the route configured rate limits, the virtual host :ref:`rate_limits <envoy_api_field_config.route.v3.VirtualHost.rate_limits>` are not applied to the request.<br>This field is deprecated. Please use :ref:`vh_rate_limits <envoy_v3_api_field_extensions.filters.http.ratelimit.v3.RateLimitPerRoute.vh_rate_limits>` |
  | hashPolicy | [][envoy.config.route.v3.RouteAction.HashPolicy]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RouteAction.HashPolicy" >}}) | repeated | Specifies a list of hash policies to use for ring hash load balancing. Each hash policy is evaluated individually and the combined result is used to route the request. The method of combination is deterministic such that identical lists of hash policies will produce the same hash. Since a hash policy examines specific parts of a request, it can fail to produce a hash (i.e. if the hashed header is not present). If (and only if) all configured hash policies fail to generate a hash, no hash will be produced for the route. In this case, the behavior is the same as if no hash policies were specified (i.e. the ring hash load balancer will choose a random backend). If a hash policy has the "terminal" attribute set to true, and there is already a hash generated, the hash is returned immediately, ignoring the rest of the hash policy list. |
  | cors | [envoy.config.route.v3.CorsPolicy]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.CorsPolicy" >}}) |  | Indicates that the route has a CORS policy. |
  | maxGrpcTimeout | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Deprecated by :ref:`grpc_timeout_header_max <envoy_api_field_config.route.v3.RouteAction.MaxStreamDuration.grpc_timeout_header_max>` If present, and the request is a gRPC request, use the `grpc-timeout header <https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md>`_, or its default value (infinity) instead of :ref:`timeout <envoy_api_field_config.route.v3.RouteAction.timeout>`, but limit the applied timeout to the maximum value specified here. If configured as 0, the maximum allowed timeout for gRPC requests is infinity. If not configured at all, the `grpc-timeout` header is not used and gRPC requests time out like any other requests using :ref:`timeout <envoy_api_field_config.route.v3.RouteAction.timeout>` or its default. This can be used to prevent unexpected upstream request timeouts due to potentially long time gaps between gRPC request and response in gRPC streaming mode.<br>.. note::<br>   If a timeout is specified using :ref:`config_http_filters_router_x-envoy-upstream-rq-timeout-ms`, it takes    precedence over `grpc-timeout header <https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md>`_, when    both are present. See also    :ref:`config_http_filters_router_x-envoy-upstream-rq-timeout-ms`,    :ref:`config_http_filters_router_x-envoy-upstream-rq-per-try-timeout-ms`, and the    :ref:`retry overview <arch_overview_http_routing_retry>`. |
  | grpcTimeoutOffset | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Deprecated by :ref:`grpc_timeout_header_offset <envoy_api_field_config.route.v3.RouteAction.MaxStreamDuration.grpc_timeout_header_offset>`. If present, Envoy will adjust the timeout provided by the `grpc-timeout` header by subtracting the provided duration from the header. This is useful in allowing Envoy to set its global timeout to be less than that of the deadline imposed by the calling client, which makes it more likely that Envoy will handle the timeout instead of having the call canceled by the client. The offset will only be applied if the provided grpc_timeout is greater than the offset. This ensures that the offset will only ever decrease the timeout and never set it to 0 (meaning infinity). |
  | upgradeConfigs | [][envoy.config.route.v3.RouteAction.UpgradeConfig]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RouteAction.UpgradeConfig" >}}) | repeated |  |
  | internalRedirectPolicy | [envoy.config.route.v3.InternalRedirectPolicy]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.InternalRedirectPolicy" >}}) |  | If present, Envoy will try to follow an upstream redirect response instead of proxying the response back to the downstream. An upstream redirect response is defined by :ref:`redirect_response_codes <envoy_api_field_config.route.v3.InternalRedirectPolicy.redirect_response_codes>`. |
  | internalRedirectAction | [envoy.config.route.v3.RouteAction.InternalRedirectAction]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RouteAction.InternalRedirectAction" >}}) |  |  |
  | maxInternalRedirects | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | An internal redirect is handled, iff the number of previous internal redirects that a downstream request has encountered is lower than this value, and :ref:`internal_redirect_action <envoy_api_field_config.route.v3.RouteAction.internal_redirect_action>` is set to :ref:`HANDLE_INTERNAL_REDIRECT <envoy_api_enum_value_config.route.v3.RouteAction.InternalRedirectAction.HANDLE_INTERNAL_REDIRECT>` In the case where a downstream request is bounced among multiple routes by internal redirect, the first route that hits this threshold, or has :ref:`internal_redirect_action <envoy_api_field_config.route.v3.RouteAction.internal_redirect_action>` set to :ref:`PASS_THROUGH_INTERNAL_REDIRECT <envoy_api_enum_value_config.route.v3.RouteAction.InternalRedirectAction.PASS_THROUGH_INTERNAL_REDIRECT>` will pass the redirect back to downstream.<br>If not specified, at most one redirect will be followed. |
  | hedgePolicy | [envoy.config.route.v3.HedgePolicy]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.HedgePolicy" >}}) |  | Indicates that the route has a hedge policy. Note that if this is set, it'll take precedence over the virtual host level hedge policy entirely (e.g.: policies are not merged, most internal one becomes the enforced policy). |
  | maxStreamDuration | [envoy.config.route.v3.RouteAction.MaxStreamDuration]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RouteAction.MaxStreamDuration" >}}) |  | Specifies the maximum stream duration for this route. |
  





<a name="envoy.config.route.v3.RouteAction.HashPolicy"></a>

### RouteAction.HashPolicy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| header | [envoy.config.route.v3.RouteAction.HashPolicy.Header]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RouteAction.HashPolicy.Header" >}}) |  | Header hash policy. |
  | cookie | [envoy.config.route.v3.RouteAction.HashPolicy.Cookie]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RouteAction.HashPolicy.Cookie" >}}) |  | Cookie hash policy. |
  | connectionProperties | [envoy.config.route.v3.RouteAction.HashPolicy.ConnectionProperties]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RouteAction.HashPolicy.ConnectionProperties" >}}) |  | Connection properties hash policy. |
  | queryParameter | [envoy.config.route.v3.RouteAction.HashPolicy.QueryParameter]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RouteAction.HashPolicy.QueryParameter" >}}) |  | Query parameter hash policy. |
  | filterState | [envoy.config.route.v3.RouteAction.HashPolicy.FilterState]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RouteAction.HashPolicy.FilterState" >}}) |  | Filter state hash policy. |
  | terminal | bool |  | The flag that short-circuits the hash computing. This field provides a 'fallback' style of configuration: "if a terminal policy doesn't work, fallback to rest of the policy list", it saves time when the terminal policy works.<br>If true, and there is already a hash computed, ignore rest of the list of hash polices. For example, if the following hash methods are configured:<br> ========= ========  specifier terminal  ========= ========  Header A  true  Header B  false  Header C  false  ========= ========<br>The generateHash process ends if policy "header A" generates a hash, as it's a terminal policy. |
  





<a name="envoy.config.route.v3.RouteAction.HashPolicy.ConnectionProperties"></a>

### RouteAction.HashPolicy.ConnectionProperties



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sourceIp | bool |  | Hash on source IP address. |
  





<a name="envoy.config.route.v3.RouteAction.HashPolicy.Cookie"></a>

### RouteAction.HashPolicy.Cookie



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The name of the cookie that will be used to obtain the hash key. If the cookie is not present and ttl below is not set, no hash will be produced. |
  | ttl | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | If specified, a cookie with the TTL will be generated if the cookie is not present. If the TTL is present and zero, the generated cookie will be a session cookie. |
  | path | string |  | The name of the path for the cookie. If no path is specified here, no path will be set for the cookie. |
  





<a name="envoy.config.route.v3.RouteAction.HashPolicy.FilterState"></a>

### RouteAction.HashPolicy.FilterState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  | The name of the Object in the per-request filterState, which is an Envoy::Http::Hashable object. If there is no data associated with the key, or the stored object is not Envoy::Http::Hashable, no hash will be produced. |
  





<a name="envoy.config.route.v3.RouteAction.HashPolicy.Header"></a>

### RouteAction.HashPolicy.Header



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headerName | string |  | The name of the request header that will be used to obtain the hash key. If the request header is not present, no hash will be produced. |
  | regexRewrite | [envoy.type.matcher.v3.RegexMatchAndSubstitute]({{< versioned_link_path fromRoot="/reference/api/envoy.type.matcher.v3.regex#envoy.type.matcher.v3.RegexMatchAndSubstitute" >}}) |  | If specified, the request header value will be rewritten and used to produce the hash key. |
  





<a name="envoy.config.route.v3.RouteAction.HashPolicy.QueryParameter"></a>

### RouteAction.HashPolicy.QueryParameter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The name of the URL query parameter that will be used to obtain the hash key. If the parameter is not present, no hash will be produced. Query parameter names are case-sensitive. |
  





<a name="envoy.config.route.v3.RouteAction.MaxStreamDuration"></a>

### RouteAction.MaxStreamDuration



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| maxStreamDuration | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Specifies the maximum duration allowed for streams on the route. If not specified, the value from the :ref:`max_stream_duration <envoy_api_field_config.core.v3.HttpProtocolOptions.max_stream_duration>` field in :ref:`HttpConnectionManager.common_http_protocol_options <envoy_api_field_extensions.filters.network.http_connection_manager.v3.HttpConnectionManager.common_http_protocol_options>` is used. If this field is set explicitly to zero, any HttpConnectionManager max_stream_duration timeout will be disabled for this route. |
  | grpcTimeoutHeaderMax | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | If present, and the request contains a `grpc-timeout header <https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md>`_, use that value as the *max_stream_duration*, but limit the applied timeout to the maximum value specified here. If set to 0, the `grpc-timeout` header is used without modification. |
  | grpcTimeoutHeaderOffset | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | If present, Envoy will adjust the timeout provided by the `grpc-timeout` header by subtracting the provided duration from the header. This is useful for allowing Envoy to set its global timeout to be less than that of the deadline imposed by the calling client, which makes it more likely that Envoy will handle the timeout instead of having the call canceled by the client. If, after applying the offset, the resulting timeout is zero or negative, the stream will timeout immediately. |
  





<a name="envoy.config.route.v3.RouteAction.RequestMirrorPolicy"></a>

### RouteAction.RequestMirrorPolicy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cluster | string |  | Specifies the cluster that requests will be mirrored to. The cluster must exist in the cluster manager configuration. |
  | runtimeFraction | [envoy.config.core.v3.RuntimeFractionalPercent]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.RuntimeFractionalPercent" >}}) |  | If not specified, all requests to the target cluster will be mirrored.<br>If specified, this field takes precedence over the `runtime_key` field and requests must also fall under the percentage of matches indicated by this field.<br>For some fraction N/D, a random number in the range [0,D) is selected. If the number is <= the value of the numerator N, or if the key is not present, the default value, the request will be mirrored. |
  | traceSampled | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | Determines if the trace span should be sampled. Defaults to true. |
  





<a name="envoy.config.route.v3.RouteAction.UpgradeConfig"></a>

### RouteAction.UpgradeConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| upgradeType | string |  | The case-insensitive name of this upgrade, e.g. "websocket". For each upgrade type present in upgrade_configs, requests with Upgrade: [upgrade_type] will be proxied upstream. |
  | enabled | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | Determines if upgrades are available on this route. Defaults to true. |
  | connectConfig | [envoy.config.route.v3.RouteAction.UpgradeConfig.ConnectConfig]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RouteAction.UpgradeConfig.ConnectConfig" >}}) |  | Configuration for sending data upstream as a raw data payload. This is used for CONNECT requests, when forwarding CONNECT payload as raw TCP. Note that CONNECT support is currently considered alpha in Envoy. [#comment:TODO(htuch): Replace the above comment with an alpha tag. |
  





<a name="envoy.config.route.v3.RouteAction.UpgradeConfig.ConnectConfig"></a>

### RouteAction.UpgradeConfig.ConnectConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| proxyProtocolConfig | [envoy.config.core.v3.ProxyProtocolConfig]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.proxy_protocol#envoy.config.core.v3.ProxyProtocolConfig" >}}) |  | If present, the proxy protocol header will be prepended to the CONNECT payload sent upstream. |
  





<a name="envoy.config.route.v3.RouteMatch"></a>

### RouteMatch



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| prefix | string |  | If specified, the route is a prefix rule meaning that the prefix must match the beginning of the *:path* header. |
  | path | string |  | If specified, the route is an exact path rule meaning that the path must exactly match the *:path* header once the query string is removed. |
  | safeRegex | [envoy.type.matcher.v3.RegexMatcher]({{< versioned_link_path fromRoot="/reference/api/envoy.type.matcher.v3.regex#envoy.type.matcher.v3.RegexMatcher" >}}) |  | If specified, the route is a regular expression rule meaning that the regex must match the *:path* header once the query string is removed. The entire path (without the query string) must match the regex. The rule will not match if only a subsequence of the *:path* header matches the regex.<br>[#next-major-version: In the v3 API we should redo how path specification works such that we utilize StringMatcher, and additionally have consistent options around whether we strip query strings, do a case sensitive match, etc. In the interim it will be too disruptive to deprecate the existing options. We should even consider whether we want to do away with path_specifier entirely and just rely on a set of header matchers which can already match on :path, etc. The issue with that is it is unclear how to generically deal with query string stripping. This needs more thought.] |
  | connectMatcher | [envoy.config.route.v3.RouteMatch.ConnectMatcher]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RouteMatch.ConnectMatcher" >}}) |  | If this is used as the matcher, the matcher will only match CONNECT requests. Note that this will not match HTTP/2 upgrade-style CONNECT requests (WebSocket and the like) as they are normalized in Envoy as HTTP/1.1 style upgrades. This is the only way to match CONNECT requests for HTTP/1.1. For HTTP/2, where Extended CONNECT requests may have a path, the path matchers will work if there is a path present. Note that CONNECT support is currently considered alpha in Envoy. [#comment:TODO(htuch): Replace the above comment with an alpha tag. |
  | caseSensitive | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | Indicates that prefix/path matching should be case sensitive. The default is true. |
  | runtimeFraction | [envoy.config.core.v3.RuntimeFractionalPercent]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.RuntimeFractionalPercent" >}}) |  | Indicates that the route should additionally match on a runtime key. Every time the route is considered for a match, it must also fall under the percentage of matches indicated by this field. For some fraction N/D, a random number in the range [0,D) is selected. If the number is <= the value of the numerator N, or if the key is not present, the default value, the router continues to evaluate the remaining match criteria. A runtime_fraction route configuration can be used to roll out route changes in a gradual manner without full code/config deploys. Refer to the :ref:`traffic shifting <config_http_conn_man_route_table_traffic_splitting_shift>` docs for additional documentation.<br>.. note::<br>   Parsing this field is implemented such that the runtime key's data may be represented    as a FractionalPercent proto represented as JSON/YAML and may also be represented as an    integer with the assumption that the value is an integral percentage out of 100. For    instance, a runtime key lookup returning the value "42" would parse as a FractionalPercent    whose numerator is 42 and denominator is HUNDRED. This preserves legacy semantics. |
  | headers | [][envoy.config.route.v3.HeaderMatcher]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.HeaderMatcher" >}}) | repeated | Specifies a set of headers that the route should match on. The router will check the request’s headers against all the specified headers in the route config. A match will happen if all the headers in the route are present in the request with the same values (or based on presence if the value field is not in the config). |
  | queryParameters | [][envoy.config.route.v3.QueryParameterMatcher]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.QueryParameterMatcher" >}}) | repeated | Specifies a set of URL query parameters on which the route should match. The router will check the query string from the *path* header against all the specified query parameters. If the number of specified query parameters is nonzero, they all must match the *path* header's query string for a match to occur. |
  | grpc | [envoy.config.route.v3.RouteMatch.GrpcRouteMatchOptions]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RouteMatch.GrpcRouteMatchOptions" >}}) |  | If specified, only gRPC requests will be matched. The router will check that the content-type header has a application/grpc or one of the various application/grpc+ values. |
  | tlsContext | [envoy.config.route.v3.RouteMatch.TlsContextMatchOptions]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RouteMatch.TlsContextMatchOptions" >}}) |  | If specified, the client tls context will be matched against the defined match options.<br>[#next-major-version: unify with RBAC] |
  





<a name="envoy.config.route.v3.RouteMatch.ConnectMatcher"></a>

### RouteMatch.ConnectMatcher







<a name="envoy.config.route.v3.RouteMatch.GrpcRouteMatchOptions"></a>

### RouteMatch.GrpcRouteMatchOptions







<a name="envoy.config.route.v3.RouteMatch.TlsContextMatchOptions"></a>

### RouteMatch.TlsContextMatchOptions



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| presented | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | If specified, the route will match against whether or not a certificate is presented. If not specified, certificate presentation status (true or false) will not be considered when route matching. |
  | validated | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | If specified, the route will match against whether or not a certificate is validated. If not specified, certificate validation status (true or false) will not be considered when route matching. |
  





<a name="envoy.config.route.v3.Tracing"></a>

### Tracing



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| clientSampling | [envoy.type.v3.FractionalPercent]({{< versioned_link_path fromRoot="/reference/api/envoy.type.v3.percent#envoy.type.v3.FractionalPercent" >}}) |  | Target percentage of requests managed by this HTTP connection manager that will be force traced if the :ref:`x-client-trace-id <config_http_conn_man_headers_x-client-trace-id>` header is set. This field is a direct analog for the runtime variable 'tracing.client_sampling' in the :ref:`HTTP Connection Manager <config_http_conn_man_runtime>`. Default: 100% |
  | randomSampling | [envoy.type.v3.FractionalPercent]({{< versioned_link_path fromRoot="/reference/api/envoy.type.v3.percent#envoy.type.v3.FractionalPercent" >}}) |  | Target percentage of requests managed by this HTTP connection manager that will be randomly selected for trace generation, if not requested by the client or not forced. This field is a direct analog for the runtime variable 'tracing.random_sampling' in the :ref:`HTTP Connection Manager <config_http_conn_man_runtime>`. Default: 100% |
  | overallSampling | [envoy.type.v3.FractionalPercent]({{< versioned_link_path fromRoot="/reference/api/envoy.type.v3.percent#envoy.type.v3.FractionalPercent" >}}) |  | Target percentage of requests managed by this HTTP connection manager that will be traced after all other sampling checks have been applied (client-directed, force tracing, random sampling). This field functions as an upper limit on the total configured sampling rate. For instance, setting client_sampling to 100% but overall_sampling to 1% will result in only 1% of client requests with the appropriate headers to be force traced. This field is a direct analog for the runtime variable 'tracing.global_enabled' in the :ref:`HTTP Connection Manager <config_http_conn_man_runtime>`. Default: 100% |
  | customTags | [][envoy.type.tracing.v3.CustomTag]({{< versioned_link_path fromRoot="/reference/api/envoy.type.tracing.v3.custom_tag#envoy.type.tracing.v3.CustomTag" >}}) | repeated | A list of custom tags with unique tag name to create tags for the active span. It will take effect after merging with the :ref:`corresponding configuration <envoy_api_field_extensions.filters.network.http_connection_manager.v3.HttpConnectionManager.Tracing.custom_tags>` configured in the HTTP connection manager. If two tags with the same name are configured each in the HTTP connection manager and the route level, the one configured here takes priority. |
  





<a name="envoy.config.route.v3.VirtualCluster"></a>

### VirtualCluster



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [][envoy.config.route.v3.HeaderMatcher]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.HeaderMatcher" >}}) | repeated | Specifies a list of header matchers to use for matching requests. Each specified header must match. The pseudo-headers `:path` and `:method` can be used to match the request path and method, respectively. |
  | name | string |  | Specifies the name of the virtual cluster. The virtual cluster name as well as the virtual host name are used when emitting statistics. The statistics are emitted by the router filter and are documented :ref:`here <config_http_filters_router_stats>`. |
  





<a name="envoy.config.route.v3.VirtualHost"></a>

### VirtualHost



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The logical name of the virtual host. This is used when emitting certain statistics but is not relevant for routing. |
  | domains | []string | repeated | A list of domains (host/authority header) that will be matched to this virtual host. Wildcard hosts are supported in the suffix or prefix form.<br>Domain search order:  1. Exact domain names: ``www.foo.com``.  2. Suffix domain wildcards: ``*.foo.com`` or ``*-bar.foo.com``.  3. Prefix domain wildcards: ``foo.*`` or ``foo-*``.  4. Special wildcard ``*`` matching any domain.<br>.. note::<br>  The wildcard will not match the empty string.   e.g. ``*-bar.foo.com`` will match ``baz-bar.foo.com`` but not ``-bar.foo.com``.   The longest wildcards match first.   Only a single virtual host in the entire route configuration can match on ``*``. A domain   must be unique across all virtual hosts or the config will fail to load.<br>Domains cannot contain control characters. This is validated by the well_known_regex HTTP_HEADER_VALUE. |
  | routes | [][envoy.config.route.v3.Route]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.Route" >}}) | repeated | The list of routes that will be matched, in order, for incoming requests. The first route that matches will be used. |
  | requireTls | [envoy.config.route.v3.VirtualHost.TlsRequirementType]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.VirtualHost.TlsRequirementType" >}}) |  | Specifies the type of TLS enforcement the virtual host expects. If this option is not specified, there is no TLS requirement for the virtual host. |
  | virtualClusters | [][envoy.config.route.v3.VirtualCluster]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.VirtualCluster" >}}) | repeated | A list of virtual clusters defined for this virtual host. Virtual clusters are used for additional statistics gathering. |
  | rateLimits | [][envoy.config.route.v3.RateLimit]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RateLimit" >}}) | repeated | Specifies a set of rate limit configurations that will be applied to the virtual host. |
  | requestHeadersToAdd | [][envoy.config.core.v3.HeaderValueOption]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.HeaderValueOption" >}}) | repeated | Specifies a list of HTTP headers that should be added to each request handled by this virtual host. Headers specified at this level are applied after headers from enclosed :ref:`envoy_api_msg_config.route.v3.Route` and before headers from the enclosing :ref:`envoy_api_msg_config.route.v3.RouteConfiguration`. For more information, including details on header value syntax, see the documentation on :ref:`custom request headers <config_http_conn_man_headers_custom_request_headers>`. |
  | requestHeadersToRemove | []string | repeated | Specifies a list of HTTP headers that should be removed from each request handled by this virtual host. |
  | responseHeadersToAdd | [][envoy.config.core.v3.HeaderValueOption]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.HeaderValueOption" >}}) | repeated | Specifies a list of HTTP headers that should be added to each response handled by this virtual host. Headers specified at this level are applied after headers from enclosed :ref:`envoy_api_msg_config.route.v3.Route` and before headers from the enclosing :ref:`envoy_api_msg_config.route.v3.RouteConfiguration`. For more information, including details on header value syntax, see the documentation on :ref:`custom request headers <config_http_conn_man_headers_custom_request_headers>`. |
  | responseHeadersToRemove | []string | repeated | Specifies a list of HTTP headers that should be removed from each response handled by this virtual host. |
  | cors | [envoy.config.route.v3.CorsPolicy]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.CorsPolicy" >}}) |  | Indicates that the virtual host has a CORS policy. |
  | typedPerFilterConfig | [][envoy.config.route.v3.VirtualHost.TypedPerFilterConfigEntry]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.VirtualHost.TypedPerFilterConfigEntry" >}}) | repeated | The per_filter_config field can be used to provide virtual host-specific configurations for filters. The key should match the filter name, such as *envoy.filters.http.buffer* for the HTTP buffer filter. Use of this field is filter specific; see the :ref:`HTTP filter documentation <config_http_filters>` for if and how it is utilized. |
  | includeRequestAttemptCount | bool |  | Decides whether the :ref:`x-envoy-attempt-count <config_http_filters_router_x-envoy-attempt-count>` header should be included in the upstream request. Setting this option will cause it to override any existing header value, so in the case of two Envoys on the request path with this option enabled, the upstream will see the attempt count as perceived by the second Envoy. Defaults to false. This header is unaffected by the :ref:`suppress_envoy_headers <envoy_api_field_extensions.filters.http.router.v3.Router.suppress_envoy_headers>` flag.<br>[#next-major-version: rename to include_attempt_count_in_request.] |
  | includeAttemptCountInResponse | bool |  | Decides whether the :ref:`x-envoy-attempt-count <config_http_filters_router_x-envoy-attempt-count>` header should be included in the downstream response. Setting this option will cause the router to override any existing header value, so in the case of two Envoys on the request path with this option enabled, the downstream will see the attempt count as perceived by the Envoy closest upstream from itself. Defaults to false. This header is unaffected by the :ref:`suppress_envoy_headers <envoy_api_field_extensions.filters.http.router.v3.Router.suppress_envoy_headers>` flag. |
  | retryPolicy | [envoy.config.route.v3.RetryPolicy]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.RetryPolicy" >}}) |  | Indicates the retry policy for all routes in this virtual host. Note that setting a route level entry will take precedence over this config and it'll be treated independently (e.g.: values are not inherited). |
  | retryPolicyTypedConfig | [google.protobuf.Any]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.any#google.protobuf.Any" >}}) |  | [#not-implemented-hide:] Specifies the configuration for retry policy extension. Note that setting a route level entry will take precedence over this config and it'll be treated independently (e.g.: values are not inherited). :ref:`Retry policy <envoy_api_field_config.route.v3.VirtualHost.retry_policy>` should not be set if this field is used. |
  | hedgePolicy | [envoy.config.route.v3.HedgePolicy]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.HedgePolicy" >}}) |  | Indicates the hedge policy for all routes in this virtual host. Note that setting a route level entry will take precedence over this config and it'll be treated independently (e.g.: values are not inherited). |
  | perRequestBufferLimitBytes | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | The maximum bytes which will be buffered for retries and shadowing. If set and a route-specific limit is not set, the bytes actually buffered will be the minimum value of this and the listener per_connection_buffer_limit_bytes. |
  





<a name="envoy.config.route.v3.VirtualHost.TypedPerFilterConfigEntry"></a>

### VirtualHost.TypedPerFilterConfigEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [google.protobuf.Any]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.any#google.protobuf.Any" >}}) |  |  |
  





<a name="envoy.config.route.v3.WeightedCluster"></a>

### WeightedCluster



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| clusters | [][envoy.config.route.v3.WeightedCluster.ClusterWeight]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.WeightedCluster.ClusterWeight" >}}) | repeated | Specifies one or more upstream clusters associated with the route. |
  | totalWeight | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | Specifies the total weight across all clusters. The sum of all cluster weights must equal this value, which must be greater than 0. Defaults to 100. |
  | runtimeKeyPrefix | string |  | Specifies the runtime key prefix that should be used to construct the runtime keys associated with each cluster. When the *runtime_key_prefix* is specified, the router will look for weights associated with each upstream cluster under the key *runtime_key_prefix* + "." + *cluster[i].name* where *cluster[i]* denotes an entry in the clusters array field. If the runtime key for the cluster does not exist, the value specified in the configuration file will be used as the default weight. See the :ref:`runtime documentation <operations_runtime>` for how key names map to the underlying implementation. |
  





<a name="envoy.config.route.v3.WeightedCluster.ClusterWeight"></a>

### WeightedCluster.ClusterWeight



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Name of the upstream cluster. The cluster must exist in the :ref:`cluster manager configuration <config_cluster_manager>`. |
  | weight | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | An integer between 0 and :ref:`total_weight <envoy_api_field_config.route.v3.WeightedCluster.total_weight>`. When a request matches the route, the choice of an upstream cluster is determined by its weight. The sum of weights across all entries in the clusters array must add up to the total_weight, which defaults to 100. |
  | metadataMatch | [envoy.config.core.v3.Metadata]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.Metadata" >}}) |  | Optional endpoint metadata match criteria used by the subset load balancer. Only endpoints in the upstream cluster with metadata matching what is set in this field will be considered for load balancing. Note that this will be merged with what's provided in :ref:`RouteAction.metadata_match <envoy_api_field_config.route.v3.RouteAction.metadata_match>`, with values here taking precedence. The filter name should be specified as *envoy.lb*. |
  | requestHeadersToAdd | [][envoy.config.core.v3.HeaderValueOption]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.HeaderValueOption" >}}) | repeated | Specifies a list of headers to be added to requests when this cluster is selected through the enclosing :ref:`envoy_api_msg_config.route.v3.RouteAction`. Headers specified at this level are applied before headers from the enclosing :ref:`envoy_api_msg_config.route.v3.Route`, :ref:`envoy_api_msg_config.route.v3.VirtualHost`, and :ref:`envoy_api_msg_config.route.v3.RouteConfiguration`. For more information, including details on header value syntax, see the documentation on :ref:`custom request headers <config_http_conn_man_headers_custom_request_headers>`. |
  | requestHeadersToRemove | []string | repeated | Specifies a list of HTTP headers that should be removed from each request when this cluster is selected through the enclosing :ref:`envoy_api_msg_config.route.v3.RouteAction`. |
  | responseHeadersToAdd | [][envoy.config.core.v3.HeaderValueOption]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.HeaderValueOption" >}}) | repeated | Specifies a list of headers to be added to responses when this cluster is selected through the enclosing :ref:`envoy_api_msg_config.route.v3.RouteAction`. Headers specified at this level are applied before headers from the enclosing :ref:`envoy_api_msg_config.route.v3.Route`, :ref:`envoy_api_msg_config.route.v3.VirtualHost`, and :ref:`envoy_api_msg_config.route.v3.RouteConfiguration`. For more information, including details on header value syntax, see the documentation on :ref:`custom request headers <config_http_conn_man_headers_custom_request_headers>`. |
  | responseHeadersToRemove | []string | repeated | Specifies a list of headers to be removed from responses when this cluster is selected through the enclosing :ref:`envoy_api_msg_config.route.v3.RouteAction`. |
  | typedPerFilterConfig | [][envoy.config.route.v3.WeightedCluster.ClusterWeight.TypedPerFilterConfigEntry]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.WeightedCluster.ClusterWeight.TypedPerFilterConfigEntry" >}}) | repeated | The per_filter_config field can be used to provide weighted cluster-specific configurations for filters. The key should match the filter name, such as *envoy.filters.http.buffer* for the HTTP buffer filter. Use of this field is filter specific; see the :ref:`HTTP filter documentation <config_http_filters>` for if and how it is utilized. |
  





<a name="envoy.config.route.v3.WeightedCluster.ClusterWeight.TypedPerFilterConfigEntry"></a>

### WeightedCluster.ClusterWeight.TypedPerFilterConfigEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [google.protobuf.Any]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.any#google.protobuf.Any" >}}) |  |  |
  




 <!-- end messages -->


<a name="envoy.config.route.v3.RateLimit.Action.MetaData.Source"></a>

### RateLimit.Action.MetaData.Source


| Name | Number | Description |
| ---- | ------ | ----------- |
| DYNAMIC | 0 | Query :ref:`dynamic metadata <well_known_dynamic_metadata>` |
| ROUTE_ENTRY | 1 | Query :ref:`route entry metadata <envoy_api_field_config.route.v3.Route.metadata>` |



<a name="envoy.config.route.v3.RedirectAction.RedirectResponseCode"></a>

### RedirectAction.RedirectResponseCode


| Name | Number | Description |
| ---- | ------ | ----------- |
| MOVED_PERMANENTLY | 0 | Moved Permanently HTTP Status Code - 301. |
| FOUND | 1 | Found HTTP Status Code - 302. |
| SEE_OTHER | 2 | See Other HTTP Status Code - 303. |
| TEMPORARY_REDIRECT | 3 | Temporary Redirect HTTP Status Code - 307. |
| PERMANENT_REDIRECT | 4 | Permanent Redirect HTTP Status Code - 308. |



<a name="envoy.config.route.v3.RetryPolicy.ResetHeaderFormat"></a>

### RetryPolicy.ResetHeaderFormat


| Name | Number | Description |
| ---- | ------ | ----------- |
| SECONDS | 0 |  |
| UNIX_TIMESTAMP | 1 |  |



<a name="envoy.config.route.v3.RouteAction.ClusterNotFoundResponseCode"></a>

### RouteAction.ClusterNotFoundResponseCode


| Name | Number | Description |
| ---- | ------ | ----------- |
| SERVICE_UNAVAILABLE | 0 | HTTP status code - 503 Service Unavailable. |
| NOT_FOUND | 1 | HTTP status code - 404 Not Found. |



<a name="envoy.config.route.v3.RouteAction.InternalRedirectAction"></a>

### RouteAction.InternalRedirectAction


| Name | Number | Description |
| ---- | ------ | ----------- |
| PASS_THROUGH_INTERNAL_REDIRECT | 0 |  |
| HANDLE_INTERNAL_REDIRECT | 1 |  |



<a name="envoy.config.route.v3.VirtualHost.TlsRequirementType"></a>

### VirtualHost.TlsRequirementType


| Name | Number | Description |
| ---- | ------ | ----------- |
| NONE | 0 | No TLS requirement for the virtual host. |
| EXTERNAL_ONLY | 1 | External requests must use TLS. If a request is external and it is not using TLS, a 301 redirect will be sent telling the client to use HTTPS. |
| ALL | 2 | All requests must use TLS. If a request is not using TLS, a 301 redirect will be sent telling the client to use HTTPS. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

