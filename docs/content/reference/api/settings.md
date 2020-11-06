
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
  - [SettingsStatus](#settings.mesh.gloo.solo.io.SettingsStatus)







<a name="settings.mesh.gloo.solo.io.NetworkingExtensionsServer"></a>

### NetworkingExtensionsServer
Options for connecting to an external gRPC NetworkingExternsions server


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| address | string |  | TCP address of the Networking Extensions Server (including port) |
| insecure | bool |  | Communicate over HTTP rather than HTTPS |
| reconnectOnNetworkFailures | bool |  | Instruct SMH to automatically reconnect to the server on network failures |






<a name="settings.mesh.gloo.solo.io.SettingsSpec"></a>

### SettingsSpec
Configure global settings and defaults.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| mtls | networking.mesh.gloo.solo.io.TrafficPolicySpec.MTLS |  | Configure default mTLS settings for TrafficTargets (MTLS declared in TrafficPolicies take precedence) |
| networkingExtensionServers | []settings.mesh.gloo.solo.io.NetworkingExtensionsServer | repeated | Configure SMH Networking to communicate with one or more external gRPC NetworkingExtensions servers. Updates will be applied by the servers in the order they are listed (servers towards the end of the list take precedence). Note: Extension Servers have full write access to the output objects written by Gloo Mesh. |






<a name="settings.mesh.gloo.solo.io.SettingsStatus"></a>

### SettingsStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the Settings metadata. If the observedGeneration does not match generation, the controller has not processed the most recent version of this resource. |
| state | networking.mesh.gloo.solo.io.ApprovalState |  | The state of the overall resource. It will only show accepted if no processing errors encountered. |
| errors | []string | repeated | Any errors encountered while processing Settings object. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

