
---

---

## Package : `extauth.api.solo.io`



<a name="top"></a>

<a name="API Reference for extauth.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## extauth.proto


## Table of Contents
  - [AuthPlugin](#extauth.api.solo.io.AuthPlugin)
  - [BasicAuth](#extauth.api.solo.io.BasicAuth)
  - [BasicAuth.Apr](#extauth.api.solo.io.BasicAuth.Apr)
  - [BasicAuth.Apr.SaltedHashedPassword](#extauth.api.solo.io.BasicAuth.Apr.SaltedHashedPassword)
  - [BasicAuth.Apr.UsersEntry](#extauth.api.solo.io.BasicAuth.Apr.UsersEntry)
  - [DiscoveryOverride](#extauth.api.solo.io.DiscoveryOverride)
  - [ExtAuthConfigSpec](#extauth.api.solo.io.ExtAuthConfigSpec)
  - [ExtAuthConfigSpec.AccessTokenValidationConfig](#extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig)
  - [ExtAuthConfigSpec.AccessTokenValidationConfig.IntrospectionValidation](#extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig.IntrospectionValidation)
  - [ExtAuthConfigSpec.AccessTokenValidationConfig.JwtValidation](#extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig.JwtValidation)
  - [ExtAuthConfigSpec.AccessTokenValidationConfig.JwtValidation.LocalJwks](#extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig.JwtValidation.LocalJwks)
  - [ExtAuthConfigSpec.AccessTokenValidationConfig.JwtValidation.RemoteJwks](#extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig.JwtValidation.RemoteJwks)
  - [ExtAuthConfigSpec.AccessTokenValidationConfig.ScopeList](#extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig.ScopeList)
  - [ExtAuthConfigSpec.ApiKeyAuthConfig](#extauth.api.solo.io.ExtAuthConfigSpec.ApiKeyAuthConfig)
  - [ExtAuthConfigSpec.ApiKeyAuthConfig.HeadersFromKeyMetadataEntry](#extauth.api.solo.io.ExtAuthConfigSpec.ApiKeyAuthConfig.HeadersFromKeyMetadataEntry)
  - [ExtAuthConfigSpec.ApiKeyAuthConfig.KeyMetadata](#extauth.api.solo.io.ExtAuthConfigSpec.ApiKeyAuthConfig.KeyMetadata)
  - [ExtAuthConfigSpec.ApiKeyAuthConfig.KeyMetadata.MetadataEntry](#extauth.api.solo.io.ExtAuthConfigSpec.ApiKeyAuthConfig.KeyMetadata.MetadataEntry)
  - [ExtAuthConfigSpec.ApiKeyAuthConfig.ValidApiKeysEntry](#extauth.api.solo.io.ExtAuthConfigSpec.ApiKeyAuthConfig.ValidApiKeysEntry)
  - [ExtAuthConfigSpec.Config](#extauth.api.solo.io.ExtAuthConfigSpec.Config)
  - [ExtAuthConfigSpec.OAuth2Config](#extauth.api.solo.io.ExtAuthConfigSpec.OAuth2Config)
  - [ExtAuthConfigSpec.OidcAuthorizationCodeConfig](#extauth.api.solo.io.ExtAuthConfigSpec.OidcAuthorizationCodeConfig)
  - [ExtAuthConfigSpec.OidcAuthorizationCodeConfig.AuthEndpointQueryParamsEntry](#extauth.api.solo.io.ExtAuthConfigSpec.OidcAuthorizationCodeConfig.AuthEndpointQueryParamsEntry)
  - [ExtAuthConfigSpec.OidcAuthorizationCodeConfig.TokenEndpointQueryParamsEntry](#extauth.api.solo.io.ExtAuthConfigSpec.OidcAuthorizationCodeConfig.TokenEndpointQueryParamsEntry)
  - [ExtAuthConfigSpec.OpaAuthConfig](#extauth.api.solo.io.ExtAuthConfigSpec.OpaAuthConfig)
  - [ExtAuthConfigSpec.OpaAuthConfig.ModulesEntry](#extauth.api.solo.io.ExtAuthConfigSpec.OpaAuthConfig.ModulesEntry)
  - [ExtAuthConfigStatus](#extauth.api.solo.io.ExtAuthConfigStatus)
  - [HeaderConfiguration](#extauth.api.solo.io.HeaderConfiguration)
  - [JwksOnDemandCacheRefreshPolicy](#extauth.api.solo.io.JwksOnDemandCacheRefreshPolicy)
  - [Ldap](#extauth.api.solo.io.Ldap)
  - [Ldap.ConnectionPool](#extauth.api.solo.io.Ldap.ConnectionPool)
  - [PassThroughAuth](#extauth.api.solo.io.PassThroughAuth)
  - [PassThroughGrpc](#extauth.api.solo.io.PassThroughGrpc)
  - [RedisOptions](#extauth.api.solo.io.RedisOptions)
  - [UserSession](#extauth.api.solo.io.UserSession)
  - [UserSession.CookieOptions](#extauth.api.solo.io.UserSession.CookieOptions)
  - [UserSession.InternalSession](#extauth.api.solo.io.UserSession.InternalSession)
  - [UserSession.RedisSession](#extauth.api.solo.io.UserSession.RedisSession)

  - [ExtAuthConfigStatus.State](#extauth.api.solo.io.ExtAuthConfigStatus.State)






<a name="extauth.api.solo.io.AuthPlugin"></a>

### AuthPlugin



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Name of the plugin |
  | pluginFileName | string |  | Name of the compiled plugin file. If not specified, Gloo Edge will look for an ".so" file with same name as the plugin. |
  | exportedSymbolName | string |  | Name of the exported symbol that implements the plugin interface in the plugin. If not specified, defaults to the name of the plugin |
  | config | [google.protobuf.Struct]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.struct#google.protobuf.Struct" >}}) |  |  |
  





<a name="extauth.api.solo.io.BasicAuth"></a>

### BasicAuth



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| realm | string |  |  |
  | apr | [extauth.api.solo.io.BasicAuth.Apr]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.BasicAuth.Apr" >}}) |  |  |
  





<a name="extauth.api.solo.io.BasicAuth.Apr"></a>

### BasicAuth.Apr



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| users | [][extauth.api.solo.io.BasicAuth.Apr.UsersEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.BasicAuth.Apr.UsersEntry" >}}) | repeated |  |
  





<a name="extauth.api.solo.io.BasicAuth.Apr.SaltedHashedPassword"></a>

### BasicAuth.Apr.SaltedHashedPassword



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| salt | string |  |  |
  | hashedPassword | string |  |  |
  





<a name="extauth.api.solo.io.BasicAuth.Apr.UsersEntry"></a>

### BasicAuth.Apr.UsersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [extauth.api.solo.io.BasicAuth.Apr.SaltedHashedPassword]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.BasicAuth.Apr.SaltedHashedPassword" >}}) |  |  |
  





<a name="extauth.api.solo.io.DiscoveryOverride"></a>

### DiscoveryOverride



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| authEndpoint | string |  | url of the provider authorization endpoint |
  | tokenEndpoint | string |  | url of the provider token endpoint |
  | jwksUri | string |  | url of the provider json web key set |
  | scopes | []string | repeated | list of scope values that the provider supports |
  | responseTypes | []string | repeated | list of response types that the provider supports |
  | subjects | []string | repeated | list of subject identifier types that the provider supports |
  | idTokenAlgs | []string | repeated | list of json web signature signing algorithms that the provider supports for encoding claims in a jwt |
  | authMethods | []string | repeated | list of client authentication methods supported by the provider token endpoint |
  | claims | []string | repeated | list of claim types that the provider supports |
  





<a name="extauth.api.solo.io.ExtAuthConfigSpec"></a>

### ExtAuthConfigSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| authConfigRefName | string |  | @solo-kit:resource.name This is the identifier of the AuthConfig resource that this configuration is associated with. Any request to the external auth server includes an identifier that is matched against this field to determine which AuthConfig should be applied to it. |
  | configs | [][extauth.api.solo.io.ExtAuthConfigSpec.Config]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.ExtAuthConfigSpec.Config" >}}) | repeated | List of auth configs to be checked for requests on a route referencing this auth config, By default, every config must be authorized for the entire request to be authorized. This behavior can be changed by defining names for each config and defining `boolean_expr` below.<br>State is shared between successful requests on the chain, i.e., the headers returned from each successful auth service get appended into the final auth response. |
  | booleanExpr | [google.protobuf.StringValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.StringValue" >}}) |  | How to handle processing of named configs within an auth config chain. An example config might be: `( basic1 || basic2 || (oidc1 && !oidc2) )` The boolean expression is evaluated left to right but honors parenthesis and short-circuiting. |
  





<a name="extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig"></a>

### ExtAuthConfigSpec.AccessTokenValidationConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| jwt | [extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig.JwtValidation]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig.JwtValidation" >}}) |  | Validate access tokens that conform to the [JSON Web Token (JWT)](https://tools.ietf.org/html/rfc7519) specification. |
  | introspection | [extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig.IntrospectionValidation]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig.IntrospectionValidation" >}}) |  | Defines how (opaque) access tokens, received from the oauth authorization endpoint, are validated [OAuth2.0 Token Introspection](https://tools.ietf.org/html/rfc7662) specification. |
  | userinfoUrl | string |  | The URL for the OIDC userinfo endpoint. If provided, the (opaque) access token provided or received from the oauth endpoint will be queried and the userinfo response (or cached response) will be added to the `AuthorizationRequest` state under the "introspection" key. This can be useful to leverage the userinfo response in, for example, an external auth server plugin. |
  | cacheTimeout | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | How long the token introspection and userinfo endpoint response for a specific access token should be kept in the in-memory cache. The result will be invalidated at this timeout, or at "exp" time from the introspection result, whichever comes sooner. If omitted, defaults to 10 minutes. If zero, then no caching will be done. |
  | requiredScopes | [extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig.ScopeList]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig.ScopeList" >}}) |  | Require access token to have all of the scopes in the given list. This configuration applies to both opaque and JWT tokens. In the case of opaque tokens, this will check the scopes returned in the "scope" member of introspection response (as described in [Section 2.2 of RFC7662](https://tools.ietf.org/html/rfc7662#section-2.2). In case of JWTs the scopes to be validated are expected to be contained in the "scope" claim of the token in the form of a space-separated string. Omitting this field means that scope validation will be skipped. |
  





<a name="extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig.IntrospectionValidation"></a>

### ExtAuthConfigSpec.AccessTokenValidationConfig.IntrospectionValidation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| introspectionUrl | string |  | The URL for the [OAuth2.0 Token Introspection](https://tools.ietf.org/html/rfc7662) endpoint. If provided, the (opaque) access token provided or received from the oauth authorization endpoint will be validated against this endpoint, or locally cached responses for this access token. |
  | clientId | string |  | Your client id as registered with the issuer. Optional: Use if the token introspection url requires client authentication. |
  | clientSecret | string |  | Your client secret as registered with the issuer. Optional: Use if the token introspection url requires client authentication. |
  | userIdAttributeName | string |  | The name of the [introspection response](https://tools.ietf.org/html/rfc7662#section-2.2) attribute that contains the ID of the resource owner (e.g. `sub`, `username`). If specified, the external auth server will use the value of the attribute as the identifier of the authenticated user and add it to the request headers and/or dynamic metadata (depending on how the server is configured); if the field is set and the attribute cannot be found, the request will be denied. This field is optional and by default the server will not try to derive the user ID. |
  





<a name="extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig.JwtValidation"></a>

### ExtAuthConfigSpec.AccessTokenValidationConfig.JwtValidation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| remoteJwks | [extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig.JwtValidation.RemoteJwks]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig.JwtValidation.RemoteJwks" >}}) |  | Fetches the JWKS from a remote location. |
  | localJwks | [extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig.JwtValidation.LocalJwks]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig.JwtValidation.LocalJwks" >}}) |  | Loads the JWKS from a local data source. |
  | issuer | string |  | Allow only tokens that have been issued by this principal (i.e. whose "iss" claim matches this value). If empty, issuer validation will be skipped. |
  





<a name="extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig.JwtValidation.LocalJwks"></a>

### ExtAuthConfigSpec.AccessTokenValidationConfig.JwtValidation.LocalJwks



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| inlineString | string |  | JWKS is embedded as a string. |
  





<a name="extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig.JwtValidation.RemoteJwks"></a>

### ExtAuthConfigSpec.AccessTokenValidationConfig.JwtValidation.RemoteJwks



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| url | string |  | The HTTP URI to fetch the JWKS. |
  | refreshInterval | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | The frequency at which the JWKS should be refreshed. If not specified, the default value is 5 minutes. |
  





<a name="extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig.ScopeList"></a>

### ExtAuthConfigSpec.AccessTokenValidationConfig.ScopeList



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| scope | []string | repeated |  |
  





<a name="extauth.api.solo.io.ExtAuthConfigSpec.ApiKeyAuthConfig"></a>

### ExtAuthConfigSpec.ApiKeyAuthConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| validApiKeys | [][extauth.api.solo.io.ExtAuthConfigSpec.ApiKeyAuthConfig.ValidApiKeysEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.ExtAuthConfigSpec.ApiKeyAuthConfig.ValidApiKeysEntry" >}}) | repeated | A mapping of valid API keys to their associated metadata. This map is automatically populated with the information from the relevant `ApiKeySecret`s. |
  | headerName | string |  | (Optional) When receiving a request, the Gloo Edge Enterprise external auth server will look for an API key in a header with this name. This field is optional; if not provided it defaults to `api-key`. |
  | headersFromKeyMetadata | [][extauth.api.solo.io.ExtAuthConfigSpec.ApiKeyAuthConfig.HeadersFromKeyMetadataEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.ExtAuthConfigSpec.ApiKeyAuthConfig.HeadersFromKeyMetadataEntry" >}}) | repeated | Determines the key metadata that will be included as headers on the upstream request. Each entry represents a header to add: the key is the name of the header, and the value is the key that will be used to look up the data entry in the key metadata. |
  





<a name="extauth.api.solo.io.ExtAuthConfigSpec.ApiKeyAuthConfig.HeadersFromKeyMetadataEntry"></a>

### ExtAuthConfigSpec.ApiKeyAuthConfig.HeadersFromKeyMetadataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="extauth.api.solo.io.ExtAuthConfigSpec.ApiKeyAuthConfig.KeyMetadata"></a>

### ExtAuthConfigSpec.ApiKeyAuthConfig.KeyMetadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| username | string |  | The user is mapped as the name of `Secret` which contains the `ApiKeySecret` |
  | metadata | [][extauth.api.solo.io.ExtAuthConfigSpec.ApiKeyAuthConfig.KeyMetadata.MetadataEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.ExtAuthConfigSpec.ApiKeyAuthConfig.KeyMetadata.MetadataEntry" >}}) | repeated | The metadata present on the `ApiKeySecret`. |
  





<a name="extauth.api.solo.io.ExtAuthConfigSpec.ApiKeyAuthConfig.KeyMetadata.MetadataEntry"></a>

### ExtAuthConfigSpec.ApiKeyAuthConfig.KeyMetadata.MetadataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="extauth.api.solo.io.ExtAuthConfigSpec.ApiKeyAuthConfig.ValidApiKeysEntry"></a>

### ExtAuthConfigSpec.ApiKeyAuthConfig.ValidApiKeysEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [extauth.api.solo.io.ExtAuthConfigSpec.ApiKeyAuthConfig.KeyMetadata]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.ExtAuthConfigSpec.ApiKeyAuthConfig.KeyMetadata" >}}) |  |  |
  





<a name="extauth.api.solo.io.ExtAuthConfigSpec.Config"></a>

### ExtAuthConfigSpec.Config



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [google.protobuf.StringValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.StringValue" >}}) |  | optional: used when defining complex boolean logic, if `boolean_expr` is defined below. Also used in logging. If omitted, an automatically generated name will be used (e.g. config_0, of the pattern 'config_$INDEX_IN_CHAIN'). In the case of plugin auth, this field is ignored in favor of the name assigned on the plugin config itself. |
  | oauth2 | [extauth.api.solo.io.ExtAuthConfigSpec.OAuth2Config]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.ExtAuthConfigSpec.OAuth2Config" >}}) |  |  |
  | basicAuth | [extauth.api.solo.io.BasicAuth]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.BasicAuth" >}}) |  |  |
  | apiKeyAuth | [extauth.api.solo.io.ExtAuthConfigSpec.ApiKeyAuthConfig]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.ExtAuthConfigSpec.ApiKeyAuthConfig" >}}) |  |  |
  | pluginAuth | [extauth.api.solo.io.AuthPlugin]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.AuthPlugin" >}}) |  |  |
  | opaAuth | [extauth.api.solo.io.ExtAuthConfigSpec.OpaAuthConfig]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.ExtAuthConfigSpec.OpaAuthConfig" >}}) |  |  |
  | ldap | [extauth.api.solo.io.Ldap]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.Ldap" >}}) |  |  |
  | jwt | [google.protobuf.Empty]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.empty#google.protobuf.Empty" >}}) |  | This is a "dummy" extauth service which can be used to support multiple auth mechanisms with JWT authentication. If Jwt authentication is to be used in the [boolean expression](https://docs.solo.io/gloo-edge/latest/reference/api/github.com/solo-io/gloo/projects/gloo/api/v1/enterprise/options/extauth/v1/extauth.proto.sk/#authconfig) in an AuthConfig, you can use this auth config type to include Jwt as an Auth config. In addition, `allow_missing_or_failed_jwt` must be set on the Virtual Host or Route that uses JWT auth or else the JWT filter will short circuit this behaviour. |
  | passThroughAuth | [extauth.api.solo.io.PassThroughAuth]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.PassThroughAuth" >}}) |  |  |
  





<a name="extauth.api.solo.io.ExtAuthConfigSpec.OAuth2Config"></a>

### ExtAuthConfigSpec.OAuth2Config



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| oidcAuthorizationCode | [extauth.api.solo.io.ExtAuthConfigSpec.OidcAuthorizationCodeConfig]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.ExtAuthConfigSpec.OidcAuthorizationCodeConfig" >}}) |  | provide issuer location and let gloo handle OIDC flow for you. requests authorized by validating the contents of ID token. can also authorize the access token if configured. |
  | accessTokenValidationConfig | [extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.ExtAuthConfigSpec.AccessTokenValidationConfig" >}}) |  | provide the access token on the request and let gloo handle authorization.<br>according to https://tools.ietf.org/html/rfc6750 you can pass tokens through: - form-encoded body parameter. recommended, more likely to appear. e.g.: Authorization: Bearer mytoken123 - URI query parameter e.g. access_token=mytoken123 - and (preferably) secure cookies |
  





<a name="extauth.api.solo.io.ExtAuthConfigSpec.OidcAuthorizationCodeConfig"></a>

### ExtAuthConfigSpec.OidcAuthorizationCodeConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| clientId | string |  | your client id as registered with the issuer |
  | clientSecret | string |  | your client secret as registered with the issuer |
  | issuerUrl | string |  | The url of the issuer. We will look for OIDC information in issuerUrl+ ".well-known/openid-configuration" |
  | authEndpointQueryParams | [][extauth.api.solo.io.ExtAuthConfigSpec.OidcAuthorizationCodeConfig.AuthEndpointQueryParamsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.ExtAuthConfigSpec.OidcAuthorizationCodeConfig.AuthEndpointQueryParamsEntry" >}}) | repeated | extra query parameters to apply to the Ext-Auth service's authorization request to the identity provider. this can be useful for flows such as PKCE (https://www.oauth.com/oauth2-servers/pkce/authorization-request/) to set the `code_challenge` and `code_challenge_method`. |
  | tokenEndpointQueryParams | [][extauth.api.solo.io.ExtAuthConfigSpec.OidcAuthorizationCodeConfig.TokenEndpointQueryParamsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.ExtAuthConfigSpec.OidcAuthorizationCodeConfig.TokenEndpointQueryParamsEntry" >}}) | repeated | extra query parameters to apply to the Ext-Auth service's token request to the identity provider. this can be useful for flows such as PKCE (https://www.oauth.com/oauth2-servers/pkce/authorization-request/) to set the `code_verifier`. |
  | appUrl | string |  | we to redirect after successful auth, if we can't determine the original url this should be your publicly available app url. |
  | callbackPath | string |  | a callback path relative to app url that will be used for OIDC callbacks. needs to not be used by the application |
  | logoutPath | string |  | a path relative to app url that will be used for logging out from an OIDC session. should not be used by the application. If not provided, logout functionality will be disabled. |
  | afterLogoutUrl | string |  | url to redirect to after logout. This should be a publicly available URL. If not provided, will default to the `app_url`. |
  | scopes | []string | repeated | scopes to request in addition to the openid scope. |
  | session | [extauth.api.solo.io.UserSession]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.UserSession" >}}) |  |  |
  | headers | [extauth.api.solo.io.HeaderConfiguration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.HeaderConfiguration" >}}) |  | Configures headers added to requests. |
  | discoveryOverride | [extauth.api.solo.io.DiscoveryOverride]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.DiscoveryOverride" >}}) |  | OIDC configuration is discovered at <issuerUrl>/.well-known/openid-configuration The configuration override defines any properties that should override this discovery configuration For example, the following AuthConfig CRD could be defined as:    ```yaml    apiVersion: enterprise.gloo.solo.io/v1    kind: AuthConfig    metadata:      name: google-oidc      namespace: gloo-system    spec:      configs:      - oauth:          app_url: http://localhost:8080          callback_path: /callback          client_id: $CLIENT_ID          client_secret_ref:            name: google            namespace: gloo-system          issuer_url: https://accounts.google.com          discovery_override:            token_endpoint: "https://token.url/gettoken"    ```<br>And this will ensure that regardless of what value is discovered at <issuerUrl>/.well-known/openid-configuration, "https://token.url/gettoken" will be used as the token endpoint |
  | discoveryPollInterval | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | The interval at which OIDC configuration is discovered at <issuerUrl>/.well-known/openid-configuration If not specified, the default value is 30 minutes. |
  | jwksCacheRefreshPolicy | [extauth.api.solo.io.JwksOnDemandCacheRefreshPolicy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.JwksOnDemandCacheRefreshPolicy" >}}) |  | If a user executes a request with a key that is not found in the JWKS, it could be that the keys have rotated on the remote source, and not yet in the local cache. This policy lets you define the behavior for how to refresh the local cache during a request where an invalid key is provided |
  | sessionIdHeaderName | string |  | If set, the randomly generated session id will be sent to the token endpoint as part of the code exchange The session id is used as the key for sessions in Redis |
  





<a name="extauth.api.solo.io.ExtAuthConfigSpec.OidcAuthorizationCodeConfig.AuthEndpointQueryParamsEntry"></a>

### ExtAuthConfigSpec.OidcAuthorizationCodeConfig.AuthEndpointQueryParamsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="extauth.api.solo.io.ExtAuthConfigSpec.OidcAuthorizationCodeConfig.TokenEndpointQueryParamsEntry"></a>

### ExtAuthConfigSpec.OidcAuthorizationCodeConfig.TokenEndpointQueryParamsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="extauth.api.solo.io.ExtAuthConfigSpec.OpaAuthConfig"></a>

### ExtAuthConfigSpec.OpaAuthConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| modules | [][extauth.api.solo.io.ExtAuthConfigSpec.OpaAuthConfig.ModulesEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.ExtAuthConfigSpec.OpaAuthConfig.ModulesEntry" >}}) | repeated | An optional modules (filename, module content) maps containing modules assist in the resolution of `query`. |
  | query | string |  | The query that determines the auth decision. The result of this query must be either a boolean or an array with boolean as the first element. A boolean `true` value means that the request will be authorized. Any other value, or error, means that the request will be denied. |
  





<a name="extauth.api.solo.io.ExtAuthConfigSpec.OpaAuthConfig.ModulesEntry"></a>

### ExtAuthConfigSpec.OpaAuthConfig.ModulesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="extauth.api.solo.io.ExtAuthConfigStatus"></a>

### ExtAuthConfigStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| state | [extauth.api.solo.io.ExtAuthConfigStatus.State]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.ExtAuthConfigStatus.State" >}}) |  | The current state of the `ExtAuthConfig`. |
  | message | string |  | A human-readable string explaining the status. |
  | observedGeneration | int64 |  | The observed generation of the resource. When this matches the metadata.generation of the resource, it indicates the status is up-to-date. |
  





<a name="extauth.api.solo.io.HeaderConfiguration"></a>

### HeaderConfiguration



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| idTokenHeader | string |  | If set, the id token will be forward upstream using this header name. |
  | accessTokenHeader | string |  | If set, the access token will be forward upstream using this header name. |
  





<a name="extauth.api.solo.io.JwksOnDemandCacheRefreshPolicy"></a>

### JwksOnDemandCacheRefreshPolicy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| never | [google.protobuf.Empty]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.empty#google.protobuf.Empty" >}}) |  | Never refresh the local JWKS cache on demand. If a key is not in the cache, it is assumed to be malicious. This is the default policy since we assume that IdPs publish keys before they rotate them, and frequent polling finds the newest keys. |
  | always | [google.protobuf.Empty]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.empty#google.protobuf.Empty" >}}) |  | If a key is not in the cache, fetch the most recent keys from the IdP and update the cache. NOTE: This should only be done in trusted environments, since missing keys will each trigger a request to the IdP. Using this in an environment exposed to the internet will allow malicious agents to execute a DDoS attack by spamming protected endpoints with tokens signed by invalid keys. |
  | maxIdpReqPerPollingInterval | uint32 |  | If a key is not in the cache, fetch the most recent keys from the IdP and update the cache. This value sets the number of requests to the IdP per polling interval. If that limit is exceeded, we will stop fetching from the IdP for the remainder of the polling interval. |
  





<a name="extauth.api.solo.io.Ldap"></a>

### Ldap



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| address | string |  | Address of the LDAP server to query. Should be in the form ADDRESS:PORT, e.g. `ldap.default.svc.cluster.local:389`. |
  | userDnTemplate | string |  | Template to build user entry distinguished names (DN). This must contains a single occurrence of the "%s" placeholder. When processing a request, Gloo will substitute the name of the user (extracted from the auth header) for the placeholder and issue a search request with the resulting DN as baseDN (and 'base' search scope). E.g. "uid=%s,ou=people,dc=solo,dc=io" |
  | membershipAttributeName | string |  | Case-insensitive name of the attribute that contains the names of the groups an entry is member of. Gloo will look for attributes with the given name to determine which groups the user entry belongs to. Defaults to 'memberOf' if not provided. |
  | allowedGroups | []string | repeated | In order for the request to be authenticated, the membership attribute (e.g. *memberOf*) on the user entry must contain at least of one of the group DNs specified via this option. E.g. []string{ "cn=managers,ou=groups,dc=solo,dc=io", "cn=developers,ou=groups,dc=solo,dc=io" } |
  | pool | [extauth.api.solo.io.Ldap.ConnectionPool]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.Ldap.ConnectionPool" >}}) |  | Use this property to tune the pool of connections to the LDAP server that Gloo maintains. |
  





<a name="extauth.api.solo.io.Ldap.ConnectionPool"></a>

### Ldap.ConnectionPool



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| maxSize | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | Maximum number connections that are pooled at any give time. The default value is 5. |
  | initialSize | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | Number of connections that the pool will be pre-populated with upon initialization. The default value is 2. |
  





<a name="extauth.api.solo.io.PassThroughAuth"></a>

### PassThroughAuth



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| grpc | [extauth.api.solo.io.PassThroughGrpc]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.PassThroughGrpc" >}}) |  |  |
  | config | [google.protobuf.Struct]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.struct#google.protobuf.Struct" >}}) |  | Custom config to be passed per request to the passthrough auth service. |
  





<a name="extauth.api.solo.io.PassThroughGrpc"></a>

### PassThroughGrpc



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| address | string |  | Address of the auth server to query. Should be in the form ADDRESS:PORT, e.g. `default.svc.cluster.local:389`. |
  | connectionTimeout | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Timeout for the auth server to respond. Defaults to 5s |
  





<a name="extauth.api.solo.io.RedisOptions"></a>

### RedisOptions



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| host | string |  | address of the redis. can be address:port or unix://path/to/unix.sock |
  | db | int32 |  | db to use. can leave unset for db 0. |
  | poolSize | int32 |  | size of the connection pool. can leave unset for default. defaults to 10 connections per every CPU |
  





<a name="extauth.api.solo.io.UserSession"></a>

### UserSession



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| failOnFetchFailure | bool |  | should we fail auth flow when failing to get a session from redis, or allow it to continue, potentially starting a new auth flow and setting a new session. |
  | cookieOptions | [extauth.api.solo.io.UserSession.CookieOptions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.UserSession.CookieOptions" >}}) |  | Set-Cookie options |
  | cookie | [extauth.api.solo.io.UserSession.InternalSession]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.UserSession.InternalSession" >}}) |  | Set the tokens in the cookie itself. No need for server side state. |
  | redis | [extauth.api.solo.io.UserSession.RedisSession]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.UserSession.RedisSession" >}}) |  | Use redis to store the tokens and just store a random id in the cookie. |
  





<a name="extauth.api.solo.io.UserSession.CookieOptions"></a>

### UserSession.CookieOptions



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| maxAge | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | Max age for the cookie. Leave unset for a default of 30 days (2592000 seconds). To disable cookie expiry, set explicitly to 0. |
  | notSecure | bool |  | Use a non-secure cookie. Note - this should only be used for testing and in trusted environments. |
  | path | [google.protobuf.StringValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.StringValue" >}}) |  | Path of the cookie. If unset, defaults to "/". Set it explicitly to "" to avoid setting a path. |
  | domain | string |  | Cookie domain |
  





<a name="extauth.api.solo.io.UserSession.InternalSession"></a>

### UserSession.InternalSession







<a name="extauth.api.solo.io.UserSession.RedisSession"></a>

### UserSession.RedisSession



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| options | [extauth.api.solo.io.RedisOptions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.RedisOptions" >}}) |  | Options to connect to redis |
  | keyPrefix | string |  | Key prefix inside redis |
  | cookieName | string |  | Cookie name to set and store the session id. If empty the default "__session" is used. |
  | allowRefreshing | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | When set, refresh expired id-tokens using the refresh-token. Defaults to true. Explicitly set to false to disable refreshing. |
  




 <!-- end messages -->


<a name="extauth.api.solo.io.ExtAuthConfigStatus.State"></a>

### ExtAuthConfigStatus.State


| Name | Number | Description |
| ---- | ------ | ----------- |
| PENDING | 0 |  |
| ACCEPTED | 1 |  |
| REJECTED | 2 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

