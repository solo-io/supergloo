
---

title: "settings.proto"

---

## Package : `settings.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for settings.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## settings.proto


## Table of Contents
  - [DiscoverySettings](#settings.mesh.gloo.solo.io.DiscoverySettings)
  - [DiscoverySettings.Istio](#settings.mesh.gloo.solo.io.DiscoverySettings.Istio)
  - [DiscoverySettings.Istio.IngressGatewayDetector](#settings.mesh.gloo.solo.io.DiscoverySettings.Istio.IngressGatewayDetector)
  - [DiscoverySettings.Istio.IngressGatewayDetector.GatewayWorkloadLabelsEntry](#settings.mesh.gloo.solo.io.DiscoverySettings.Istio.IngressGatewayDetector.GatewayWorkloadLabelsEntry)
  - [DiscoverySettings.Istio.IngressGatewayDetectorsEntry](#settings.mesh.gloo.solo.io.DiscoverySettings.Istio.IngressGatewayDetectorsEntry)
  - [GrpcServer](#settings.mesh.gloo.solo.io.GrpcServer)
  - [RelaySettings](#settings.mesh.gloo.solo.io.RelaySettings)
  - [SettingsSpec](#settings.mesh.gloo.solo.io.SettingsSpec)
  - [SettingsStatus](#settings.mesh.gloo.solo.io.SettingsStatus)







<a name="settings.mesh.gloo.solo.io.DiscoverySettings"></a>

### DiscoverySettings
Settings for Gloo Mesh discovery.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| istio | [settings.mesh.gloo.solo.io.DiscoverySettings.Istio]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1alpha2.settings#settings.mesh.gloo.solo.io.DiscoverySettings.Istio" >}}) |  | Istio-specific discovery settings |
  





<a name="settings.mesh.gloo.solo.io.DiscoverySettings.Istio"></a>

### DiscoverySettings.Istio
Istio-specific discovery settings


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ingressGatewayDetectors | [][settings.mesh.gloo.solo.io.DiscoverySettings.Istio.IngressGatewayDetectorsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1alpha2.settings#settings.mesh.gloo.solo.io.DiscoverySettings.Istio.IngressGatewayDetectorsEntry" >}}) | repeated | Configure discovery of ingress gateways per cluster. The key to the map is either a Gloo Mesh cluster name or `*` denoting all clusters. If an entry is found for a given cluster, it will be used. Otherwise, the wildcard entry will be used if it exists. Lastly, we will fall back to a set of default values. |
  





<a name="settings.mesh.gloo.solo.io.DiscoverySettings.Istio.IngressGatewayDetector"></a>

### DiscoverySettings.Istio.IngressGatewayDetector
Configure discovery of ingress gateways.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gatewayWorkloadLabels | [][settings.mesh.gloo.solo.io.DiscoverySettings.Istio.IngressGatewayDetector.GatewayWorkloadLabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1alpha2.settings#settings.mesh.gloo.solo.io.DiscoverySettings.Istio.IngressGatewayDetector.GatewayWorkloadLabelsEntry" >}}) | repeated | Workload labels used to detect ingress gateways for an Istio deployment. If not specified, will default to `{"istio": "ingressgateway"}`. |
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
  | value | [settings.mesh.gloo.solo.io.DiscoverySettings.Istio.IngressGatewayDetector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1alpha2.settings#settings.mesh.gloo.solo.io.DiscoverySettings.Istio.IngressGatewayDetector" >}}) |  |  |
  





<a name="settings.mesh.gloo.solo.io.GrpcServer"></a>

### GrpcServer
Options for connecting to an external gRPC server.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| address | string |  | TCP address of the gRPC Server (including port). |
  | insecure | bool |  | If true communicate over HTTP rather than HTTPS. |
  | reconnectOnNetworkFailures | bool |  | If true Gloo Mesh will automatically attempt to reconnect to the server after encountering network failures. |
  





<a name="settings.mesh.gloo.solo.io.RelaySettings"></a>

### RelaySettings
RelaySettings contains options for configuring Gloo Mesh to use Relay for cluster management. Relay provides a way for connecting Gloo Mesh to remote Kubernetes Clusters without the need to share credentials and access to remote Kube API Servers from the management cluster (the Gloo Mesh controllers).<br>Relay instead uses a streaming gRPC API to pass discovery data from remote clusters to the management cluster, and push configuration from the management cluster to the remote clusters.<br>Architecturally, it includes a Relay-agent which is installed to remote Kube clusters at registration time, which then connects directly to the Relay Server in the management cluster. to push its discovery data and pull its mesh configuration.<br> To configure Gloo Mesh to use Relay, make sure to read the [relay installation guide]({{< versioned_link_path fromRoot="/guides/setup/install_gloo_mesh" >}}) and [relay cluster registration guide]({{< versioned_link_path fromRoot="/guides/setup/register_cluster" >}}).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| enabled | bool |  | Enable the use of Relay for cluster management. If relay is enabled, make sure to follow the [relay cluster registration guide]({{< versioned_link_path fromRoot="/guides/setup/register_cluster#relay" >}}) for registering your clusters. |
  | server | [settings.mesh.gloo.solo.io.GrpcServer]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1alpha2.settings#settings.mesh.gloo.solo.io.GrpcServer" >}}) |  | TODO(harveyxia) can we remove this? Connection info for the Relay Server. Gloo Mesh will fetch discovery resources from this server and push translated outputs to this server. Note: currently this field has no effect as the relay server runs in-process of the networking Pod. |
  





<a name="settings.mesh.gloo.solo.io.SettingsSpec"></a>

### SettingsSpec
Configure system-wide settings and defaults. Settings specified in networking policies take precedence over those specified here.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| mtls | [networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.MTLS]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.MTLS" >}}) |  | Configure default mTLS settings for TrafficTargets. |
  | networkingExtensionServers | [][settings.mesh.gloo.solo.io.GrpcServer]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1alpha2.settings#settings.mesh.gloo.solo.io.GrpcServer" >}}) | repeated | Configure Gloo Mesh networking to communicate with one or more external gRPC NetworkingExtensions servers. Updates will be applied by the servers in the order they are listed (servers towards the end of the list take precedence). Note: Extension Servers have full write access to the output objects written by Gloo Mesh. |
  | discovery | [settings.mesh.gloo.solo.io.DiscoverySettings]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1alpha2.settings#settings.mesh.gloo.solo.io.DiscoverySettings" >}}) |  | Settings for Gloo Mesh discovery. |
  | relay | [settings.mesh.gloo.solo.io.RelaySettings]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.settings.v1alpha2.settings#settings.mesh.gloo.solo.io.RelaySettings" >}}) |  | Enable and configure use of Relay mode to communicate with remote clusters. This is an enterprise-only feature. |
  





<a name="settings.mesh.gloo.solo.io.SettingsStatus"></a>

### SettingsStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the Settings metadata. If the `observedGeneration` does not match `metadata.generation`, Gloo Mesh has not processed the most recent version of this resource. |
  | state | [networking.mesh.gloo.solo.io.ApprovalState]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.validation_state#networking.mesh.gloo.solo.io.ApprovalState" >}}) |  | The state of the overall resource. It will only show accepted if no processing errors encountered. |
  | errors | []string | repeated | Any errors encountered while processing Settings object. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

