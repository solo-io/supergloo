
---

---

## Package : `envoy.config.core.v3`



<a name="top"></a>

<a name="API Reference for grpc_service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## grpc_service.proto


## Table of Contents
  - [GrpcService](#envoy.config.core.v3.GrpcService)
  - [GrpcService.EnvoyGrpc](#envoy.config.core.v3.GrpcService.EnvoyGrpc)
  - [GrpcService.GoogleGrpc](#envoy.config.core.v3.GrpcService.GoogleGrpc)
  - [GrpcService.GoogleGrpc.CallCredentials](#envoy.config.core.v3.GrpcService.GoogleGrpc.CallCredentials)
  - [GrpcService.GoogleGrpc.CallCredentials.GoogleIAMCredentials](#envoy.config.core.v3.GrpcService.GoogleGrpc.CallCredentials.GoogleIAMCredentials)
  - [GrpcService.GoogleGrpc.CallCredentials.MetadataCredentialsFromPlugin](#envoy.config.core.v3.GrpcService.GoogleGrpc.CallCredentials.MetadataCredentialsFromPlugin)
  - [GrpcService.GoogleGrpc.CallCredentials.ServiceAccountJWTAccessCredentials](#envoy.config.core.v3.GrpcService.GoogleGrpc.CallCredentials.ServiceAccountJWTAccessCredentials)
  - [GrpcService.GoogleGrpc.CallCredentials.StsService](#envoy.config.core.v3.GrpcService.GoogleGrpc.CallCredentials.StsService)
  - [GrpcService.GoogleGrpc.ChannelArgs](#envoy.config.core.v3.GrpcService.GoogleGrpc.ChannelArgs)
  - [GrpcService.GoogleGrpc.ChannelArgs.ArgsEntry](#envoy.config.core.v3.GrpcService.GoogleGrpc.ChannelArgs.ArgsEntry)
  - [GrpcService.GoogleGrpc.ChannelArgs.Value](#envoy.config.core.v3.GrpcService.GoogleGrpc.ChannelArgs.Value)
  - [GrpcService.GoogleGrpc.ChannelCredentials](#envoy.config.core.v3.GrpcService.GoogleGrpc.ChannelCredentials)
  - [GrpcService.GoogleGrpc.GoogleLocalCredentials](#envoy.config.core.v3.GrpcService.GoogleGrpc.GoogleLocalCredentials)
  - [GrpcService.GoogleGrpc.SslCredentials](#envoy.config.core.v3.GrpcService.GoogleGrpc.SslCredentials)







<a name="envoy.config.core.v3.GrpcService"></a>

### GrpcService



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| envoyGrpc | [envoy.config.core.v3.GrpcService.EnvoyGrpc]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.grpc_service#envoy.config.core.v3.GrpcService.EnvoyGrpc" >}}) |  | Envoy's in-built gRPC client. See the :ref:`gRPC services overview <arch_overview_grpc_services>` documentation for discussion on gRPC client selection. |
  | googleGrpc | [envoy.config.core.v3.GrpcService.GoogleGrpc]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.grpc_service#envoy.config.core.v3.GrpcService.GoogleGrpc" >}}) |  | `Google C++ gRPC client <https://github.com/grpc/grpc>`_ See the :ref:`gRPC services overview <arch_overview_grpc_services>` documentation for discussion on gRPC client selection. |
  | timeout | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | The timeout for the gRPC request. This is the timeout for a specific request. |
  | initialMetadata | [][envoy.config.core.v3.HeaderValue]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.HeaderValue" >}}) | repeated | Additional metadata to include in streams initiated to the GrpcService. This can be used for scenarios in which additional ad hoc authorization headers (e.g. ``x-foo-bar: baz-key``) are to be injected. For more information, including details on header value syntax, see the documentation on :ref:`custom request headers <config_http_conn_man_headers_custom_request_headers>`. |
  





<a name="envoy.config.core.v3.GrpcService.EnvoyGrpc"></a>

### GrpcService.EnvoyGrpc



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| clusterName | string |  | The name of the upstream gRPC cluster. SSL credentials will be supplied in the :ref:`Cluster <envoy_api_msg_config.cluster.v3.Cluster>` :ref:`transport_socket <envoy_api_field_config.cluster.v3.Cluster.transport_socket>`. |
  | authority | string |  | The `:authority` header in the grpc request. If this field is not set, the authority header value will be `cluster_name`. Note that this authority does not override the SNI. The SNI is provided by the transport socket of the cluster. |
  





<a name="envoy.config.core.v3.GrpcService.GoogleGrpc"></a>

### GrpcService.GoogleGrpc



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| targetUri | string |  | The target URI when using the `Google C++ gRPC client <https://github.com/grpc/grpc>`_. SSL credentials will be supplied in :ref:`channel_credentials <envoy_api_field_config.core.v3.GrpcService.GoogleGrpc.channel_credentials>`. |
  | channelCredentials | [envoy.config.core.v3.GrpcService.GoogleGrpc.ChannelCredentials]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.grpc_service#envoy.config.core.v3.GrpcService.GoogleGrpc.ChannelCredentials" >}}) |  |  |
  | callCredentials | [][envoy.config.core.v3.GrpcService.GoogleGrpc.CallCredentials]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.grpc_service#envoy.config.core.v3.GrpcService.GoogleGrpc.CallCredentials" >}}) | repeated | A set of call credentials that can be composed with `channel credentials <https://grpc.io/docs/guides/auth.html#credential-types>`_. |
  | statPrefix | string |  | The human readable prefix to use when emitting statistics for the gRPC service.<br>.. csv-table::    :header: Name, Type, Description    :widths: 1, 1, 2<br>   streams_total, Counter, Total number of streams opened    streams_closed_<gRPC status code>, Counter, Total streams closed with <gRPC status code> |
  | credentialsFactoryName | string |  | The name of the Google gRPC credentials factory to use. This must have been registered with Envoy. If this is empty, a default credentials factory will be used that sets up channel credentials based on other configuration parameters. |
  | config | [google.protobuf.Struct]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.struct#google.protobuf.Struct" >}}) |  | Additional configuration for site-specific customizations of the Google gRPC library. |
  | perStreamBufferLimitBytes | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | How many bytes each stream can buffer internally. If not set an implementation defined default is applied (1MiB). |
  | channelArgs | [envoy.config.core.v3.GrpcService.GoogleGrpc.ChannelArgs]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.grpc_service#envoy.config.core.v3.GrpcService.GoogleGrpc.ChannelArgs" >}}) |  | Custom channels args. |
  





<a name="envoy.config.core.v3.GrpcService.GoogleGrpc.CallCredentials"></a>

### GrpcService.GoogleGrpc.CallCredentials



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| accessToken | string |  | Access token credentials. https://grpc.io/grpc/cpp/namespacegrpc.html#ad3a80da696ffdaea943f0f858d7a360d. |
  | googleComputeEngine | [google.protobuf.Empty]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.empty#google.protobuf.Empty" >}}) |  | Google Compute Engine credentials. https://grpc.io/grpc/cpp/namespacegrpc.html#a6beb3ac70ff94bd2ebbd89b8f21d1f61 |
  | googleRefreshToken | string |  | Google refresh token credentials. https://grpc.io/grpc/cpp/namespacegrpc.html#a96901c997b91bc6513b08491e0dca37c. |
  | serviceAccountJwtAccess | [envoy.config.core.v3.GrpcService.GoogleGrpc.CallCredentials.ServiceAccountJWTAccessCredentials]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.grpc_service#envoy.config.core.v3.GrpcService.GoogleGrpc.CallCredentials.ServiceAccountJWTAccessCredentials" >}}) |  | Service Account JWT Access credentials. https://grpc.io/grpc/cpp/namespacegrpc.html#a92a9f959d6102461f66ee973d8e9d3aa. |
  | googleIam | [envoy.config.core.v3.GrpcService.GoogleGrpc.CallCredentials.GoogleIAMCredentials]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.grpc_service#envoy.config.core.v3.GrpcService.GoogleGrpc.CallCredentials.GoogleIAMCredentials" >}}) |  | Google IAM credentials. https://grpc.io/grpc/cpp/namespacegrpc.html#a9fc1fc101b41e680d47028166e76f9d0. |
  | fromPlugin | [envoy.config.core.v3.GrpcService.GoogleGrpc.CallCredentials.MetadataCredentialsFromPlugin]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.grpc_service#envoy.config.core.v3.GrpcService.GoogleGrpc.CallCredentials.MetadataCredentialsFromPlugin" >}}) |  | Custom authenticator credentials. https://grpc.io/grpc/cpp/namespacegrpc.html#a823c6a4b19ffc71fb33e90154ee2ad07. https://grpc.io/docs/guides/auth.html#extending-grpc-to-support-other-authentication-mechanisms. |
  | stsService | [envoy.config.core.v3.GrpcService.GoogleGrpc.CallCredentials.StsService]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.grpc_service#envoy.config.core.v3.GrpcService.GoogleGrpc.CallCredentials.StsService" >}}) |  | Custom security token service which implements OAuth 2.0 token exchange. https://tools.ietf.org/html/draft-ietf-oauth-token-exchange-16 See https://github.com/grpc/grpc/pull/19587. |
  





<a name="envoy.config.core.v3.GrpcService.GoogleGrpc.CallCredentials.GoogleIAMCredentials"></a>

### GrpcService.GoogleGrpc.CallCredentials.GoogleIAMCredentials



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| authorizationToken | string |  |  |
  | authoritySelector | string |  |  |
  





<a name="envoy.config.core.v3.GrpcService.GoogleGrpc.CallCredentials.MetadataCredentialsFromPlugin"></a>

### GrpcService.GoogleGrpc.CallCredentials.MetadataCredentialsFromPlugin



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  |  |
  | typedConfig | [google.protobuf.Any]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.any#google.protobuf.Any" >}}) |  |  |
  





<a name="envoy.config.core.v3.GrpcService.GoogleGrpc.CallCredentials.ServiceAccountJWTAccessCredentials"></a>

### GrpcService.GoogleGrpc.CallCredentials.ServiceAccountJWTAccessCredentials



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| jsonKey | string |  |  |
  | tokenLifetimeSeconds | uint64 |  |  |
  





<a name="envoy.config.core.v3.GrpcService.GoogleGrpc.CallCredentials.StsService"></a>

### GrpcService.GoogleGrpc.CallCredentials.StsService



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tokenExchangeServiceUri | string |  | URI of the token exchange service that handles token exchange requests. [#comment:TODO(asraa): Add URI validation when implemented. Tracked by https://github.com/envoyproxy/protoc-gen-validate/issues/303] |
  | resource | string |  | Location of the target service or resource where the client intends to use the requested security token. |
  | audience | string |  | Logical name of the target service where the client intends to use the requested security token. |
  | scope | string |  | The desired scope of the requested security token in the context of the service or resource where the token will be used. |
  | requestedTokenType | string |  | Type of the requested security token. |
  | subjectTokenPath | string |  | The path of subject token, a security token that represents the identity of the party on behalf of whom the request is being made. |
  | subjectTokenType | string |  | Type of the subject token. |
  | actorTokenPath | string |  | The path of actor token, a security token that represents the identity of the acting party. The acting party is authorized to use the requested security token and act on behalf of the subject. |
  | actorTokenType | string |  | Type of the actor token. |
  





<a name="envoy.config.core.v3.GrpcService.GoogleGrpc.ChannelArgs"></a>

### GrpcService.GoogleGrpc.ChannelArgs



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| args | [][envoy.config.core.v3.GrpcService.GoogleGrpc.ChannelArgs.ArgsEntry]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.grpc_service#envoy.config.core.v3.GrpcService.GoogleGrpc.ChannelArgs.ArgsEntry" >}}) | repeated | See grpc_types.h GRPC_ARG #defines for keys that work here. |
  





<a name="envoy.config.core.v3.GrpcService.GoogleGrpc.ChannelArgs.ArgsEntry"></a>

### GrpcService.GoogleGrpc.ChannelArgs.ArgsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [envoy.config.core.v3.GrpcService.GoogleGrpc.ChannelArgs.Value]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.grpc_service#envoy.config.core.v3.GrpcService.GoogleGrpc.ChannelArgs.Value" >}}) |  |  |
  





<a name="envoy.config.core.v3.GrpcService.GoogleGrpc.ChannelArgs.Value"></a>

### GrpcService.GoogleGrpc.ChannelArgs.Value



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| stringValue | string |  |  |
  | intValue | int64 |  |  |
  





<a name="envoy.config.core.v3.GrpcService.GoogleGrpc.ChannelCredentials"></a>

### GrpcService.GoogleGrpc.ChannelCredentials



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sslCredentials | [envoy.config.core.v3.GrpcService.GoogleGrpc.SslCredentials]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.grpc_service#envoy.config.core.v3.GrpcService.GoogleGrpc.SslCredentials" >}}) |  |  |
  | googleDefault | [google.protobuf.Empty]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.empty#google.protobuf.Empty" >}}) |  | https://grpc.io/grpc/cpp/namespacegrpc.html#a6beb3ac70ff94bd2ebbd89b8f21d1f61 |
  | localCredentials | [envoy.config.core.v3.GrpcService.GoogleGrpc.GoogleLocalCredentials]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.grpc_service#envoy.config.core.v3.GrpcService.GoogleGrpc.GoogleLocalCredentials" >}}) |  |  |
  





<a name="envoy.config.core.v3.GrpcService.GoogleGrpc.GoogleLocalCredentials"></a>

### GrpcService.GoogleGrpc.GoogleLocalCredentials







<a name="envoy.config.core.v3.GrpcService.GoogleGrpc.SslCredentials"></a>

### GrpcService.GoogleGrpc.SslCredentials



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rootCerts | [envoy.config.core.v3.DataSource]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.DataSource" >}}) |  | PEM encoded server root certificates. |
  | privateKey | [envoy.config.core.v3.DataSource]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.DataSource" >}}) |  | PEM encoded client private key. |
  | certChain | [envoy.config.core.v3.DataSource]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.base#envoy.config.core.v3.DataSource" >}}) |  | PEM encoded client certificate chain. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

