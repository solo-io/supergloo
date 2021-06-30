
---

title: "rate_limit.proto"

---

## Package : `ratelimit.networking.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for rate_limit.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## rate_limit.proto


## Table of Contents
  - [GatewayRateLimit](#ratelimit.networking.mesh.gloo.solo.io.GatewayRateLimit)
  - [RouteRateLimit](#ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit)
  - [RouteRateLimit.AdvancedRateLimit](#ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.AdvancedRateLimit)
  - [RouteRateLimit.AdvancedRateLimit.RateLimitActions](#ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.AdvancedRateLimit.RateLimitActions)
  - [RouteRateLimit.BasicRateLimit](#ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.BasicRateLimit)
  - [RouteRateLimit.BasicRateLimit.RateLimitRatio](#ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.BasicRateLimit.RateLimitRatio)

  - [RouteRateLimit.BasicRateLimit.RateLimitRatio.Unit](#ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.BasicRateLimit.RateLimitRatio.Unit)






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
  | configRefs | [core.skv2.solo.io.ObjectRefList]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRefList" >}}) |  | References to RateLimitConfig resources. This is used to configure the GlooE rate limit server. Only one of `ratelimit` or `rate_limit_configs` can be set. |
  





<a name="ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.AdvancedRateLimit"></a>

### RouteRateLimit.AdvancedRateLimit
Use this field if you want to inline the Envoy rate limits for this VirtualHost. Note that this does not configure the rate limit server. If you are running Gloo Mesh, you need to specify the server configuration via the appropriate field in the Gloo Mesh `GatewayRateLimit` resource on the gateway. If you are running a custom rate limit server you need to configure it yourself.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| actions | [][ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.AdvancedRateLimit.RateLimitActions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.AdvancedRateLimit.RateLimitActions" >}}) | repeated | Define individual rate limits here. Each rate limit will be evaluated, if any rate limit would be throttled, the entire request returns a 429 (gets throttled) |
  





<a name="ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit.AdvancedRateLimit.RateLimitActions"></a>

### RouteRateLimit.AdvancedRateLimit.RateLimitActions
Each action and setAction in the lists maps part of the request (or its context) to a descriptor. The tuple or set of descriptors generated by the provided actions is sent to the rate limit server and matched against rate limit rules. Order matters on provided actions but not on setActions, e.g. the following actions: - actions:   - requestHeaders:      descriptorKey: account_id      headerName: x-account-id   - requestHeaders:      descriptorKey: plan      headerName: x-plan define an ordered descriptor tuple like so: ('account_id', '<x-account-id value>'), ('plan', '<x-plan value>')<br>While the current form matches, the same tuple in reverse order would not match the following descriptor:<br>descriptors: - key: account_id   descriptors:   - key: plan     value: BASIC     rateLimit:       requestsPerUnit: 1       unit: MINUTE  - key: plan    value: PLUS    rateLimit:      requestsPerUnit: 20      unit: MINUTE<br>Similarly, the following setActions: - setActions:   - requestHeaders:      descriptorKey: account_id      headerName: x-account-id   - requestHeaders:      descriptorKey: plan      headerName: x-plan define an unordered descriptor set like so: {('account_id', '<x-account-id value>'), ('plan', '<x-plan value>')}<br>This set would match the following setDescriptor:<br>setDescriptors: - simpleDescriptors:   - key: plan     value: BASIC   - key: account_id  rateLimit:    requestsPerUnit: 20    unit: MINUTE<br>It would also match the following setDescriptor which includes only a subset of the setActions enumerated:<br>setDescriptors: - simpleDescriptors:   - key: account_id  rateLimit:    requestsPerUnit: 20    unit: MINUTE<br>It would even match the following setDescriptor. Any setActions list would match this setDescriptor which has simpleDescriptors omitted entirely:<br>setDescriptors: - rateLimit:    requestsPerUnit: 20    unit: MINUTE


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| actions | [][ratelimit.api.solo.io.Action]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.Action" >}}) | repeated | Defines an ordered descriptor set that maps part of the request to a descriptor sent to the rate limit server and matched against the rate limit rules. |
  | setActions | [][ratelimit.api.solo.io.Action]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.Action" >}}) | repeated | Defines an unordered descriptor set that maps part of the request to a descriptor sent to the rate limit server and matched against the rate limit rules. |
  





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

