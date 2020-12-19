
---
title: "settings.proto"
---

## Package : `settings.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for settings.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## settings.proto


## Table of Contents
  - [NetworkingExtensionsServer](#settings.mesh.gloo.solo.io.NetworkingExtensionsServer)
  - [SettingsSpec](#settings.mesh.gloo.solo.io.SettingsSpec)
  - [SettingsSpec.Istio](#settings.mesh.gloo.solo.io.SettingsSpec.Istio)
  - [SettingsSpec.Istio.IngressGatewayDetector](#settings.mesh.gloo.solo.io.SettingsSpec.Istio.IngressGatewayDetector)
  - [SettingsSpec.Istio.IngressGatewayDetector.GatewayWorkloadLabelsEntry](#settings.mesh.gloo.solo.io.SettingsSpec.Istio.IngressGatewayDetector.GatewayWorkloadLabelsEntry)
  - [SettingsSpec.Istio.IngressGatewayDetectorsEntry](#settings.mesh.gloo.solo.io.SettingsSpec.Istio.IngressGatewayDetectorsEntry)
  - [SettingsStatus](#settings.mesh.gloo.solo.io.SettingsStatus)







<a name="settings.mesh.gloo.solo.io.NetworkingExtensionsServer"></a>

### NetworkingExtensionsServer
Options for connecting to an external gRPC NetworkingExtensions server


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| address | string |  | TCP address of the Networking Extensions Server (including port) |
  | insecure | bool |  | Communicate over HTTP rather than HTTPS |
  | reconnectOnNetworkFailures | bool |  | Instruct Gloo Mesh to automatically reconnect to the server on network failures |
  





<a name="settings.mesh.gloo.solo.io.SettingsSpec"></a>

### SettingsSpec
Configure global settings and defaults.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| mtls | [networking.mesh.gloo.solo.io.TrafficPolicySpec.MTLS]({{< ref "traffic_policy.md#networking.mesh.gloo.solo.io.TrafficPolicySpec.MTLS" >}}) |  | Configure default mTLS settings for TrafficTargets (MTLS declared in TrafficPolicies take precedence) |
  | networkingExtensionServers | [][settings.mesh.gloo.solo.io.NetworkingExtensionsServer]({{< ref "settings.md#settings.mesh.gloo.solo.io.NetworkingExtensionsServer" >}}) | repeated | Configure Gloo Mesh networking to communicate with one or more external gRPC NetworkingExtensions servers. Updates will be applied by the servers in the order they are listed (servers towards the end of the list take precedence). Note: Extension Servers have full write access to the output objects written by Gloo Mesh. |
  | istio | [settings.mesh.gloo.solo.io.SettingsSpec.Istio]({{< ref "settings.md#settings.mesh.gloo.solo.io.SettingsSpec.Istio" >}}) |  | Istio-specific discovery settings |
  





<a name="settings.mesh.gloo.solo.io.SettingsSpec.Istio"></a>

### SettingsSpec.Istio



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ingressGatewayDetectors | [][settings.mesh.gloo.solo.io.SettingsSpec.Istio.IngressGatewayDetectorsEntry]({{< ref "settings.md#settings.mesh.gloo.solo.io.SettingsSpec.Istio.IngressGatewayDetectorsEntry" >}}) | repeated | Ingress gateway detectors for each cluster. The key to the map is either a k8s cluster name or the wildcard `*` meaning all clusters. If an entry is found for a given cluster, it will be used. Otherwise, the wildcard entry will be used if it exists. Lastly, we will fall back to the default values. |
  





<a name="settings.mesh.gloo.solo.io.SettingsSpec.Istio.IngressGatewayDetector"></a>

### SettingsSpec.Istio.IngressGatewayDetector
Workload labels and TLS port name used during discovery to detect ingress gateways for a mesh.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gatewayWorkloadLabels | [][settings.mesh.gloo.solo.io.SettingsSpec.Istio.IngressGatewayDetector.GatewayWorkloadLabelsEntry]({{< ref "settings.md#settings.mesh.gloo.solo.io.SettingsSpec.Istio.IngressGatewayDetector.GatewayWorkloadLabelsEntry" >}}) | repeated | The workload labels used during discovery to detect ingress gateways for a mesh. If not specified, will default to `{"istio": "ingressgateway"}`. |
  | gatewayTlsPortName | string |  | The name of the TLS port used to detect ingress gateways. Services must have a port with this name in order to be recognized as an ingress gateway during discovery. If not specified, will default to `tls`. |
  





<a name="settings.mesh.gloo.solo.io.SettingsSpec.Istio.IngressGatewayDetector.GatewayWorkloadLabelsEntry"></a>

### SettingsSpec.Istio.IngressGatewayDetector.GatewayWorkloadLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="settings.mesh.gloo.solo.io.SettingsSpec.Istio.IngressGatewayDetectorsEntry"></a>

### SettingsSpec.Istio.IngressGatewayDetectorsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [settings.mesh.gloo.solo.io.SettingsSpec.Istio.IngressGatewayDetector]({{< ref "settings.md#settings.mesh.gloo.solo.io.SettingsSpec.Istio.IngressGatewayDetector" >}}) |  |  |
  





<a name="settings.mesh.gloo.solo.io.SettingsStatus"></a>

### SettingsStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the Settings metadata. If the observedGeneration does not match generation, the controller has not processed the most recent version of this resource. |
  | state | [networking.mesh.gloo.solo.io.ApprovalState]({{< ref "validation_state.md#networking.mesh.gloo.solo.io.ApprovalState" >}}) |  | The state of the overall resource. It will only show accepted if no processing errors encountered. |
  | errors | []string | repeated | Any errors encountered while processing Settings object. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

