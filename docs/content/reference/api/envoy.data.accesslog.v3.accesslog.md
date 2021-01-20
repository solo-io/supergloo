
---

---

## Package : `envoy.data.accesslog.v3`



<a name="top"></a>

<a name="API Reference for accesslog.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## accesslog.proto


## Table of Contents
  - [AccessLogCommon](#envoy.data.accesslog.v3.AccessLogCommon)
  - [AccessLogCommon.FilterStateObjectsEntry](#envoy.data.accesslog.v3.AccessLogCommon.FilterStateObjectsEntry)
  - [ConnectionProperties](#envoy.data.accesslog.v3.ConnectionProperties)
  - [HTTPAccessLogEntry](#envoy.data.accesslog.v3.HTTPAccessLogEntry)
  - [HTTPRequestProperties](#envoy.data.accesslog.v3.HTTPRequestProperties)
  - [HTTPRequestProperties.RequestHeadersEntry](#envoy.data.accesslog.v3.HTTPRequestProperties.RequestHeadersEntry)
  - [HTTPResponseProperties](#envoy.data.accesslog.v3.HTTPResponseProperties)
  - [HTTPResponseProperties.ResponseHeadersEntry](#envoy.data.accesslog.v3.HTTPResponseProperties.ResponseHeadersEntry)
  - [HTTPResponseProperties.ResponseTrailersEntry](#envoy.data.accesslog.v3.HTTPResponseProperties.ResponseTrailersEntry)
  - [ResponseFlags](#envoy.data.accesslog.v3.ResponseFlags)
  - [ResponseFlags.Unauthorized](#envoy.data.accesslog.v3.ResponseFlags.Unauthorized)
  - [TCPAccessLogEntry](#envoy.data.accesslog.v3.TCPAccessLogEntry)
  - [TLSProperties](#envoy.data.accesslog.v3.TLSProperties)
  - [TLSProperties.CertificateProperties](#envoy.data.accesslog.v3.TLSProperties.CertificateProperties)
  - [TLSProperties.CertificateProperties.SubjectAltName](#envoy.data.accesslog.v3.TLSProperties.CertificateProperties.SubjectAltName)

  - [HTTPAccessLogEntry.HTTPVersion](#envoy.data.accesslog.v3.HTTPAccessLogEntry.HTTPVersion)
  - [ResponseFlags.Unauthorized.Reason](#envoy.data.accesslog.v3.ResponseFlags.Unauthorized.Reason)
  - [TLSProperties.TLSVersion](#envoy.data.accesslog.v3.TLSProperties.TLSVersion)






<a name="envoy.data.accesslog.v3.AccessLogCommon"></a>

### AccessLogCommon



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sampleRate | double |  | [#not-implemented-hide:] This field indicates the rate at which this log entry was sampled. Valid range is (0.0, 1.0]. |
  | downstreamRemoteAddress | [envoy.config.core.v3.Address]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.address#envoy.config.core.v3.Address" >}}) |  | This field is the remote/origin address on which the request from the user was received. Note: This may not be the physical peer. E.g, if the remote address is inferred from for example the x-forwarder-for header, proxy protocol, etc. |
  | downstreamLocalAddress | [envoy.config.core.v3.Address]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.address#envoy.config.core.v3.Address" >}}) |  | This field is the local/destination address on which the request from the user was received. |
  | tlsProperties | [envoy.data.accesslog.v3.TLSProperties]({{< versioned_link_path fromRoot="/reference/api/envoy.data.accesslog.v3.accesslog#envoy.data.accesslog.v3.TLSProperties" >}}) |  | If the connection is secure,S this field will contain TLS properties. |
  | startTime | [google.protobuf.Timestamp]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.timestamp#google.protobuf.Timestamp" >}}) |  | The time that Envoy started servicing this request. This is effectively the time that the first downstream byte is received. |
  | timeToLastRxByte | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Interval between the first downstream byte received and the last downstream byte received (i.e. time it takes to receive a request). |
  | timeToFirstUpstreamTxByte | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Interval between the first downstream byte received and the first upstream byte sent. There may by considerable delta between *time_to_last_rx_byte* and this value due to filters. Additionally, the same caveats apply as documented in *time_to_last_downstream_tx_byte* about not accounting for kernel socket buffer time, etc. |
  | timeToLastUpstreamTxByte | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Interval between the first downstream byte received and the last upstream byte sent. There may by considerable delta between *time_to_last_rx_byte* and this value due to filters. Additionally, the same caveats apply as documented in *time_to_last_downstream_tx_byte* about not accounting for kernel socket buffer time, etc. |
  | timeToFirstUpstreamRxByte | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Interval between the first downstream byte received and the first upstream byte received (i.e. time it takes to start receiving a response). |
  | timeToLastUpstreamRxByte | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Interval between the first downstream byte received and the last upstream byte received (i.e. time it takes to receive a complete response). |
  | timeToFirstDownstreamTxByte | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Interval between the first downstream byte received and the first downstream byte sent. There may be a considerable delta between the *time_to_first_upstream_rx_byte* and this field due to filters. Additionally, the same caveats apply as documented in *time_to_last_downstream_tx_byte* about not accounting for kernel socket buffer time, etc. |
  | timeToLastDownstreamTxByte | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Interval between the first downstream byte received and the last downstream byte sent. Depending on protocol, buffering, windowing, filters, etc. there may be a considerable delta between *time_to_last_upstream_rx_byte* and this field. Note also that this is an approximate time. In the current implementation it does not include kernel socket buffer time. In the current implementation it also does not include send window buffering inside the HTTP/2 codec. In the future it is likely that work will be done to make this duration more accurate. |
  | upstreamRemoteAddress | [envoy.config.core.v3.Address]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.address#envoy.config.core.v3.Address" >}}) |  | The upstream remote/destination address that handles this exchange. This does not include retries. |
  | upstreamLocalAddress | [envoy.config.core.v3.Address]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.address#envoy.config.core.v3.Address" >}}) |  | The upstream local/origin address that handles this exchange. This does not include retries. |
  | upstreamCluster | string |  | The upstream cluster that *upstream_remote_address* belongs to. |
  | responseFlags | [envoy.data.accesslog.v3.ResponseFlags]({{< versioned_link_path fromRoot="/reference/api/envoy.data.accesslog.v3.accesslog#envoy.data.accesslog.v3.ResponseFlags" >}}) |  | Flags indicating occurrences during request/response processing. |
  | metadata | [envoy.config.core.v3.Metadata]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.Metadata" >}}) |  | All metadata encountered during request processing, including endpoint selection.<br>This can be used to associate IDs attached to the various configurations used to process this request with the access log entry. For example, a route created from a higher level forwarding rule with some ID can place that ID in this field and cross reference later. It can also be used to determine if a canary endpoint was used or not. |
  | upstreamTransportFailureReason | string |  | If upstream connection failed due to transport socket (e.g. TLS handshake), provides the failure reason from the transport socket. The format of this field depends on the configured upstream transport socket. Common TLS failures are in :ref:`TLS trouble shooting <arch_overview_ssl_trouble_shooting>`. |
  | routeName | string |  | The name of the route |
  | downstreamDirectRemoteAddress | [envoy.config.core.v3.Address]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.address#envoy.config.core.v3.Address" >}}) |  | This field is the downstream direct remote address on which the request from the user was received. Note: This is always the physical peer, even if the remote address is inferred from for example the x-forwarder-for header, proxy protocol, etc. |
  | filterStateObjects | [][envoy.data.accesslog.v3.AccessLogCommon.FilterStateObjectsEntry]({{< versioned_link_path fromRoot="/reference/api/envoy.data.accesslog.v3.accesslog#envoy.data.accesslog.v3.AccessLogCommon.FilterStateObjectsEntry" >}}) | repeated | Map of filter state in stream info that have been configured to be logged. If the filter state serialized to any message other than `google.protobuf.Any` it will be packed into `google.protobuf.Any`. |
  





<a name="envoy.data.accesslog.v3.AccessLogCommon.FilterStateObjectsEntry"></a>

### AccessLogCommon.FilterStateObjectsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [google.protobuf.Any]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.any#google.protobuf.Any" >}}) |  |  |
  





<a name="envoy.data.accesslog.v3.ConnectionProperties"></a>

### ConnectionProperties



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| receivedBytes | uint64 |  | Number of bytes received from downstream. |
  | sentBytes | uint64 |  | Number of bytes sent to downstream. |
  





<a name="envoy.data.accesslog.v3.HTTPAccessLogEntry"></a>

### HTTPAccessLogEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| commonProperties | [envoy.data.accesslog.v3.AccessLogCommon]({{< versioned_link_path fromRoot="/reference/api/envoy.data.accesslog.v3.accesslog#envoy.data.accesslog.v3.AccessLogCommon" >}}) |  | Common properties shared by all Envoy access logs. |
  | protocolVersion | [envoy.data.accesslog.v3.HTTPAccessLogEntry.HTTPVersion]({{< versioned_link_path fromRoot="/reference/api/envoy.data.accesslog.v3.accesslog#envoy.data.accesslog.v3.HTTPAccessLogEntry.HTTPVersion" >}}) |  |  |
  | request | [envoy.data.accesslog.v3.HTTPRequestProperties]({{< versioned_link_path fromRoot="/reference/api/envoy.data.accesslog.v3.accesslog#envoy.data.accesslog.v3.HTTPRequestProperties" >}}) |  | Description of the incoming HTTP request. |
  | response | [envoy.data.accesslog.v3.HTTPResponseProperties]({{< versioned_link_path fromRoot="/reference/api/envoy.data.accesslog.v3.accesslog#envoy.data.accesslog.v3.HTTPResponseProperties" >}}) |  | Description of the outgoing HTTP response. |
  





<a name="envoy.data.accesslog.v3.HTTPRequestProperties"></a>

### HTTPRequestProperties



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| requestMethod | [envoy.config.core.v3.RequestMethod]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.RequestMethod" >}}) |  | The request method (RFC 7231/2616). |
  | scheme | string |  | The scheme portion of the incoming request URI. |
  | authority | string |  | HTTP/2 ``:authority`` or HTTP/1.1 ``Host`` header value. |
  | port | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | The port of the incoming request URI (unused currently, as port is composed onto authority). |
  | path | string |  | The path portion from the incoming request URI. |
  | userAgent | string |  | Value of the ``User-Agent`` request header. |
  | referer | string |  | Value of the ``Referer`` request header. |
  | forwardedFor | string |  | Value of the ``X-Forwarded-For`` request header. |
  | requestId | string |  | Value of the ``X-Request-Id`` request header<br>This header is used by Envoy to uniquely identify a request. It will be generated for all external requests and internal requests that do not already have a request ID. |
  | originalPath | string |  | Value of the ``X-Envoy-Original-Path`` request header. |
  | requestHeadersBytes | uint64 |  | Size of the HTTP request headers in bytes.<br>This value is captured from the OSI layer 7 perspective, i.e. it does not include overhead from framing or encoding at other networking layers. |
  | requestBodyBytes | uint64 |  | Size of the HTTP request body in bytes.<br>This value is captured from the OSI layer 7 perspective, i.e. it does not include overhead from framing or encoding at other networking layers. |
  | requestHeaders | [][envoy.data.accesslog.v3.HTTPRequestProperties.RequestHeadersEntry]({{< versioned_link_path fromRoot="/reference/api/envoy.data.accesslog.v3.accesslog#envoy.data.accesslog.v3.HTTPRequestProperties.RequestHeadersEntry" >}}) | repeated | Map of additional headers that have been configured to be logged. |
  





<a name="envoy.data.accesslog.v3.HTTPRequestProperties.RequestHeadersEntry"></a>

### HTTPRequestProperties.RequestHeadersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="envoy.data.accesslog.v3.HTTPResponseProperties"></a>

### HTTPResponseProperties



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| responseCode | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | The HTTP response code returned by Envoy. |
  | responseHeadersBytes | uint64 |  | Size of the HTTP response headers in bytes.<br>This value is captured from the OSI layer 7 perspective, i.e. it does not include overhead from framing or encoding at other networking layers. |
  | responseBodyBytes | uint64 |  | Size of the HTTP response body in bytes.<br>This value is captured from the OSI layer 7 perspective, i.e. it does not include overhead from framing or encoding at other networking layers. |
  | responseHeaders | [][envoy.data.accesslog.v3.HTTPResponseProperties.ResponseHeadersEntry]({{< versioned_link_path fromRoot="/reference/api/envoy.data.accesslog.v3.accesslog#envoy.data.accesslog.v3.HTTPResponseProperties.ResponseHeadersEntry" >}}) | repeated | Map of additional headers configured to be logged. |
  | responseTrailers | [][envoy.data.accesslog.v3.HTTPResponseProperties.ResponseTrailersEntry]({{< versioned_link_path fromRoot="/reference/api/envoy.data.accesslog.v3.accesslog#envoy.data.accesslog.v3.HTTPResponseProperties.ResponseTrailersEntry" >}}) | repeated | Map of trailers configured to be logged. |
  | responseCodeDetails | string |  | The HTTP response code details. |
  





<a name="envoy.data.accesslog.v3.HTTPResponseProperties.ResponseHeadersEntry"></a>

### HTTPResponseProperties.ResponseHeadersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="envoy.data.accesslog.v3.HTTPResponseProperties.ResponseTrailersEntry"></a>

### HTTPResponseProperties.ResponseTrailersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="envoy.data.accesslog.v3.ResponseFlags"></a>

### ResponseFlags



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| failedLocalHealthcheck | bool |  | Indicates local server healthcheck failed. |
  | noHealthyUpstream | bool |  | Indicates there was no healthy upstream. |
  | upstreamRequestTimeout | bool |  | Indicates an there was an upstream request timeout. |
  | localReset | bool |  | Indicates local codec level reset was sent on the stream. |
  | upstreamRemoteReset | bool |  | Indicates remote codec level reset was received on the stream. |
  | upstreamConnectionFailure | bool |  | Indicates there was a local reset by a connection pool due to an initial connection failure. |
  | upstreamConnectionTermination | bool |  | Indicates the stream was reset due to an upstream connection termination. |
  | upstreamOverflow | bool |  | Indicates the stream was reset because of a resource overflow. |
  | noRouteFound | bool |  | Indicates no route was found for the request. |
  | delayInjected | bool |  | Indicates that the request was delayed before proxying. |
  | faultInjected | bool |  | Indicates that the request was aborted with an injected error code. |
  | rateLimited | bool |  | Indicates that the request was rate-limited locally. |
  | unauthorizedDetails | [envoy.data.accesslog.v3.ResponseFlags.Unauthorized]({{< versioned_link_path fromRoot="/reference/api/envoy.data.accesslog.v3.accesslog#envoy.data.accesslog.v3.ResponseFlags.Unauthorized" >}}) |  | Indicates if the request was deemed unauthorized and the reason for it. |
  | rateLimitServiceError | bool |  | Indicates that the request was rejected because there was an error in rate limit service. |
  | downstreamConnectionTermination | bool |  | Indicates the stream was reset due to a downstream connection termination. |
  | upstreamRetryLimitExceeded | bool |  | Indicates that the upstream retry limit was exceeded, resulting in a downstream error. |
  | streamIdleTimeout | bool |  | Indicates that the stream idle timeout was hit, resulting in a downstream 408. |
  | invalidEnvoyRequestHeaders | bool |  | Indicates that the request was rejected because an envoy request header failed strict validation. |
  | downstreamProtocolError | bool |  | Indicates there was an HTTP protocol error on the downstream request. |
  | upstreamMaxStreamDurationReached | bool |  | Indicates there was a max stream duration reached on the upstream request. |
  | responseFromCacheFilter | bool |  | Indicates the response was served from a cache filter. |
  | noFilterConfigFound | bool |  | Indicates that a filter configuration is not available. |
  | durationTimeout | bool |  | Indicates that request or connection exceeded the downstream connection duration. |
  





<a name="envoy.data.accesslog.v3.ResponseFlags.Unauthorized"></a>

### ResponseFlags.Unauthorized



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| reason | [envoy.data.accesslog.v3.ResponseFlags.Unauthorized.Reason]({{< versioned_link_path fromRoot="/reference/api/envoy.data.accesslog.v3.accesslog#envoy.data.accesslog.v3.ResponseFlags.Unauthorized.Reason" >}}) |  |  |
  





<a name="envoy.data.accesslog.v3.TCPAccessLogEntry"></a>

### TCPAccessLogEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| commonProperties | [envoy.data.accesslog.v3.AccessLogCommon]({{< versioned_link_path fromRoot="/reference/api/envoy.data.accesslog.v3.accesslog#envoy.data.accesslog.v3.AccessLogCommon" >}}) |  | Common properties shared by all Envoy access logs. |
  | connectionProperties | [envoy.data.accesslog.v3.ConnectionProperties]({{< versioned_link_path fromRoot="/reference/api/envoy.data.accesslog.v3.accesslog#envoy.data.accesslog.v3.ConnectionProperties" >}}) |  | Properties of the TCP connection. |
  





<a name="envoy.data.accesslog.v3.TLSProperties"></a>

### TLSProperties



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tlsVersion | [envoy.data.accesslog.v3.TLSProperties.TLSVersion]({{< versioned_link_path fromRoot="/reference/api/envoy.data.accesslog.v3.accesslog#envoy.data.accesslog.v3.TLSProperties.TLSVersion" >}}) |  | Version of TLS that was negotiated. |
  | tlsCipherSuite | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | TLS cipher suite negotiated during handshake. The value is a four-digit hex code defined by the IANA TLS Cipher Suite Registry (e.g. ``009C`` for ``TLS_RSA_WITH_AES_128_GCM_SHA256``).<br>Here it is expressed as an integer. |
  | tlsSniHostname | string |  | SNI hostname from handshake. |
  | localCertificateProperties | [envoy.data.accesslog.v3.TLSProperties.CertificateProperties]({{< versioned_link_path fromRoot="/reference/api/envoy.data.accesslog.v3.accesslog#envoy.data.accesslog.v3.TLSProperties.CertificateProperties" >}}) |  | Properties of the local certificate used to negotiate TLS. |
  | peerCertificateProperties | [envoy.data.accesslog.v3.TLSProperties.CertificateProperties]({{< versioned_link_path fromRoot="/reference/api/envoy.data.accesslog.v3.accesslog#envoy.data.accesslog.v3.TLSProperties.CertificateProperties" >}}) |  | Properties of the peer certificate used to negotiate TLS. |
  | tlsSessionId | string |  | The TLS session ID. |
  





<a name="envoy.data.accesslog.v3.TLSProperties.CertificateProperties"></a>

### TLSProperties.CertificateProperties



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| subjectAltName | [][envoy.data.accesslog.v3.TLSProperties.CertificateProperties.SubjectAltName]({{< versioned_link_path fromRoot="/reference/api/envoy.data.accesslog.v3.accesslog#envoy.data.accesslog.v3.TLSProperties.CertificateProperties.SubjectAltName" >}}) | repeated | SANs present in the certificate. |
  | subject | string |  | The subject field of the certificate. |
  





<a name="envoy.data.accesslog.v3.TLSProperties.CertificateProperties.SubjectAltName"></a>

### TLSProperties.CertificateProperties.SubjectAltName



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uri | string |  |  |
  | dns | string |  | [#not-implemented-hide:] |
  




 <!-- end messages -->


<a name="envoy.data.accesslog.v3.HTTPAccessLogEntry.HTTPVersion"></a>

### HTTPAccessLogEntry.HTTPVersion


| Name | Number | Description |
| ---- | ------ | ----------- |
| PROTOCOL_UNSPECIFIED | 0 |  |
| HTTP10 | 1 |  |
| HTTP11 | 2 |  |
| HTTP2 | 3 |  |
| HTTP3 | 4 |  |



<a name="envoy.data.accesslog.v3.ResponseFlags.Unauthorized.Reason"></a>

### ResponseFlags.Unauthorized.Reason


| Name | Number | Description |
| ---- | ------ | ----------- |
| REASON_UNSPECIFIED | 0 |  |
| EXTERNAL_SERVICE | 1 | The request was denied by the external authorization service. |



<a name="envoy.data.accesslog.v3.TLSProperties.TLSVersion"></a>

### TLSProperties.TLSVersion


| Name | Number | Description |
| ---- | ------ | ----------- |
| VERSION_UNSPECIFIED | 0 |  |
| TLSv1 | 1 |  |
| TLSv1_1 | 2 |  |
| TLSv1_2 | 3 |  |
| TLSv1_3 | 4 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

