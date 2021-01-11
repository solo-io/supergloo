
---

---

## Package : `envoy.config.accesslog.v3`



<a name="top"></a>

<a name="API Reference for accesslog.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## accesslog.proto


## Table of Contents
  - [AccessLog](#envoy.config.accesslog.v3.AccessLog)
  - [AccessLogFilter](#envoy.config.accesslog.v3.AccessLogFilter)
  - [AndFilter](#envoy.config.accesslog.v3.AndFilter)
  - [ComparisonFilter](#envoy.config.accesslog.v3.ComparisonFilter)
  - [DurationFilter](#envoy.config.accesslog.v3.DurationFilter)
  - [ExtensionFilter](#envoy.config.accesslog.v3.ExtensionFilter)
  - [GrpcStatusFilter](#envoy.config.accesslog.v3.GrpcStatusFilter)
  - [HeaderFilter](#envoy.config.accesslog.v3.HeaderFilter)
  - [MetadataFilter](#envoy.config.accesslog.v3.MetadataFilter)
  - [NotHealthCheckFilter](#envoy.config.accesslog.v3.NotHealthCheckFilter)
  - [OrFilter](#envoy.config.accesslog.v3.OrFilter)
  - [ResponseFlagFilter](#envoy.config.accesslog.v3.ResponseFlagFilter)
  - [RuntimeFilter](#envoy.config.accesslog.v3.RuntimeFilter)
  - [StatusCodeFilter](#envoy.config.accesslog.v3.StatusCodeFilter)
  - [TraceableFilter](#envoy.config.accesslog.v3.TraceableFilter)

  - [ComparisonFilter.Op](#envoy.config.accesslog.v3.ComparisonFilter.Op)
  - [GrpcStatusFilter.Status](#envoy.config.accesslog.v3.GrpcStatusFilter.Status)






<a name="envoy.config.accesslog.v3.AccessLog"></a>

### AccessLog



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The name of the access log implementation to instantiate. The name must match a statically registered access log. Current built-in loggers include:<br>#. "envoy.access_loggers.file" #. "envoy.access_loggers.http_grpc" #. "envoy.access_loggers.tcp_grpc" |
  | filter | [envoy.config.accesslog.v3.AccessLogFilter]({{< versioned_link_path fromRoot="/reference/api/envoy.config.accesslog.v3.accesslog#envoy.config.accesslog.v3.AccessLogFilter" >}}) |  | Filter which is used to determine if the access log needs to be written. |
  | typedConfig | [google.protobuf.Any]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.any#google.protobuf.Any" >}}) |  |  |
  





<a name="envoy.config.accesslog.v3.AccessLogFilter"></a>

### AccessLogFilter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| statusCodeFilter | [envoy.config.accesslog.v3.StatusCodeFilter]({{< versioned_link_path fromRoot="/reference/api/envoy.config.accesslog.v3.accesslog#envoy.config.accesslog.v3.StatusCodeFilter" >}}) |  | Status code filter. |
  | durationFilter | [envoy.config.accesslog.v3.DurationFilter]({{< versioned_link_path fromRoot="/reference/api/envoy.config.accesslog.v3.accesslog#envoy.config.accesslog.v3.DurationFilter" >}}) |  | Duration filter. |
  | notHealthCheckFilter | [envoy.config.accesslog.v3.NotHealthCheckFilter]({{< versioned_link_path fromRoot="/reference/api/envoy.config.accesslog.v3.accesslog#envoy.config.accesslog.v3.NotHealthCheckFilter" >}}) |  | Not health check filter. |
  | traceableFilter | [envoy.config.accesslog.v3.TraceableFilter]({{< versioned_link_path fromRoot="/reference/api/envoy.config.accesslog.v3.accesslog#envoy.config.accesslog.v3.TraceableFilter" >}}) |  | Traceable filter. |
  | runtimeFilter | [envoy.config.accesslog.v3.RuntimeFilter]({{< versioned_link_path fromRoot="/reference/api/envoy.config.accesslog.v3.accesslog#envoy.config.accesslog.v3.RuntimeFilter" >}}) |  | Runtime filter. |
  | andFilter | [envoy.config.accesslog.v3.AndFilter]({{< versioned_link_path fromRoot="/reference/api/envoy.config.accesslog.v3.accesslog#envoy.config.accesslog.v3.AndFilter" >}}) |  | And filter. |
  | orFilter | [envoy.config.accesslog.v3.OrFilter]({{< versioned_link_path fromRoot="/reference/api/envoy.config.accesslog.v3.accesslog#envoy.config.accesslog.v3.OrFilter" >}}) |  | Or filter. |
  | headerFilter | [envoy.config.accesslog.v3.HeaderFilter]({{< versioned_link_path fromRoot="/reference/api/envoy.config.accesslog.v3.accesslog#envoy.config.accesslog.v3.HeaderFilter" >}}) |  | Header filter. |
  | responseFlagFilter | [envoy.config.accesslog.v3.ResponseFlagFilter]({{< versioned_link_path fromRoot="/reference/api/envoy.config.accesslog.v3.accesslog#envoy.config.accesslog.v3.ResponseFlagFilter" >}}) |  | Response flag filter. |
  | grpcStatusFilter | [envoy.config.accesslog.v3.GrpcStatusFilter]({{< versioned_link_path fromRoot="/reference/api/envoy.config.accesslog.v3.accesslog#envoy.config.accesslog.v3.GrpcStatusFilter" >}}) |  | gRPC status filter. |
  | extensionFilter | [envoy.config.accesslog.v3.ExtensionFilter]({{< versioned_link_path fromRoot="/reference/api/envoy.config.accesslog.v3.accesslog#envoy.config.accesslog.v3.ExtensionFilter" >}}) |  | Extension filter. |
  | metadataFilter | [envoy.config.accesslog.v3.MetadataFilter]({{< versioned_link_path fromRoot="/reference/api/envoy.config.accesslog.v3.accesslog#envoy.config.accesslog.v3.MetadataFilter" >}}) |  | Metadata Filter |
  





<a name="envoy.config.accesslog.v3.AndFilter"></a>

### AndFilter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| filters | [][envoy.config.accesslog.v3.AccessLogFilter]({{< versioned_link_path fromRoot="/reference/api/envoy.config.accesslog.v3.accesslog#envoy.config.accesslog.v3.AccessLogFilter" >}}) | repeated |  |
  





<a name="envoy.config.accesslog.v3.ComparisonFilter"></a>

### ComparisonFilter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| op | [envoy.config.accesslog.v3.ComparisonFilter.Op]({{< versioned_link_path fromRoot="/reference/api/envoy.config.accesslog.v3.accesslog#envoy.config.accesslog.v3.ComparisonFilter.Op" >}}) |  | Comparison operator. |
  | value | [envoy.config.core.v3.RuntimeUInt32]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.RuntimeUInt32" >}}) |  | Value to compare against. |
  





<a name="envoy.config.accesslog.v3.DurationFilter"></a>

### DurationFilter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| comparison | [envoy.config.accesslog.v3.ComparisonFilter]({{< versioned_link_path fromRoot="/reference/api/envoy.config.accesslog.v3.accesslog#envoy.config.accesslog.v3.ComparisonFilter" >}}) |  | Comparison. |
  





<a name="envoy.config.accesslog.v3.ExtensionFilter"></a>

### ExtensionFilter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The name of the filter implementation to instantiate. The name must match a statically registered filter. |
  | typedConfig | [google.protobuf.Any]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.any#google.protobuf.Any" >}}) |  |  |
  





<a name="envoy.config.accesslog.v3.GrpcStatusFilter"></a>

### GrpcStatusFilter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| statuses | [][envoy.config.accesslog.v3.GrpcStatusFilter.Status]({{< versioned_link_path fromRoot="/reference/api/envoy.config.accesslog.v3.accesslog#envoy.config.accesslog.v3.GrpcStatusFilter.Status" >}}) | repeated | Logs only responses that have any one of the gRPC statuses in this field. |
  | exclude | bool |  | If included and set to true, the filter will instead block all responses with a gRPC status or inferred gRPC status enumerated in statuses, and allow all other responses. |
  





<a name="envoy.config.accesslog.v3.HeaderFilter"></a>

### HeaderFilter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| header | [envoy.config.route.v3.HeaderMatcher]({{< versioned_link_path fromRoot="/reference/api/envoy.config.route.v3.route_components#envoy.config.route.v3.HeaderMatcher" >}}) |  | Only requests with a header which matches the specified HeaderMatcher will pass the filter check. |
  





<a name="envoy.config.accesslog.v3.MetadataFilter"></a>

### MetadataFilter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| matcher | [envoy.type.matcher.v3.MetadataMatcher]({{< versioned_link_path fromRoot="/reference/api/envoy.type.matcher.v3.metadata#envoy.type.matcher.v3.MetadataMatcher" >}}) |  | Matcher to check metadata for specified value. For example, to match on the access_log_hint metadata, set the filter to "envoy.common" and the path to "access_log_hint", and the value to "true". |
  | matchIfKeyNotFound | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | Default result if the key does not exist in dynamic metadata: if unset or true, then log; if false, then don't log. |
  





<a name="envoy.config.accesslog.v3.NotHealthCheckFilter"></a>

### NotHealthCheckFilter







<a name="envoy.config.accesslog.v3.OrFilter"></a>

### OrFilter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| filters | [][envoy.config.accesslog.v3.AccessLogFilter]({{< versioned_link_path fromRoot="/reference/api/envoy.config.accesslog.v3.accesslog#envoy.config.accesslog.v3.AccessLogFilter" >}}) | repeated |  |
  





<a name="envoy.config.accesslog.v3.ResponseFlagFilter"></a>

### ResponseFlagFilter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| flags | []string | repeated | Only responses with the any of the flags listed in this field will be logged. This field is optional. If it is not specified, then any response flag will pass the filter check. |
  





<a name="envoy.config.accesslog.v3.RuntimeFilter"></a>

### RuntimeFilter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| runtimeKey | string |  | Runtime key to get an optional overridden numerator for use in the *percent_sampled* field. If found in runtime, this value will replace the default numerator. |
  | percentSampled | [envoy.type.v3.FractionalPercent]({{< versioned_link_path fromRoot="/reference/api/envoy.type.v3.percent#envoy.type.v3.FractionalPercent" >}}) |  | The default sampling percentage. If not specified, defaults to 0% with denominator of 100. |
  | useIndependentRandomness | bool |  | By default, sampling pivots on the header :ref:`x-request-id<config_http_conn_man_headers_x-request-id>` being present. If :ref:`x-request-id<config_http_conn_man_headers_x-request-id>` is present, the filter will consistently sample across multiple hosts based on the runtime key value and the value extracted from :ref:`x-request-id<config_http_conn_man_headers_x-request-id>`. If it is missing, or *use_independent_randomness* is set to true, the filter will randomly sample based on the runtime key value alone. *use_independent_randomness* can be used for logging kill switches within complex nested :ref:`AndFilter <envoy_api_msg_config.accesslog.v3.AndFilter>` and :ref:`OrFilter <envoy_api_msg_config.accesslog.v3.OrFilter>` blocks that are easier to reason about from a probability perspective (i.e., setting to true will cause the filter to behave like an independent random variable when composed within logical operator filters). |
  





<a name="envoy.config.accesslog.v3.StatusCodeFilter"></a>

### StatusCodeFilter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| comparison | [envoy.config.accesslog.v3.ComparisonFilter]({{< versioned_link_path fromRoot="/reference/api/envoy.config.accesslog.v3.accesslog#envoy.config.accesslog.v3.ComparisonFilter" >}}) |  | Comparison. |
  





<a name="envoy.config.accesslog.v3.TraceableFilter"></a>

### TraceableFilter






 <!-- end messages -->


<a name="envoy.config.accesslog.v3.ComparisonFilter.Op"></a>

### ComparisonFilter.Op


| Name | Number | Description |
| ---- | ------ | ----------- |
| EQ | 0 | = |
| GE | 1 | >= |
| LE | 2 | <= |



<a name="envoy.config.accesslog.v3.GrpcStatusFilter.Status"></a>

### GrpcStatusFilter.Status


| Name | Number | Description |
| ---- | ------ | ----------- |
| OK | 0 |  |
| CANCELED | 1 |  |
| UNKNOWN | 2 |  |
| INVALID_ARGUMENT | 3 |  |
| DEADLINE_EXCEEDED | 4 |  |
| NOT_FOUND | 5 |  |
| ALREADY_EXISTS | 6 |  |
| PERMISSION_DENIED | 7 |  |
| RESOURCE_EXHAUSTED | 8 |  |
| FAILED_PRECONDITION | 9 |  |
| ABORTED | 10 |  |
| OUT_OF_RANGE | 11 |  |
| UNIMPLEMENTED | 12 |  |
| INTERNAL | 13 |  |
| UNAVAILABLE | 14 |  |
| DATA_LOSS | 15 |  |
| UNAUTHENTICATED | 16 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

