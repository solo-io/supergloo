
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
  - [GlobalServiceSpec.MeshList](#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.MeshList)
  - [GlobalServiceSpec.Port](#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.Port)
  - [GlobalServiceStatus](#networking.enterprise.mesh.gloo.solo.io.GlobalServiceStatus)
  - [GlobalServiceStatus.MeshesEntry](#networking.enterprise.mesh.gloo.solo.io.GlobalServiceStatus.MeshesEntry)







<a name="networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec"></a>

### GlobalServiceSpec
A GlobalService creates a new hostname to which services can send requests. Requests will be routed based on a list of backing traffic targets ordered by decreasing priority. When outlier detection detects that a traffic target in the list is in an unhealthy state, requests sent to the GlobalService will be routed to the next healthy traffic target in the list. For each traffic target referenced in the GlobalService's BackingServices list, outlier detection must be configured using a TrafficPolicy.<br>Currently this feature only supports Services backed by Istio.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hostname | string |  | The DNS name of the GlobalService. Must be unique within the service mesh instance since it is used as the hostname with which clients communicate. |
  | port | [networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.Port]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.global_service#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.Port" >}}) |  | The port on which the GlobalService listens. |
  | trafficTargetSelectors | [][networking.mesh.gloo.solo.io.TrafficTargetSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.selectors#networking.mesh.gloo.solo.io.TrafficTargetSelector" >}}) | repeated | Selectors for the services backing the GlobalService. The selected services are considered equivalent, and Gloo Mesh will route to the optimal service instance based on the locality failover configuration. |
  | virtualMesh | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | The VirtualMesh that this GlobalService will be visible to. |
  | meshes | [networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.MeshList]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.global_service#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.MeshList" >}}) |  | The meshes that this GlobalService will be visible to. If multiple meshes are specified, they must all belong to the same VirtualMesh. |
  | backingServices | [networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.BackingServiceList]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.global_service#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.BackingServiceList" >}}) |  |  |
  | localityConfig | [networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.LocalityConfig]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.global_service#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.LocalityConfig" >}}) |  | Locality Failover configuration. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.BackingServiceList"></a>

### GlobalServiceSpec.BackingServiceList



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| backingServices | [][networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.BackingServiceList.BackingService]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.global_service#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.BackingServiceList.BackingService" >}}) | repeated | The list of services backing the GlobalService, ordered by decreasing priority. All services must be backed by either the same service mesh instance or backed by service meshes that are grouped under a common VirtualMesh. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.BackingServiceList.BackingService"></a>

### GlobalServiceSpec.BackingServiceList.BackingService
The traffic targets that comprise the GlobalService.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeService | [core.skv2.solo.io.ClusterObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ClusterObjectRef" >}}) |  | Name/namespace/cluster of a kubernetes service. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.LocalityConfig"></a>

### GlobalServiceSpec.LocalityConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| localities | [][networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.LocalityConfig.Locality]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.global_service#networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.LocalityConfig.Locality" >}}) | repeated |  |
  





<a name="networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.LocalityConfig.Locality"></a>

### GlobalServiceSpec.LocalityConfig.Locality



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| from | string |  |  |
  | to | []string | repeated |  |
  





<a name="networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.MeshList"></a>

### GlobalServiceSpec.MeshList



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| meshRefs | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated |  |
  





<a name="networking.enterprise.mesh.gloo.solo.io.GlobalServiceSpec.Port"></a>

### GlobalServiceSpec.Port
The port on which the GlobalService listens.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| number | uint32 |  | Port number. |
  | protocol | string |  | Protocol of the requests sent to the GlobalService, must be one of HTTP, HTTPS, GRPC, HTTP2, MONGO, TCP, TLS. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.GlobalServiceStatus"></a>

### GlobalServiceStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the GlobalService metadata. If the observedGeneration does not match generation, the controller has not received the most recent version of this resource. |
  | state | [networking.mesh.gloo.solo.io.ApprovalState]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.validation_state#networking.mesh.gloo.solo.io.ApprovalState" >}}) |  | The state of the overall resource, will only show accepted if it has been successfully applied to all target meshes. |
  | meshes | [][networking.enterprise.mesh.gloo.solo.io.GlobalServiceStatus.MeshesEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.global_service#networking.enterprise.mesh.gloo.solo.io.GlobalServiceStatus.MeshesEntry" >}}) | repeated | The status of the GlobalService for each Mesh to which it has been applied. |
  | errors | []string | repeated | Any errors found while processing this generation of the resource. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.GlobalServiceStatus.MeshesEntry"></a>

### GlobalServiceStatus.MeshesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [networking.mesh.gloo.solo.io.ApprovalStatus]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.validation_state#networking.mesh.gloo.solo.io.ApprovalStatus" >}}) |  |  |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

