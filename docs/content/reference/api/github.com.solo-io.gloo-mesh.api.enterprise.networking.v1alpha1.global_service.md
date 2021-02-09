
---

title: "global_service.proto"

---

## Package : `networking.enterprise.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for global_service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## global_service.proto


## Table of Contents
  - [GlobalServiceSpec](#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec)
  - [GlobalServiceSpec.BackingServiceList](#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.BackingServiceList)
  - [GlobalServiceSpec.BackingServiceList.BackingService](#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.BackingServiceList.BackingService)
  - [GlobalServiceSpec.LocalityConfig](#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.LocalityConfig)
  - [GlobalServiceSpec.LocalityConfig.Locality](#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.LocalityConfig.Locality)
  - [GlobalServiceSpec.LocalityConfig.LocalityFailoverDirective](#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.LocalityConfig.LocalityFailoverDirective)
  - [GlobalServiceSpec.MeshList](#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.MeshList)
  - [GlobalServiceSpec.Port](#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.Port)
  - [GlobalServiceStatus](#networking.enterprise.mesh.gloo.solo.io.GlobalServiceStatus)
  - [GlobalServiceStatus.MeshesEntry](#networking.enterprise.mesh.gloo.solo.io.GlobalServiceStatus.MeshesEntry)
  - [GlobalServiceStatus.SelectedTrafficTarget](#networking.enterprise.mesh.gloo.solo.io.GlobalServiceStatus.SelectedTrafficTarget)
  - [GlobalServiceStatus.SelectedTrafficTarget.BackingService](#networking.enterprise.mesh.gloo.solo.io.GlobalServiceStatus.SelectedTrafficTarget.BackingService)







<a name="networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec"></a>

### GlobalServiceSpec
A GlobalService creates a new hostname to which client workloads can send requests. Requests will be routed based on either a list of backing traffic targets ordered by priority, or a list of locality directives. Each traffic target backing the GlobalService must be configured with outlier detection using a traffic policy.<br>Currently this feature only supports Services backed by Istio.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hostname | string |  | The DNS name of the GlobalService. Must be unique within the service mesh instance since it is used as the hostname with which clients communicate. |
  | port | [networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.Port]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.global_service#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.Port" >}}) |  | The port on which the GlobalService listens. |
  | virtualMesh | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | The VirtualMesh that this GlobalService will be visible to. |
  | meshList | [networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.MeshList]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.global_service#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.MeshList" >}}) |  | The meshes that this GlobalService will be visible to. If multiple meshes are specified, they must all belong to the same VirtualMesh. |
  | static | [networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.BackingServiceList]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.global_service#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.BackingServiceList" >}}) |  | List of backing services in priority order. |
  | localized | [networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.LocalityConfig]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.global_service#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.LocalityConfig" >}}) |  | Locality failover configuration. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.BackingServiceList"></a>

### GlobalServiceSpec.BackingServiceList
Enables failover based on a list of services. When outlier detection detects that a traffic target in the list is in an unhealthy state, requests sent to the GlobalService will be routed to the next healthy traffic target in the list.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| services | [][networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.BackingServiceList.BackingService]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.global_service#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.BackingServiceList.BackingService" >}}) | repeated | The list of services backing the GlobalService, ordered by decreasing priority. All services must be either in the same mesh or in meshes that are grouped under a common VirtualMesh. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.BackingServiceList.BackingService"></a>

### GlobalServiceSpec.BackingServiceList.BackingService
A service represented by a TrafficTarget


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeService | [core.skv2.solo.io.ClusterObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ClusterObjectRef" >}}) |  | Name/namespace/cluster of a kubernetes service. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.LocalityConfig"></a>

### GlobalServiceSpec.LocalityConfig
Enables failover based on locality. When a client workload makes a request to the GlobalService, Gloo Mesh will first try to direct traffic to the service instance geographically closest to the client workload. If outlier detection detects that the closest traffic target is in an unhealthy state, requests will instead be routed to a service instance in one of the localities specified in the `to` field. Currently, each locality in the `to` field will be routed to with equal probability if the local instance is unhealthy.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| serviceSelectors | [][networking.mesh.gloo.solo.io.TrafficTargetSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.selectors#networking.mesh.gloo.solo.io.TrafficTargetSelector" >}}) | repeated | Selectors for the services backing the GlobalService. All services must be either in the same mesh or in meshes that are grouped under a common VirtualMesh. |
  | failoverDirectives | [][networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.LocalityConfig.LocalityFailoverDirective]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.global_service#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.LocalityConfig.LocalityFailoverDirective" >}}) | repeated | Directives describing the locality failover behavior. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.LocalityConfig.Locality"></a>

### GlobalServiceSpec.LocalityConfig.Locality
A geographic location defined by a region, zone, and sub-zone.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| region | string |  | The locality's region. |
  | zone | string |  | The locality's zone. Currently this value is not used. |
  | subZone | string |  | The locality's sub-zone. Currently this value is not used. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.LocalityConfig.LocalityFailoverDirective"></a>

### GlobalServiceSpec.LocalityConfig.LocalityFailoverDirective



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| from | [networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.LocalityConfig.Locality]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.global_service#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.LocalityConfig.Locality" >}}) |  | The locality of a client workload. |
  | to | [][networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.LocalityConfig.Locality]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.global_service#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.LocalityConfig.Locality" >}}) | repeated | The list of traffic target localities that can be routed to if the instance local to the client workload is not available. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.MeshList"></a>

### GlobalServiceSpec.MeshList
A list of mesh references.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| meshes | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated |  |
  





<a name="networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.Port"></a>

### GlobalServiceSpec.Port
The port on which the GlobalService listens.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| number | uint32 |  | Port number. |
  | protocol | string |  | Protocol of the requests sent to the GlobalService. Must be one of HTTP, HTTPS, GRPC, HTTP2, MONGO, TCP, TLS. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.GlobalServiceStatus"></a>

### GlobalServiceStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the GlobalService metadata. If the observedGeneration does not match generation, the controller has not received the most recent version of this resource. |
  | state | [networking.mesh.gloo.solo.io.ApprovalState]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.validation_state#networking.mesh.gloo.solo.io.ApprovalState" >}}) |  | The state of the overall resource, will only show accepted if it has been successfully applied to all target meshes. |
  | meshes | [][networking.enterprise.mesh.gloo.solo.io.GlobalServiceStatus.MeshesEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.global_service#networking.enterprise.mesh.gloo.solo.io.GlobalServiceStatus.MeshesEntry" >}}) | repeated | The status of the GlobalService for each Mesh to which it has been applied. |
  | selectedTrafficTargets | [][networking.enterprise.mesh.gloo.solo.io.GlobalServiceStatus.SelectedTrafficTarget]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.global_service#networking.enterprise.mesh.gloo.solo.io.GlobalServiceStatus.SelectedTrafficTarget" >}}) | repeated | The traffic targets that comprise this Global Service. |
  | errors | []string | repeated | Any errors found while processing this generation of the resource. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.GlobalServiceStatus.MeshesEntry"></a>

### GlobalServiceStatus.MeshesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [networking.mesh.gloo.solo.io.ApprovalStatus]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.validation_state#networking.mesh.gloo.solo.io.ApprovalStatus" >}}) |  |  |
  





<a name="networking.enterprise.mesh.gloo.solo.io.GlobalServiceStatus.SelectedTrafficTarget"></a>

### GlobalServiceStatus.SelectedTrafficTarget



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.skv2.solo.io.ClusterObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ClusterObjectRef" >}}) |  | Reference to the traffic target. |
  | service | [networking.enterprise.mesh.gloo.solo.io.GlobalServiceStatus.SelectedTrafficTarget.BackingService]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.global_service#networking.enterprise.mesh.gloo.solo.io.GlobalServiceStatus.SelectedTrafficTarget.BackingService" >}}) |  | The service that the traffic target represents. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.GlobalServiceStatus.SelectedTrafficTarget.BackingService"></a>

### GlobalServiceStatus.SelectedTrafficTarget.BackingService
A service represented by a TrafficTarget


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeService | [core.skv2.solo.io.ClusterObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ClusterObjectRef" >}}) |  | Name/namespace/cluster of a kubernetes service. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

