
---

title: "destination.proto"

---

## Package : `discovery.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for destination.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## destination.proto


## Table of Contents
  - [DestinationSpec](#discovery.mesh.gloo.solo.io.DestinationSpec)
  - [DestinationSpec.KubeService](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService)
  - [DestinationSpec.KubeService.EndpointsSubset](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset)
  - [DestinationSpec.KubeService.EndpointsSubset.Endpoint](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset.Endpoint)
  - [DestinationSpec.KubeService.EndpointsSubset.Endpoint.LabelsEntry](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset.Endpoint.LabelsEntry)
  - [DestinationSpec.KubeService.EndpointsSubset.Endpoint.SubLocality](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset.Endpoint.SubLocality)
  - [DestinationSpec.KubeService.KubeServicePort](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.KubeServicePort)
  - [DestinationSpec.KubeService.LabelsEntry](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.LabelsEntry)
  - [DestinationSpec.KubeService.Subset](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.Subset)
  - [DestinationSpec.KubeService.SubsetsEntry](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.SubsetsEntry)
  - [DestinationSpec.KubeService.WorkloadSelectorLabelsEntry](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.WorkloadSelectorLabelsEntry)
  - [DestinationStatus](#discovery.mesh.gloo.solo.io.DestinationStatus)
  - [DestinationStatus.AppliedAccessPolicy](#discovery.mesh.gloo.solo.io.DestinationStatus.AppliedAccessPolicy)
  - [DestinationStatus.AppliedFederation](#discovery.mesh.gloo.solo.io.DestinationStatus.AppliedFederation)
  - [DestinationStatus.AppliedTrafficPolicy](#discovery.mesh.gloo.solo.io.DestinationStatus.AppliedTrafficPolicy)







<a name="discovery.mesh.gloo.solo.io.DestinationSpec"></a>

### DestinationSpec
The Destination is an abstraction for any entity capable of receiving networking requests.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeService | [discovery.mesh.gloo.solo.io.DestinationSpec.KubeService]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService" >}}) |  | Describes the Kubernetes service backing this Destination. |
  | mesh | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | The mesh that controls this Destination. |
  





<a name="discovery.mesh.gloo.solo.io.DestinationSpec.KubeService"></a>

### DestinationSpec.KubeService
Describes a Kubernetes service.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.skv2.solo.io.ClusterObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ClusterObjectRef" >}}) |  | Reference to the Kubernetes service object. |
  | workloadSelectorLabels | [][discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.WorkloadSelectorLabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.WorkloadSelectorLabelsEntry" >}}) | repeated | Selectors for the set of pods targeted by the Kubernetes service. |
  | labels | [][discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.LabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.LabelsEntry" >}}) | repeated | Labels on the Kubernetes service. |
  | ports | [][discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.KubeServicePort]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.KubeServicePort" >}}) | repeated | The ports exposed by the underlying service. |
  | subsets | [][discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.SubsetsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.SubsetsEntry" >}}) | repeated | Subsets for routing, based on labels. |
  | region | string |  | The region the service resides in, typically representing a large geographic area. |
  | endpointSubsets | [][discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset" >}}) | repeated | Each endpoints subset is a group of endpoints arranged in terms of IP/port pairs. This API mirrors the [Kubernetes Endpoints API](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#endpoints-v1-core). |
  





<a name="discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset"></a>

### DestinationSpec.KubeService.EndpointsSubset
A series of IP addresses and their associated ports. The list of IP and port pairs is the cartesian product of the endpoint and port lists.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| endpoints | [][discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset.Endpoint]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset.Endpoint" >}}) | repeated |  |
  | ports | [][discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.KubeServicePort]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.KubeServicePort" >}}) | repeated |  |
  





<a name="discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset.Endpoint"></a>

### DestinationSpec.KubeService.EndpointsSubset.Endpoint
An endpoint exposed by this service.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ipAddress | string |  |  |
  | labels | [][discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset.Endpoint.LabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset.Endpoint.LabelsEntry" >}}) | repeated | Labels which belong to this IP. These are taken from the backing workload instance. |
  | subLocality | [discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset.Endpoint.SubLocality]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset.Endpoint.SubLocality" >}}) |  | The zone and sub-zone (if controlled by Istio) of the endpoint. |
  





<a name="discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset.Endpoint.LabelsEntry"></a>

### DestinationSpec.KubeService.EndpointsSubset.Endpoint.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset.Endpoint.SubLocality"></a>

### DestinationSpec.KubeService.EndpointsSubset.Endpoint.SubLocality
A subdivision of a region representing a set of physically colocated compute resources.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| zone | string |  | A subdivision of a geographical region, see [here](https://kubernetes.io/docs/reference/labels-annotations-taints/#topologykubernetesiozone) for more information. |
  | subZone | string |  | A subdivision of zone. Only applies to Istio-controlled Destinations, see [here](https://istio.io/latest/docs/tasks/traffic-management/locality-load-balancing/) for more information. |
  





<a name="discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.KubeServicePort"></a>

### DestinationSpec.KubeService.KubeServicePort
Describes the service's ports. See [here](https://kubernetes.io/docs/concepts/services-networking/service/#multi-port-services) for more information.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | uint32 |  | External-facing port for this Kubernetes service (*not* the service's target port on the targeted pods). |
  | name | string |  |  |
  | protocol | string |  |  |
  | appProtocol | string |  | Available in Kubernetes 1.18+, describes the application protocol. |
  





<a name="discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.LabelsEntry"></a>

### DestinationSpec.KubeService.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.Subset"></a>

### DestinationSpec.KubeService.Subset
Subsets for routing, based on labels.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| values | []string | repeated |  |
  





<a name="discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.SubsetsEntry"></a>

### DestinationSpec.KubeService.SubsetsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.Subset]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.Subset" >}}) |  |  |
  





<a name="discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.WorkloadSelectorLabelsEntry"></a>

### DestinationSpec.KubeService.WorkloadSelectorLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="discovery.mesh.gloo.solo.io.DestinationStatus"></a>

### DestinationStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the TrafficPolicy metadata. If the `observedGeneration` does not match `metadata.generation`, Gloo Mesh has not processed the most recent version of this resource. |
  | appliedTrafficPolicies | [][discovery.mesh.gloo.solo.io.DestinationStatus.AppliedTrafficPolicy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationStatus.AppliedTrafficPolicy" >}}) | repeated | The set of TrafficPolicies that have been applied to this Destination. |
  | appliedAccessPolicies | [][discovery.mesh.gloo.solo.io.DestinationStatus.AppliedAccessPolicy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationStatus.AppliedAccessPolicy" >}}) | repeated | The set of AccessPolicies that have been applied to this Destination. |
  | localFqdn | string |  | The fully qualified domain name for requests originating from a source *coloated* with this Destination. For Kubernetes services, "colocated" means within the same Kubernetes cluster. |
  | appliedFederation | [discovery.mesh.gloo.solo.io.DestinationStatus.AppliedFederation]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationStatus.AppliedFederation" >}}) |  | Federation metadata. Only populated if this Destination is federated through a VirtualMesh. |
  





<a name="discovery.mesh.gloo.solo.io.DestinationStatus.AppliedAccessPolicy"></a>

### DestinationStatus.AppliedAccessPolicy
Describes an [AccessPolicy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.access_policy" >}}) that applies to this Destination. If an existing AccessPolicy becomes invalid, the last valid applied policy will be used.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | Reference to the AccessPolicy object. |
  | observedGeneration | int64 |  | The observed generation of the accepted AccessPolicy. |
  | spec | [networking.mesh.gloo.solo.io.AccessPolicySpec]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.access_policy#networking.mesh.gloo.solo.io.AccessPolicySpec" >}}) |  | The spec of the last known valid AccessPolicy. |
  





<a name="discovery.mesh.gloo.solo.io.DestinationStatus.AppliedFederation"></a>

### DestinationStatus.AppliedFederation
Describes the federation configuration applied to this Destination through a [VirtualMesh]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.virtual_mesh" >}}). Federation allows access to the Destination from other meshes/clusters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| federatedHostname | string |  | For any Workload that this Destination has been federated to (i.e., any Workload controlled by a Mesh whose reference appears in `federated_to_meshes`), that Workload will be able to reach this Destination using this DNS name. For Kubernetes Destinations this includes Workloads on clusters other than the one hosting this Destination. |
  | federatedToMeshes | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated | The list of Meshes which are able to resolve this Destination's `multicluster_dns_name`. |
  | flatNetwork | bool |  | Whether or not the Destination has been federated to the given meshes using a VirtualMesh where [Federation.FlatNetwork]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.virtual_mesh/#virtualmeshspecfederation" >}}) is true. |
  





<a name="discovery.mesh.gloo.solo.io.DestinationStatus.AppliedTrafficPolicy"></a>

### DestinationStatus.AppliedTrafficPolicy
Describes a [TrafficPolicy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.traffic_policy" >}}) that applies to the Destination. If an existing TrafficPolicy becomes invalid, the last valid applied TrafficPolicy will be used.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | Reference to the TrafficPolicy object. |
  | observedGeneration | int64 |  | The observed generation of the accepted TrafficPolicy. |
  | spec | [networking.mesh.gloo.solo.io.TrafficPolicySpec]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec" >}}) |  | The spec of the last known valid TrafficPolicy. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

