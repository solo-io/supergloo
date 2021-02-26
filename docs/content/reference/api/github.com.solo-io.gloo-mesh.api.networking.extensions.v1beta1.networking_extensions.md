
---

title: "networking_extensions.proto"

---

## Package : `extensions.networking.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for networking_extensions.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## networking_extensions.proto


## Table of Contents
  - [DestinationObject](#extensions.networking.mesh.gloo.solo.io.DestinationObject)
  - [DiscoverySnapshot](#extensions.networking.mesh.gloo.solo.io.DiscoverySnapshot)
  - [ExtensionPatchRequest](#extensions.networking.mesh.gloo.solo.io.ExtensionPatchRequest)
  - [ExtensionPatchResponse](#extensions.networking.mesh.gloo.solo.io.ExtensionPatchResponse)
  - [GeneratedObject](#extensions.networking.mesh.gloo.solo.io.GeneratedObject)
  - [GeneratedObject.ConfigMap](#extensions.networking.mesh.gloo.solo.io.GeneratedObject.ConfigMap)
  - [GeneratedObject.ConfigMap.DataEntry](#extensions.networking.mesh.gloo.solo.io.GeneratedObject.ConfigMap.DataEntry)
  - [MeshObject](#extensions.networking.mesh.gloo.solo.io.MeshObject)
  - [ObjectMeta](#extensions.networking.mesh.gloo.solo.io.ObjectMeta)
  - [ObjectMeta.AnnotationsEntry](#extensions.networking.mesh.gloo.solo.io.ObjectMeta.AnnotationsEntry)
  - [ObjectMeta.LabelsEntry](#extensions.networking.mesh.gloo.solo.io.ObjectMeta.LabelsEntry)
  - [PushNotification](#extensions.networking.mesh.gloo.solo.io.PushNotification)
  - [WatchPushNotificationsRequest](#extensions.networking.mesh.gloo.solo.io.WatchPushNotificationsRequest)
  - [WorkloadObject](#extensions.networking.mesh.gloo.solo.io.WorkloadObject)



  - [NetworkingExtensions](#extensions.networking.mesh.gloo.solo.io.NetworkingExtensions)




<a name="extensions.networking.mesh.gloo.solo.io.DestinationObject"></a>

### DestinationObject
a proto-serializable representation of a Destination object


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [extensions.networking.mesh.gloo.solo.io.ObjectMeta]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.extensions.v1beta1.networking_extensions#extensions.networking.mesh.gloo.solo.io.ObjectMeta" >}}) |  | metadata of the object |
  | spec | [discovery.mesh.gloo.solo.io.DestinationSpec]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationSpec" >}}) |  | the spec of the object |
  | status | [discovery.mesh.gloo.solo.io.DestinationStatus]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationStatus" >}}) |  | the status of the object |
  





<a name="extensions.networking.mesh.gloo.solo.io.DiscoverySnapshot"></a>

### DiscoverySnapshot
a Protobuf representation of the set of Discovery objects used to produce the Networking outputs.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| meshes | [][extensions.networking.mesh.gloo.solo.io.MeshObject]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.extensions.v1beta1.networking_extensions#extensions.networking.mesh.gloo.solo.io.MeshObject" >}}) | repeated | all meshes in the discovery snapshot |
  | destinations | [][extensions.networking.mesh.gloo.solo.io.DestinationObject]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.extensions.v1beta1.networking_extensions#extensions.networking.mesh.gloo.solo.io.DestinationObject" >}}) | repeated | all Destinations in the discovery snapshot |
  | workloads | [][extensions.networking.mesh.gloo.solo.io.WorkloadObject]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.extensions.v1beta1.networking_extensions#extensions.networking.mesh.gloo.solo.io.WorkloadObject" >}}) | repeated | all workloads in the discovery snapshot |
  





<a name="extensions.networking.mesh.gloo.solo.io.ExtensionPatchRequest"></a>

### ExtensionPatchRequest
the parameters provided to the Extensions server when requesting patches


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| inputs | [extensions.networking.mesh.gloo.solo.io.DiscoverySnapshot]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.extensions.v1beta1.networking_extensions#extensions.networking.mesh.gloo.solo.io.DiscoverySnapshot" >}}) |  | the set of discovery objects provided as inputs for the Gloo Mesh translation |
  | outputs | [][extensions.networking.mesh.gloo.solo.io.GeneratedObject]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.extensions.v1beta1.networking_extensions#extensions.networking.mesh.gloo.solo.io.GeneratedObject" >}}) | repeated | the base set of output objects translated by Gloo Mesh. these may have been operated upon by a previous Extension server if multiple servers have been configured. |
  





<a name="extensions.networking.mesh.gloo.solo.io.ExtensionPatchResponse"></a>

### ExtensionPatchResponse
the set of patches the server wishes to apply to the Gloo Mesh Networking outputs. Any objects provided here will be inserted into the final Gloo Mesh snapshot. If an object already exists in the snapshot, it will be overridden by the version provided here. If multiple extensions servers are configured, this response may be operated upon by Extension patches provided by subsequent servers.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| patchedOutputs | [][extensions.networking.mesh.gloo.solo.io.GeneratedObject]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.extensions.v1beta1.networking_extensions#extensions.networking.mesh.gloo.solo.io.GeneratedObject" >}}) | repeated | the set of modified/added output objects desired by the Extension server. |
  





<a name="extensions.networking.mesh.gloo.solo.io.GeneratedObject"></a>

### GeneratedObject
a generated object can be of any output type supported by Gloo Mesh. the content of the type field should be used to determine the type of the output object. TODO(ilackarms): consider parameterizing Gloo Mesh to allow excluding GeneratedObjects from patch requests in the case where an implementer only performs additions (no updates required).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [extensions.networking.mesh.gloo.solo.io.ObjectMeta]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.extensions.v1beta1.networking_extensions#extensions.networking.mesh.gloo.solo.io.ObjectMeta" >}}) |  | metadata of the object |
  | destinationRule | [istio.networking.v1alpha3.DestinationRule]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.DestinationRule" >}}) |  |  |
  | envoyFilter | [istio.networking.v1alpha3.EnvoyFilter]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.envoy_filter#istio.networking.v1alpha3.EnvoyFilter" >}}) |  |  |
  | serviceEntry | [istio.networking.v1alpha3.ServiceEntry]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.service_entry#istio.networking.v1alpha3.ServiceEntry" >}}) |  |  |
  | virtualService | [istio.networking.v1alpha3.VirtualService]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.VirtualService" >}}) |  |  |
  | configMap | [extensions.networking.mesh.gloo.solo.io.GeneratedObject.ConfigMap]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.extensions.v1beta1.networking_extensions#extensions.networking.mesh.gloo.solo.io.GeneratedObject.ConfigMap" >}}) |  |  |
  | xdsConfig | [xds.agent.enterprise.mesh.gloo.solo.io.XdsConfigSpec]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.xds.agent.v1beta1.xds_config#xds.agent.enterprise.mesh.gloo.solo.io.XdsConfigSpec" >}}) |  | TODO(ilackarms): add more types here. note that we may need to support non-proto resourecs here in the future, in which case we will probably use a proto Struct to represent the object. |
  





<a name="extensions.networking.mesh.gloo.solo.io.GeneratedObject.ConfigMap"></a>

### GeneratedObject.ConfigMap



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| data | [][extensions.networking.mesh.gloo.solo.io.GeneratedObject.ConfigMap.DataEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.extensions.v1beta1.networking_extensions#extensions.networking.mesh.gloo.solo.io.GeneratedObject.ConfigMap.DataEntry" >}}) | repeated |  |
  





<a name="extensions.networking.mesh.gloo.solo.io.GeneratedObject.ConfigMap.DataEntry"></a>

### GeneratedObject.ConfigMap.DataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="extensions.networking.mesh.gloo.solo.io.MeshObject"></a>

### MeshObject
a proto-serializable representation of a Mesh object


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [extensions.networking.mesh.gloo.solo.io.ObjectMeta]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.extensions.v1beta1.networking_extensions#extensions.networking.mesh.gloo.solo.io.ObjectMeta" >}}) |  | metadata of the object |
  | spec | [discovery.mesh.gloo.solo.io.MeshSpec]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.mesh#discovery.mesh.gloo.solo.io.MeshSpec" >}}) |  | the spec of the object |
  | status | [discovery.mesh.gloo.solo.io.MeshStatus]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.mesh#discovery.mesh.gloo.solo.io.MeshStatus" >}}) |  | the status of the object |
  





<a name="extensions.networking.mesh.gloo.solo.io.ObjectMeta"></a>

### ObjectMeta
ObjectMeta is a simplified clone of the Kubernetes ObjectMeta used to represent object metadata for Kubernetes objects passed as messages in the NetworkingExtensions API.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | the Kubernetes name of the object |
  | namespace | string |  | the Kubernetes namespace of the object |
  | clusterName | string |  | the Kubernetes clusterName of the object (used internally by Gloo Mesh) |
  | labels | [][extensions.networking.mesh.gloo.solo.io.ObjectMeta.LabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.extensions.v1beta1.networking_extensions#extensions.networking.mesh.gloo.solo.io.ObjectMeta.LabelsEntry" >}}) | repeated | the Kubernetes labels on the object |
  | annotations | [][extensions.networking.mesh.gloo.solo.io.ObjectMeta.AnnotationsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.extensions.v1beta1.networking_extensions#extensions.networking.mesh.gloo.solo.io.ObjectMeta.AnnotationsEntry" >}}) | repeated | the Kubernetes annotations on the object |
  





<a name="extensions.networking.mesh.gloo.solo.io.ObjectMeta.AnnotationsEntry"></a>

### ObjectMeta.AnnotationsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="extensions.networking.mesh.gloo.solo.io.ObjectMeta.LabelsEntry"></a>

### ObjectMeta.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="extensions.networking.mesh.gloo.solo.io.PushNotification"></a>

### PushNotification
triggers a resync of Gloo Mesh objects






<a name="extensions.networking.mesh.gloo.solo.io.WatchPushNotificationsRequest"></a>

### WatchPushNotificationsRequest
request to initiate push notifications






<a name="extensions.networking.mesh.gloo.solo.io.WorkloadObject"></a>

### WorkloadObject
a proto-serializable representation of a Workload object


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [extensions.networking.mesh.gloo.solo.io.ObjectMeta]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.extensions.v1beta1.networking_extensions#extensions.networking.mesh.gloo.solo.io.ObjectMeta" >}}) |  | metadata of the object |
  | spec | [discovery.mesh.gloo.solo.io.WorkloadSpec]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.workload#discovery.mesh.gloo.solo.io.WorkloadSpec" >}}) |  | the spec of the object |
  | status | [discovery.mesh.gloo.solo.io.WorkloadStatus]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.workload#discovery.mesh.gloo.solo.io.WorkloadStatus" >}}) |  | the status of the object |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="extensions.networking.mesh.gloo.solo.io.NetworkingExtensions"></a>

### NetworkingExtensions
NetworkingExtensions provides customizable patches to Gloo Mesh-generated configuration. Gloo Mesh uses a NetworkingExtensions client to request optional patches from a pluggable NetworkingExtensions server. The server can return a set of patches which Gloo Mesh will apply before writing configuration to the cluster.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetExtensionPatches | [ExtensionPatchRequest](#extensions.networking.mesh.gloo.solo.io.ExtensionPatchRequest) | [ExtensionPatchResponse](#extensions.networking.mesh.gloo.solo.io.ExtensionPatchResponse) | GetExtensionPatches fetches a set of patches to the output configuration from the Extensions server. The current discovery snapshot and translated outputs are provided in the ExtensionPatchRequest |
| WatchPushNotifications | [WatchPushNotificationsRequest](#extensions.networking.mesh.gloo.solo.io.WatchPushNotificationsRequest) | [PushNotification](#extensions.networking.mesh.gloo.solo.io.PushNotification) stream | WatchPushNotifications initiates a streaming connection which allows the NetworkingExtensions server to push notifications to Gloo Mesh telling it to resync its configuration. This allows a NetworkingExtensions server to trigger Gloo Mesh to resync its state for events triggered by objects not watched by Gloo Mesh. |

 <!-- end services -->

