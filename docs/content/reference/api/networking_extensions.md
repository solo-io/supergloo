
---
title: "networking_extensions.proto"
---

## Package : `extensions.networking.smh.solo.io`



<a name="top"></a>

<a name="API Reference for networking_extensions.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## networking_extensions.proto


## Table of Contents
  - [DiscoverySnapshot](#extensions.networking.smh.solo.io.DiscoverySnapshot)
  - [ExtensionPatchRequest](#extensions.networking.smh.solo.io.ExtensionPatchRequest)
  - [ExtensionPatchResponse](#extensions.networking.smh.solo.io.ExtensionPatchResponse)
  - [GeneratedObject](#extensions.networking.smh.solo.io.GeneratedObject)
  - [MeshObject](#extensions.networking.smh.solo.io.MeshObject)
  - [ObjectMeta](#extensions.networking.smh.solo.io.ObjectMeta)
  - [ObjectMeta.AnnotationsEntry](#extensions.networking.smh.solo.io.ObjectMeta.AnnotationsEntry)
  - [ObjectMeta.LabelsEntry](#extensions.networking.smh.solo.io.ObjectMeta.LabelsEntry)
  - [PushNotification](#extensions.networking.smh.solo.io.PushNotification)
  - [TrafficTargetObject](#extensions.networking.smh.solo.io.TrafficTargetObject)
  - [WatchPushNotificationsRequest](#extensions.networking.smh.solo.io.WatchPushNotificationsRequest)
  - [WorkloadObject](#extensions.networking.smh.solo.io.WorkloadObject)



  - [NetworkingExtensions](#extensions.networking.smh.solo.io.NetworkingExtensions)




<a name="extensions.networking.smh.solo.io.DiscoverySnapshot"></a>

### DiscoverySnapshot
a Protobuf representation of the set of Discovery objects used to produce the Networking outputs.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| meshes | []extensions.networking.smh.solo.io.MeshObject | repeated | all meshes in the discovery snapshot |
| trafficTargets | []extensions.networking.smh.solo.io.TrafficTargetObject | repeated | all traffic targets in the discovery snapshot |
| workloads | []extensions.networking.smh.solo.io.WorkloadObject | repeated | all workloads in the discovery snapshot |






<a name="extensions.networking.smh.solo.io.ExtensionPatchRequest"></a>

### ExtensionPatchRequest
the parameters provided to the Extensions server when requesting patches


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| inputs | extensions.networking.smh.solo.io.DiscoverySnapshot |  | the set of discovery objects provided as inputs for the SMH translation |
| outputs | []extensions.networking.smh.solo.io.GeneratedObject | repeated | the base set of output objects translated by SMH. these may have been operated upon by a previous Extension server if multiple servers have been configured. |






<a name="extensions.networking.smh.solo.io.ExtensionPatchResponse"></a>

### ExtensionPatchResponse
the set of patches the server wishes to apply to the SMH Networking outputs. Any objects provided here will be inserted into the final SMH snapshot. If an object already exists in the snapshot, it will be overridden by the version provided here. If multiple extensions servers are configured, this response may be operated upon by Extension patches provided by subsequent servers.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| patchedOutputs | []extensions.networking.smh.solo.io.GeneratedObject | repeated | the set of modified/added output objects desired by the Extension server. |






<a name="extensions.networking.smh.solo.io.GeneratedObject"></a>

### GeneratedObject
a generated object can be of any output type supported by SMH. the content of the type field should be used to determine the type of the output object. TODO(ilackarms): consider parameterizing SMH to allow excluding GeneratedObjects from patch requests in the case where an implementer only performs additions (no updates required).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | extensions.networking.smh.solo.io.ObjectMeta |  | metadata of the object |
| destinationRule | istio.networking.v1alpha3.DestinationRule |  |  |
| envoyFilter | istio.networking.v1alpha3.EnvoyFilter |  |  |
| serviceEntry | istio.networking.v1alpha3.ServiceEntry |  |  |
| virtualService | istio.networking.v1alpha3.VirtualService |  | TODO(ilackarms): add more types here. note that we may need to support non-proto resourecs here in the future, in which case we will probably use a proto Struct to represent the object. |






<a name="extensions.networking.smh.solo.io.MeshObject"></a>

### MeshObject
a proto-serializable representation of a Mesh object


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | extensions.networking.smh.solo.io.ObjectMeta |  | metadata of the object |
| spec | discovery.smh.solo.io.MeshSpec |  | the spec of the object |
| status | discovery.smh.solo.io.MeshStatus |  | the status of the object |






<a name="extensions.networking.smh.solo.io.ObjectMeta"></a>

### ObjectMeta
ObjectMeta is a simplified clone of the kubernetes ObjectMeta used to represent object metadata for K8s objects passed as messages in the NetworkingExtensions API.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | the kubernetes name of the object |
| namespace | string |  | the kubernetes namespace of the object |
| clusterName | string |  | the kubernetes clusterName of the object (used internally by Service Mesh Hub) |
| labels | []extensions.networking.smh.solo.io.ObjectMeta.LabelsEntry | repeated | the kubernetes labels on the object |
| annotations | []extensions.networking.smh.solo.io.ObjectMeta.AnnotationsEntry | repeated | the kubernetes annotations on the object |






<a name="extensions.networking.smh.solo.io.ObjectMeta.AnnotationsEntry"></a>

### ObjectMeta.AnnotationsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
| value | string |  |  |






<a name="extensions.networking.smh.solo.io.ObjectMeta.LabelsEntry"></a>

### ObjectMeta.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
| value | string |  |  |






<a name="extensions.networking.smh.solo.io.PushNotification"></a>

### PushNotification
triggers a resync of SMH objects






<a name="extensions.networking.smh.solo.io.TrafficTargetObject"></a>

### TrafficTargetObject
a proto-serializable representation of a TrafficTarget object


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | extensions.networking.smh.solo.io.ObjectMeta |  | metadata of the object |
| spec | discovery.smh.solo.io.TrafficTargetSpec |  | the spec of the object |
| status | discovery.smh.solo.io.TrafficTargetStatus |  | the status of the object |






<a name="extensions.networking.smh.solo.io.WatchPushNotificationsRequest"></a>

### WatchPushNotificationsRequest
request to initiate push notifications






<a name="extensions.networking.smh.solo.io.WorkloadObject"></a>

### WorkloadObject
a proto-serializable representation of a Workload object


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | extensions.networking.smh.solo.io.ObjectMeta |  | metadata of the object |
| spec | discovery.smh.solo.io.WorkloadSpec |  | the spec of the object |
| status | discovery.smh.solo.io.WorkloadStatus |  | the status of the object |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="extensions.networking.smh.solo.io.NetworkingExtensions"></a>

### NetworkingExtensions
NetworkingExtensions provides customizeable patches to Service Mesh Hub-generated configuration. Service Mesh Hub uses a NetworkingExtensions client to request optional patches from a pluggable NetworkingExtensions server. The server can return a set of patches which SMH will apply before writing configuration to the cluster.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetExtensionPatches | [ExtensionPatchRequest](#extensions.networking.smh.solo.io.ExtensionPatchRequest) | [ExtensionPatchResponse](#extensions.networking.smh.solo.io.ExtensionPatchResponse) | GetExtensionPatches fetches a set of patches to the output configuration from the Extensions server. The current discovery snapshot and translated outputs are provided in the ExtensionPatchRequest |
| WatchPushNotifications | [WatchPushNotificationsRequest](#extensions.networking.smh.solo.io.WatchPushNotificationsRequest) | [PushNotification](#extensions.networking.smh.solo.io.PushNotification) stream | WatchPushNotifications initiates a streaming connection which allows the NetworkingExtensions server to push notifications to Service Mesh Hub telling it to resync its configuration. This allows a NetworkingExtensions server to trigger SMH to resync its state for events triggered by objects not watched by SMH. |

 <!-- end services -->

