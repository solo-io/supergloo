
---

title: "rate_limit_server_config.proto"

---

## Package : `networking.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for rate_limit_server_config.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## rate_limit_server_config.proto


## Table of Contents
  - [RateLimiterConfigSpec](#networking.mesh.gloo.solo.io.RateLimiterConfigSpec)
  - [RateLimiterConfigStatus](#networking.mesh.gloo.solo.io.RateLimiterConfigStatus)

  - [RateLimiterConfigStatus.State](#networking.mesh.gloo.solo.io.RateLimiterConfigStatus.State)






<a name="networking.mesh.gloo.solo.io.RateLimiterConfigSpec"></a>

### RateLimiterConfigSpec
RateLimiterConfig contains the configuration for the Gloo Rate Limiter, the external rate-limiting server used by mesh proxies to rate-limit HTTP requests. One or more rate limiter servers may be deployed in order to rate limit traffic across East-West and North-South routes. The RateLimiterConfig allows users to map a single rate-limiter configuration to multiple rate-limiter server instances, deployed across managed clusters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| serverConfigRefs | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated | The per-server rate limit config objects will be generated from the given config for each provided ref. Each rate limit server must be configured to read its server configuration from one of these refs. |
  | rateLimitConfig | [ratelimit.api.solo.io.RateLimitConfigSpec]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.rate-limiter.v1alpha1.ratelimit#ratelimit.api.solo.io.RateLimitConfigSpec" >}}) |  | the configuration which will be deployed to the selected rate limit servers. |
  





<a name="networking.mesh.gloo.solo.io.RateLimiterConfigStatus"></a>

### RateLimiterConfigStatus
The current status of the `RateLimitConfig`.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| state | [networking.mesh.gloo.solo.io.RateLimiterConfigStatus.State]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.rate_limit_server_config#networking.mesh.gloo.solo.io.RateLimiterConfigStatus.State" >}}) |  | The current state of the `RateLimitConfig`. |
  | message | string |  | A human-readable string explaining the status. |
  | configuredServers | [][core.skv2.solo.io.ClusterObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ClusterObjectRef" >}}) | repeated | a list of rate limit server workloads which have been configured with this RateLimiterConfig |
  | observedGeneration | int64 |  | The observed generation of the resource. When this matches the metadata.generation of the resource, it indicates the status is up-to-date. |
  




 <!-- end messages -->


<a name="networking.mesh.gloo.solo.io.RateLimiterConfigStatus.State"></a>

### RateLimiterConfigStatus.State


| Name | Number | Description |
| ---- | ------ | ----------- |
| PENDING | 0 |  |
| ACCEPTED | 1 |  |
| REJECTED | 2 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

