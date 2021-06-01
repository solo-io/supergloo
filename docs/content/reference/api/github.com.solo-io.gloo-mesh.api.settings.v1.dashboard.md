
---

title: "dashboard.proto"

---

## Package : `settings.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for dashboard.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## dashboard.proto


## Table of Contents
  - [DashboardSpec](#settings.mesh.gloo.solo.io.DashboardSpec)
  - [DashboardSpec.AuthConfig](#settings.mesh.gloo.solo.io.DashboardSpec.AuthConfig)
  - [DashboardStatus](#settings.mesh.gloo.solo.io.DashboardStatus)
  - [JwksOnDemandCacheRefreshPolicy](#settings.mesh.gloo.solo.io.JwksOnDemandCacheRefreshPolicy)
  - [OidcConfig](#settings.mesh.gloo.solo.io.OidcConfig)
  - [OidcConfig.AuthEndpointQueryParamsEntry](#settings.mesh.gloo.solo.io.OidcConfig.AuthEndpointQueryParamsEntry)
  - [OidcConfig.DiscoveryOverride](#settings.mesh.gloo.solo.io.OidcConfig.DiscoveryOverride)
  - [OidcConfig.HeaderConfig](#settings.mesh.gloo.solo.io.OidcConfig.HeaderConfig)
  - [OidcConfig.TokenEndpointQueryParamsEntry](#settings.mesh.gloo.solo.io.OidcConfig.TokenEndpointQueryParamsEntry)
  - [SessionConfig](#settings.mesh.gloo.solo.io.SessionConfig)
  - [SessionConfig.CookieOptions](#settings.mesh.gloo.solo.io.SessionConfig.CookieOptions)
  - [SessionConfig.CookieSession](#settings.mesh.gloo.solo.io.SessionConfig.CookieSession)
  - [SessionConfig.RedisSession](#settings.mesh.gloo.solo.io.SessionConfig.RedisSession)







<a name="settings.mesh.gloo.solo.io.DashboardSpec"></a>

### DashboardSpec
Configure settings for the dashboard.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| auth | [settings.mesh.gloo.solo.io.DashboardSpec.AuthConfig]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.dashboard#settings.mesh.gloo.solo.io.DashboardSpec.AuthConfig" >}}) |  | Configuration used to authenticate incoming requests. |
  





<a name="settings.mesh.gloo.solo.io.DashboardSpec.AuthConfig"></a>

### DashboardSpec.AuthConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| oidc | [settings.mesh.gloo.solo.io.OidcConfig]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.dashboard#settings.mesh.gloo.solo.io.OidcConfig" >}}) |  |  |
  





<a name="settings.mesh.gloo.solo.io.DashboardStatus"></a>

### DashboardStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the Dashboard metadata. If the `observedGeneration` does not match `metadata.generation`, Gloo Mesh has not processed the most recent version of this resource. |
  | state | [common.mesh.gloo.solo.io.ApprovalState]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.validation_state#common.mesh.gloo.solo.io.ApprovalState" >}}) |  | The state of the overall resource. It will only show accepted if no processing errors encountered. |
  | errors | []string | repeated | Any errors encountered while processing Settings object. |
  





<a name="settings.mesh.gloo.solo.io.JwksOnDemandCacheRefreshPolicy"></a>

### JwksOnDemandCacheRefreshPolicy
The json web key set (JWKS) (https://tools.ietf.org/html/rfc7517) is discovered at an interval from a remote source. When keys rotate in the remote source, there may be a delay in the local source picking up those new keys. Therefore, a user could execute a request with a token that has been signed by a key in the remote JWKS, but the local cache doesn't have the key yet. The request would fail because the key isn't contained in the local set. Since most IdPs publish key keys in their remote JWKS before they are used, this is not an issue most of the time. This policy lets you define the behavior for when a user has a token with a key not yet in the local cache.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| never | [google.protobuf.Empty]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.empty#google.protobuf.Empty" >}}) |  | Never refresh the local JWKS cache on demand. If a key is not in the cache, it is assumed to be malicious. This is the default policy since we assume that IdPs publish keys before they rotate them, and frequent polling finds the newest keys. |
  | always | [google.protobuf.Empty]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.empty#google.protobuf.Empty" >}}) |  | If a key is not in the cache, fetch the most recent keys from the IdP and update the cache. NOTE: This should only be done in trusted environments, since missing keys will each trigger a request to the IdP. Using this in an environment exposed to the internet will allow malicious agents to execute a DDoS attack by spamming protected endpoints with tokens signed by invalid keys. |
  | maxIdpReqPerPollingInterval | uint32 |  | If a key is not in the cache, fetch the most recent keys from the IdP and update the cache. This value sets the number of requests to the IdP per polling interval. If that limit is exceeded, we will stop fetching from the IdP for the remainder of the polling interval. |
  





<a name="settings.mesh.gloo.solo.io.OidcConfig"></a>

### OidcConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| clientId | string |  | The client ID from the issuer |
  | clientSecret | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | THe client secret from the issuer |
  | issuerUrl | string |  | The url of the issuer. We will look for OIDC information in:   {{ issuerURL }}/.well-known/openid-configuration |
  | authEndpointQueryParams | [][settings.mesh.gloo.solo.io.OidcConfig.AuthEndpointQueryParamsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.dashboard#settings.mesh.gloo.solo.io.OidcConfig.AuthEndpointQueryParamsEntry" >}}) | repeated | Extra query parameters to apply to the authorization request to the identity provider. For example, using the PKCE flow (https://www.oauth.com/oauth2-servers/pkce/authorization-request/) by setting `code_challenge` and `code_challenge_method`. |
  | tokenEndpointQueryParams | [][settings.mesh.gloo.solo.io.OidcConfig.TokenEndpointQueryParamsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.dashboard#settings.mesh.gloo.solo.io.OidcConfig.TokenEndpointQueryParamsEntry" >}}) | repeated | Extra query parameters to apply to the token request to the identity provider. For example, using the PKCE flow (https://www.oauth.com/oauth2-servers/pkce/authorization-request/) by setting `code_challenge` and `code_challenge_method`. |
  | appUrl | string |  | URL to redirect to after successful auth. |
  | callbackPath | string |  | Path to handle the OIDC callback. |
  | logoutPath | string |  | Path used to logout. If not provided, logout will be disabled. |
  | scopes | []string | repeated | Scopes to request in addition to 'openid'. |
  | session | [settings.mesh.gloo.solo.io.SessionConfig]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.dashboard#settings.mesh.gloo.solo.io.SessionConfig" >}}) |  | Configuration for session storage. |
  | header | [settings.mesh.gloo.solo.io.OidcConfig.HeaderConfig]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.dashboard#settings.mesh.gloo.solo.io.OidcConfig.HeaderConfig" >}}) |  | Additional headers. |
  | discoveryOverride | [settings.mesh.gloo.solo.io.OidcConfig.DiscoveryOverride]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.dashboard#settings.mesh.gloo.solo.io.OidcConfig.DiscoveryOverride" >}}) |  | Ensure that certain values are set regardless of what the OIDC provider returns. |
  | discoveryPollInterval | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | How often to poll the OIDC issuer for new configuration. |
  | jwksCacheRefreshPolicy | [settings.mesh.gloo.solo.io.JwksOnDemandCacheRefreshPolicy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.dashboard#settings.mesh.gloo.solo.io.JwksOnDemandCacheRefreshPolicy" >}}) |  | If a user executes a request with a key that is not found in the JWKS, it could be that the keys have rotated on the remote source, and not yet in the local cache. This policy lets you define the behavior for how to refresh the local cache during a request where an invalid key is provided |
  





<a name="settings.mesh.gloo.solo.io.OidcConfig.AuthEndpointQueryParamsEntry"></a>

### OidcConfig.AuthEndpointQueryParamsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="settings.mesh.gloo.solo.io.OidcConfig.DiscoveryOverride"></a>

### OidcConfig.DiscoveryOverride
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
  





<a name="settings.mesh.gloo.solo.io.OidcConfig.HeaderConfig"></a>

### OidcConfig.HeaderConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| idTokenHeader | string |  | If set, the ID token will be sent upstream with this header. |
  | accessTokenHeader | string |  | If set, the access token will be sent upstream with this header. |
  





<a name="settings.mesh.gloo.solo.io.OidcConfig.TokenEndpointQueryParamsEntry"></a>

### OidcConfig.TokenEndpointQueryParamsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="settings.mesh.gloo.solo.io.SessionConfig"></a>

### SessionConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cookieOptions | [settings.mesh.gloo.solo.io.SessionConfig.CookieOptions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.dashboard#settings.mesh.gloo.solo.io.SessionConfig.CookieOptions" >}}) |  | Set-Cookie options |
  | cookie | [settings.mesh.gloo.solo.io.SessionConfig.CookieSession]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.dashboard#settings.mesh.gloo.solo.io.SessionConfig.CookieSession" >}}) |  | Store all session data in the cookie itself |
  | redis | [settings.mesh.gloo.solo.io.SessionConfig.RedisSession]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1.dashboard#settings.mesh.gloo.solo.io.SessionConfig.RedisSession" >}}) |  | Store the session data in a Redis instance. |
  





<a name="settings.mesh.gloo.solo.io.SessionConfig.CookieOptions"></a>

### SessionConfig.CookieOptions



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| maxAge | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | Max age of the cookie. If unset, the default of 30 days will be used. To disable expiration, set explicitly to 0. |
  | notSecure | bool |  | Use an insecure cookie. Should only be used for testing and in trusted environments. |
  | path | [google.protobuf.StringValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.StringValue" >}}) |  | Path of the cookie. Defaults to "/", set to "" to disable the option. |
  | domain | string |  | Domain of the cookie. |
  





<a name="settings.mesh.gloo.solo.io.SessionConfig.CookieSession"></a>

### SessionConfig.CookieSession







<a name="settings.mesh.gloo.solo.io.SessionConfig.RedisSession"></a>

### SessionConfig.RedisSession






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

