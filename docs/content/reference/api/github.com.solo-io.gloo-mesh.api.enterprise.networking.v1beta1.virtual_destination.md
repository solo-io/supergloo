
---

title: "virtual_destination.proto"

---

## Package : `networking.enterprise.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for virtual_destination.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## virtual_destination.proto


## Table of Contents
  - [VirtualDestinationBackingDestination](#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationBackingDestination)
  - [VirtualDestinationSpec](#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec)
  - [VirtualDestinationSpec.BackingDestinationList](#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.BackingDestinationList)
  - [VirtualDestinationSpec.LocalityConfig](#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig)
  - [VirtualDestinationSpec.LocalityConfig.Locality](#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig.Locality)
  - [VirtualDestinationSpec.LocalityConfig.LocalityFailoverDirective](#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig.LocalityFailoverDirective)
  - [VirtualDestinationSpec.MeshList](#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.MeshList)
  - [VirtualDestinationSpec.Port](#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.Port)
  - [VirtualDestinationStatus](#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationStatus)
  - [VirtualDestinationStatus.MeshesEntry](#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationStatus.MeshesEntry)
  - [VirtualDestinationStatus.SelectedDestinations](#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationStatus.SelectedDestinations)







<a name="networking.enterprise.mesh.gloo.solo.io.VirtualDestinationBackingDestination"></a>

### VirtualDestinationBackingDestination
A backing Destination. Has to be at the top level, as cue does not function well with referencing nested messages.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeService | [core.skv2.solo.io.ClusterObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ClusterObjectRef" >}}) |  | Reference to a Kubernetes Service. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec"></a>

### VirtualDestinationSpec
A VirtualDestination creates a new hostname to which client workloads can send requests. The hostname abstracts over a set of underlying Destinations and provides failover functionality between them. Failover order is determined by either an explicitly defined priority (`static`), or a list of locality directives (`localized`).<br>Each Destination backing the VirtualDestination must be configured with a [TrafficPolicy's outlier detection]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.traffic_policy/" >}}). Currently this feature only supports Destinations backed by Istio.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hostname | string |  | The DNS name of the VirtualDestination. Must be unique within the service mesh instance. |
  | port | [networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.Port]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.Port" >}}) |  | The port on which the VirtualDestination listens. |
  | virtualMesh | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | The VirtualMesh that this VirtualDestination will be visible to. |
  | meshList | [networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.MeshList]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.MeshList" >}}) |  | The Meshes that this VirtualDestination will be visible to. If multiple meshes are specified, they must all belong to the same VirtualMesh. |
  | static | [networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.BackingDestinationList]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.BackingDestinationList" >}}) |  | Failover priority is determined by an explicitly provided static ordering of Destinations. |
  | localized | [networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig" >}}) |  | Failover priority is determined by the localities of the traffic source and destination. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.BackingDestinationList"></a>

### VirtualDestinationSpec.BackingDestinationList
Failover priority is determined by an explicitly provided static ordering of Destinations. When a Destination in the list is in an unhealthy state (as determined by its configured outlier detection), requests sent to the VirtualDestination will be routed to the next healthy Destination in the list.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| destinations | [][networking.enterprise.mesh.gloo.solo.io.VirtualDestinationBackingDestination]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationBackingDestination" >}}) | repeated | The list of Destinations backing the VirtualDestination, ordered by decreasing priority. All Destinations must be either in the same Mesh or in Meshes that are grouped under a common VirtualMesh. Required, cannot be omitted. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig"></a>

### VirtualDestinationSpec.LocalityConfig
Enables failover based on locality. When a client workload makes a request to the VirtualDestination, Gloo Mesh will first try to direct traffic to the Destination instance geographically closest to the client workload. If outlier detection detects that the closest Destination is in an unhealthy state, requests will instead be routed to a Destination in one of the localities specified in the `to` field.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| destinationSelectors | [][common.mesh.gloo.solo.io.DestinationSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.DestinationSelector" >}}) | repeated | Selectors for the Destinations backing the VirtualDestination. All Destinations must be either in the same mesh or in meshes that are grouped under a common VirtualMesh. Currently only one Destination per cluster can be selected, more than one per cluster will be considered invalid. Required, cannot be omitted. |
  | failoverDirectives | [][networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig.LocalityFailoverDirective]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig.LocalityFailoverDirective" >}}) | repeated | Directives describing the locality failover behavior. |
  | outlierDetection | [networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.OutlierDetection]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.OutlierDetection" >}}) |  | Outlier detection to determine the health of the selected services. If not set will default to the folling: consecutiveGatewayErrors: 10 consecutive5XXErrors: 10 interval: 5s baseEjectionTime: 120s |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig.Locality"></a>

### VirtualDestinationSpec.LocalityConfig.Locality
A geographic location defined by a region, zone, and sub-zone.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| region | string |  | The locality's region. |
  | zone | string |  | The locality's zone. Currently this value is not used. |
  | subZone | string |  | The locality's sub-zone. Currently this value is not used. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig.LocalityFailoverDirective"></a>

### VirtualDestinationSpec.LocalityConfig.LocalityFailoverDirective



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| from | [networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig.Locality]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig.Locality" >}}) |  | The locality of the client workload. |
  | to | [][networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig.Locality]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig.Locality" >}}) | repeated | The list of Destination localities that can be routed to if the instance local to the client workload is not available. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.MeshList"></a>

### VirtualDestinationSpec.MeshList
A list of Mesh references.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| meshes | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated |  |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.Port"></a>

### VirtualDestinationSpec.Port
VirtualDestination port infomation. Contains information about which port to listen on, as well as which backend port to target.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| number | uint32 |  | REQUIRED: Port number which the VirutalDestination will listen on. |
  | protocol | string |  | OPTIONAL: Protocol of the requests sent to the VirtualDestination. Must be one of `HTTP`, `HTTPS`, `GRPC`, `HTTP2`, `MONGO`, `TCP`, `TLS`. |
  | targetName | string |  | If the target_name is specified, the VirtualDestination will attempt to find a port by this name on all backing services |
  | targetNumber | string |  | If the target_number is specified, the VirtualDestination will attempt to find a port by this number on all backing services |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualDestinationStatus"></a>

### VirtualDestinationStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the VirtualDestination metadata. If the observedGeneration does not match `metadata.generation`, Gloo Mesh has not processed the most recent version of this resource. |
  | state | [common.mesh.gloo.solo.io.ApprovalState]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.validation_state#common.mesh.gloo.solo.io.ApprovalState" >}}) |  | The state of the overall resource, will only show accepted if it has been successfully applied to all exported to Meshes. |
  | meshes | [][networking.enterprise.mesh.gloo.solo.io.VirtualDestinationStatus.MeshesEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationStatus.MeshesEntry" >}}) | repeated | The status of the VirtualDestination for each Mesh to which it has been exported to. |
  | selectedDestinations | [][networking.enterprise.mesh.gloo.solo.io.VirtualDestinationStatus.SelectedDestinations]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationStatus.SelectedDestinations" >}}) | repeated | The Destinations that comprise this VirtualDestination. |
  | errors | []string | repeated | Any errors found while processing this generation of the resource. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualDestinationStatus.MeshesEntry"></a>

### VirtualDestinationStatus.MeshesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [networking.mesh.gloo.solo.io.ApprovalStatus]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.status#networking.mesh.gloo.solo.io.ApprovalStatus" >}}) |  |  |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualDestinationStatus.SelectedDestinations"></a>

### VirtualDestinationStatus.SelectedDestinations



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.skv2.solo.io.ClusterObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ClusterObjectRef" >}}) |  | Reference to the Destination object. |
  | destination | [networking.enterprise.mesh.gloo.solo.io.VirtualDestinationBackingDestination]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationBackingDestination" >}}) |  | The platform-specific destination that the Destination object represents. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

