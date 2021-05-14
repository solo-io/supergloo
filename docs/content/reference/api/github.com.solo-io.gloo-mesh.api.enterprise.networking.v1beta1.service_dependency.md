
---

title: "service_dependency.proto"

---

## Package : `networking.enterprise.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for service_dependency.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## service_dependency.proto


## Table of Contents
  - [ServiceDependencySpec](#networking.enterprise.mesh.gloo.solo.io.ServiceDependencySpec)
  - [ServiceDependencyStatus](#networking.enterprise.mesh.gloo.solo.io.ServiceDependencyStatus)
  - [ServiceDependencyStatus.WorkloadsEntry](#networking.enterprise.mesh.gloo.solo.io.ServiceDependencyStatus.WorkloadsEntry)







<a name="networking.enterprise.mesh.gloo.solo.io.ServiceDependencySpec"></a>

### ServiceDependencySpec
A ServiceDependency specifies explicit dependencies between traffic sources and destinations in a service graph. Depending on the underlying service mesh, explicitly describing dependencies can improve the performance of the data plane by pruning away any unneeded networking configuration from the relevant proxies.<br>The complete set of service dependencies for a given traffic source is the aggregation of all unique Destinations selected by any applicable ServiceDependencies. If a traffic source has no applied ServiceDependencies, its service dependency configuration defaults to the behavior of the underlying service mesh.<br>Note that in order to block communication between sources and destinations not explicitly declared in a ServiceDependency, additional configuration on the underlying service mesh may be required. For instance, Istio must be configured with `outboundTrafficPolicy.Mode` set to `REGISTRY_ONLY` (see [here](https://istio.io/latest/docs/reference/config/istio.mesh.v1alpha1/#MeshConfig-OutboundTrafficPolicy)) to enforce this behavior.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sourceSelectors | [][common.mesh.gloo.solo.io.WorkloadSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.WorkloadSelector" >}}) | repeated | Select the traffic sources (i.e. Workloads) for this network ServiceDependency. If omitted, selects all Workloads. |
  | destinationSelectors | [][common.mesh.gloo.solo.io.DestinationSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.DestinationSelector" >}}) | repeated | Select the traffic targets (i.e. Destination) for this network ServiceDependency. If omitted, selects all Destinations. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.ServiceDependencyStatus"></a>

### ServiceDependencyStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the ServiceDependency metadata. If the observedGeneration does not match `metadata.generation`, Gloo Mesh has not processed the most recent version of this resource. |
  | state | [common.mesh.gloo.solo.io.ApprovalState]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.validation_state#common.mesh.gloo.solo.io.ApprovalState" >}}) |  | The state of the overall resource, will only show accepted if it has been successfully applied to all exported to Meshes. |
  | workloads | [][networking.enterprise.mesh.gloo.solo.io.ServiceDependencyStatus.WorkloadsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.service_dependency#networking.enterprise.mesh.gloo.solo.io.ServiceDependencyStatus.WorkloadsEntry" >}}) | repeated | The status of the ServiceDependency for each selected Workload. A ServiceDependency may have different statuses for each Workload it applies to. |
  | errors | []string | repeated | Any errors found while processing this generation of the resource. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.ServiceDependencyStatus.WorkloadsEntry"></a>

### ServiceDependencyStatus.WorkloadsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [networking.mesh.gloo.solo.io.ApprovalStatus]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.status#networking.mesh.gloo.solo.io.ApprovalStatus" >}}) |  |  |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

