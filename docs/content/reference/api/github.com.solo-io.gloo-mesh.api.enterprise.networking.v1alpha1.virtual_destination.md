
---

title: "virtual_destination.proto"

---

## Package : `networking.enterprise.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for virtual_destination.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## virtual_destination.proto


## Table of Contents
  - [BackingService](#networking.enterprise.mesh.gloo.solo.io.BackingService)
  - [VirtualDestinationSpec](#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec)
  - [VirtualDestinationSpec.BackingServiceList](#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.BackingServiceList)
  - [VirtualDestinationSpec.LocalityConfig](#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig)
  - [VirtualDestinationSpec.LocalityConfig.Locality](#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig.Locality)
  - [VirtualDestinationSpec.LocalityConfig.LocalityFailoverDirective](#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig.LocalityFailoverDirective)
  - [VirtualDestinationSpec.MeshList](#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.MeshList)
  - [VirtualDestinationSpec.Port](#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.Port)
  - [VirtualDestinationStatus](#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationStatus)
  - [VirtualDestinationStatus.MeshesEntry](#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationStatus.MeshesEntry)
  - [VirtualDestinationStatus.SelectedTrafficTarget](#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationStatus.SelectedTrafficTarget)







<a name="networking.enterprise.mesh.gloo.solo.io.BackingService"></a>

### BackingService
A service represented by a TrafficTarget


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeService | [core.skv2.solo.io.ClusterObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ClusterObjectRef" >}}) |  | Name/namespace/cluster of a kubernetes service. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec"></a>

### VirtualDestinationSpec
A VirtualDestination creates a new hostname to which client workloads can send requests. Requests will be routed based on either a list of backing traffic targets ordered by explicit priority, or a list of locality directives. Each TrafficTarget backing the VirtualDestination must be configured with outlier detection through a TrafficPolicy.<br>Currently this feature only supports TrafficTargets backed by Istio.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hostname | string |  | The DNS name of the VirtualDestination. Must be unique within the service mesh instance. |
  | port | [networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.Port]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.Port" >}}) |  | The port on which the VirtualDestination listens. |
  | virtualMesh | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | The VirtualMesh that this VirtualDestination will be visible to. |
  | meshList | [networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.MeshList]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.MeshList" >}}) |  | The meshes that this VirtualDestination will be visible to. If multiple meshes are specified, they must all belong to the same VirtualMesh. |
  | static | [networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.BackingServiceList]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.BackingServiceList" >}}) |  | Failover priority is determined by an explicitly provided static ordering of TrafficTargets. |
  | localized | [networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig" >}}) |  | Failover priority is determined by the localities of the traffic source and destination. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.BackingServiceList"></a>

### VirtualDestinationSpec.BackingServiceList
Configure failover based on a list of TrafficTargets. When a TrafficTarget in the list is in an unhealthy state (as determined by its outlier detection configuration), requests sent to the VirtualDestination will be routed to the next healthy TrafficTarget in the list.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| services | [][networking.enterprise.mesh.gloo.solo.io.BackingService]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.BackingService" >}}) | repeated | The list of services backing the VirtualDestination, ordered by decreasing priority. All services must be either in the same mesh or in meshes that are grouped under a common VirtualMesh. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig"></a>

### VirtualDestinationSpec.LocalityConfig
Enables failover based on locality. When a client workload makes a request to the VirtualDestination, Gloo Mesh will first try to direct traffic to the service instance geographically closest to the client workload. If outlier detection detects that the closest traffic target is in an unhealthy state, requests will instead be routed to a service instance in one of the localities specified in the `to` field. Currently, each locality in the `to` field will be routed to with equal probability if the local instance is unhealthy.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| serviceSelectors | [][networking.mesh.gloo.solo.io.TrafficTargetSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.selectors#networking.mesh.gloo.solo.io.TrafficTargetSelector" >}}) | repeated | Selectors for the services backing the VirtualDestination. All services must be either in the same mesh or in meshes that are grouped under a common VirtualMesh. Currently only one service per cluster can be selected, more than one per cluster will be considered invalid. |
  | failoverDirectives | [][networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig.LocalityFailoverDirective]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig.LocalityFailoverDirective" >}}) | repeated | Directives describing the locality failover behavior. |
  | outlierDetection | [networking.mesh.gloo.solo.io.TrafficPolicySpec.OutlierDetection]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec.OutlierDetection" >}}) |  | Outlier detection to determine the health of the selected services. If not set will default to the folling: consecutiveGatewayErrors: 10 consecutive5XXErrors: 10 interval: 5s baseEjectionTime: 120s |
  





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
| from | [networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig.Locality]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig.Locality" >}}) |  | The locality of a client workload. |
  | to | [][networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig.Locality]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.LocalityConfig.Locality" >}}) | repeated | The list of traffic target localities that can be routed to if the instance local to the client workload is not available. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.MeshList"></a>

### VirtualDestinationSpec.MeshList
A list of mesh references.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| meshes | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated |  |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualDestinationSpec.Port"></a>

### VirtualDestinationSpec.Port
The port on which the VirtualDestination listens.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| number | uint32 |  | Port number. |
  | protocol | string |  | Protocol of the requests sent to the VirtualDestination. Must be one of HTTP, HTTPS, GRPC, HTTP2, MONGO, TCP, TLS. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualDestinationStatus"></a>

### VirtualDestinationStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the VirtualDestination metadata. If the observedGeneration does not match generation, the controller has not received the most recent version of this resource. |
  | state | [networking.mesh.gloo.solo.io.ApprovalState]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.validation_state#networking.mesh.gloo.solo.io.ApprovalState" >}}) |  | The state of the overall resource, will only show accepted if it has been successfully applied to all target meshes. |
  | meshes | [][networking.enterprise.mesh.gloo.solo.io.VirtualDestinationStatus.MeshesEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationStatus.MeshesEntry" >}}) | repeated | The status of the VirtualDestination for each Mesh to which it has been applied. |
  | selectedTrafficTargets | [][networking.enterprise.mesh.gloo.solo.io.VirtualDestinationStatus.SelectedTrafficTarget]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationStatus.SelectedTrafficTarget" >}}) | repeated | The traffic targets that comprise this Virtual Destination. |
  | errors | []string | repeated | Any errors found while processing this generation of the resource. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualDestinationStatus.MeshesEntry"></a>

### VirtualDestinationStatus.MeshesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [networking.mesh.gloo.solo.io.ApprovalStatus]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.validation_state#networking.mesh.gloo.solo.io.ApprovalStatus" >}}) |  |  |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualDestinationStatus.SelectedTrafficTarget"></a>

### VirtualDestinationStatus.SelectedTrafficTarget



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.skv2.solo.io.ClusterObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ClusterObjectRef" >}}) |  | Reference to the traffic target. |
  | service | [networking.enterprise.mesh.gloo.solo.io.BackingService]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.BackingService" >}}) |  | The service that the traffic target represents. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

