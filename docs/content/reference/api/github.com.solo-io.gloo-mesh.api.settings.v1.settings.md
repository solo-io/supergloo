
---

title: "settings.proto"

---

## Package : `settings.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for settings.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## settings.proto


## Table of Contents
  - [DashboardSettings](#settings.mesh.gloo.solo.io.DashboardSettings)
  - [DashboardSettings.AuthConfig](#settings.mesh.gloo.solo.io.DashboardSettings.AuthConfig)
  - [DashboardSettings.NoAuth](#settings.mesh.gloo.solo.io.DashboardSettings.NoAuth)
  - [DashboardSettings.OidcConfig](#settings.mesh.gloo.solo.io.DashboardSettings.OidcConfig)
  - [DashboardSettings.OidcConfig.AuthEndpointQueryParamsEntry](#settings.mesh.gloo.solo.io.DashboardSettings.OidcConfig.AuthEndpointQueryParamsEntry)
  - [DashboardSettings.OidcConfig.DiscoveryOverride](#settings.mesh.gloo.solo.io.DashboardSettings.OidcConfig.DiscoveryOverride)
  - [DashboardSettings.OidcConfig.HeaderConfig](#settings.mesh.gloo.solo.io.DashboardSettings.OidcConfig.HeaderConfig)
  - [DashboardSettings.OidcConfig.TokenEndpointQueryParamsEntry](#settings.mesh.gloo.solo.io.DashboardSettings.OidcConfig.TokenEndpointQueryParamsEntry)
  - [DashboardSettings.SessionConfig](#settings.mesh.gloo.solo.io.DashboardSettings.SessionConfig)
  - [DashboardSettings.SessionConfig.CookieOptions](#settings.mesh.gloo.solo.io.DashboardSettings.SessionConfig.CookieOptions)
  - [DashboardSettings.SessionConfig.CookieSession](#settings.mesh.gloo.solo.io.DashboardSettings.SessionConfig.CookieSession)
  - [DashboardSettings.SessionConfig.RedisSession](#settings.mesh.gloo.solo.io.DashboardSettings.SessionConfig.RedisSession)
  - [DiscoverySettings](#settings.mesh.gloo.solo.io.DiscoverySettings)
  - [DiscoverySettings.Istio](#settings.mesh.gloo.solo.io.DiscoverySettings.Istio)
  - [DiscoverySettings.Istio.IngressGatewayDetector](#settings.mesh.gloo.solo.io.DiscoverySettings.Istio.IngressGatewayDetector)
  - [DiscoverySettings.Istio.IngressGatewayDetector.GatewayWorkloadLabelsEntry](#settings.mesh.gloo.solo.io.DiscoverySettings.Istio.IngressGatewayDetector.GatewayWorkloadLabelsEntry)
  - [DiscoverySettings.Istio.IngressGatewayDetectorsEntry](#settings.mesh.gloo.solo.io.DiscoverySettings.Istio.IngressGatewayDetectorsEntry)
  - [GrpcServer](#settings.mesh.gloo.solo.io.GrpcServer)
  - [JwksOnDemandCacheRefreshPolicy](#settings.mesh.gloo.solo.io.JwksOnDemandCacheRefreshPolicy)
  - [RelaySettings](#settings.mesh.gloo.solo.io.RelaySettings)
  - [SettingsSpec](#settings.mesh.gloo.solo.io.SettingsSpec)
  - [SettingsStatus](#settings.mesh.gloo.solo.io.SettingsStatus)







<a name="settings.mesh.gloo.solo.io.DashboardSettings"></a>

### DashboardSettings



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| authConfigs | [][settings.mesh.gloo.solo.io.DashboardSettings.AuthConfig]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.settings#settings.mesh.gloo.solo.io.DashboardSettings.AuthConfig" >}}) | repeated |  |
  





<a name="settings.mesh.gloo.solo.io.DashboardSettings.AuthConfig"></a>

### DashboardSettings.AuthConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [google.protobuf.StringValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.StringValue" >}}) |  |  |
  | none | [settings.mesh.gloo.solo.io.DashboardSettings.NoAuth]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.settings#settings.mesh.gloo.solo.io.DashboardSettings.NoAuth" >}}) |  |  |
  | oidc | [settings.mesh.gloo.solo.io.DashboardSettings.OidcConfig]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.settings#settings.mesh.gloo.solo.io.DashboardSettings.OidcConfig" >}}) |  |  |
  





<a name="settings.mesh.gloo.solo.io.DashboardSettings.NoAuth"></a>

### DashboardSettings.NoAuth







<a name="settings.mesh.gloo.solo.io.DashboardSettings.OidcConfig"></a>

### DashboardSettings.OidcConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| clientId | string |  | The client ID from the issuer |
  | clientSecret | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | THe client secret from the issuer |
  | issuerUrl | string |  | The url of the issuer. We will look for OIDC information in:   {{ issuerURL }}/.well-known/openid-configuration |
  | authEndpointQueryParams | [][settings.mesh.gloo.solo.io.DashboardSettings.OidcConfig.AuthEndpointQueryParamsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.settings#settings.mesh.gloo.solo.io.DashboardSettings.OidcConfig.AuthEndpointQueryParamsEntry" >}}) | repeated | Extra query parameters to apply to the authorization request to the identity provider. For example, using the PKCE flow (https://www.oauth.com/oauth2-servers/pkce/authorization-request/) by setting `code_challenge` and `code_challenge_method`. |
  | tokenEndpointQueryParams | [][settings.mesh.gloo.solo.io.DashboardSettings.OidcConfig.TokenEndpointQueryParamsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.settings#settings.mesh.gloo.solo.io.DashboardSettings.OidcConfig.TokenEndpointQueryParamsEntry" >}}) | repeated | Extra query parameters to apply to the token request to the identity provider. For example, using the PKCE flow (https://www.oauth.com/oauth2-servers/pkce/authorization-request/) by setting `code_challenge` and `code_challenge_method`. |
  | appUrl | string |  | URL to redirect to after successful auth. |
  | callbackPath | string |  | Path to handle the OIDC callback. |
  | logoutPath | string |  | Path used to logout. If not provided, logout will be disabled. |
  | scopes | []string | repeated | Scopes to request in addition to 'openid'. |
  | header | [settings.mesh.gloo.solo.io.DashboardSettings.OidcConfig.HeaderConfig]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.settings#settings.mesh.gloo.solo.io.DashboardSettings.OidcConfig.HeaderConfig" >}}) |  | Additional headers. |
  | discoveryOverride | [settings.mesh.gloo.solo.io.DashboardSettings.OidcConfig.DiscoveryOverride]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.settings#settings.mesh.gloo.solo.io.DashboardSettings.OidcConfig.DiscoveryOverride" >}}) |  | Ensure that certain values are set regardless of what the OIDC provider returns. |
  | discoveryPollInterval | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | How often to poll the OIDC issuer for new configuration. |
  | jwksCacheRefreshPolicy | [settings.mesh.gloo.solo.io.JwksOnDemandCacheRefreshPolicy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.settings#settings.mesh.gloo.solo.io.JwksOnDemandCacheRefreshPolicy" >}}) |  | If a user executes a request with a key that is not found in the JWKS, it could be that the keys have rotated on the remote source, and not yet in the local cache. This policy lets you define the behavior for how to refresh the local cache during a request where an invalid key is provided |
  





<a name="settings.mesh.gloo.solo.io.DashboardSettings.OidcConfig.AuthEndpointQueryParamsEntry"></a>

### DashboardSettings.OidcConfig.AuthEndpointQueryParamsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="settings.mesh.gloo.solo.io.DashboardSettings.OidcConfig.DiscoveryOverride"></a>

### DashboardSettings.OidcConfig.DiscoveryOverride
OIDC configuration is discovered at <issuerUrl>/.well-known/openid-configuration The discovery override defines any properties that should override this discovery configuration https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| authEndpoint | string |  | URL of the provider authorization endpoint. |
  | tokenEndpoint | string |  | URL of the provider token endpoint. |
  | jwksUri | string |  | URL of the provider JSON web key set. |
  | scopes | []string | repeated | List of scope values that the provider supports. |
  | responseTypes | []string | repeated | List of response types that the provider supports. |
  | subjects | []string | repeated | List of subject identifier types that the provider supports. |
  | idTokenAlgs | []string | repeated | List of json web signature signing algorithms that the provider supports for encoding claims in a JWT. |
  | authMethods | []string | repeated | List of client authentication methods supported by the provider token endpoint. |
  | claims | []string | repeated | List of claim types that the provider supports. |
  





<a name="settings.mesh.gloo.solo.io.DashboardSettings.OidcConfig.HeaderConfig"></a>

### DashboardSettings.OidcConfig.HeaderConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| idTokenHeader | string |  | If set, the ID token will be sent upstream with this header. |
  | accessTokenHeader | string |  | If set, the access token will be sent upstream with this header. |
  





<a name="settings.mesh.gloo.solo.io.DashboardSettings.OidcConfig.TokenEndpointQueryParamsEntry"></a>

### DashboardSettings.OidcConfig.TokenEndpointQueryParamsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="settings.mesh.gloo.solo.io.DashboardSettings.SessionConfig"></a>

### DashboardSettings.SessionConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cookieOptions | [settings.mesh.gloo.solo.io.DashboardSettings.SessionConfig.CookieOptions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.settings#settings.mesh.gloo.solo.io.DashboardSettings.SessionConfig.CookieOptions" >}}) |  | Set-Cookie options |
  | cookie | [settings.mesh.gloo.solo.io.DashboardSettings.SessionConfig.CookieSession]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.settings#settings.mesh.gloo.solo.io.DashboardSettings.SessionConfig.CookieSession" >}}) |  | Store all session data in the cookie itself |
  | redis | [settings.mesh.gloo.solo.io.DashboardSettings.SessionConfig.RedisSession]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.settings#settings.mesh.gloo.solo.io.DashboardSettings.SessionConfig.RedisSession" >}}) |  | Store the session data in a Redis instance. |
  





<a name="settings.mesh.gloo.solo.io.DashboardSettings.SessionConfig.CookieOptions"></a>

### DashboardSettings.SessionConfig.CookieOptions



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| maxAge | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | Max age of the cookie. If unset, the default of 30 days will be used. To disable expiration, set explicitly to 0. |
  | notSecure | bool |  | Use an insecure cookie. Should only be used for testing and in trusted environments. |
  | path | [google.protobuf.StringValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.StringValue" >}}) |  | Path of the cookie. Defaults to "/", set to "" to disable the option. |
  | domain | string |  | Domain of the cookie. |
  





<a name="settings.mesh.gloo.solo.io.DashboardSettings.SessionConfig.CookieSession"></a>

### DashboardSettings.SessionConfig.CookieSession







<a name="settings.mesh.gloo.solo.io.DashboardSettings.SessionConfig.RedisSession"></a>

### DashboardSettings.SessionConfig.RedisSession







<a name="settings.mesh.gloo.solo.io.DiscoverySettings"></a>

### DiscoverySettings
Settings for Gloo Mesh discovery.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| istio | [settings.mesh.gloo.solo.io.DiscoverySettings.Istio]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.settings#settings.mesh.gloo.solo.io.DiscoverySettings.Istio" >}}) |  | Istio-specific discovery settings |
  





<a name="settings.mesh.gloo.solo.io.DiscoverySettings.Istio"></a>

### DiscoverySettings.Istio
Istio-specific discovery settings


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ingressGatewayDetectors | [][settings.mesh.gloo.solo.io.DiscoverySettings.Istio.IngressGatewayDetectorsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.settings#settings.mesh.gloo.solo.io.DiscoverySettings.Istio.IngressGatewayDetectorsEntry" >}}) | repeated | Configure discovery of ingress gateways per cluster. The key to the map is either a Gloo Mesh cluster name or `*` denoting all clusters. If an entry is found for a given cluster, it will be used. Otherwise, the wildcard entry will be used if it exists. Lastly, we will fall back to a set of default values. |
  





<a name="settings.mesh.gloo.solo.io.DiscoverySettings.Istio.IngressGatewayDetector"></a>

### DiscoverySettings.Istio.IngressGatewayDetector
Configure discovery of ingress gateways.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gatewayWorkloadLabels | [][settings.mesh.gloo.solo.io.DiscoverySettings.Istio.IngressGatewayDetector.GatewayWorkloadLabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.settings#settings.mesh.gloo.solo.io.DiscoverySettings.Istio.IngressGatewayDetector.GatewayWorkloadLabelsEntry" >}}) | repeated | Workload labels used to detect ingress gateways for an Istio deployment. If not specified, will default to `{"istio": "ingressgateway"}`. |
  | gatewayTlsPortName | string |  | The name of the TLS port used to detect ingress gateways. Kubernetes services must have a port with this name in order to be recognized as an ingress gateway. If not specified, will default to `tls`. |
  





<a name="settings.mesh.gloo.solo.io.DiscoverySettings.Istio.IngressGatewayDetector.GatewayWorkloadLabelsEntry"></a>

### DiscoverySettings.Istio.IngressGatewayDetector.GatewayWorkloadLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="settings.mesh.gloo.solo.io.DiscoverySettings.Istio.IngressGatewayDetectorsEntry"></a>

### DiscoverySettings.Istio.IngressGatewayDetectorsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [settings.mesh.gloo.solo.io.DiscoverySettings.Istio.IngressGatewayDetector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.settings#settings.mesh.gloo.solo.io.DiscoverySettings.Istio.IngressGatewayDetector" >}}) |  |  |
  





<a name="settings.mesh.gloo.solo.io.GrpcServer"></a>

### GrpcServer
Options for connecting to an external gRPC server.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| address | string |  | TCP address of the gRPC Server (including port). |
  | insecure | bool |  | If true communicate over HTTP rather than HTTPS. |
  | reconnectOnNetworkFailures | bool |  | If true Gloo Mesh will automatically attempt to reconnect to the server after encountering network failures. |
  





<a name="settings.mesh.gloo.solo.io.JwksOnDemandCacheRefreshPolicy"></a>

### JwksOnDemandCacheRefreshPolicy
The json web key set (JWKS) (https://tools.ietf.org/html/rfc7517) is discovered at an interval from a remote source. When keys rotate in the remote source, there may be a delay in the local source picking up those new keys. Therefore, a user could execute a request with a token that has been signed by a key in the remote JWKS, but the local cache doesn't have the key yet. The request would fail because the key isn't contained in the local set. Since most IdPs publish key keys in their remote JWKS before they are used, this is not an issue most of the time. This policy lets you define the behavior for when a user has a token with a key not yet in the local cache.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| never | [google.protobuf.Empty]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.empty#google.protobuf.Empty" >}}) |  | Never refresh the local JWKS cache on demand. If a key is not in the cache, it is assumed to be malicious. This is the default policy since we assume that IdPs publish keys before they rotate them, and frequent polling finds the newest keys. |
  | always | [google.protobuf.Empty]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.empty#google.protobuf.Empty" >}}) |  | If a key is not in the cache, fetch the most recent keys from the IdP and update the cache. NOTE: This should only be done in trusted environments, since missing keys will each trigger a request to the IdP. Using this in an environment exposed to the internet will allow malicious agents to execute a DDoS attack by spamming protected endpoints with tokens signed by invalid keys. |
  | maxIdpReqPerPollingInterval | uint32 |  | If a key is not in the cache, fetch the most recent keys from the IdP and update the cache. This value sets the number of requests to the IdP per polling interval. If that limit is exceeded, we will stop fetching from the IdP for the remainder of the polling interval. |
  





<a name="settings.mesh.gloo.solo.io.RelaySettings"></a>

### RelaySettings
RelaySettings contains options for configuring Gloo Mesh to use Relay for cluster management. Relay provides a way for connecting Gloo Mesh to remote Kubernetes Clusters without the need to share credentials and access to remote Kube API Servers from the management cluster (the Gloo Mesh controllers).<br>Relay instead uses a streaming gRPC API to pass discovery data from remote clusters to the management cluster, and push configuration from the management cluster to the remote clusters.<br>Architecturally, it includes a Relay-agent which is installed to remote Kube clusters at registration time, which then connects directly to the Relay Server in the management cluster. to push its discovery data and pull its mesh configuration.<br> To configure Gloo Mesh to use Relay, make sure to read the [relay installation guide]({{< versioned_link_path fromRoot="/guides/setup/install_gloo_mesh" >}}) and [relay cluster registration guide]({{< versioned_link_path fromRoot="/guides/setup/register_cluster" >}}).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| enabled | bool |  | Enable the use of Relay for cluster management. If relay is enabled, make sure to follow the [relay cluster registration guide]({{< versioned_link_path fromRoot="/guides/setup/register_cluster#relay" >}}) for registering your clusters. |
  | server | [settings.mesh.gloo.solo.io.GrpcServer]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.settings#settings.mesh.gloo.solo.io.GrpcServer" >}}) |  | Connection info for the Relay Server. Gloo Mesh will fetch discovery resources from this server and push translated outputs to this server. Note: currently this field has no effect as the relay server runs in-process of the networking Pod. |
  





<a name="settings.mesh.gloo.solo.io.SettingsSpec"></a>

### SettingsSpec
Configure system-wide settings and defaults. Settings specified in networking policies take precedence over those specified here.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| mtls | [networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.MTLS]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.MTLS" >}}) |  | Configure default mTLS settings for Destinations. |
  | networkingExtensionServers | [][settings.mesh.gloo.solo.io.GrpcServer]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.settings#settings.mesh.gloo.solo.io.GrpcServer" >}}) | repeated | Configure Gloo Mesh networking to communicate with one or more external gRPC NetworkingExtensions servers. Updates will be applied by the servers in the order they are listed (servers towards the end of the list take precedence). Note: Extension Servers have full write access to the output objects written by Gloo Mesh. |
  | discovery | [settings.mesh.gloo.solo.io.DiscoverySettings]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.settings#settings.mesh.gloo.solo.io.DiscoverySettings" >}}) |  | Settings for Gloo Mesh discovery. |
  | relay | [settings.mesh.gloo.solo.io.RelaySettings]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.settings#settings.mesh.gloo.solo.io.RelaySettings" >}}) |  | Enable and configure use of Relay mode to communicate with remote clusters. This is an enterprise-only feature. |
  | dashboard | [settings.mesh.gloo.solo.io.DashboardSettings]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.settings#settings.mesh.gloo.solo.io.DashboardSettings" >}}) |  | Configure the enterprise dashboard. |
  





<a name="settings.mesh.gloo.solo.io.SettingsStatus"></a>

### SettingsStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the Settings metadata. If the `observedGeneration` does not match `metadata.generation`, Gloo Mesh has not processed the most recent version of this resource. |
  | state | [common.mesh.gloo.solo.io.ApprovalState]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.validation_state#common.mesh.gloo.solo.io.ApprovalState" >}}) |  | The state of the overall resource. It will only show accepted if no processing errors encountered. |
  | errors | []string | repeated | Any errors encountered while processing Settings object. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

