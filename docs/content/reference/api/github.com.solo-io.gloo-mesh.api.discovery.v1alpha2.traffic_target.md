
---

title: "traffic_target.proto"

---

## Package : `discovery.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for traffic_target.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## traffic_target.proto


## Table of Contents
  - [TrafficTargetSpec](#discovery.mesh.gloo.solo.io.TrafficTargetSpec)
  - [TrafficTargetSpec.ExternalService](#discovery.mesh.gloo.solo.io.TrafficTargetSpec.ExternalService)
  - [TrafficTargetSpec.ExternalService.ExtEndpoint](#discovery.mesh.gloo.solo.io.TrafficTargetSpec.ExternalService.ExtEndpoint)
  - [TrafficTargetSpec.ExternalService.ExtEndpoint.PortsEntry](#discovery.mesh.gloo.solo.io.TrafficTargetSpec.ExternalService.ExtEndpoint.PortsEntry)
  - [TrafficTargetSpec.ExternalService.ServicePort](#discovery.mesh.gloo.solo.io.TrafficTargetSpec.ExternalService.ServicePort)
  - [TrafficTargetSpec.KubeService](#discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService)
  - [TrafficTargetSpec.KubeService.EndpointsSubset](#discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.EndpointsSubset)
  - [TrafficTargetSpec.KubeService.EndpointsSubset.Endpoint](#discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.EndpointsSubset.Endpoint)
  - [TrafficTargetSpec.KubeService.EndpointsSubset.Endpoint.LabelsEntry](#discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.EndpointsSubset.Endpoint.LabelsEntry)
  - [TrafficTargetSpec.KubeService.KubeServicePort](#discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.KubeServicePort)
  - [TrafficTargetSpec.KubeService.LabelsEntry](#discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.LabelsEntry)
  - [TrafficTargetSpec.KubeService.Subset](#discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.Subset)
  - [TrafficTargetSpec.KubeService.SubsetsEntry](#discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.SubsetsEntry)
  - [TrafficTargetSpec.KubeService.WorkloadSelectorLabelsEntry](#discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.WorkloadSelectorLabelsEntry)
  - [TrafficTargetStatus](#discovery.mesh.gloo.solo.io.TrafficTargetStatus)
  - [TrafficTargetStatus.AppliedAccessPolicy](#discovery.mesh.gloo.solo.io.TrafficTargetStatus.AppliedAccessPolicy)
  - [TrafficTargetStatus.AppliedFederation](#discovery.mesh.gloo.solo.io.TrafficTargetStatus.AppliedFederation)
  - [TrafficTargetStatus.AppliedTrafficPolicy](#discovery.mesh.gloo.solo.io.TrafficTargetStatus.AppliedTrafficPolicy)







<a name="discovery.mesh.gloo.solo.io.TrafficTargetSpec"></a>

### TrafficTargetSpec
The TrafficTarget is an abstraction for a traffic target which we have discovered to be part of a given mesh.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeService | [discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.traffic_target#discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService" >}}) |  | KubeService is a kube-native Traffic Target representing a kubernetes service running inside of a kubernetes cluster. |
  | externalService | [discovery.mesh.gloo.solo.io.TrafficTargetSpec.ExternalService]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.traffic_target#discovery.mesh.gloo.solo.io.TrafficTargetSpec.ExternalService" >}}) |  | ExternalService is a Traffic Target representing a service external to the Mesh. It can be used to expose a given hostname or IP address to all clusters in the Virtual Mesh. |
  | mesh | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | The mesh with which this traffic target is associated. Can be omitted if the traffic target isn't associated with any particular mesh, eg for externalservices. |
  





<a name="discovery.mesh.gloo.solo.io.TrafficTargetSpec.ExternalService"></a>

### TrafficTargetSpec.ExternalService



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  |  |
  | hosts | []string | repeated |  |
  | addresses | []string | repeated |  |
  | ports | [][discovery.mesh.gloo.solo.io.TrafficTargetSpec.ExternalService.ServicePort]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.traffic_target#discovery.mesh.gloo.solo.io.TrafficTargetSpec.ExternalService.ServicePort" >}}) | repeated |  |
  | endpoints | [][discovery.mesh.gloo.solo.io.TrafficTargetSpec.ExternalService.ExtEndpoint]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.traffic_target#discovery.mesh.gloo.solo.io.TrafficTargetSpec.ExternalService.ExtEndpoint" >}}) | repeated |  |
  





<a name="discovery.mesh.gloo.solo.io.TrafficTargetSpec.ExternalService.ExtEndpoint"></a>

### TrafficTargetSpec.ExternalService.ExtEndpoint



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| address | string |  |  |
  | ports | [][discovery.mesh.gloo.solo.io.TrafficTargetSpec.ExternalService.ExtEndpoint.PortsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.traffic_target#discovery.mesh.gloo.solo.io.TrafficTargetSpec.ExternalService.ExtEndpoint.PortsEntry" >}}) | repeated |  |
  





<a name="discovery.mesh.gloo.solo.io.TrafficTargetSpec.ExternalService.ExtEndpoint.PortsEntry"></a>

### TrafficTargetSpec.ExternalService.ExtEndpoint.PortsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | uint32 |  |  |
  





<a name="discovery.mesh.gloo.solo.io.TrafficTargetSpec.ExternalService.ServicePort"></a>

### TrafficTargetSpec.ExternalService.ServicePort



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| number | uint32 |  |  |
  | name | string |  |  |
  | protocol | string |  |  |
  





<a name="discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService"></a>

### TrafficTargetSpec.KubeService



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.skv2.solo.io.ClusterObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ClusterObjectRef" >}}) |  | A reference to the kube-native traffic target that this TrafficTarget represents. |
  | workloadSelectorLabels | [][discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.WorkloadSelectorLabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.traffic_target#discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.WorkloadSelectorLabelsEntry" >}}) | repeated | Selectors for the set of pods targeted by the k8s Service. |
  | labels | [][discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.LabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.traffic_target#discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.LabelsEntry" >}}) | repeated | Labels on the underlying k8s Service itself. |
  | ports | [][discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.KubeServicePort]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.traffic_target#discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.KubeServicePort" >}}) | repeated | The ports exposed by the underlying service. |
  | subsets | [][discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.SubsetsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.traffic_target#discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.SubsetsEntry" >}}) | repeated | Subsets for routing, based on labels. |
  | endpointSubsets | [][discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.EndpointsSubset]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.traffic_target#discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.EndpointsSubset" >}}) | repeated | Each endpoints subset is a group of endpoint arranged in terms of ip/port pairs. This API mirrors the kubernetes Endpoints API. https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#endpoints-v1-core |
  





<a name="discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.EndpointsSubset"></a>

### TrafficTargetSpec.KubeService.EndpointsSubset
A series of IP addresses and their associated ports. The list of ip + port pairs is the cartesian product of the following lists


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| endpoints | [][discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.EndpointsSubset.Endpoint]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.traffic_target#discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.EndpointsSubset.Endpoint" >}}) | repeated |  |
  | ports | [][discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.KubeServicePort]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.traffic_target#discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.KubeServicePort" >}}) | repeated |  |
  





<a name="discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.EndpointsSubset.Endpoint"></a>

### TrafficTargetSpec.KubeService.EndpointsSubset.Endpoint
An endpoint exposed by the service


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ipAddress | string |  | IP address |
  | labels | [][discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.EndpointsSubset.Endpoint.LabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.traffic_target#discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.EndpointsSubset.Endpoint.LabelsEntry" >}}) | repeated | Labels which belong to this IP, these are taken from the backing workload instance. |
  





<a name="discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.EndpointsSubset.Endpoint.LabelsEntry"></a>

### TrafficTargetSpec.KubeService.EndpointsSubset.Endpoint.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.KubeServicePort"></a>

### TrafficTargetSpec.KubeService.KubeServicePort



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | uint32 |  | External-facing port for this k8s service (NOT the service's target port on the backing pods). |
  | name | string |  |  |
  | protocol | string |  |  |
  | appProtocol | string |  | Available in k8s 1.18+, specifies the application protocol. |
  





<a name="discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.LabelsEntry"></a>

### TrafficTargetSpec.KubeService.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.Subset"></a>

### TrafficTargetSpec.KubeService.Subset
Subsets for routing, based on labels.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| values | []string | repeated |  |
  





<a name="discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.SubsetsEntry"></a>

### TrafficTargetSpec.KubeService.SubsetsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.Subset]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.traffic_target#discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.Subset" >}}) |  |  |
  





<a name="discovery.mesh.gloo.solo.io.TrafficTargetSpec.KubeService.WorkloadSelectorLabelsEntry"></a>

### TrafficTargetSpec.KubeService.WorkloadSelectorLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="discovery.mesh.gloo.solo.io.TrafficTargetStatus"></a>

### TrafficTargetStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the TrafficPolicy metadata. if the observedGeneration does not match generation, the controller has not received the most recent version of this resource. |
  | appliedTrafficPolicies | [][discovery.mesh.gloo.solo.io.TrafficTargetStatus.AppliedTrafficPolicy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.traffic_target#discovery.mesh.gloo.solo.io.TrafficTargetStatus.AppliedTrafficPolicy" >}}) | repeated | The set of Traffic Policies that have been applied to this TrafficTarget |
  | appliedAccessPolicies | [][discovery.mesh.gloo.solo.io.TrafficTargetStatus.AppliedAccessPolicy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.traffic_target#discovery.mesh.gloo.solo.io.TrafficTargetStatus.AppliedAccessPolicy" >}}) | repeated | The set of Access Policies that have been applied to this TrafficTarget |
  | localFqdn | string |  | The local fully qualified domain |
  | appliedFederation | [discovery.mesh.gloo.solo.io.TrafficTargetStatus.AppliedFederation]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.traffic_target#discovery.mesh.gloo.solo.io.TrafficTargetStatus.AppliedFederation" >}}) |  | Federation metadata. Only populated if this traffic target is federated through a VirtualMesh. |
  





<a name="discovery.mesh.gloo.solo.io.TrafficTargetStatus.AppliedAccessPolicy"></a>

### TrafficTargetStatus.AppliedAccessPolicy
AppliedAccessPolicy represents a access policy that has been applied to the TrafficTarget. if an existing Access Policy becomes invalid, the last applied policy will be used


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | reference to the access policy |
  | observedGeneration | int64 |  | the observed generation of the accepted access policy |
  | spec | [networking.mesh.gloo.solo.io.AccessPolicySpec]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.access_policy#networking.mesh.gloo.solo.io.AccessPolicySpec" >}}) |  | the last known valid spec of the access policy |
  





<a name="discovery.mesh.gloo.solo.io.TrafficTargetStatus.AppliedFederation"></a>

### TrafficTargetStatus.AppliedFederation
Federation policy applied to this TrafficTarget, allowing access to the traffic target from other meshes/clusters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| federatedHostname | string |  | For any workload that this traffic target has federated to (i.e., any Workload controlled by a mesh whose ref appears in `federated_to_meshes`), a client in that workload will be able to reach this traffic target at this DNS name. This includes workloads on clusters other than the one hosting this service. |
  | federatedToMeshes | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated | The list of Meshes which are able to resolve this service's `multicluster_dns_name`. |
  | flatNetwork | bool |  | Whether or not the TrafficTarget has been federated via a flat-network to the given meshes. |
  





<a name="discovery.mesh.gloo.solo.io.TrafficTargetStatus.AppliedTrafficPolicy"></a>

### TrafficTargetStatus.AppliedTrafficPolicy
AppliedTrafficPolicy represents a traffic policy that has been applied to the TrafficTarget. if an existing Traffic Policy becomes invalid, the last applied policy will be used


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | reference to the traffic policy |
  | observedGeneration | int64 |  | the observed generation of the accepted traffic policy |
  | spec | [networking.mesh.gloo.solo.io.TrafficPolicySpec]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec" >}}) |  | the last known valid spec of the traffic policy |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

