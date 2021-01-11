
---

---

## Package : `envoy.config.core.v3`



<a name="top"></a>

<a name="API Reference for config_source.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## config_source.proto


## Table of Contents
  - [AggregatedConfigSource](#envoy.config.core.v3.AggregatedConfigSource)
  - [ApiConfigSource](#envoy.config.core.v3.ApiConfigSource)
  - [ConfigSource](#envoy.config.core.v3.ConfigSource)
  - [RateLimitSettings](#envoy.config.core.v3.RateLimitSettings)
  - [SelfConfigSource](#envoy.config.core.v3.SelfConfigSource)

  - [ApiConfigSource.ApiType](#envoy.config.core.v3.ApiConfigSource.ApiType)
  - [ApiVersion](#envoy.config.core.v3.ApiVersion)






<a name="envoy.config.core.v3.AggregatedConfigSource"></a>

### AggregatedConfigSource







<a name="envoy.config.core.v3.ApiConfigSource"></a>

### ApiConfigSource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| apiType | [envoy.config.core.v3.ApiConfigSource.ApiType]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.config_source#envoy.config.core.v3.ApiConfigSource.ApiType" >}}) |  | API type (gRPC, REST, delta gRPC) |
  | transportApiVersion | [envoy.config.core.v3.ApiVersion]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.config_source#envoy.config.core.v3.ApiVersion" >}}) |  | API version for xDS transport protocol. This describes the xDS gRPC/REST endpoint and version of [Delta]DiscoveryRequest/Response used on the wire. |
  | clusterNames | []string | repeated | Cluster names should be used only with REST. If > 1 cluster is defined, clusters will be cycled through if any kind of failure occurs.<br>.. note::<br> The cluster with name ``cluster_name`` must be statically defined and its  type must not be ``EDS``. |
  | grpcServices | [][envoy.config.core.v3.GrpcService]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.grpc_service#envoy.config.core.v3.GrpcService" >}}) | repeated | Multiple gRPC services be provided for GRPC. If > 1 cluster is defined, services will be cycled through if any kind of failure occurs. |
  | refreshDelay | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | For REST APIs, the delay between successive polls. |
  | requestTimeout | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | For REST APIs, the request timeout. If not set, a default value of 1s will be used. |
  | rateLimitSettings | [envoy.config.core.v3.RateLimitSettings]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.config_source#envoy.config.core.v3.RateLimitSettings" >}}) |  | For GRPC APIs, the rate limit settings. If present, discovery requests made by Envoy will be rate limited. |
  | setNodeOnFirstMessageOnly | bool |  | Skip the node identifier in subsequent discovery requests for streaming gRPC config types. |
  





<a name="envoy.config.core.v3.ConfigSource"></a>

### ConfigSource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| authorities | [][xds.core.v3.Authority]({{< versioned_link_path fromRoot="/reference/api/github.com.cncf.udpa.xds.core.v3.authority#xds.core.v3.Authority" >}}) | repeated | Authorities that this config source may be used for. An authority specified in a xdstp:// URL is resolved to a *ConfigSource* prior to configuration fetch. This field provides the association between authority name and configuration source. [#not-implemented-hide:] |
  | path | string |  | Path on the filesystem to source and watch for configuration updates. When sourcing configuration for :ref:`secret <envoy_api_msg_extensions.transport_sockets.tls.v3.Secret>`, the certificate and key files are also watched for updates.<br>.. note::<br> The path to the source must exist at config load time.<br>.. note::<br>  Envoy will only watch the file path for *moves.* This is because in general only moves   are atomic. The same method of swapping files as is demonstrated in the   :ref:`runtime documentation <config_runtime_symbolic_link_swap>` can be used here also. |
  | apiConfigSource | [envoy.config.core.v3.ApiConfigSource]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.config_source#envoy.config.core.v3.ApiConfigSource" >}}) |  | API configuration source. |
  | ads | [envoy.config.core.v3.AggregatedConfigSource]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.config_source#envoy.config.core.v3.AggregatedConfigSource" >}}) |  | When set, ADS will be used to fetch resources. The ADS API configuration source in the bootstrap configuration is used. |
  | self | [envoy.config.core.v3.SelfConfigSource]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.config_source#envoy.config.core.v3.SelfConfigSource" >}}) |  | [#not-implemented-hide:] When set, the client will access the resources from the same server it got the ConfigSource from, although not necessarily from the same stream. This is similar to the :ref:`ads<envoy_api_field.ConfigSource.ads>` field, except that the client may use a different stream to the same server. As a result, this field can be used for things like LRS that cannot be sent on an ADS stream. It can also be used to link from (e.g.) LDS to RDS on the same server without requiring the management server to know its name or required credentials. [#next-major-version: In xDS v3, consider replacing the ads field with this one, since this field can implicitly mean to use the same stream in the case where the ConfigSource is provided via ADS and the specified data can also be obtained via ADS.] |
  | initialFetchTimeout | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | When this timeout is specified, Envoy will wait no longer than the specified time for first config response on this xDS subscription during the :ref:`initialization process <arch_overview_initialization>`. After reaching the timeout, Envoy will move to the next initialization phase, even if the first config is not delivered yet. The timer is activated when the xDS API subscription starts, and is disarmed on first config update or on error. 0 means no timeout - Envoy will wait indefinitely for the first xDS config (unless another timeout applies). The default is 15s. |
  | resourceApiVersion | [envoy.config.core.v3.ApiVersion]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.config_source#envoy.config.core.v3.ApiVersion" >}}) |  | API version for xDS resources. This implies the type URLs that the client will request for resources and the resource type that the client will in turn expect to be delivered. |
  





<a name="envoy.config.core.v3.RateLimitSettings"></a>

### RateLimitSettings



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| maxTokens | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | Maximum number of tokens to be used for rate limiting discovery request calls. If not set, a default value of 100 will be used. |
  | fillRate | [google.protobuf.DoubleValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.DoubleValue" >}}) |  | Rate at which tokens will be filled per second. If not set, a default fill rate of 10 tokens per second will be used. |
  





<a name="envoy.config.core.v3.SelfConfigSource"></a>

### SelfConfigSource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| transportApiVersion | [envoy.config.core.v3.ApiVersion]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.config_source#envoy.config.core.v3.ApiVersion" >}}) |  | API version for xDS transport protocol. This describes the xDS gRPC/REST endpoint and version of [Delta]DiscoveryRequest/Response used on the wire. |
  




 <!-- end messages -->


<a name="envoy.config.core.v3.ApiConfigSource.ApiType"></a>

### ApiConfigSource.ApiType


| Name | Number | Description |
| ---- | ------ | ----------- |
| DEPRECATED_AND_UNAVAILABLE_DO_NOT_USE | 0 | Ideally this would be 'reserved 0' but one can't reserve the default value. Instead we throw an exception if this is ever used. |
| REST | 1 | REST-JSON v2 API. The `canonical JSON encoding <https://developers.google.com/protocol-buffers/docs/proto3#json>`_ for the v2 protos is used. |
| GRPC | 2 | SotW gRPC service. |
| DELTA_GRPC | 3 | Using the delta xDS gRPC service, i.e. DeltaDiscovery{Request,Response} rather than Discovery{Request,Response}. Rather than sending Envoy the entire state with every update, the xDS server only sends what has changed since the last update. |
| AGGREGATED_GRPC | 5 | SotW xDS gRPC with ADS. All resources which resolve to this configuration source will be multiplexed on a single connection to an ADS endpoint. [#not-implemented-hide:] |
| AGGREGATED_DELTA_GRPC | 6 | Delta xDS gRPC with ADS. All resources which resolve to this configuration source will be multiplexed on a single connection to an ADS endpoint. [#not-implemented-hide:] |



<a name="envoy.config.core.v3.ApiVersion"></a>

### ApiVersion


| Name | Number | Description |
| ---- | ------ | ----------- |
| AUTO | 0 | When not specified, we assume v2, to ease migration to Envoy's stable API versioning. If a client does not support v2 (e.g. due to deprecation), this is an invalid value. |
| V2 | 1 | Use xDS v2 API. |
| V3 | 2 | Use xDS v3 API. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

