
---

title: "extauth.proto"

---

## Package : `extauth.networking.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for extauth.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## extauth.proto


## Table of Contents
  - [BufferSettings](#extauth.networking.mesh.gloo.solo.io.BufferSettings)
  - [CustomAuth](#extauth.networking.mesh.gloo.solo.io.CustomAuth)
  - [CustomAuth.ContextExtensionsEntry](#extauth.networking.mesh.gloo.solo.io.CustomAuth.ContextExtensionsEntry)
  - [GatewayExtauth](#extauth.networking.mesh.gloo.solo.io.GatewayExtauth)
  - [RouteExtauth](#extauth.networking.mesh.gloo.solo.io.RouteExtauth)

  - [GatewayExtauth.ApiVersion](#extauth.networking.mesh.gloo.solo.io.GatewayExtauth.ApiVersion)






<a name="extauth.networking.mesh.gloo.solo.io.BufferSettings"></a>

### BufferSettings
Configuration for buffering the request data.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| maxRequestBytes | uint32 |  | Sets the maximum size of a message body that the filter will hold in memory. Envoy will return *HTTP 413* and will *not* initiate the authorization process when buffer reaches the number set in this field. Note that this setting will have precedence over failure_mode_allow. Defaults to 4KB. |
  | allowPartialMessage | bool |  | When this field is true, Envoy will buffer the message until *max_request_bytes* is reached. The authorization request will be dispatched and no 413 HTTP error will be returned by the filter. |
  | packAsBytes | bool |  | When this field is true, Envoy will send the body sent to the external authorization service with raw bytes. |
  





<a name="extauth.networking.mesh.gloo.solo.io.CustomAuth"></a>

### CustomAuth
Gloo Mesh is not expected to configure the ext auth server in this case. This is used with custom auth servers.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| contextExtensions | [][extauth.networking.mesh.gloo.solo.io.CustomAuth.ContextExtensionsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.extauth.extauth#extauth.networking.mesh.gloo.solo.io.CustomAuth.ContextExtensionsEntry" >}}) | repeated | When a request matches the route or traffic policy on which this configuration is defined, Gloo Mesh will add the given context_extensions to the request that is sent to the external authorization server. This allows the server to base the auth decision on metadata that you define on the source of the request.<br>This attribute is analogous to Envoy's config.filter.http.ext_authz.v2.CheckSettings. See the official [Envoy documentation](https://www.envoyproxy.io/docs/envoy/latest/api-v2/config/filter/http/ext_authz/v2/ext_authz.proto.html?highlight=ext_authz#config-filter-http-ext-authz-v2-checksettings) for more details. |
  | name | string |  | TODO(kdorosh) re-evaluate if this is needed for MVP [Enterprise-only] Only required in the case where multiple auth servers are configured in Settings This name must match a key in the named_extauth Settings. |
  





<a name="extauth.networking.mesh.gloo.solo.io.CustomAuth.ContextExtensionsEntry"></a>

### CustomAuth.ContextExtensionsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="extauth.networking.mesh.gloo.solo.io.GatewayExtauth"></a>

### GatewayExtauth
Configure the Extauth Filter on a Gateway<br>TODO(kdorosh): this is hard-coded. make configurable, allow health checks / mTLS, use EnvoyFilter? The upstream to ask about auth decisions .core.skv2.solo.io.ObjectRef extauthz_server_ref = 1;


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| userIdHeader | string |  | If the auth server trusted id of the user, it will be set in this header. Specifically this means that this header will be sanitized form the incoming request. |
  | requestTimeout | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Timeout for the ext auth service to respond. Defaults to 200ms |
  | failureModeAllow | bool |  | In case of a failure or timeout querying the auth server, normally a request is denied. if this is set to true, the request will be allowed. |
  | requestBody | [extauth.networking.mesh.gloo.solo.io.BufferSettings]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.extauth.extauth#extauth.networking.mesh.gloo.solo.io.BufferSettings" >}}) |  | Set this if you also want to send the body of the request, and not just the headers. |
  | clearRouteCache | bool |  | Clears route cache in order to allow the external authorization service to correctly affect routing decisions. Filter clears all cached routes when:<br>1. The field is set to *true*.<br>2. The status returned from the authorization service is a HTTP 200 or gRPC 0.<br>3. At least one *authorization response header* is added to the client request, or is used for altering another client request header. |
  | statusOnError | uint32 |  | Sets the HTTP status that is returned to the client when there is a network error between the filter and the authorization server. The default status is HTTP 403 Forbidden. If set, this must be one of the following: - 100 - 200 201 202 203 204 205 206 207 208 226 - 300 301 302 303 304 305 307 308 - 400 401 402 403 404 405 406 407 408 409 410 411 412 413 414 415 416 417 421 422 423 424 426 428 429 431 - 500 501 502 503 504 505 506 507 508 510 511 |
  | transportApiVersion | [extauth.networking.mesh.gloo.solo.io.GatewayExtauth.ApiVersion]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.extauth.extauth#extauth.networking.mesh.gloo.solo.io.GatewayExtauth.ApiVersion" >}}) |  | Determines the API version for the `ext_authz` transport protocol that will be used by Envoy to communicate with the auth server. Defaults to `V2`. For more info, see the `transport_api_version` field [here](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_authz/v3/ext_authz.proto#extensions-filters-http-ext-authz-v3-extauthz). |
  | statPrefix | string |  | Optional additional prefix to use when emitting statistics. This allows to distinguish emitted statistics between configured ext_authz filters in an HTTP filter chain. |
  





<a name="extauth.networking.mesh.gloo.solo.io.RouteExtauth"></a>

### RouteExtauth
Extauth configuration for a Route or TrafficPolicy. Configures extauth for individual HTTP routes


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| disable | bool |  | Set to true to disable auth on the route. |
  | configRef | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | A reference to an AuthConfig. This is used to configure the Gloo Mesh Gateway extauth server. |
  | customAuth | [extauth.networking.mesh.gloo.solo.io.CustomAuth]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.extauth.extauth#extauth.networking.mesh.gloo.solo.io.CustomAuth" >}}) |  | Use this field if you are running your own custom extauth server. |
  




 <!-- end messages -->


<a name="extauth.networking.mesh.gloo.solo.io.GatewayExtauth.ApiVersion"></a>

### GatewayExtauth.ApiVersion
Describes the transport protocol version to use when connecting to the ext auth server.

| Name | Number | Description |
| ---- | ------ | ----------- |
| V3 | 0 | Use v3 API. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

