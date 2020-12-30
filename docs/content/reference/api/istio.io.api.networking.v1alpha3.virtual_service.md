
---

---

## Package : `istio.networking.v1alpha3`



<a name="top"></a>

<a name="API Reference for virtual_service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## virtual_service.proto


## Table of Contents
  - [CorsPolicy](#istio.networking.v1alpha3.CorsPolicy)
  - [Delegate](#istio.networking.v1alpha3.Delegate)
  - [Destination](#istio.networking.v1alpha3.Destination)
  - [HTTPFaultInjection](#istio.networking.v1alpha3.HTTPFaultInjection)
  - [HTTPFaultInjection.Abort](#istio.networking.v1alpha3.HTTPFaultInjection.Abort)
  - [HTTPFaultInjection.Delay](#istio.networking.v1alpha3.HTTPFaultInjection.Delay)
  - [HTTPMatchRequest](#istio.networking.v1alpha3.HTTPMatchRequest)
  - [HTTPMatchRequest.HeadersEntry](#istio.networking.v1alpha3.HTTPMatchRequest.HeadersEntry)
  - [HTTPMatchRequest.QueryParamsEntry](#istio.networking.v1alpha3.HTTPMatchRequest.QueryParamsEntry)
  - [HTTPMatchRequest.SourceLabelsEntry](#istio.networking.v1alpha3.HTTPMatchRequest.SourceLabelsEntry)
  - [HTTPMatchRequest.WithoutHeadersEntry](#istio.networking.v1alpha3.HTTPMatchRequest.WithoutHeadersEntry)
  - [HTTPRedirect](#istio.networking.v1alpha3.HTTPRedirect)
  - [HTTPRetry](#istio.networking.v1alpha3.HTTPRetry)
  - [HTTPRewrite](#istio.networking.v1alpha3.HTTPRewrite)
  - [HTTPRoute](#istio.networking.v1alpha3.HTTPRoute)
  - [HTTPRouteDestination](#istio.networking.v1alpha3.HTTPRouteDestination)
  - [Headers](#istio.networking.v1alpha3.Headers)
  - [Headers.HeaderOperations](#istio.networking.v1alpha3.Headers.HeaderOperations)
  - [Headers.HeaderOperations.AddEntry](#istio.networking.v1alpha3.Headers.HeaderOperations.AddEntry)
  - [Headers.HeaderOperations.SetEntry](#istio.networking.v1alpha3.Headers.HeaderOperations.SetEntry)
  - [L4MatchAttributes](#istio.networking.v1alpha3.L4MatchAttributes)
  - [L4MatchAttributes.SourceLabelsEntry](#istio.networking.v1alpha3.L4MatchAttributes.SourceLabelsEntry)
  - [Percent](#istio.networking.v1alpha3.Percent)
  - [PortSelector](#istio.networking.v1alpha3.PortSelector)
  - [RouteDestination](#istio.networking.v1alpha3.RouteDestination)
  - [StringMatch](#istio.networking.v1alpha3.StringMatch)
  - [TCPRoute](#istio.networking.v1alpha3.TCPRoute)
  - [TLSMatchAttributes](#istio.networking.v1alpha3.TLSMatchAttributes)
  - [TLSMatchAttributes.SourceLabelsEntry](#istio.networking.v1alpha3.TLSMatchAttributes.SourceLabelsEntry)
  - [TLSRoute](#istio.networking.v1alpha3.TLSRoute)
  - [VirtualService](#istio.networking.v1alpha3.VirtualService)







<a name="istio.networking.v1alpha3.CorsPolicy"></a>

### CorsPolicy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| allowOrigin | []string | repeated | The list of origins that are allowed to perform CORS requests. The content will be serialized into the Access-Control-Allow-Origin header. Wildcard * will allow all origins. $hide_from_docs |
  | allowOrigins | [][istio.networking.v1alpha3.StringMatch]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.StringMatch" >}}) | repeated | String patterns that match allowed origins. An origin is allowed if any of the string matchers match. If a match is found, then the outgoing Access-Control-Allow-Origin would be set to the origin as provided by the client. |
  | allowMethods | []string | repeated | List of HTTP methods allowed to access the resource. The content will be serialized into the Access-Control-Allow-Methods header. |
  | allowHeaders | []string | repeated | List of HTTP headers that can be used when requesting the resource. Serialized to Access-Control-Allow-Headers header. |
  | exposeHeaders | []string | repeated | A list of HTTP headers that the browsers are allowed to access. Serialized into Access-Control-Expose-Headers header. |
  | maxAge | [google.protobuf.Duration]({{< versioned_link_path fromRoot="github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Specifies how long the results of a preflight request can be cached. Translates to the `Access-Control-Max-Age` header. |
  | allowCredentials | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | Indicates whether the caller is allowed to send the actual request (not the preflight) using credentials. Translates to `Access-Control-Allow-Credentials` header. |
  





<a name="istio.networking.v1alpha3.Delegate"></a>

### Delegate



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Name specifies the name of the delegate VirtualService. |
  | namespace | string |  | Namespace specifies the namespace where the delegate VirtualService resides. By default, it is same to the root's. |
  





<a name="istio.networking.v1alpha3.Destination"></a>

### Destination



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| host | string |  | The name of a service from the service registry. Service names are looked up from the platform's service registry (e.g., Kubernetes services, Consul services, etc.) and from the hosts declared by [ServiceEntry](https://istio.io/docs/reference/config/networking/service-entry/#ServiceEntry). Traffic forwarded to destinations that are not found in either of the two, will be dropped.<br>*Note for Kubernetes users*: When short names are used (e.g. "reviews" instead of "reviews.default.svc.cluster.local"), Istio will interpret the short name based on the namespace of the rule, not the service. A rule in the "default" namespace containing a host "reviews will be interpreted as "reviews.default.svc.cluster.local", irrespective of the actual namespace associated with the reviews service. To avoid potential misconfiguration, it is recommended to always use fully qualified domain names over short names. |
  | subset | string |  | The name of a subset within the service. Applicable only to services within the mesh. The subset must be defined in a corresponding DestinationRule. |
  | port | [istio.networking.v1alpha3.PortSelector]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.PortSelector" >}}) |  | Specifies the port on the host that is being addressed. If a service exposes only a single port it is not required to explicitly select the port. |
  





<a name="istio.networking.v1alpha3.HTTPFaultInjection"></a>

### HTTPFaultInjection



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| delay | [istio.networking.v1alpha3.HTTPFaultInjection.Delay]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.HTTPFaultInjection.Delay" >}}) |  | Delay requests before forwarding, emulating various failures such as network issues, overloaded upstream service, etc. |
  | abort | [istio.networking.v1alpha3.HTTPFaultInjection.Abort]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.HTTPFaultInjection.Abort" >}}) |  | Abort Http request attempts and return error codes back to downstream service, giving the impression that the upstream service is faulty. |
  





<a name="istio.networking.v1alpha3.HTTPFaultInjection.Abort"></a>

### HTTPFaultInjection.Abort



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| httpStatus | int32 |  | HTTP status code to use to abort the Http request. |
  | grpcStatus | string |  | $hide_from_docs |
  | http2Error | string |  | $hide_from_docs |
  | percentage | [istio.networking.v1alpha3.Percent]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.Percent" >}}) |  | Percentage of requests to be aborted with the error code provided. |
  





<a name="istio.networking.v1alpha3.HTTPFaultInjection.Delay"></a>

### HTTPFaultInjection.Delay



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| percent | int32 |  | Percentage of requests on which the delay will be injected (0-100). Use of integer `percent` value is deprecated. Use the double `percentage` field instead. |
  | fixedDelay | [google.protobuf.Duration]({{< versioned_link_path fromRoot="github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Add a fixed delay before forwarding the request. Format: 1h/1m/1s/1ms. MUST be >=1ms. |
  | exponentialDelay | [google.protobuf.Duration]({{< versioned_link_path fromRoot="github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | $hide_from_docs |
  | percentage | [istio.networking.v1alpha3.Percent]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.Percent" >}}) |  | Percentage of requests on which the delay will be injected. |
  





<a name="istio.networking.v1alpha3.HTTPMatchRequest"></a>

### HTTPMatchRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The name assigned to a match. The match's name will be concatenated with the parent route's name and will be logged in the access logs for requests matching this route. |
  | uri | [istio.networking.v1alpha3.StringMatch]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.StringMatch" >}}) |  | URI to match values are case-sensitive and formatted as follows:<br>- `exact: "value"` for exact string match<br>- `prefix: "value"` for prefix-based match<br>- `regex: "value"` for ECMAscript style regex-based match<br>**Note:** Case-insensitive matching could be enabled via the `ignore_uri_case` flag. |
  | scheme | [istio.networking.v1alpha3.StringMatch]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.StringMatch" >}}) |  | URI Scheme values are case-sensitive and formatted as follows:<br>- `exact: "value"` for exact string match<br>- `prefix: "value"` for prefix-based match<br>- `regex: "value"` for ECMAscript style regex-based match |
  | method | [istio.networking.v1alpha3.StringMatch]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.StringMatch" >}}) |  | HTTP Method values are case-sensitive and formatted as follows:<br>- `exact: "value"` for exact string match<br>- `prefix: "value"` for prefix-based match<br>- `regex: "value"` for ECMAscript style regex-based match |
  | authority | [istio.networking.v1alpha3.StringMatch]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.StringMatch" >}}) |  | HTTP Authority values are case-sensitive and formatted as follows:<br>- `exact: "value"` for exact string match<br>- `prefix: "value"` for prefix-based match<br>- `regex: "value"` for ECMAscript style regex-based match |
  | headers | [][istio.networking.v1alpha3.HTTPMatchRequest.HeadersEntry]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.HTTPMatchRequest.HeadersEntry" >}}) | repeated | The header keys must be lowercase and use hyphen as the separator, e.g. _x-request-id_.<br>Header values are case-sensitive and formatted as follows:<br>- `exact: "value"` for exact string match<br>- `prefix: "value"` for prefix-based match<br>- `regex: "value"` for ECMAscript style regex-based match<br>If the value is empty and only the name of header is specfied, presence of the header is checked. **Note:** The keys `uri`, `scheme`, `method`, and `authority` will be ignored. |
  | port | uint32 |  | Specifies the ports on the host that is being addressed. Many services only expose a single port or label ports with the protocols they support, in these cases it is not required to explicitly select the port. |
  | sourceLabels | [][istio.networking.v1alpha3.HTTPMatchRequest.SourceLabelsEntry]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.HTTPMatchRequest.SourceLabelsEntry" >}}) | repeated | One or more labels that constrain the applicability of a rule to workloads with the given labels. If the VirtualService has a list of gateways specified in the top-level `gateways` field, it must include the reserved gateway `mesh` for this field to be applicable. |
  | gateways | []string | repeated | Names of gateways where the rule should be applied. Gateway names in the top-level `gateways` field of the VirtualService (if any) are overridden. The gateway match is independent of sourceLabels. |
  | queryParams | [][istio.networking.v1alpha3.HTTPMatchRequest.QueryParamsEntry]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.HTTPMatchRequest.QueryParamsEntry" >}}) | repeated | Query parameters for matching.<br>Ex: - For a query parameter like "?key=true", the map key would be "key" and   the string match could be defined as `exact: "true"`. - For a query parameter like "?key", the map key would be "key" and the   string match could be defined as `exact: ""`. - For a query parameter like "?key=123", the map key would be "key" and the   string match could be defined as `regex: "\d+$"`. Note that this   configuration will only match values like "123" but not "a123" or "123a".<br>**Note:** `prefix` matching is currently not supported. |
  | ignoreUriCase | bool |  | Flag to specify whether the URI matching should be case-insensitive.<br>**Note:** The case will be ignored only in the case of `exact` and `prefix` URI matches. |
  | withoutHeaders | [][istio.networking.v1alpha3.HTTPMatchRequest.WithoutHeadersEntry]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.HTTPMatchRequest.WithoutHeadersEntry" >}}) | repeated | withoutHeader has the same syntax with the header, but has opposite meaning. If a header is matched with a matching rule among withoutHeader, the traffic becomes not matched one. |
  | sourceNamespace | string |  | Source namespace constraining the applicability of a rule to workloads in that namespace. If the VirtualService has a list of gateways specified in the top-level `gateways` field, it must include the reserved gateway `mesh` for this field to be applicable. |
  





<a name="istio.networking.v1alpha3.HTTPMatchRequest.HeadersEntry"></a>

### HTTPMatchRequest.HeadersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [istio.networking.v1alpha3.StringMatch]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.StringMatch" >}}) |  |  |
  





<a name="istio.networking.v1alpha3.HTTPMatchRequest.QueryParamsEntry"></a>

### HTTPMatchRequest.QueryParamsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [istio.networking.v1alpha3.StringMatch]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.StringMatch" >}}) |  |  |
  





<a name="istio.networking.v1alpha3.HTTPMatchRequest.SourceLabelsEntry"></a>

### HTTPMatchRequest.SourceLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="istio.networking.v1alpha3.HTTPMatchRequest.WithoutHeadersEntry"></a>

### HTTPMatchRequest.WithoutHeadersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [istio.networking.v1alpha3.StringMatch]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.StringMatch" >}}) |  |  |
  





<a name="istio.networking.v1alpha3.HTTPRedirect"></a>

### HTTPRedirect



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uri | string |  | On a redirect, overwrite the Path portion of the URL with this value. Note that the entire path will be replaced, irrespective of the request URI being matched as an exact path or prefix. |
  | authority | string |  | On a redirect, overwrite the Authority/Host portion of the URL with this value. |
  | redirectCode | uint32 |  | On a redirect, Specifies the HTTP status code to use in the redirect response. The default response code is MOVED_PERMANENTLY (301). |
  





<a name="istio.networking.v1alpha3.HTTPRetry"></a>

### HTTPRetry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| attempts | int32 |  | Number of retries for a given request. The interval between retries will be determined automatically (25ms+). Actual number of retries attempted depends on the request `timeout` of the [HTTP route](https://istio.io/docs/reference/config/networking/virtual-service/#HTTPRoute). |
  | perTryTimeout | [google.protobuf.Duration]({{< versioned_link_path fromRoot="github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Timeout per retry attempt for a given request. format: 1h/1m/1s/1ms. MUST BE >=1ms. |
  | retryOn | string |  | Specifies the conditions under which retry takes place. One or more policies can be specified using a ‘,’ delimited list. See the [retry policies](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/router_filter#x-envoy-retry-on) and [gRPC retry policies](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/router_filter#x-envoy-retry-grpc-on) for more details. |
  | retryRemoteLocalities | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | Flag to specify whether the retries should retry to other localities. See the [retry plugin configuration](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/http/http_connection_management#retry-plugin-configuration) for more details. |
  





<a name="istio.networking.v1alpha3.HTTPRewrite"></a>

### HTTPRewrite



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uri | string |  | rewrite the path (or the prefix) portion of the URI with this value. If the original URI was matched based on prefix, the value provided in this field will replace the corresponding matched prefix. |
  | authority | string |  | rewrite the Authority/Host header with this value. |
  





<a name="istio.networking.v1alpha3.HTTPRoute"></a>

### HTTPRoute



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The name assigned to the route for debugging purposes. The route's name will be concatenated with the match's name and will be logged in the access logs for requests matching this route/match. |
  | match | [][istio.networking.v1alpha3.HTTPMatchRequest]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.HTTPMatchRequest" >}}) | repeated | Match conditions to be satisfied for the rule to be activated. All conditions inside a single match block have AND semantics, while the list of match blocks have OR semantics. The rule is matched if any one of the match blocks succeed. |
  | route | [][istio.networking.v1alpha3.HTTPRouteDestination]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.HTTPRouteDestination" >}}) | repeated | A HTTP rule can either redirect or forward (default) traffic. The forwarding target can be one of several versions of a service (see glossary in beginning of document). Weights associated with the service version determine the proportion of traffic it receives. |
  | redirect | [istio.networking.v1alpha3.HTTPRedirect]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.HTTPRedirect" >}}) |  | A HTTP rule can either redirect or forward (default) traffic. If traffic passthrough option is specified in the rule, route/redirect will be ignored. The redirect primitive can be used to send a HTTP 301 redirect to a different URI or Authority. |
  | delegate | [istio.networking.v1alpha3.Delegate]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.Delegate" >}}) |  | Delegate is used to specify the particular VirtualService which can be used to define delegate HTTPRoute. It can be set only when `Route` and `Redirect` are empty, and the route rules of the delegate VirtualService will be merged with that in the current one. **NOTE**:    1. Only one level delegation is supported.    2. The delegate's HTTPMatchRequest must be a strict subset of the root's,       otherwise there is a conflict and the HTTPRoute will not take effect. |
  | rewrite | [istio.networking.v1alpha3.HTTPRewrite]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.HTTPRewrite" >}}) |  | Rewrite HTTP URIs and Authority headers. Rewrite cannot be used with Redirect primitive. Rewrite will be performed before forwarding. |
  | timeout | [google.protobuf.Duration]({{< versioned_link_path fromRoot="github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Timeout for HTTP requests. |
  | retries | [istio.networking.v1alpha3.HTTPRetry]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.HTTPRetry" >}}) |  | Retry policy for HTTP requests. |
  | fault | [istio.networking.v1alpha3.HTTPFaultInjection]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.HTTPFaultInjection" >}}) |  | Fault injection policy to apply on HTTP traffic at the client side. Note that timeouts or retries will not be enabled when faults are enabled on the client side. |
  | mirror | [istio.networking.v1alpha3.Destination]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.Destination" >}}) |  | Mirror HTTP traffic to a another destination in addition to forwarding the requests to the intended destination. Mirrored traffic is on a best effort basis where the sidecar/gateway will not wait for the mirrored cluster to respond before returning the response from the original destination.  Statistics will be generated for the mirrored destination. |
  | mirrorPercent | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | Percentage of the traffic to be mirrored by the `mirror` field. Use of integer `mirror_percent` value is deprecated. Use the double `mirror_percentage` field instead |
  | mirrorPercentage | [istio.networking.v1alpha3.Percent]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.Percent" >}}) |  | Percentage of the traffic to be mirrored by the `mirror` field. If this field is absent, all the traffic (100%) will be mirrored. Max value is 100. |
  | corsPolicy | [istio.networking.v1alpha3.CorsPolicy]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.CorsPolicy" >}}) |  | Cross-Origin Resource Sharing policy (CORS). Refer to [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) for further details about cross origin resource sharing. |
  | headers | [istio.networking.v1alpha3.Headers]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.Headers" >}}) |  | Header manipulation rules |
  





<a name="istio.networking.v1alpha3.HTTPRouteDestination"></a>

### HTTPRouteDestination



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| destination | [istio.networking.v1alpha3.Destination]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.Destination" >}}) |  | Destination uniquely identifies the instances of a service to which the request/connection should be forwarded to. |
  | weight | int32 |  | The proportion of traffic to be forwarded to the service version. (0-100). Sum of weights across destinations SHOULD BE == 100. If there is only one destination in a rule, the weight value is assumed to be 100. |
  | headers | [istio.networking.v1alpha3.Headers]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.Headers" >}}) |  | Header manipulation rules |
  





<a name="istio.networking.v1alpha3.Headers"></a>

### Headers



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request | [istio.networking.v1alpha3.Headers.HeaderOperations]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.Headers.HeaderOperations" >}}) |  | Header manipulation rules to apply before forwarding a request to the destination service |
  | response | [istio.networking.v1alpha3.Headers.HeaderOperations]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.Headers.HeaderOperations" >}}) |  | Header manipulation rules to apply before returning a response to the caller |
  





<a name="istio.networking.v1alpha3.Headers.HeaderOperations"></a>

### Headers.HeaderOperations



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| set | [][istio.networking.v1alpha3.Headers.HeaderOperations.SetEntry]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.Headers.HeaderOperations.SetEntry" >}}) | repeated | Overwrite the headers specified by key with the given values |
  | add | [][istio.networking.v1alpha3.Headers.HeaderOperations.AddEntry]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.Headers.HeaderOperations.AddEntry" >}}) | repeated | Append the given values to the headers specified by keys (will create a comma-separated list of values) |
  | remove | []string | repeated | Remove a the specified headers |
  





<a name="istio.networking.v1alpha3.Headers.HeaderOperations.AddEntry"></a>

### Headers.HeaderOperations.AddEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="istio.networking.v1alpha3.Headers.HeaderOperations.SetEntry"></a>

### Headers.HeaderOperations.SetEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="istio.networking.v1alpha3.L4MatchAttributes"></a>

### L4MatchAttributes



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| destinationSubnets | []string | repeated | IPv4 or IPv6 ip addresses of destination with optional subnet.  E.g., a.b.c.d/xx form or just a.b.c.d. |
  | port | uint32 |  | Specifies the port on the host that is being addressed. Many services only expose a single port or label ports with the protocols they support, in these cases it is not required to explicitly select the port. |
  | sourceSubnet | string |  | IPv4 or IPv6 ip address of source with optional subnet. E.g., a.b.c.d/xx form or just a.b.c.d $hide_from_docs |
  | sourceLabels | [][istio.networking.v1alpha3.L4MatchAttributes.SourceLabelsEntry]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.L4MatchAttributes.SourceLabelsEntry" >}}) | repeated | One or more labels that constrain the applicability of a rule to workloads with the given labels. If the VirtualService has a list of gateways specified in the top-level `gateways` field, it should include the reserved gateway `mesh` in order for this field to be applicable. |
  | gateways | []string | repeated | Names of gateways where the rule should be applied. Gateway names in the top-level `gateways` field of the VirtualService (if any) are overridden. The gateway match is independent of sourceLabels. |
  | sourceNamespace | string |  | Source namespace constraining the applicability of a rule to workloads in that namespace. If the VirtualService has a list of gateways specified in the top-level `gateways` field, it must include the reserved gateway `mesh` for this field to be applicable. |
  





<a name="istio.networking.v1alpha3.L4MatchAttributes.SourceLabelsEntry"></a>

### L4MatchAttributes.SourceLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="istio.networking.v1alpha3.Percent"></a>

### Percent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | double |  |  |
  





<a name="istio.networking.v1alpha3.PortSelector"></a>

### PortSelector



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| number | uint32 |  | Valid port number |
  





<a name="istio.networking.v1alpha3.RouteDestination"></a>

### RouteDestination



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| destination | [istio.networking.v1alpha3.Destination]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.Destination" >}}) |  | Destination uniquely identifies the instances of a service to which the request/connection should be forwarded to. |
  | weight | int32 |  | The proportion of traffic to be forwarded to the service version. If there is only one destination in a rule, all traffic will be routed to it irrespective of the weight. |
  





<a name="istio.networking.v1alpha3.StringMatch"></a>

### StringMatch



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| exact | string |  | exact string match |
  | prefix | string |  | prefix-based match |
  | regex | string |  | RE2 style regex-based match (https://github.com/google/re2/wiki/Syntax). |
  





<a name="istio.networking.v1alpha3.TCPRoute"></a>

### TCPRoute



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| match | [][istio.networking.v1alpha3.L4MatchAttributes]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.L4MatchAttributes" >}}) | repeated | Match conditions to be satisfied for the rule to be activated. All conditions inside a single match block have AND semantics, while the list of match blocks have OR semantics. The rule is matched if any one of the match blocks succeed. |
  | route | [][istio.networking.v1alpha3.RouteDestination]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.RouteDestination" >}}) | repeated | The destination to which the connection should be forwarded to. |
  





<a name="istio.networking.v1alpha3.TLSMatchAttributes"></a>

### TLSMatchAttributes



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sniHosts | []string | repeated | SNI (server name indicator) to match on. Wildcard prefixes can be used in the SNI value, e.g., *.com will match foo.example.com as well as example.com. An SNI value must be a subset (i.e., fall within the domain) of the corresponding virtual serivce's hosts. |
  | destinationSubnets | []string | repeated | IPv4 or IPv6 ip addresses of destination with optional subnet.  E.g., a.b.c.d/xx form or just a.b.c.d. |
  | port | uint32 |  | Specifies the port on the host that is being addressed. Many services only expose a single port or label ports with the protocols they support, in these cases it is not required to explicitly select the port. |
  | sourceLabels | [][istio.networking.v1alpha3.TLSMatchAttributes.SourceLabelsEntry]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.TLSMatchAttributes.SourceLabelsEntry" >}}) | repeated | One or more labels that constrain the applicability of a rule to workloads with the given labels. If the VirtualService has a list of gateways specified in the top-level `gateways` field, it should include the reserved gateway `mesh` in order for this field to be applicable. |
  | gateways | []string | repeated | Names of gateways where the rule should be applied. Gateway names in the top-level `gateways` field of the VirtualService (if any) are overridden. The gateway match is independent of sourceLabels. |
  | sourceNamespace | string |  | Source namespace constraining the applicability of a rule to workloads in that namespace. If the VirtualService has a list of gateways specified in the top-level `gateways` field, it must include the reserved gateway `mesh` for this field to be applicable. |
  





<a name="istio.networking.v1alpha3.TLSMatchAttributes.SourceLabelsEntry"></a>

### TLSMatchAttributes.SourceLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="istio.networking.v1alpha3.TLSRoute"></a>

### TLSRoute



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| match | [][istio.networking.v1alpha3.TLSMatchAttributes]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.TLSMatchAttributes" >}}) | repeated | Match conditions to be satisfied for the rule to be activated. All conditions inside a single match block have AND semantics, while the list of match blocks have OR semantics. The rule is matched if any one of the match blocks succeed. |
  | route | [][istio.networking.v1alpha3.RouteDestination]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.RouteDestination" >}}) | repeated | The destination to which the connection should be forwarded to. |
  





<a name="istio.networking.v1alpha3.VirtualService"></a>

### VirtualService



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hosts | []string | repeated | The destination hosts to which traffic is being sent. Could be a DNS name with wildcard prefix or an IP address.  Depending on the platform, short-names can also be used instead of a FQDN (i.e. has no dots in the name). In such a scenario, the FQDN of the host would be derived based on the underlying platform.<br>A single VirtualService can be used to describe all the traffic properties of the corresponding hosts, including those for multiple HTTP and TCP ports. Alternatively, the traffic properties of a host can be defined using more than one VirtualService, with certain caveats. Refer to the [Operations Guide](https://istio.io/docs/ops/best-practices/traffic-management/#split-virtual-services) for details.<br>*Note for Kubernetes users*: When short names are used (e.g. "reviews" instead of "reviews.default.svc.cluster.local"), Istio will interpret the short name based on the namespace of the rule, not the service. A rule in the "default" namespace containing a host "reviews" will be interpreted as "reviews.default.svc.cluster.local", irrespective of the actual namespace associated with the reviews service. _To avoid potential misconfigurations, it is recommended to always use fully qualified domain names over short names._<br>The hosts field applies to both HTTP and TCP services. Service inside the mesh, i.e., those found in the service registry, must always be referred to using their alphanumeric names. IP addresses are allowed only for services defined via the Gateway.<br>*Note*: It must be empty for a delegate VirtualService. |
  | gateways | []string | repeated | The names of gateways and sidecars that should apply these routes. Gateways in other namespaces may be referred to by `<gateway namespace>/<gateway name>`; specifying a gateway with no namespace qualifier is the same as specifying the VirtualService's namespace. A single VirtualService is used for sidecars inside the mesh as well as for one or more gateways. The selection condition imposed by this field can be overridden using the source field in the match conditions of protocol-specific routes. The reserved word `mesh` is used to imply all the sidecars in the mesh. When this field is omitted, the default gateway (`mesh`) will be used, which would apply the rule to all sidecars in the mesh. If a list of gateway names is provided, the rules will apply only to the gateways. To apply the rules to both gateways and sidecars, specify `mesh` as one of the gateway names. |
  | http | [][istio.networking.v1alpha3.HTTPRoute]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.HTTPRoute" >}}) | repeated | An ordered list of route rules for HTTP traffic. HTTP routes will be applied to platform service ports named 'http-*'/'http2-*'/'grpc-*', gateway ports with protocol HTTP/HTTP2/GRPC/ TLS-terminated-HTTPS and service entry ports using HTTP/HTTP2/GRPC protocols.  The first rule matching an incoming request is used. |
  | tls | [][istio.networking.v1alpha3.TLSRoute]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.TLSRoute" >}}) | repeated | An ordered list of route rule for non-terminated TLS & HTTPS traffic. Routing is typically performed using the SNI value presented by the ClientHello message. TLS routes will be applied to platform service ports named 'https-*', 'tls-*', unterminated gateway ports using HTTPS/TLS protocols (i.e. with "passthrough" TLS mode) and service entry ports using HTTPS/TLS protocols.  The first rule matching an incoming request is used.  NOTE: Traffic 'https-*' or 'tls-*' ports without associated virtual service will be treated as opaque TCP traffic. |
  | tcp | [][istio.networking.v1alpha3.TCPRoute]({{< versioned_link_path fromRoot="istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.TCPRoute" >}}) | repeated | An ordered list of route rules for opaque TCP traffic. TCP routes will be applied to any port that is not a HTTP or TLS port. The first rule matching an incoming request is used. |
  | exportTo | []string | repeated | A list of namespaces to which this virtual service is exported. Exporting a virtual service allows it to be used by sidecars and gateways defined in other namespaces. This feature provides a mechanism for service owners and mesh administrators to control the visibility of virtual services across namespace boundaries.<br>If no namespaces are specified then the virtual service is exported to all namespaces by default.<br>The value "." is reserved and defines an export to the same namespace that the virtual service is declared in. Similarly the value "*" is reserved and defines an export to all namespaces.<br>NOTE: in the current release, the `exportTo` value is restricted to "." or "*" (i.e., the current namespace or all namespaces). |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

