
---

title: "rate_limit_server_config.proto"

---

## Package : `networking.enterprise.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for rate_limit_server_config.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## rate_limit_server_config.proto


## Table of Contents
  - [RateLimiterServerConfigSpec](#networking.enterprise.mesh.gloo.solo.io.RateLimiterServerConfigSpec)
  - [RateLimiterServerConfigStatus](#networking.enterprise.mesh.gloo.solo.io.RateLimiterServerConfigStatus)







<a name="networking.enterprise.mesh.gloo.solo.io.RateLimiterServerConfigSpec"></a>

### RateLimiterServerConfigSpec
RateLimiterConfig contains the configuration for the Gloo Rate Limiter, the external rate-limiting server used by mesh proxies to rate-limit HTTP requests. One or more rate limiter servers may be deployed in order to rate limit traffic across East-West and North-South routes. The RateLimiterConfig allows users to map a single rate-limiter configuration to multiple rate-limiter server instances, deployed across managed clusters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| serverConfigRefs | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated | The per-server rate limit config objects will be generated from the given config for each provided ref. Each rate limit server must be configured to read its server configuration from one of these refs. |
  | rateLimitConfig | [ratelimit.api.solo.io.RateLimitConfigSpec]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.RateLimitConfigSpec" >}}) |  | the configuration which will be deployed to the selected rate limit servers. TODO: move disable validation annotation into solo-apis |
  





<a name="networking.enterprise.mesh.gloo.solo.io.RateLimiterServerConfigStatus"></a>

### RateLimiterServerConfigStatus
The current status of the `RateLimitConfig`.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the RateLimiterServerConfig metadata. If the `observedGeneration` does not match `metadata.generation`, Gloo Mesh has not processed the most recent version of this resource. |
  | errors | []string | repeated | Any errors found while processing this generation of the resource. |
  | warnings | []string | repeated | Any warnings found while processing this generation of the resource. |
  | configuredServers | [][core.skv2.solo.io.ClusterObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ClusterObjectRef" >}}) | repeated | a list of rate limit server workloads which have been configured with this RateLimiterConfig |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

