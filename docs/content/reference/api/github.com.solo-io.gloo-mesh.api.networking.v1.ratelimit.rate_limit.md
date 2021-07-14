
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
  - [RateLimitClient](#ratelimit.networking.mesh.gloo.solo.io.RateLimitClient)
  - [RateLimitClient.AdvancedRateLimit](#ratelimit.networking.mesh.gloo.solo.io.RateLimitClient.AdvancedRateLimit)
  - [RateLimitClient.BasicRateLimit](#ratelimit.networking.mesh.gloo.solo.io.RateLimitClient.BasicRateLimit)
  - [RouteRateLimit](#ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit)







<a name="ratelimit.networking.mesh.gloo.solo.io.GatewayRateLimit"></a>

### GatewayRateLimit
Configure the Rate-Limit Filter on a Gateway


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ratelimitServerRef | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  |  |
  | requestTimeout | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  |  |
  | denyOnFail | bool |  |  |
  | rateLimitBeforeAuth | bool |  | Set this is set to true if you would like to rate limit traffic before applying external auth to it. *Note*: When this is true, you will lose some features like being able to rate limit a request based on its auth state |
  





<a name="ratelimit.networking.mesh.gloo.solo.io.RateLimitClient"></a>

### RateLimitClient



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| advanced | [ratelimit.networking.mesh.gloo.solo.io.RateLimitClient.AdvancedRateLimit]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.RateLimitClient.AdvancedRateLimit" >}}) |  |  |
  | basic | [ratelimit.networking.mesh.gloo.solo.io.RateLimitClient.BasicRateLimit]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.RateLimitClient.BasicRateLimit" >}}) |  |  |
  





<a name="ratelimit.networking.mesh.gloo.solo.io.RateLimitClient.AdvancedRateLimit"></a>

### RateLimitClient.AdvancedRateLimit
Use this field if you want to inline the Envoy rate limits for this VirtualHost. Note that this does not configure the rate limit server. If you are running Gloo Mesh, you need to specify the server configuration via the appropriate field in the Gloo Mesh `GatewayRateLimit` resource on the gateway. If you are running a custom rate limit server you need to configure it yourself.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| actions | [][ratelimit.api.solo.io.RateLimitActions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.RateLimitActions" >}}) | repeated | Actions specify how the client (Envoy) will compose the descriptors that will be sent to the server to make a rate limiting decision. |
  





<a name="ratelimit.networking.mesh.gloo.solo.io.RateLimitClient.BasicRateLimit"></a>

### RateLimitClient.BasicRateLimit
Basic rate-limiting API






<a name="ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit"></a>

### RouteRateLimit
Rate limit configuration for a Route or TrafficPolicy. Configures rate limits for individual HTTP routes


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ratelimitConfigRef | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | RateLimitConfig ref |
  | client | [ratelimit.networking.mesh.gloo.solo.io.RateLimitClient]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.RateLimitClient" >}}) |  |  |
  | ratelimitClientConfigRef | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  |  |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

