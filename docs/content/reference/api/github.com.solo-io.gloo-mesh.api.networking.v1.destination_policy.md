
---

title: "destination_policy.proto"

---

## Package : `networking.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for destination_policy.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## destination_policy.proto


## Table of Contents
  - [DestinationPolicySpec](#networking.mesh.gloo.solo.io.DestinationPolicySpec)
  - [DestinationPolicySpec.Policy](#networking.mesh.gloo.solo.io.DestinationPolicySpec.Policy)
  - [DestinationPolicySpec.Policy.MTLS](#networking.mesh.gloo.solo.io.DestinationPolicySpec.Policy.MTLS)
  - [DestinationPolicySpec.Policy.MTLS.Istio](#networking.mesh.gloo.solo.io.DestinationPolicySpec.Policy.MTLS.Istio)
  - [DestinationPolicySpec.Policy.OutlierDetection](#networking.mesh.gloo.solo.io.DestinationPolicySpec.Policy.OutlierDetection)
  - [DestinationPolicyStatus](#networking.mesh.gloo.solo.io.DestinationPolicyStatus)

  - [DestinationPolicySpec.Policy.MTLS.Istio.TLSmode](#networking.mesh.gloo.solo.io.DestinationPolicySpec.Policy.MTLS.Istio.TLSmode)






<a name="networking.mesh.gloo.solo.io.DestinationPolicySpec"></a>

### DestinationPolicySpec
Applies L7 routing and post-routing configuration on selected network edges.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| destinationSelector | [][common.mesh.gloo.solo.io.DestinationSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.DestinationSelector" >}}) | repeated | Specify the Destinations (destinations) this DestinationPolicy applies to. Omit to apply to all Destinations. |
  | policy | [networking.mesh.gloo.solo.io.DestinationPolicySpec.Policy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.destination_policy#networking.mesh.gloo.solo.io.DestinationPolicySpec.Policy" >}}) |  | Specify L7 routing and post-routing configuration. |
  





<a name="networking.mesh.gloo.solo.io.DestinationPolicySpec.Policy"></a>

### DestinationPolicySpec.Policy
Specify L7 routing and post-routing configuration.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| outlierDetection | [networking.mesh.gloo.solo.io.DestinationPolicySpec.Policy.OutlierDetection]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.destination_policy#networking.mesh.gloo.solo.io.DestinationPolicySpec.Policy.OutlierDetection" >}}) |  | Configure [outlier detection](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/outlier) on the selected destinations. |
  | mtls | [networking.mesh.gloo.solo.io.DestinationPolicySpec.Policy.MTLS]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.destination_policy#networking.mesh.gloo.solo.io.DestinationPolicySpec.Policy.MTLS" >}}) |  | Configure mTLS settings. If specified will override global default defined in Settings. |
  





<a name="networking.mesh.gloo.solo.io.DestinationPolicySpec.Policy.MTLS"></a>

### DestinationPolicySpec.Policy.MTLS
Configure mTLS settings on destinations. If specified this overrides the global default defined in Settings.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| istio | [networking.mesh.gloo.solo.io.DestinationPolicySpec.Policy.MTLS.Istio]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.destination_policy#networking.mesh.gloo.solo.io.DestinationPolicySpec.Policy.MTLS.Istio" >}}) |  | Istio TLS settings. |
  





<a name="networking.mesh.gloo.solo.io.DestinationPolicySpec.Policy.MTLS.Istio"></a>

### DestinationPolicySpec.Policy.MTLS.Istio
Istio TLS settings.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tlsMode | [networking.mesh.gloo.solo.io.DestinationPolicySpec.Policy.MTLS.Istio.TLSmode]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.destination_policy#networking.mesh.gloo.solo.io.DestinationPolicySpec.Policy.MTLS.Istio.TLSmode" >}}) |  | TLS connection mode |
  





<a name="networking.mesh.gloo.solo.io.DestinationPolicySpec.Policy.OutlierDetection"></a>

### DestinationPolicySpec.Policy.OutlierDetection
Configure [outlier detection](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/outlier) on the selected destinations.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| consecutiveErrors | uint32 |  | The number of errors before a host is ejected from the connection pool. A default will be used if not set. |
  | interval | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | The time interval between ejection sweep analysis. Format: `1h`/`1m`/`1s`/`1ms`. Must be >= `1ms`. A default will be used if not set. |
  | baseEjectionTime | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | The minimum ejection duration. Format: `1h`/`1m`/`1s`/`1ms`. Must be >= `1ms`. A default will be used if not set. |
  | maxEjectionPercent | uint32 |  | The maximum percentage of hosts that can be ejected from the load balancing pool. At least one host will be ejected regardless of the value. Must be between 0 and 100. A default will be used if not set. |
  





<a name="networking.mesh.gloo.solo.io.DestinationPolicyStatus"></a>

### DestinationPolicyStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the DestinationPolicy metadata. If the `observedGeneration` does not match `metadata.generation`, Gloo Mesh has not processed the most recent version of this resource. |
  | state | [common.mesh.gloo.solo.io.ApprovalState]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.validation_state#common.mesh.gloo.solo.io.ApprovalState" >}}) |  | The state of the overall resource. It will only show accepted if it has been successfully applied to all selected Destinations. |
  | errors | []string | repeated | Any errors found while processing this generation of the resource. |
  




 <!-- end messages -->


<a name="networking.mesh.gloo.solo.io.DestinationPolicySpec.Policy.MTLS.Istio.TLSmode"></a>

### DestinationPolicySpec.Policy.MTLS.Istio.TLSmode
TLS connection mode. Enums correspond to those [defined here](https://github.com/istio/api/blob/00636152b9d9254b614828a65723840282a177d3/networking/v1beta1/destination_rule.proto#L886)

| Name | Number | Description |
| ---- | ------ | ----------- |
| DISABLE | 0 | Do not originate a TLS connection to the upstream endpoint. |
| SIMPLE | 1 | Originate a TLS connection to the upstream endpoint. |
| ISTIO_MUTUAL | 2 | Secure connections to the upstream using mutual TLS by presenting client certificates for authentication. This mode uses certificates generated automatically by Istio for mTLS authentication. When this mode is used, all other fields in `ClientTLSSettings` should be empty. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

