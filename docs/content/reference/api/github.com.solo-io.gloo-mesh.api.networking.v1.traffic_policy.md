
---

title: "traffic_policy.proto"

---

## Package : `networking.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for traffic_policy.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## traffic_policy.proto


## Table of Contents
  - [TrafficPolicySpec](#networking.mesh.gloo.solo.io.TrafficPolicySpec)
  - [TrafficPolicySpec.Policy](#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy)
  - [TrafficPolicySpec.Policy.CorsPolicy](#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.CorsPolicy)
  - [TrafficPolicySpec.Policy.DLPPolicy](#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.DLPPolicy)
  - [TrafficPolicySpec.Policy.ExtAuth](#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.ExtAuth)
  - [TrafficPolicySpec.Policy.FaultInjection](#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.FaultInjection)
  - [TrafficPolicySpec.Policy.FaultInjection.Abort](#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.FaultInjection.Abort)
  - [TrafficPolicySpec.Policy.MTLS](#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.MTLS)
  - [TrafficPolicySpec.Policy.MTLS.Istio](#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.MTLS.Istio)
  - [TrafficPolicySpec.Policy.Mirror](#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.Mirror)
  - [TrafficPolicySpec.Policy.MultiDestination](#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.MultiDestination)
  - [TrafficPolicySpec.Policy.OutlierDetection](#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.OutlierDetection)
  - [TrafficPolicySpec.Policy.RetryPolicy](#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.RetryPolicy)
  - [TrafficPolicySpec.Policy.Transform](#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.Transform)
  - [TrafficPolicyStatus](#networking.mesh.gloo.solo.io.TrafficPolicyStatus)
  - [TrafficPolicyStatus.DestinationsEntry](#networking.mesh.gloo.solo.io.TrafficPolicyStatus.DestinationsEntry)

  - [TrafficPolicySpec.Policy.MTLS.Istio.TLSmode](#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.MTLS.Istio.TLSmode)






<a name="networking.mesh.gloo.solo.io.TrafficPolicySpec"></a>

### TrafficPolicySpec
Applies L7 routing and post-routing configuration on selected network edges.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sourceSelector | [][common.mesh.gloo.solo.io.WorkloadSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.WorkloadSelector" >}}) | repeated | Specify the Workloads (traffic sources) this TrafficPolicy applies to. Omit to apply to all Workloads. |
  | destinationSelector | [][common.mesh.gloo.solo.io.DestinationSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.DestinationSelector" >}}) | repeated | Specify the Destinations (destinations) this TrafficPolicy applies to. Omit to apply to all Destinations. |
  | httpRequestMatchers | [][networking.mesh.gloo.solo.io.HttpMatcher]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.request_matchers#networking.mesh.gloo.solo.io.HttpMatcher" >}}) | repeated | Specify criteria that HTTP requests must satisfy for the TrafficPolicy to apply. Conditions defined within a single matcher are conjunctive, i.e. all conditions must be satisfied for a match to occur. Conditions defined between different matchers are disjunctive, i.e. at least one matcher must be satisfied for the TrafficPolicy to apply. Omit to apply to any HTTP request. |
  | policy | [networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy" >}}) |  | Specify L7 routing and post-routing configuration. |
  





<a name="networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy"></a>

### TrafficPolicySpec.Policy
Specify L7 routing and post-routing configuration.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| trafficShift | [networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.MultiDestination]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.MultiDestination" >}}) |  | Shift traffic to a different destination. |
  | faultInjection | [networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.FaultInjection]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.FaultInjection" >}}) |  | Inject faulty responses. |
  | requestTimeout | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Set a timeout on requests. |
  | retries | [networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.RetryPolicy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.RetryPolicy" >}}) |  | Set a retry policy on requests. |
  | corsPolicy | [networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.CorsPolicy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.CorsPolicy" >}}) |  | Set a Cross-Origin Resource Sharing policy (CORS) for requests. Refer to [this link](https://developer.mozilla.org/en-US/docs/Web/HTTP/Access_control_CORS) for further details about cross origin resource sharing. |
  | mirror | [networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.Mirror]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.Mirror" >}}) |  | Mirror traffic to a another destination (traffic will be sent to its original destination in addition to the mirrored destinations). |
  | headerManipulation | [networking.mesh.gloo.solo.io.HeaderManipulation]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.weighed_destination#networking.mesh.gloo.solo.io.HeaderManipulation" >}}) |  | Manipulate request and response headers. |
  | outlierDetection | [networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.OutlierDetection]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.OutlierDetection" >}}) |  | Configure [outlier detection](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/outlier) on the selected destinations. Specifying this field requires an empty `source_selector` because it must apply to all traffic. |
  | mtls | [networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.MTLS]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.MTLS" >}}) |  | Configure mTLS settings. If specified will override global default defined in Settings. |
  | csrf | [csrf.networking.mesh.gloo.solo.io.CsrfPolicy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.csrf.csrf#csrf.networking.mesh.gloo.solo.io.CsrfPolicy" >}}) |  | Configure the Envoy based CSRF filter |
  | rateLimit | [ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.ratelimit.rate_limit#ratelimit.networking.mesh.gloo.solo.io.RouteRateLimit" >}}) |  | Config the Envoy based Ratelimit filter |
  





<a name="networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.CorsPolicy"></a>

### TrafficPolicySpec.Policy.CorsPolicy
Specify Cross-Origin Resource Sharing policy (CORS) for requests. Refer to [this link](https://developer.mozilla.org/en-US/docs/Web/HTTP/Access_control_CORS) for further details about cross origin resource sharing.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| allowOrigins | [][common.mesh.gloo.solo.io.StringMatch]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.string_match#common.mesh.gloo.solo.io.StringMatch" >}}) | repeated | String patterns that match allowed origins. An origin is allowed if any of the string matchers match. |
  | allowMethods | []string | repeated | List of HTTP methods allowed to access the resource. The content will be serialized to the `Access-Control-Allow-Methods` header. |
  | allowHeaders | []string | repeated | List of HTTP headers that can be used when requesting the resource. Serialized to the `Access-Control-Allow-Headers` header. |
  | exposeHeaders | []string | repeated | A list of HTTP headers that browsers are allowed to access. Serialized to the `Access-Control-Expose-Headers` header. |
  | maxAge | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Specify how long the results of a preflight request can be cached. Serialized to the `Access-Control-Max-Age` header. |
  | allowCredentials | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | Indicates whether the caller is allowed to send the actual request (not the preflight) using credentials. Translates to the `Access-Control-Allow-Credentials` header. |
  





<a name="networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.DLPPolicy"></a>

### TrafficPolicySpec.Policy.DLPPolicy
DLP filter config.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| todo | string |  | TODO: implement |
  





<a name="networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.ExtAuth"></a>

### TrafficPolicySpec.Policy.ExtAuth
ExtAuth filter config.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| todo | string |  | TODO: implement |
  





<a name="networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.FaultInjection"></a>

### TrafficPolicySpec.Policy.FaultInjection
Specify one or more faults to inject for the selected network edge.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fixedDelay | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Add a delay of a fixed duration before sending the request. Format: `1h`/`1m`/`1s`/`1ms`. MUST be >=1ms. |
  | abort | [networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.FaultInjection.Abort]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.FaultInjection.Abort" >}}) |  | Abort the request and return the specified error code back to traffic source. |
  | percentage | double |  | Percentage of requests to be faulted. Values range between 0 and 100. If omitted all requests will be faulted. |
  





<a name="networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.FaultInjection.Abort"></a>

### TrafficPolicySpec.Policy.FaultInjection.Abort
Abort the request and return the specified error code back to traffic source.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| httpStatus | int32 |  | Required. HTTP status code to use to abort the request. |
  





<a name="networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.MTLS"></a>

### TrafficPolicySpec.Policy.MTLS
Configure mTLS settings on destinations. If specified this overrides the global default defined in Settings.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| istio | [networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.MTLS.Istio]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.MTLS.Istio" >}}) |  | Istio TLS settings. |
  





<a name="networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.MTLS.Istio"></a>

### TrafficPolicySpec.Policy.MTLS.Istio
Istio TLS settings.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tlsMode | [networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.MTLS.Istio.TLSmode]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.MTLS.Istio.TLSmode" >}}) |  | TLS connection mode |
  





<a name="networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.Mirror"></a>

### TrafficPolicySpec.Policy.Mirror
Mirror traffic to a another destination (traffic will be sent to its original destination in addition to the mirrored destinations).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeService | [core.skv2.solo.io.ClusterObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ClusterObjectRef" >}}) |  | Reference (name, namespace, Gloo Mesh cluster) to a Kubernetes service. |
  | percentage | double |  | Percentage of traffic to mirror. If omitted all traffic will be mirrored. Values must be between 0 and 100. |
  | port | uint32 |  | Port on the destination to receive traffic. Required if the destination exposes multiple ports. |
  





<a name="networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.MultiDestination"></a>

### TrafficPolicySpec.Policy.MultiDestination
Specify a traffic shift destination.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| destinations | [][networking.mesh.gloo.solo.io.WeightedDestination]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.weighed_destination#networking.mesh.gloo.solo.io.WeightedDestination" >}}) | repeated | Specify weighted traffic shift destinations. |
  





<a name="networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.OutlierDetection"></a>

### TrafficPolicySpec.Policy.OutlierDetection
Configure [outlier detection](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/outlier) on the selected destinations. Specifying this field requires an empty `source_selector` because it must apply to all traffic.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| consecutiveErrors | uint32 |  | The number of errors before a host is ejected from the connection pool. A default will be used if not set. |
  | interval | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | The time interval between ejection sweep analysis. Format: `1h`/`1m`/`1s`/`1ms`. Must be >= `1ms`. A default will be used if not set. |
  | baseEjectionTime | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | The minimum ejection duration. Format: `1h`/`1m`/`1s`/`1ms`. Must be >= `1ms`. A default will be used if not set. |
  | maxEjectionPercent | uint32 |  | The maximum percentage of hosts that can be ejected from the load balancing pool. At least one host will be ejected regardless of the value. Must be between 0 and 100. A default will be used if not set. |
  





<a name="networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.RetryPolicy"></a>

### TrafficPolicySpec.Policy.RetryPolicy
Specify retries for failed requests.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| attempts | int32 |  | Number of retries for a given request |
  | perTryTimeout | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Timeout per retry attempt for a given request. Format: `1h`/`1m`/`1s`/`1ms`. *Must be >= 1ms*. |
  





<a name="networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.Transform"></a>

### TrafficPolicySpec.Policy.Transform
Transform filter config.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| todo | string |  | TODO: implement |
  





<a name="networking.mesh.gloo.solo.io.TrafficPolicyStatus"></a>

### TrafficPolicyStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the TrafficPolicy metadata. If the `observedGeneration` does not match `metadata.generation`, Gloo Mesh has not processed the most recent version of this resource. |
  | state | [common.mesh.gloo.solo.io.ApprovalState]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.validation_state#common.mesh.gloo.solo.io.ApprovalState" >}}) |  | The state of the overall resource. It will only show accepted if it has been successfully applied to all selected Destinations. |
  | destinations | [][networking.mesh.gloo.solo.io.TrafficPolicyStatus.DestinationsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicyStatus.DestinationsEntry" >}}) | repeated | The status of the TrafficPolicy for each selected Destination. A TrafficPolicy may be Accepted for some Destinations and rejected for others. |
  | workloads | []string | repeated | The list of selected Workloads for which this policy has been applied. |
  | errors | []string | repeated | Any errors found while processing this generation of the resource. |
  





<a name="networking.mesh.gloo.solo.io.TrafficPolicyStatus.DestinationsEntry"></a>

### TrafficPolicyStatus.DestinationsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [networking.mesh.gloo.solo.io.ApprovalStatus]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.status#networking.mesh.gloo.solo.io.ApprovalStatus" >}}) |  |  |
  




 <!-- end messages -->


<a name="networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.MTLS.Istio.TLSmode"></a>

### TrafficPolicySpec.Policy.MTLS.Istio.TLSmode
TLS connection mode. Enums correspond to those [defined here](https://github.com/istio/api/blob/00636152b9d9254b614828a65723840282a177d3/networking/v1beta1/destination_rule.proto#L886)

| Name | Number | Description |
| ---- | ------ | ----------- |
| DISABLE | 0 | Do not originate a TLS connection to the upstream endpoint. |
| SIMPLE | 1 | Originate a TLS connection to the upstream endpoint. |
| ISTIO_MUTUAL | 2 | Secure connections to the upstream using mutual TLS by presenting client certificates for authentication. This mode uses certificates generated automatically by Istio for mTLS authentication. When this mode is used, all other fields in `ClientTLSSettings` should be empty. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

