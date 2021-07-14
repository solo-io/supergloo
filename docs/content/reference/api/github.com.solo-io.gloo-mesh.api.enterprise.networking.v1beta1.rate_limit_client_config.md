
---

title: "rate_limit_client_config.proto"

---

## Package : `networking.enterprise.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for rate_limit_client_config.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## rate_limit_client_config.proto


## Table of Contents
  - [RateLimitClientConfigSpec](#networking.enterprise.mesh.gloo.solo.io.RateLimitClientConfigSpec)
  - [RateLimitClientConfigStatus](#networking.enterprise.mesh.gloo.solo.io.RateLimitClientConfigStatus)







<a name="networking.enterprise.mesh.gloo.solo.io.RateLimitClientConfigSpec"></a>

### RateLimitClientConfigSpec
RateLimitClientConfig contains the client configuration for the Gloo Rate Limiter


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rateLimits | [][ratelimit.networking.mesh.gloo.solo.io.RateLimitClient]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.RateLimitClient" >}}) | repeated | The RateLimitClient specifies the ratelimit Actions which the client (Envoy) will use to compose the descriptors that will be sent to the server to make a rate limiting decision. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.RateLimitClientConfigStatus"></a>

### RateLimitClientConfigStatus
The current status of the `RateLimitClientConfig`.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the RateLimiterClientConfig metadata. If the `observedGeneration` does not match `metadata.generation`, Gloo Mesh has not processed the most recent version of this resource. |
  | errors | []string | repeated | Any errors found while processing this generation of the resource. |
  | warnings | []string | repeated | Any warnings found while processing this generation of the resource. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

