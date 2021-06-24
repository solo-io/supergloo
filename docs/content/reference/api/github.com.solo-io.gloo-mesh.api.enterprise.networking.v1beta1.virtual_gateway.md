
---

title: "virtual_gateway.proto"

---

## Package : `networking.enterprise.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for virtual_gateway.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## virtual_gateway.proto


## Table of Contents
  - [SDSConfig](#networking.enterprise.mesh.gloo.solo.io.SDSConfig)
  - [SDSConfig.CallCredentials](#networking.enterprise.mesh.gloo.solo.io.SDSConfig.CallCredentials)
  - [SDSConfig.CallCredentials.FileCredentialSource](#networking.enterprise.mesh.gloo.solo.io.SDSConfig.CallCredentials.FileCredentialSource)
  - [SelectedGatewayWorkload](#networking.enterprise.mesh.gloo.solo.io.SelectedGatewayWorkload)
  - [SelectedGatewayWorkload.LabelsEntry](#networking.enterprise.mesh.gloo.solo.io.SelectedGatewayWorkload.LabelsEntry)
  - [VirtualGatewaySpec](#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec)
  - [VirtualGatewaySpec.ConnectionHandler](#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler)
  - [VirtualGatewaySpec.ConnectionHandler.ConnectionMatch](#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.ConnectionMatch)
  - [VirtualGatewaySpec.ConnectionHandler.ConnectionOptions](#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.ConnectionOptions)
  - [VirtualGatewaySpec.ConnectionHandler.ConnectionOptions.TlsTerminationOptions](#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.ConnectionOptions.TlsTerminationOptions)
  - [VirtualGatewaySpec.ConnectionHandler.HttpRoutes](#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.HttpRoutes)
  - [VirtualGatewaySpec.ConnectionHandler.HttpRoutes.HttpOptions](#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.HttpRoutes.HttpOptions)
  - [VirtualGatewaySpec.ConnectionHandler.HttpRoutes.RouteSpecifier](#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.HttpRoutes.RouteSpecifier)
  - [VirtualGatewaySpec.ConnectionHandler.TcpRoutes](#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes)
  - [VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost](#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost)
  - [VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig](#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig)
  - [VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig.SSLFiles](#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig.SSLFiles)
  - [VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig.SslParameters](#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig.SslParameters)
  - [VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.TcpAction](#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.TcpAction)
  - [VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpOptions](#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpOptions)
  - [VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpOptions.TcpProxySettings](#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpOptions.TcpProxySettings)
  - [VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpOptions.TcpProxySettings.TunnelingConfig](#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpOptions.TcpProxySettings.TunnelingConfig)
  - [VirtualGatewaySpec.DeployToIngressGateway](#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.DeployToIngressGateway)
  - [VirtualGatewaySpec.GatewayOptions](#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.GatewayOptions)
  - [VirtualGatewayStatus](#networking.enterprise.mesh.gloo.solo.io.VirtualGatewayStatus)
  - [VirtualGatewayStatus.CreatedIstioGatewaysEntry](#networking.enterprise.mesh.gloo.solo.io.VirtualGatewayStatus.CreatedIstioGatewaysEntry)

  - [VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig.SslParameters.ProtocolVersion](#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig.SslParameters.ProtocolVersion)






<a name="networking.enterprise.mesh.gloo.solo.io.SDSConfig"></a>

### SDSConfig
Note: This message needs to be at this level (rather than nested) due to cue restrictions.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| targetUri | string |  | Target uri for the sds channel. currently only a unix domain socket is supported. |
  | callCredentials | [networking.enterprise.mesh.gloo.solo.io.SDSConfig.CallCredentials]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.SDSConfig.CallCredentials" >}}) |  | Call credentials. |
  | clusterName | string |  | The name of the sds cluster in envoy |
  | certificatesSecretName | string |  | The name of the secret containing the certificate |
  | validationContextName | string |  | The name of secret containing the validation context (i.e. root ca) |
  





<a name="networking.enterprise.mesh.gloo.solo.io.SDSConfig.CallCredentials"></a>

### SDSConfig.CallCredentials



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fileCredentialSource | [networking.enterprise.mesh.gloo.solo.io.SDSConfig.CallCredentials.FileCredentialSource]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.SDSConfig.CallCredentials.FileCredentialSource" >}}) |  | Call credentials are coming from a file, |
  





<a name="networking.enterprise.mesh.gloo.solo.io.SDSConfig.CallCredentials.FileCredentialSource"></a>

### SDSConfig.CallCredentials.FileCredentialSource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tokenFileName | string |  | File containing auth token. |
  | header | string |  | Header to carry the token. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.SelectedGatewayWorkload"></a>

### SelectedGatewayWorkload
a gateway workload (e.g. Deployment) where a virtual gateway will be served


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | the name of the gateway workload |
  | namespace | string |  | the namespace where the gateway workload is running |
  | clusterName | string |  | the cluster where the gateway workload is running |
  | labels | [][networking.enterprise.mesh.gloo.solo.io.SelectedGatewayWorkload.LabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.SelectedGatewayWorkload.LabelsEntry" >}}) | repeated | the labels used to identify the gateway workload |
  | externalUrl | string |  | the external URL by which the gateway can be accessed on the given workload, if it exists |
  





<a name="networking.enterprise.mesh.gloo.solo.io.SelectedGatewayWorkload.LabelsEntry"></a>

### SelectedGatewayWorkload.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec"></a>

### VirtualGatewaySpec
VirtualGateway is the top-level object for configuring ingress into a Mesh or VirtualMesh. A single VirtualGateway can apply to multiple deployed ingress pods and sidecars across meshes and clusters contained within a VirtualMesh. VirtualGateways can route traffic to destination services which live in a specific cluster or mesh. This allows VirtualGateways to route traffic from an ingress or sidecar in one mesh to a service in another. In order to perform cross-mesh routing, the Gateway Mesh and Destination mesh must be contained in a single VirtualMesh, with federation enabled.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| deployToIngressGateways | [networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.DeployToIngressGateway]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.DeployToIngressGateway" >}}) |  | deploy this VirtualGateway to one or more Ingress Gateway workloads {{/* TODO: evaluate supporting multiple ingress gateway deployments per VG */}} |
  | deployToSidecars | [][common.mesh.gloo.solo.io.WorkloadSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.WorkloadSelector" >}}) | repeated | deploy this VirtualGateway to one or more workload sidecars {{/* NOTE: unimplemented */}} |
  | connectionHandlers | [][networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler" >}}) | repeated | Each Gateway must implement one or more ConnectionHandlers. A ConnectionHandler instructs the gateway how to handle clients  which have connected to the specified bind address. Typically `connectionHandlers` will consist of a single `http` handler which serves HTTP Routes defined in a set of VirtualHosts. Multiple `connectionHandlers` can be specified to provide different behavior on the same Gateway, e.g. one for TCP and one for HTTP traffic. |
  | options | [networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.GatewayOptions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.GatewayOptions" >}}) |  | Options applied to all clients who connect to this gateway |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler"></a>

### VirtualGatewaySpec.ConnectionHandler
Each ConnnectionHandler specifies a `connectionMatch` (required if using multiple ConnectionHandlers) and a set of (HTTP or TCP) routes to serve matched connections.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| connectionMatch | [networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.ConnectionMatch]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.ConnectionMatch" >}}) |  | Additional options for matching a connection to a specific gateway. This is required when more than one `connectionHandler` is specified for a single gateway. Typically this is used to serve different |
  | http | [networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.HttpRoutes]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.HttpRoutes" >}}) |  |  |
  | tcp | [networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes" >}}) |  |  |
  | connectionOptions | [networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.ConnectionOptions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.ConnectionOptions" >}}) |  | Top level optional configuration for all routes on the ConnectionHandler. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.ConnectionMatch"></a>

### VirtualGatewaySpec.ConnectionHandler.ConnectionMatch
Look at what Envoy exposes, put them all(maybe?) here Should expose everything we can do with Envoy, ideally.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| transportProtocol | string |  | Protocol |
  | serverNames | []string | repeated | If non-empty, a list of server names (e.g. SNI for TLS protocol) to consider when determining a `connectionMatch`. Those values will be compared against the server names of a new connection, when detected by one of the listener filters.<br>The server name will be matched against all wildcard domains, i.e. `www.example.com` will be first matched against `www.example.com`, then `*.example.com`, then ``*.com`.<br>Note that partial wildcards are not supported, and values like `*w.example.com` are invalid. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.ConnectionOptions"></a>

### VirtualGatewaySpec.ConnectionHandler.ConnectionOptions
TODO: Fill ConnectionOptions


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tlsContext | [networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.ConnectionOptions.TlsTerminationOptions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.ConnectionOptions.TlsTerminationOptions" >}}) |  | TODO |
  | strictFilterManagement | bool |  | Restricts filter from being added to the corresponding Envoy Listener unless they are explicitly configured in the connection handler options |
  | enableProxyProtocol | bool |  | enable PROXY protocol for this connection handler. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.ConnectionOptions.TlsTerminationOptions"></a>

### VirtualGatewaySpec.ConnectionHandler.ConnectionOptions.TlsTerminationOptions



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| presented | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | If specified, the route will match against whether or not a certificate is presented. If not specified, certificate presentation status (true or false) will not be considered when route matching. |
  | validated | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | If specified, the route will match against whether or not a certificate is validated. If not specified, certificate validation status (true or false) will not be considered when route matching. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.HttpRoutes"></a>

### VirtualGatewaySpec.ConnectionHandler.HttpRoutes



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| routeConfig | [][networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.HttpRoutes.RouteSpecifier]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.HttpRoutes.RouteSpecifier" >}}) | repeated |  |
  | routeOptions | [networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.HttpRoutes.HttpOptions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.HttpRoutes.HttpOptions" >}}) |  | HTTP Listener Options // Root level RouteTable + VirtualHost + routes level |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.HttpRoutes.HttpOptions"></a>

### VirtualGatewaySpec.ConnectionHandler.HttpRoutes.HttpOptions



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| csrf | [networking.mesh.gloo.solo.io.CsrfPolicy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy#networking.mesh.gloo.solo.io.CsrfPolicy" >}}) |  | Configure the Envoy based CSRF filter for this Gateway. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.HttpRoutes.RouteSpecifier"></a>

### VirtualGatewaySpec.ConnectionHandler.HttpRoutes.RouteSpecifier



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| virtualHostSelector | [core.skv2.solo.io.ObjectSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectSelector" >}}) |  | RouteSelector is used to specify which VirtualHosts should be attached to this gateway. |
  | virtualHost | [networking.enterprise.mesh.gloo.solo.io.VirtualHostSpec]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_host#networking.enterprise.mesh.gloo.solo.io.VirtualHostSpec" >}}) |  | VirtualHost allows in-lining a route table directly in the Gateway Resource, for simple configs using fewer CRDs. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes"></a>

### VirtualGatewaySpec.ConnectionHandler.TcpRoutes



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tcpHosts | [][networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost" >}}) | repeated | TCP hosts that the gateway can route to |
  | options | [networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpOptions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpOptions" >}}) |  | TCP Gateway configuration |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost"></a>

### VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | the logical name of the tcp host. names must be unique for each tcp host within a listener |
  | sslConfig | [networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig" >}}) |  | If provided, the Gateway will serve TLS/SSL traffic for this set of routes |
  | destination | [networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.TcpAction]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.TcpAction" >}}) |  |  |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig"></a>

### VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig
SslConfig contains the options necessary to configure a virtual host or listener to use TLS


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| secretRef | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | SecretRef contains the secret ref to a gloo tls secret or a kubernetes tls secret. gloo tls secret can contain a root ca as well if verification is needed. |
  | sslFiles | [networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig.SSLFiles]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig.SSLFiles" >}}) |  | SSLFiles reference paths to certificates which are local to the proxy |
  | sds | [networking.enterprise.mesh.gloo.solo.io.SDSConfig]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.SDSConfig" >}}) |  | Use secret discovery service. |
  | sniDomains | []string | repeated | optional. the SNI domains that should be considered for TLS connections |
  | verifySubjectAltName | []string | repeated | Verify that the Subject Alternative Name in the peer certificate is one of the specified values. note that a root_ca must be provided if this option is used. |
  | parameters | [networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig.SslParameters]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig.SslParameters" >}}) |  |  |
  | alpnProtocols | []string | repeated | Set Application Level Protocol Negotiation If empty, defaults to ["h2", "http/1.1"]. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig.SSLFiles"></a>

### VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig.SSLFiles



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tlsCert | string |  |  |
  | tlsKey | string |  |  |
  | rootCa | string |  | for client cert validation. optional |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig.SslParameters"></a>

### VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig.SslParameters
General TLS parameters. See the [envoy docs](https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/auth/cert.proto#envoy-api-enum-auth-tlsparameters-tlsprotocol) for more information on the meaning of these values.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| minimumProtocolVersion | [networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig.SslParameters.ProtocolVersion]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig.SslParameters.ProtocolVersion" >}}) |  |  |
  | maximumProtocolVersion | [networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig.SslParameters.ProtocolVersion]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig.SslParameters.ProtocolVersion" >}}) |  |  |
  | cipherSuites | []string | repeated |  |
  | ecdhCurves | []string | repeated |  |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.TcpAction"></a>

### VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.TcpAction
Name of the destinations the gateway can route to. Note: the destination spec and subsets are not supported in this context and will be ignored.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| static | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | Reference to a gloo mesh Static Destination |
  | virtual | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | Reference to a gloo mesh VirtualDestination |
  | kube | [core.skv2.solo.io.ClusterObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ClusterObjectRef" >}}) |  | Reference to a Kubernetes Service. Note that the service must exist in the same mesh or virtual mesh (with federation enabled) as  each gateway workload which routes to this destination. |
  | forwardSniClusterName | [google.protobuf.Empty]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.empty#google.protobuf.Empty" >}}) |  | Forwards the request to a cluster name matching the TLS SNI name https://www.envoyproxy.io/docs/envoy/latest/api-v2/config/filter/network/sni_cluster/empty/sni_cluster Note: This filter will only work properly with TLS connections in which the upstream SNI domain is specified |
  | weight | uint32 |  | Relative weight of this destination to others in the same route. If omitted, all destinations in the route will be load balanced between evenly. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpOptions"></a>

### VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpOptions



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tcpProxySettings | [networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpOptions.TcpProxySettings]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpOptions.TcpProxySettings" >}}) |  |  |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpOptions.TcpProxySettings"></a>

### VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpOptions.TcpProxySettings



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| maxConnectAttempts | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | Contains various settings for Envoy's tcp proxy filter. See here for more information: https://www.envoyproxy.io/docs/envoy/v1.10.0/api-v2/config/filter/network/tcp_proxy/v2/tcp_proxy.proto#envoy-api-msg-config-filter-network-tcp-proxy-v2-tcpproxy |
  | idleTimeout | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  |  |
  | tunnelingConfig | [networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpOptions.TcpProxySettings.TunnelingConfig]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpOptions.TcpProxySettings.TunnelingConfig" >}}) |  | If set, this configures tunneling, e.g. configuration options to tunnel multiple TCP payloads over a shared HTTP tunnel. If this message is absent, the payload will be proxied upstream as per usual. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpOptions.TcpProxySettings.TunnelingConfig"></a>

### VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpOptions.TcpProxySettings.TunnelingConfig
Configuration for tunneling TCP over other transports or application layers.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hostname | string |  | The hostname to send in the synthesized CONNECT headers to the upstream proxy. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.DeployToIngressGateway"></a>

### VirtualGatewaySpec.DeployToIngressGateway
Options for deploying the VirtualGateway to an Istio Ingress Gateway


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bindAddress | string |  | The bind address the gateway should serve traffic on This maps to the Envoy Listener address. Defaults to "::" or "0.0.0.0". |
  | bindPort | uint32 |  | The bind port where the gateway workload will listen for connections. This maps to the Envoy Listener port. |
  | gatewayWorkloads | [][common.mesh.gloo.solo.io.WorkloadSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.WorkloadSelector" >}}) | repeated | Select which gateway workloads (Envoy pods / Istio ingress-gateways) this config applies to. Ingress pods selected must be in the same Mesh (or Federated VirtualMesh) as the Destination services being routed to. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.GatewayOptions"></a>

### VirtualGatewaySpec.GatewayOptions
TODO: Fill in more options<br>gateway-level options (only apply to gateway/listener)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| perConnectionBufferLimitBytes | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | Soft limit on size of the listener's new connection read and write buffers. If unspecified, defaults to 1MiB For more info, check out the [Envoy docs](https://www.envoyproxy.io/docs/envoy/v1.17.1/api-v3/config/listener/v3/listener.proto) |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualGatewayStatus"></a>

### VirtualGatewayStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the VirtualGateway metadata. If the `observedGeneration` does not match `metadata.generation`, Gloo Mesh has not processed the most recent version of this resource. |
  | errors | []string | repeated | Any errors found while processing this generation of the resource. |
  | warnings | []string | repeated | Any warnings found while processing this generation of the resource. |
  | selectedGateways | [][networking.enterprise.mesh.gloo.solo.io.SelectedGatewayWorkload]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.SelectedGatewayWorkload" >}}) | repeated |  |
  | selectedVirtualHosts | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated |  |
  | selectedRouteTables | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated | List of Delegated Route tables that this Route table delegates to |
  | createdIstioGateways | [][networking.enterprise.mesh.gloo.solo.io.VirtualGatewayStatus.CreatedIstioGatewaysEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_gateway#networking.enterprise.mesh.gloo.solo.io.VirtualGatewayStatus.CreatedIstioGatewaysEntry" >}}) | repeated | List of Istio Gateway CRs created by this VirtualGateway in each cluster |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualGatewayStatus.CreatedIstioGatewaysEntry"></a>

### VirtualGatewayStatus.CreatedIstioGatewaysEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [core.skv2.solo.io.ObjectRefList]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRefList" >}}) |  |  |
  




 <!-- end messages -->


<a name="networking.enterprise.mesh.gloo.solo.io.VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig.SslParameters.ProtocolVersion"></a>

### VirtualGatewaySpec.ConnectionHandler.TcpRoutes.TcpHost.SslConfig.SslParameters.ProtocolVersion


| Name | Number | Description |
| ---- | ------ | ----------- |
| TLS_AUTO | 0 | Envoy will choose the optimal TLS version. |
| TLSv1_0 | 1 | TLS 1.0 |
| TLSv1_1 | 2 | TLS 1.1 |
| TLSv1_2 | 3 | TLS 1.2 |
| TLSv1_3 | 4 | TLS 1.3 |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

