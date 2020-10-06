
---
title: "traffic_policy.proto"
---

## Package : `networking.smh.solo.io`



<a name="top"></a>

<a name="API Reference for traffic_policy.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## traffic_policy.proto


## Table of Contents
  - [TrafficPolicySpec](#networking.smh.solo.io.TrafficPolicySpec)
  - [TrafficPolicySpec.CorsPolicy](#networking.smh.solo.io.TrafficPolicySpec.CorsPolicy)
  - [TrafficPolicySpec.FaultInjection](#networking.smh.solo.io.TrafficPolicySpec.FaultInjection)
  - [TrafficPolicySpec.FaultInjection.Abort](#networking.smh.solo.io.TrafficPolicySpec.FaultInjection.Abort)
  - [TrafficPolicySpec.HeaderManipulation](#networking.smh.solo.io.TrafficPolicySpec.HeaderManipulation)
  - [TrafficPolicySpec.HeaderManipulation.AppendRequestHeadersEntry](#networking.smh.solo.io.TrafficPolicySpec.HeaderManipulation.AppendRequestHeadersEntry)
  - [TrafficPolicySpec.HeaderManipulation.AppendResponseHeadersEntry](#networking.smh.solo.io.TrafficPolicySpec.HeaderManipulation.AppendResponseHeadersEntry)
  - [TrafficPolicySpec.HeaderMatcher](#networking.smh.solo.io.TrafficPolicySpec.HeaderMatcher)
  - [TrafficPolicySpec.HttpMatcher](#networking.smh.solo.io.TrafficPolicySpec.HttpMatcher)
  - [TrafficPolicySpec.HttpMethod](#networking.smh.solo.io.TrafficPolicySpec.HttpMethod)
  - [TrafficPolicySpec.Mirror](#networking.smh.solo.io.TrafficPolicySpec.Mirror)
  - [TrafficPolicySpec.MultiDestination](#networking.smh.solo.io.TrafficPolicySpec.MultiDestination)
  - [TrafficPolicySpec.MultiDestination.WeightedDestination](#networking.smh.solo.io.TrafficPolicySpec.MultiDestination.WeightedDestination)
  - [TrafficPolicySpec.MultiDestination.WeightedDestination.FailoverServiceDestination](#networking.smh.solo.io.TrafficPolicySpec.MultiDestination.WeightedDestination.FailoverServiceDestination)
  - [TrafficPolicySpec.MultiDestination.WeightedDestination.FailoverServiceDestination.SubsetEntry](#networking.smh.solo.io.TrafficPolicySpec.MultiDestination.WeightedDestination.FailoverServiceDestination.SubsetEntry)
  - [TrafficPolicySpec.MultiDestination.WeightedDestination.KubeDestination](#networking.smh.solo.io.TrafficPolicySpec.MultiDestination.WeightedDestination.KubeDestination)
  - [TrafficPolicySpec.MultiDestination.WeightedDestination.KubeDestination.SubsetEntry](#networking.smh.solo.io.TrafficPolicySpec.MultiDestination.WeightedDestination.KubeDestination.SubsetEntry)
  - [TrafficPolicySpec.OutlierDetection](#networking.smh.solo.io.TrafficPolicySpec.OutlierDetection)
  - [TrafficPolicySpec.QueryParameterMatcher](#networking.smh.solo.io.TrafficPolicySpec.QueryParameterMatcher)
  - [TrafficPolicySpec.RetryPolicy](#networking.smh.solo.io.TrafficPolicySpec.RetryPolicy)
  - [TrafficPolicySpec.StringMatch](#networking.smh.solo.io.TrafficPolicySpec.StringMatch)
  - [TrafficPolicyStatus](#networking.smh.solo.io.TrafficPolicyStatus)
  - [TrafficPolicyStatus.TrafficTargetsEntry](#networking.smh.solo.io.TrafficPolicyStatus.TrafficTargetsEntry)







<a name="networking.smh.solo.io.TrafficPolicySpec"></a>

### TrafficPolicySpec
A Traffic Policy applies some L7 routing features to an existing mesh. Traffic Policies specify the following for all requests: - originating from from **source workload** - sent to **destination targets** - matching one or more **request matcher**<br>NB: If any additional TrafficPolicy action fields (i.e. non selection related fields) are added, the TrafficPolicy Merger's "AreTrafficPolicyActionsEqual" method must be updated to reflect the new field.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sourceSelector | [][WorkloadSelector](#networking.smh.solo.io.WorkloadSelector) | repeated | Requests originating from these workloads will have the rule applied. Leave empty to have all workloads in the mesh apply these rules.<br>Note: Source Selectors are ignored when TrafficPolicies are applied to pods in a Linkerd mesh. TrafficPolicies will apply to all selected destinations in Linkerd, regardless of the source.<br>Note: If using the TrafficTargetSelector.Matcher, specifying clusters is currently not supported in Istio. |
| destinationSelector | [][TrafficTargetSelector](#networking.smh.solo.io.TrafficTargetSelector) | repeated | Requests destined for these k8s services will have the rule applied. Leave empty to apply to all destination k8s services in the mesh. |
| httpRequestMatchers | [][TrafficPolicySpec.HttpMatcher](#networking.smh.solo.io.TrafficPolicySpec.HttpMatcher) | repeated | If specified, this rule will only apply to http requests matching these conditions. Within a single matcher, all conditions must be satisfied for a match to occur. Between matchers, at least one matcher must be satisfied for the TrafficPolicy to apply. NB: Linkerd only supports matching on Request Path and Method. |
| trafficShift | [TrafficPolicySpec.MultiDestination](#networking.smh.solo.io.TrafficPolicySpec.MultiDestination) |  | Enables traffic shifting, i.e. to reroute requests to a different service, to a subset of pods based on their label, and/or split traffic between multiple services. |
| faultInjection | [TrafficPolicySpec.FaultInjection](#networking.smh.solo.io.TrafficPolicySpec.FaultInjection) |  | Enable fault injection on requests. |
| requestTimeout | [google.protobuf.Duration](#google.protobuf.Duration) |  | Set a timeout on requests. |
| retries | [TrafficPolicySpec.RetryPolicy](#networking.smh.solo.io.TrafficPolicySpec.RetryPolicy) |  | Set a retry policy on requests. |
| corsPolicy | [TrafficPolicySpec.CorsPolicy](#networking.smh.solo.io.TrafficPolicySpec.CorsPolicy) |  | Set a Cross-Origin Resource Sharing policy (CORS) for requests. Refer to https://developer.mozilla.org/en-US/docs/Web/HTTP/Access_control_CORS for further details about cross origin resource sharing. |
| mirror | [TrafficPolicySpec.Mirror](#networking.smh.solo.io.TrafficPolicySpec.Mirror) |  | Mirror HTTP traffic to a another destination. Traffic will still be sent to its original destination as normal. |
| headerManipulation | [TrafficPolicySpec.HeaderManipulation](#networking.smh.solo.io.TrafficPolicySpec.HeaderManipulation) |  | Manipulate request and response headers. |
| outlierDetection | [TrafficPolicySpec.OutlierDetection](#networking.smh.solo.io.TrafficPolicySpec.OutlierDetection) |  | Configure outlier detection on the targeted services. Setting this field requires an empty source_selector because it must apply to all traffic. |






<a name="networking.smh.solo.io.TrafficPolicySpec.CorsPolicy"></a>

### TrafficPolicySpec.CorsPolicy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| allowOrigins | [][TrafficPolicySpec.StringMatch](#networking.smh.solo.io.TrafficPolicySpec.StringMatch) | repeated | String patterns that match allowed origins. An origin is allowed if any of the string matchers match. If a match is found, then the outgoing Access-Control-Allow-Origin would be set to the origin as provided by the client. |
| allowMethods | [][string](#string) | repeated | List of HTTP methods allowed to access the resource. The content will be serialized into the Access-Control-Allow-Methods header. |
| allowHeaders | [][string](#string) | repeated | List of HTTP headers that can be used when requesting the resource. Serialized to Access-Control-Allow-Headers header. |
| exposeHeaders | [][string](#string) | repeated | A white list of HTTP headers that the browsers are allowed to access. Serialized into Access-Control-Expose-Headers header. |
| maxAge | [google.protobuf.Duration](#google.protobuf.Duration) |  | Specifies how long the results of a preflight request can be cached. Translates to the `Access-Control-Max-Age` header. |
| allowCredentials | [google.protobuf.BoolValue](#google.protobuf.BoolValue) |  | Indicates whether the caller is allowed to send the actual request (not the preflight) using credentials. Translates to `Access-Control-Allow-Credentials` header. |






<a name="networking.smh.solo.io.TrafficPolicySpec.FaultInjection"></a>

### TrafficPolicySpec.FaultInjection
FaultInjection can be used to specify one or more faults to inject while forwarding http requests to the destination specified in a route. Faults include aborting the Http request from downstream service, and/or delaying proxying of requests. A fault rule MUST HAVE delay or abort.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fixedDelay | [google.protobuf.Duration](#google.protobuf.Duration) |  | Add a fixed delay before forwarding the request. Format: 1h/1m/1s/1ms. MUST be >=1ms. |
| exponentialDelay | [google.protobuf.Duration](#google.protobuf.Duration) |  | $hide_from_docs |
| abort | [TrafficPolicySpec.FaultInjection.Abort](#networking.smh.solo.io.TrafficPolicySpec.FaultInjection.Abort) |  | Abort Http request attempts and return error codes back to downstream service, giving the impression that the upstream service is faulty. |
| percentage | [double](#double) |  | Percentage of requests to be faulted with the error code provided. Values range between 0 and 100 |






<a name="networking.smh.solo.io.TrafficPolicySpec.FaultInjection.Abort"></a>

### TrafficPolicySpec.FaultInjection.Abort
The _httpStatus_ field is used to indicate the HTTP status code to return to the caller. The optional _percentage_ field can be used to only abort a certain percentage of requests. If not specified, all requests are aborted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| httpStatus | [int32](#int32) |  | REQUIRED. HTTP status code to use to abort the Http request. |






<a name="networking.smh.solo.io.TrafficPolicySpec.HeaderManipulation"></a>

### TrafficPolicySpec.HeaderManipulation
Manipulate request and response headers.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| removeResponseHeaders | [][string](#string) | repeated | HTTP headers to remove before returning a response to the caller. |
| appendResponseHeaders | [][TrafficPolicySpec.HeaderManipulation.AppendResponseHeadersEntry](#networking.smh.solo.io.TrafficPolicySpec.HeaderManipulation.AppendResponseHeadersEntry) | repeated | Additional HTTP headers to add before returning a response to the caller. |
| removeRequestHeaders | [][string](#string) | repeated | HTTP headers to remove before forwarding a request to the destination service. |
| appendRequestHeaders | [][TrafficPolicySpec.HeaderManipulation.AppendRequestHeadersEntry](#networking.smh.solo.io.TrafficPolicySpec.HeaderManipulation.AppendRequestHeadersEntry) | repeated | Additional HTTP headers to add before forwarding a request to the destination service. |






<a name="networking.smh.solo.io.TrafficPolicySpec.HeaderManipulation.AppendRequestHeadersEntry"></a>

### TrafficPolicySpec.HeaderManipulation.AppendRequestHeadersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="networking.smh.solo.io.TrafficPolicySpec.HeaderManipulation.AppendResponseHeadersEntry"></a>

### TrafficPolicySpec.HeaderManipulation.AppendResponseHeadersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="networking.smh.solo.io.TrafficPolicySpec.HeaderMatcher"></a>

### TrafficPolicySpec.HeaderMatcher



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Specifies the name of the header in the request. |
| value | [string](#string) |  | Specifies the value of the header. If the value is absent a request that has the name header will match, regardless of the header’s value. |
| regex | [bool](#bool) |  | Specifies whether the header value should be treated as regex or not. |
| invertMatch | [bool](#bool) |  | If set to true, the result of the match will be inverted. Defaults to false.<br>Examples: name=foo, invert_match=true: matches if no header named `foo` is present name=foo, value=bar, invert_match=true: matches if no header named `foo` with value `bar` is present name=foo, value=``\d{3}``, regex=true, invert_match=true: matches if no header named `foo` with a value consisting of three integers is present |






<a name="networking.smh.solo.io.TrafficPolicySpec.HttpMatcher"></a>

### TrafficPolicySpec.HttpMatcher
Parameters for matching routes. All specified conditions must be satisfied for a match to occur.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| prefix | [string](#string) |  | If specified, the route is a prefix rule meaning that the prefix must match the beginning of the *:path* header. |
| exact | [string](#string) |  | If specified, the route is an exact path rule meaning that the path must exactly match the *:path* header once the query string is removed. |
| regex | [string](#string) |  | If specified, the route is a regular expression rule meaning that the regex must match the *:path* header once the query string is removed. The entire path (without the query string) must match the regex. The rule will not match if only a sub-sequence of the *:path* header matches the regex. The regex grammar is defined `here <http://en.cppreference.com/w/cpp/regex/ecmascript>`_. |
| headers | [][TrafficPolicySpec.HeaderMatcher](#networking.smh.solo.io.TrafficPolicySpec.HeaderMatcher) | repeated | Specifies a set of headers which requests must match in entirety (all headers must match). |
| queryParameters | [][TrafficPolicySpec.QueryParameterMatcher](#networking.smh.solo.io.TrafficPolicySpec.QueryParameterMatcher) | repeated | Specifies a set of URL query parameters which requests must match in entirety (all query params must match). The router will check the query string from the *path* header against all the specified query parameters |
| method | [TrafficPolicySpec.HttpMethod](#networking.smh.solo.io.TrafficPolicySpec.HttpMethod) |  | HTTP Method/Verb to match on. If none specified, the matcher will ignore the HTTP Method |






<a name="networking.smh.solo.io.TrafficPolicySpec.HttpMethod"></a>

### TrafficPolicySpec.HttpMethod
Express an optional HttpMethod by wrapping it in a nillable message.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| method | [HttpMethodValue](#networking.smh.solo.io.HttpMethodValue) |  |  |






<a name="networking.smh.solo.io.TrafficPolicySpec.Mirror"></a>

### TrafficPolicySpec.Mirror



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeService | [core.skv2.solo.io.ClusterObjectRef](#core.skv2.solo.io.ClusterObjectRef) |  | Name/namespace/cluster of a kubernetes service. |
| percentage | [double](#double) |  | Percentage of traffic to mirror. If absent, 100% will be mirrored. Values range between 0 and 100 |
| port | [uint32](#uint32) |  | Port on the destination k8s service to receive traffic. If multiple are found, and none are specified, then the configuration will be considered invalid. |






<a name="networking.smh.solo.io.TrafficPolicySpec.MultiDestination"></a>

### TrafficPolicySpec.MultiDestination



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| destinations | [][TrafficPolicySpec.MultiDestination.WeightedDestination](#networking.smh.solo.io.TrafficPolicySpec.MultiDestination.WeightedDestination) | repeated | A traffic shift destination. |






<a name="networking.smh.solo.io.TrafficPolicySpec.MultiDestination.WeightedDestination"></a>

### TrafficPolicySpec.MultiDestination.WeightedDestination



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeService | [TrafficPolicySpec.MultiDestination.WeightedDestination.KubeDestination](#networking.smh.solo.io.TrafficPolicySpec.MultiDestination.WeightedDestination.KubeDestination) |  | The use kubeService to shift traffic a Kubernetes Service/subset. |
| failoverService | [TrafficPolicySpec.MultiDestination.WeightedDestination.FailoverServiceDestination](#networking.smh.solo.io.TrafficPolicySpec.MultiDestination.WeightedDestination.FailoverServiceDestination) |  | A traffic shift destination targeting a FailoverService. |
| weight | [uint32](#uint32) |  | Weights across all of the destinations must sum to 100. Each is interpreted as a percent from 0-100. |






<a name="networking.smh.solo.io.TrafficPolicySpec.MultiDestination.WeightedDestination.FailoverServiceDestination"></a>

### TrafficPolicySpec.MultiDestination.WeightedDestination.FailoverServiceDestination
A traffic shift destination that references a FailoverService.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | The name of the FailoverService. |
| namespace | [string](#string) |  | The namespace of the FailoverService. |
| subset | [][TrafficPolicySpec.MultiDestination.WeightedDestination.FailoverServiceDestination.SubsetEntry](#networking.smh.solo.io.TrafficPolicySpec.MultiDestination.WeightedDestination.FailoverServiceDestination.SubsetEntry) | repeated | Subset routing is currently only supported for Istio backing services. |






<a name="networking.smh.solo.io.TrafficPolicySpec.MultiDestination.WeightedDestination.FailoverServiceDestination.SubsetEntry"></a>

### TrafficPolicySpec.MultiDestination.WeightedDestination.FailoverServiceDestination.SubsetEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="networking.smh.solo.io.TrafficPolicySpec.MultiDestination.WeightedDestination.KubeDestination"></a>

### TrafficPolicySpec.MultiDestination.WeightedDestination.KubeDestination
A traffic shift destination which lives in kubernetes.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | The name of the destination service. |
| namespace | [string](#string) |  | The namespace of the destination service. |
| clusterName | [string](#string) |  | The cluster of the destination k8s service (as it is registered with Service Mesh Hub). |
| subset | [][TrafficPolicySpec.MultiDestination.WeightedDestination.KubeDestination.SubsetEntry](#networking.smh.solo.io.TrafficPolicySpec.MultiDestination.WeightedDestination.KubeDestination.SubsetEntry) | repeated | Subset routing is currently only supported on Istio. |
| port | [uint32](#uint32) |  | Port on the destination k8s service to receive traffic. Required if the service exposes more than one port. |






<a name="networking.smh.solo.io.TrafficPolicySpec.MultiDestination.WeightedDestination.KubeDestination.SubsetEntry"></a>

### TrafficPolicySpec.MultiDestination.WeightedDestination.KubeDestination.SubsetEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="networking.smh.solo.io.TrafficPolicySpec.OutlierDetection"></a>

### TrafficPolicySpec.OutlierDetection
Configure outlier detection settings on targeted services. If set, source selectors must be empty. Outlier detection settings apply to all incoming traffic.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| consecutiveErrors | [uint32](#uint32) |  | Number of errors before a host is ejected from the connection pool. Defaults to 5. |
| interval | [google.protobuf.Duration](#google.protobuf.Duration) |  | Time interval between ejection sweep analysis. Format: 1h/1m/1s/1ms. MUST BE >=1ms. Defaults to 10s. |
| baseEjectionTime | [google.protobuf.Duration](#google.protobuf.Duration) |  | Minimum ejection duration. Format: 1h/1m/1s/1ms. MUST BE >=1ms. Defaults to 30s. |
| maxEjectionPercent | [uint32](#uint32) |  | Maximum % of hosts in the load balancing pool for the upstream service that can be ejected, but will eject at least one host regardless of the value. MUST BE >= 0 and <= 100. Defaults to 100%, allowing all hosts to be ejected from the pool. |






<a name="networking.smh.solo.io.TrafficPolicySpec.QueryParameterMatcher"></a>

### TrafficPolicySpec.QueryParameterMatcher
Query parameter matching treats the query string of a request's :path header as an ampersand-separated list of keys and/or key=value elements.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Specifies the name of a key that must be present in the requested path*'s query string. |
| value | [string](#string) |  | Specifies the value of the key. If the value is absent, a request that contains the key in its query string will match, whether the key appears with a value (e.g., "?debug=true") or not (e.g., "?debug") |
| regex | [bool](#bool) |  | Specifies whether the query parameter value is a regular expression. Defaults to false. The entire query parameter value (i.e., the part to the right of the equals sign in "key=value") must match the regex. E.g., the regex "\d+$" will match "123" but not "a123" or "123a". |






<a name="networking.smh.solo.io.TrafficPolicySpec.RetryPolicy"></a>

### TrafficPolicySpec.RetryPolicy
RetryPolicy contains mesh-specific retry configuration. Different meshes support different Retry features. Service Mesh Hub's RetryPolicy exposes config for multiple meshes simultaneously, allowing the same TrafficPolicy to apply retries to different mesh types. The configuration applied to the target mesh will use the corresponding config for each type, while other config types will be ignored.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| attempts | [int32](#int32) |  | Number of retries for a given request |
| perTryTimeout | [google.protobuf.Duration](#google.protobuf.Duration) |  | Timeout per retry attempt for a given request. format: 1h/1m/1s/1ms. MUST BE >=1ms. |






<a name="networking.smh.solo.io.TrafficPolicySpec.StringMatch"></a>

### TrafficPolicySpec.StringMatch
Describes how to match a given string in HTTP headers. Match is case-sensitive.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| exact | [string](#string) |  | Exact string match. |
| prefix | [string](#string) |  | Prefix-based match. |
| regex | [string](#string) |  | ECMAscript style regex-based match. |






<a name="networking.smh.solo.io.TrafficPolicyStatus"></a>

### TrafficPolicyStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | [int64](#int64) |  | The most recent generation observed in the the TrafficPolicy metadata. if the observedGeneration does not match generation, the controller has not received the most recent version of this resource. |
| state | [ApprovalState](#networking.smh.solo.io.ApprovalState) |  | The state of the overall resource. It will only show accepted if it has been successfully applied to all target meshes. |
| trafficTargets | [][TrafficPolicyStatus.TrafficTargetsEntry](#networking.smh.solo.io.TrafficPolicyStatus.TrafficTargetsEntry) | repeated | The status of the TrafficPolicy for each TrafficTarget to which it has been applied. A TrafficPolicy may be Accepted for some TrafficTargets and rejected for others. |
| workloads | [][string](#string) | repeated | The list of Workloads to which this policy has been applied. |
| errors | [][string](#string) | repeated | Any errors found while processing this generation of the resource. |






<a name="networking.smh.solo.io.TrafficPolicyStatus.TrafficTargetsEntry"></a>

### TrafficPolicyStatus.TrafficTargetsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [ApprovalStatus](#networking.smh.solo.io.ApprovalStatus) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

