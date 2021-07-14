
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
  - [DestinationSpec.ExternalService](#discovery.mesh.gloo.solo.io.DestinationSpec.ExternalService)
  - [DestinationSpec.ExternalService.ExternalEndpoint](#discovery.mesh.gloo.solo.io.DestinationSpec.ExternalService.ExternalEndpoint)
  - [DestinationSpec.ExternalService.ExternalEndpoint.PortsEntry](#discovery.mesh.gloo.solo.io.DestinationSpec.ExternalService.ExternalEndpoint.PortsEntry)
  - [DestinationSpec.ExternalService.ServicePort](#discovery.mesh.gloo.solo.io.DestinationSpec.ExternalService.ServicePort)
  - [DestinationSpec.KubeService](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService)
  - [DestinationSpec.KubeService.EndpointPort](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointPort)
  - [DestinationSpec.KubeService.EndpointsSubset](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset)
  - [DestinationSpec.KubeService.EndpointsSubset.Endpoint](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset.Endpoint)
  - [DestinationSpec.KubeService.EndpointsSubset.Endpoint.LabelsEntry](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset.Endpoint.LabelsEntry)
  - [DestinationSpec.KubeService.EndpointsSubset.Endpoint.SubLocality](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset.Endpoint.SubLocality)
  - [DestinationSpec.KubeService.ExternalAddress](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.ExternalAddress)
  - [DestinationSpec.KubeService.KubeServicePort](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.KubeServicePort)
  - [DestinationSpec.KubeService.LabelsEntry](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.LabelsEntry)
  - [DestinationSpec.KubeService.Subset](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.Subset)
  - [DestinationSpec.KubeService.SubsetsEntry](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.SubsetsEntry)
  - [DestinationSpec.KubeService.WorkloadSelectorLabelsEntry](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.WorkloadSelectorLabelsEntry)
  - [DestinationStatus](#discovery.mesh.gloo.solo.io.DestinationStatus)
  - [DestinationStatus.AppliedAccessPolicy](#discovery.mesh.gloo.solo.io.DestinationStatus.AppliedAccessPolicy)
  - [DestinationStatus.AppliedFederation](#discovery.mesh.gloo.solo.io.DestinationStatus.AppliedFederation)
  - [RequiredSubsets](#discovery.mesh.gloo.solo.io.RequiredSubsets)

  - [DestinationSpec.KubeService.ServiceType](#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.ServiceType)






<a name="discovery.mesh.gloo.solo.io.DestinationSpec"></a>

### DestinationSpec
The Destination is an abstraction for any entity capable of receiving networking requests.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeService | [discovery.mesh.gloo.solo.io.DestinationSpec.KubeService]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService" >}}) |  | KubeService is a kube-native Destination representing a kubernetes service running inside of a kubernetes cluster. |
  | externalService | [discovery.mesh.gloo.solo.io.DestinationSpec.ExternalService]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationSpec.ExternalService" >}}) |  | ExternalService is a Destination representing a service external to the Mesh. It can be used to expose a given hostname or IP address to all clusters in the Virtual Mesh. |
  | mesh | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | The mesh that controls this Destination. Can be omitted if the Destination isn't associated with any particular mesh, eg for External Services. |
  





<a name="discovery.mesh.gloo.solo.io.DestinationSpec.ExternalService"></a>

### DestinationSpec.ExternalService
Describes a service external to the mesh


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The name of the service |
  | hosts | []string | repeated | The list of hosts which will resolve to this Destination for services within the Virtual Mesh. |
  | addresses | []string | repeated | The List of addresses which will resolve to this service for services within the Virtual Mesh. |
  | ports | [][discovery.mesh.gloo.solo.io.DestinationSpec.ExternalService.ServicePort]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationSpec.ExternalService.ServicePort" >}}) | repeated | The associated ports of the external service |
  | endpoints | [][discovery.mesh.gloo.solo.io.DestinationSpec.ExternalService.ExternalEndpoint]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationSpec.ExternalService.ExternalEndpoint" >}}) | repeated | List of endpoints, to which any requests to this Destionation will be load balanced across. |
  





<a name="discovery.mesh.gloo.solo.io.DestinationSpec.ExternalService.ExternalEndpoint"></a>

### DestinationSpec.ExternalService.ExternalEndpoint
ExternalEndpoint represents the address/port(s) of the external service which will receive requests sent to this Destination.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| address | string |  | The address of the external service. Can be a domain or an IP. |
  | ports | [][discovery.mesh.gloo.solo.io.DestinationSpec.ExternalService.ExternalEndpoint.PortsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationSpec.ExternalService.ExternalEndpoint.PortsEntry" >}}) | repeated | The port(s) of the external endpoint. Eg: `https: 443` |
  





<a name="discovery.mesh.gloo.solo.io.DestinationSpec.ExternalService.ExternalEndpoint.PortsEntry"></a>

### DestinationSpec.ExternalService.ExternalEndpoint.PortsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | uint32 |  |  |
  





<a name="discovery.mesh.gloo.solo.io.DestinationSpec.ExternalService.ServicePort"></a>

### DestinationSpec.ExternalService.ServicePort
ServicePort describes a port accessible on this Destination


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| number | uint32 |  | The port number. Must be a valid, non-negative integer port number. |
  | name | string |  | A label for the port, eg "http" |
  | protocol | string |  | The protocol used in communications with this Destination MUST BE one of HTTP|HTTPS|GRPC|HTTP2|MONGO|TCP|TLS. |
  





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
  | externalAddresses | [][discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.ExternalAddress]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.ExternalAddress" >}}) | repeated | Describes the address data for Kubernetes Services exposed to external traffic (i.e. for non ClusterIP type Services). |
  | serviceType | [discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.ServiceType]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.ServiceType" >}}) |  | Describes the Kubernetes Service type. |
  





<a name="discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointPort"></a>

### DestinationSpec.KubeService.EndpointPort
Describes the endpoints's ports. See [here](https://kubernetes.io/docs/reference/kubernetes-api/service-resources/endpoints-v1/) for more information.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | uint32 |  | Port on which the endpoints can be reached |
  | name | string |  | Name of the port |
  | protocol | string |  | Protocol on which this port serves traffic (HTTP, TCP, UDP, etc...) |
  | appProtocol | string |  | Available in Kubernetes 1.18+, describes the application protocol. |
  





<a name="discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset"></a>

### DestinationSpec.KubeService.EndpointsSubset
A series of IP addresses and their associated ports. The list of IP and port pairs is the cartesian product of the endpoint and port lists.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| endpoints | [][discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset.Endpoint]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointsSubset.Endpoint" >}}) | repeated |  |
  | ports | [][discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointPort]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.EndpointPort" >}}) | repeated |  |
  





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
  





<a name="discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.ExternalAddress"></a>

### DestinationSpec.KubeService.ExternalAddress
Describes the address data for Kubernetes Services exposed to external traffic (i.e. for non ClusterIP type Services).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| dnsName | string |  | An externally accessible DNS name. |
  | ip | string |  | An externally accessible IP address. |
  





<a name="discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.KubeServicePort"></a>

### DestinationSpec.KubeService.KubeServicePort
Describes the service's ports. See [here](https://kubernetes.io/docs/concepts/services-networking/service/#multi-port-services) for more information.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | uint32 |  | External-facing port for this Kubernetes service (*not* the service's target port on the targeted pods). |
  | name | string |  | Name of the port |
  | protocol | string |  | Protocol on which this port serves traffic (HTTP, TCP, UDP, etc...) |
  | appProtocol | string |  | Available in Kubernetes 1.18+, describes the application protocol. |
  | targetPortName | string |  | Name of the target port |
  | targetPortNumber | uint32 |  | Number of the target port |
  | nodePort | uint32 |  | Populated for NodePort or LoadBalancer Services. |
  





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
  | appliedTrafficPolicies | [][networking.mesh.gloo.solo.io.AppliedTrafficPolicy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.applied_policies#networking.mesh.gloo.solo.io.AppliedTrafficPolicy" >}}) | repeated | The set of TrafficPolicies that have been applied to this Destination. {{/* Note: validation of this field disabled because it slows down cue tremendously*/}} |
  | appliedAccessPolicies | [][discovery.mesh.gloo.solo.io.DestinationStatus.AppliedAccessPolicy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationStatus.AppliedAccessPolicy" >}}) | repeated | The set of AccessPolicies that have been applied to this Destination. |
  | localFqdn | string |  | The fully qualified domain name for requests originating from a source *coloated* with this Destination. For Kubernetes services, "colocated" means within the same Kubernetes cluster. |
  | appliedFederation | [discovery.mesh.gloo.solo.io.DestinationStatus.AppliedFederation]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.DestinationStatus.AppliedFederation" >}}) |  | Federation metadata. Only populated if this Destination is federated through a VirtualMesh. |
  | requiredSubsets | [][discovery.mesh.gloo.solo.io.RequiredSubsets]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.destination#discovery.mesh.gloo.solo.io.RequiredSubsets" >}}) | repeated | The set of TrafficPolicy traffic shifts that reference subsets on this Destination. |
  





<a name="discovery.mesh.gloo.solo.io.DestinationStatus.AppliedAccessPolicy"></a>

### DestinationStatus.AppliedAccessPolicy
Describes an [AccessPolicy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.access_policy" >}}) that applies to this Destination. If an existing AccessPolicy becomes invalid, the last valid applied policy will be used.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | Reference to the AccessPolicy object. |
  | observedGeneration | int64 |  | The observed generation of the accepted AccessPolicy. |
  | spec | [networking.mesh.gloo.solo.io.AccessPolicySpec]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.access_policy#networking.mesh.gloo.solo.io.AccessPolicySpec" >}}) |  | The spec of the last known valid AccessPolicy. |
  





<a name="discovery.mesh.gloo.solo.io.DestinationStatus.AppliedFederation"></a>

### DestinationStatus.AppliedFederation
Describes the federation configuration applied to this Destination through a [VirtualMesh]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.virtual_mesh" >}}). Federation allows access to the Destination from other meshes/clusters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| federatedHostname | string |  | For any Workload that this Destination has been federated to (i.e., any Workload controlled by a Mesh whose reference appears in `federated_to_meshes`), that Workload will be able to reach this Destination using this DNS name. For Kubernetes Destinations this includes Workloads on clusters other than the one hosting this Destination. |
  | federatedToMeshes | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated | The list of Meshes which are able to resolve this Destination's `federated_hostname`. |
  | flatNetwork | bool |  | Whether the Destination has been federated to the given meshes using a VirtualMesh where [Federation.FlatNetwork]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.virtual_mesh/#virtualmeshspecfederation" >}}) is true. |
  | virtualMeshRef | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | Reference to the VirtualMesh object. |
  | tcpKeepalive | [common.mesh.gloo.solo.io.TCPKeepalive]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.keepalive#common.mesh.gloo.solo.io.TCPKeepalive" >}}) |  | Specify a keepalive rule for all requests made within the VirtualMesh which cross clusters within that VirtualMesh, as well as any requests to externalService type destinations. |
  





<a name="discovery.mesh.gloo.solo.io.RequiredSubsets"></a>

### RequiredSubsets
Describes a [TrafficPolicy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy" >}}) that references subsets on this Destination in a traffic shift. Note: this is an Istio-specific feature.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| trafficPolicyRef | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | Reference to the TrafficPolicy object. |
  | observedGeneration | int64 |  | The observed generation of the TrafficPolicy. |
  | trafficShift | [networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.MultiDestination]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy.MultiDestination" >}}) |  | The applied traffic shift. |
  




 <!-- end messages -->


<a name="discovery.mesh.gloo.solo.io.DestinationSpec.KubeService.ServiceType"></a>

### DestinationSpec.KubeService.ServiceType
Describes the Kubernetes Service type.

| Name | Number | Description |
| ---- | ------ | ----------- |
| CLUSTER_IP | 0 | A ClusterIP Service. |
| NODE_PORT | 1 | A NodePort Service. |
| LOAD_BALANCER | 2 | A LoadBalancer Service. |
| EXTERNAL_NAME | 3 | An ExternalName Service. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

