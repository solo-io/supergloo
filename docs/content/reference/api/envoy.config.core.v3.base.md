
---

---

## Package : `envoy.config.core.v3`



<a name="top"></a>

<a name="API Reference for base.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## base.proto


## Table of Contents
  - [AsyncDataSource](#envoy.config.core.v3.AsyncDataSource)
  - [BuildVersion](#envoy.config.core.v3.BuildVersion)
  - [ControlPlane](#envoy.config.core.v3.ControlPlane)
  - [DataSource](#envoy.config.core.v3.DataSource)
  - [Extension](#envoy.config.core.v3.Extension)
  - [HeaderMap](#envoy.config.core.v3.HeaderMap)
  - [HeaderValue](#envoy.config.core.v3.HeaderValue)
  - [HeaderValueOption](#envoy.config.core.v3.HeaderValueOption)
  - [Locality](#envoy.config.core.v3.Locality)
  - [Metadata](#envoy.config.core.v3.Metadata)
  - [Metadata.FilterMetadataEntry](#envoy.config.core.v3.Metadata.FilterMetadataEntry)
  - [Node](#envoy.config.core.v3.Node)
  - [RemoteDataSource](#envoy.config.core.v3.RemoteDataSource)
  - [RetryPolicy](#envoy.config.core.v3.RetryPolicy)
  - [RuntimeDouble](#envoy.config.core.v3.RuntimeDouble)
  - [RuntimeFeatureFlag](#envoy.config.core.v3.RuntimeFeatureFlag)
  - [RuntimeFractionalPercent](#envoy.config.core.v3.RuntimeFractionalPercent)
  - [RuntimePercent](#envoy.config.core.v3.RuntimePercent)
  - [RuntimeUInt32](#envoy.config.core.v3.RuntimeUInt32)
  - [TransportSocket](#envoy.config.core.v3.TransportSocket)
  - [WatchedDirectory](#envoy.config.core.v3.WatchedDirectory)

  - [RequestMethod](#envoy.config.core.v3.RequestMethod)
  - [RoutingPriority](#envoy.config.core.v3.RoutingPriority)
  - [TrafficDirection](#envoy.config.core.v3.TrafficDirection)






<a name="envoy.config.core.v3.AsyncDataSource"></a>

### AsyncDataSource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| local | [envoy.config.core.v3.DataSource]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.DataSource" >}}) |  | Local async data source. |
  | remote | [envoy.config.core.v3.RemoteDataSource]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.RemoteDataSource" >}}) |  | Remote async data source. |
  





<a name="envoy.config.core.v3.BuildVersion"></a>

### BuildVersion



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| version | [envoy.type.v3.SemanticVersion]({{< versioned_link_path fromRoot="/reference/api/envoy.type.v3.semantic_version#envoy.type.v3.SemanticVersion" >}}) |  | SemVer version of extension. |
  | metadata | [google.protobuf.Struct]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.struct#google.protobuf.Struct" >}}) |  | Free-form build information. Envoy defines several well known keys in the source/common/version/version.h file |
  





<a name="envoy.config.core.v3.ControlPlane"></a>

### ControlPlane



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| identifier | string |  | An opaque control plane identifier that uniquely identifies an instance of control plane. This can be used to identify which control plane instance, the Envoy is connected to. |
  





<a name="envoy.config.core.v3.DataSource"></a>

### DataSource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| filename | string |  | Local filesystem data source. |
  | inlineBytes | bytes |  | Bytes inlined in the configuration. |
  | inlineString | string |  | String inlined in the configuration. |
  





<a name="envoy.config.core.v3.Extension"></a>

### Extension



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | This is the name of the Envoy filter as specified in the Envoy configuration, e.g. envoy.filters.http.router, com.acme.widget. |
  | category | string |  | Category of the extension. Extension category names use reverse DNS notation. For instance "envoy.filters.listener" for Envoy's built-in listener filters or "com.acme.filters.http" for HTTP filters from acme.com vendor. [#comment:TODO(yanavlasov): Link to the doc with existing envoy category names.] |
  | typeDescriptor | string |  | [#not-implemented-hide:] Type descriptor of extension configuration proto. [#comment:TODO(yanavlasov): Link to the doc with existing configuration protos.] [#comment:TODO(yanavlasov): Add tests when PR #9391 lands.] |
  | version | [envoy.config.core.v3.BuildVersion]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.BuildVersion" >}}) |  | The version is a property of the extension and maintained independently of other extensions and the Envoy API. This field is not set when extension did not provide version information. |
  | disabled | bool |  | Indicates that the extension is present but was disabled via dynamic configuration. |
  





<a name="envoy.config.core.v3.HeaderMap"></a>

### HeaderMap



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [][envoy.config.core.v3.HeaderValue]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.HeaderValue" >}}) | repeated |  |
  





<a name="envoy.config.core.v3.HeaderValue"></a>

### HeaderValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  | Header name. |
  | value | string |  | Header value.<br>The same :ref:`format specifier <config_access_log_format>` as used for :ref:`HTTP access logging <config_access_log>` applies here, however unknown header values are replaced with the empty string instead of `-`. |
  





<a name="envoy.config.core.v3.HeaderValueOption"></a>

### HeaderValueOption



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| header | [envoy.config.core.v3.HeaderValue]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.HeaderValue" >}}) |  | Header name/value pair that this option applies to. |
  | append | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | Should the value be appended? If true (default), the value is appended to existing values. Otherwise it replaces any existing values. |
  





<a name="envoy.config.core.v3.Locality"></a>

### Locality



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| region | string |  | Region this :ref:`zone <envoy_api_field_config.core.v3.Locality.zone>` belongs to. |
  | zone | string |  | Defines the local service zone where Envoy is running. Though optional, it should be set if discovery service routing is used and the discovery service exposes :ref:`zone data <envoy_api_field_config.endpoint.v3.LocalityLbEndpoints.locality>`, either in this message or via :option:`--service-zone`. The meaning of zone is context dependent, e.g. `Availability Zone (AZ) <https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html>`_ on AWS, `Zone <https://cloud.google.com/compute/docs/regions-zones/>`_ on GCP, etc. |
  | subZone | string |  | When used for locality of upstream hosts, this field further splits zone into smaller chunks of sub-zones so they can be load balanced independently. |
  





<a name="envoy.config.core.v3.Metadata"></a>

### Metadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| filterMetadata | [][envoy.config.core.v3.Metadata.FilterMetadataEntry]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.Metadata.FilterMetadataEntry" >}}) | repeated | Key is the reverse DNS filter name, e.g. com.acme.widget. The envoy.* namespace is reserved for Envoy's built-in filters. |
  





<a name="envoy.config.core.v3.Metadata.FilterMetadataEntry"></a>

### Metadata.FilterMetadataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [google.protobuf.Struct]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.struct#google.protobuf.Struct" >}}) |  |  |
  





<a name="envoy.config.core.v3.Node"></a>

### Node



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string |  | An opaque node identifier for the Envoy node. This also provides the local service node name. It should be set if any of the following features are used: :ref:`statsd <arch_overview_statistics>`, :ref:`CDS <config_cluster_manager_cds>`, and :ref:`HTTP tracing <arch_overview_tracing>`, either in this message or via :option:`--service-node`. |
  | cluster | string |  | Defines the local service cluster name where Envoy is running. Though optional, it should be set if any of the following features are used: :ref:`statsd <arch_overview_statistics>`, :ref:`health check cluster verification <envoy_api_field_config.core.v3.HealthCheck.HttpHealthCheck.service_name_matcher>`, :ref:`runtime override directory <envoy_api_msg_config.bootstrap.v3.Runtime>`, :ref:`user agent addition <envoy_api_field_extensions.filters.network.http_connection_manager.v3.HttpConnectionManager.add_user_agent>`, :ref:`HTTP global rate limiting <config_http_filters_rate_limit>`, :ref:`CDS <config_cluster_manager_cds>`, and :ref:`HTTP tracing <arch_overview_tracing>`, either in this message or via :option:`--service-cluster`. |
  | metadata | [google.protobuf.Struct]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.struct#google.protobuf.Struct" >}}) |  | Opaque metadata extending the node identifier. Envoy will pass this directly to the management server. |
  | locality | [envoy.config.core.v3.Locality]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.Locality" >}}) |  | Locality specifying where the Envoy instance is running. |
  | userAgentName | string |  | Free-form string that identifies the entity requesting config. E.g. "envoy" or "grpc" |
  | userAgentVersion | string |  | Free-form string that identifies the version of the entity requesting config. E.g. "1.12.2" or "abcd1234", or "SpecialEnvoyBuild" |
  | userAgentBuildVersion | [envoy.config.core.v3.BuildVersion]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.BuildVersion" >}}) |  | Structured version of the entity requesting config. |
  | extensions | [][envoy.config.core.v3.Extension]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.Extension" >}}) | repeated | List of extensions and their versions supported by the node. |
  | clientFeatures | []string | repeated | Client feature support list. These are well known features described in the Envoy API repository for a given major version of an API. Client features use reverse DNS naming scheme, for example `com.acme.feature`. See :ref:`the list of features <client_features>` that xDS client may support. |
  | listeningAddresses | [][envoy.config.core.v3.Address]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.address#envoy.config.core.v3.Address" >}}) | repeated | Known listening ports on the node as a generic hint to the management server for filtering :ref:`listeners <config_listeners>` to be returned. For example, if there is a listener bound to port 80, the list can optionally contain the SocketAddress `(0.0.0.0,80)`. The field is optional and just a hint. |
  





<a name="envoy.config.core.v3.RemoteDataSource"></a>

### RemoteDataSource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| httpUri | [envoy.config.core.v3.HttpUri]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.http_uri#envoy.config.core.v3.HttpUri" >}}) |  | The HTTP URI to fetch the remote data. |
  | sha256 | string |  | SHA256 string for verifying data. |
  | retryPolicy | [envoy.config.core.v3.RetryPolicy]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.RetryPolicy" >}}) |  | Retry policy for fetching remote data. |
  





<a name="envoy.config.core.v3.RetryPolicy"></a>

### RetryPolicy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| retryBackOff | [envoy.config.core.v3.BackoffStrategy]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.backoff#envoy.config.core.v3.BackoffStrategy" >}}) |  | Specifies parameters that control :ref:`retry backoff strategy <envoy_api_msg_config.core.v3.BackoffStrategy>`. This parameter is optional, in which case the default base interval is 1000 milliseconds. The default maximum interval is 10 times the base interval. |
  | numRetries | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | Specifies the allowed number of retries. This parameter is optional and defaults to 1. |
  





<a name="envoy.config.core.v3.RuntimeDouble"></a>

### RuntimeDouble



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| defaultValue | double |  | Default value if runtime value is not available. |
  | runtimeKey | string |  | Runtime key to get value for comparison. This value is used if defined. |
  





<a name="envoy.config.core.v3.RuntimeFeatureFlag"></a>

### RuntimeFeatureFlag



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| defaultValue | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | Default value if runtime value is not available. |
  | runtimeKey | string |  | Runtime key to get value for comparison. This value is used if defined. The boolean value must be represented via its `canonical JSON encoding <https://developers.google.com/protocol-buffers/docs/proto3#json>`_. |
  





<a name="envoy.config.core.v3.RuntimeFractionalPercent"></a>

### RuntimeFractionalPercent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| defaultValue | [envoy.type.v3.FractionalPercent]({{< versioned_link_path fromRoot="/reference/api/envoy.type.v3.percent#envoy.type.v3.FractionalPercent" >}}) |  | Default value if the runtime value's for the numerator/denominator keys are not available. |
  | runtimeKey | string |  | Runtime key for a YAML representation of a FractionalPercent. |
  





<a name="envoy.config.core.v3.RuntimePercent"></a>

### RuntimePercent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| defaultValue | [envoy.type.v3.Percent]({{< versioned_link_path fromRoot="/reference/api/envoy.type.v3.percent#envoy.type.v3.Percent" >}}) |  | Default value if runtime value is not available. |
  | runtimeKey | string |  | Runtime key to get value for comparison. This value is used if defined. |
  





<a name="envoy.config.core.v3.RuntimeUInt32"></a>

### RuntimeUInt32



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| defaultValue | uint32 |  | Default value if runtime value is not available. |
  | runtimeKey | string |  | Runtime key to get value for comparison. This value is used if defined. |
  





<a name="envoy.config.core.v3.TransportSocket"></a>

### TransportSocket



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The name of the transport socket to instantiate. The name must match a supported transport socket implementation. |
  | typedConfig | [google.protobuf.Any]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.any#google.protobuf.Any" >}}) |  |  |
  





<a name="envoy.config.core.v3.WatchedDirectory"></a>

### WatchedDirectory



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | string |  | Directory path to watch. |
  




 <!-- end messages -->


<a name="envoy.config.core.v3.RequestMethod"></a>

### RequestMethod


| Name | Number | Description |
| ---- | ------ | ----------- |
| METHOD_UNSPECIFIED | 0 |  |
| GET | 1 |  |
| HEAD | 2 |  |
| POST | 3 |  |
| PUT | 4 |  |
| DELETE | 5 |  |
| CONNECT | 6 |  |
| OPTIONS | 7 |  |
| TRACE | 8 |  |
| PATCH | 9 |  |



<a name="envoy.config.core.v3.RoutingPriority"></a>

### RoutingPriority


| Name | Number | Description |
| ---- | ------ | ----------- |
| DEFAULT | 0 |  |
| HIGH | 1 |  |



<a name="envoy.config.core.v3.TrafficDirection"></a>

### TrafficDirection


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNSPECIFIED | 0 | Default option is unspecified. |
| INBOUND | 1 | The transport is used for incoming traffic. |
| OUTBOUND | 2 | The transport is used for outgoing traffic. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

