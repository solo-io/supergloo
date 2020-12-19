
---
title: "service_entry.proto"
---

## Package : `istio.networking.v1alpha3`



<a name="top"></a>

<a name="API Reference for service_entry.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## service_entry.proto


## Table of Contents
  - [ServiceEntry](#istio.networking.v1alpha3.ServiceEntry)

  - [ServiceEntry.Location](#istio.networking.v1alpha3.ServiceEntry.Location)
  - [ServiceEntry.Resolution](#istio.networking.v1alpha3.ServiceEntry.Resolution)






<a name="istio.networking.v1alpha3.ServiceEntry"></a>

### ServiceEntry
ServiceEntry enables adding additional entries into Istio's internal service registry.<br><!-- crd generation tags +cue-gen:ServiceEntry:groupName:networking.istio.io +cue-gen:ServiceEntry:version:v1alpha3 +cue-gen:ServiceEntry:storageVersion +cue-gen:ServiceEntry:annotations:helm.sh/resource-policy=keep +cue-gen:ServiceEntry:labels:app=istio-pilot,chart=istio,heritage=Tiller,release=istio +cue-gen:ServiceEntry:subresource:status +cue-gen:ServiceEntry:scope:Namespaced +cue-gen:ServiceEntry:resource:categories=istio-io,networking-istio-io,shortNames=se,plural=serviceentries +cue-gen:ServiceEntry:printerColumn:name=Hosts,type=string,JSONPath=.spec.hosts,description="The hosts associated with the ServiceEntry" +cue-gen:ServiceEntry:printerColumn:name=Location,type=string,JSONPath=.spec.location,description="Whether the service is external to the mesh or part of the mesh (MESH_EXTERNAL or MESH_INTERNAL)" +cue-gen:ServiceEntry:printerColumn:name=Resolution,type=string,JSONPath=.spec.resolution,description="Service discovery mode for the hosts (NONE, STATIC, or DNS)" +cue-gen:ServiceEntry:printerColumn:name=Age,type=date,JSONPath=.metadata.creationTimestamp,description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC. Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata" --><br><!-- go code generation tags +kubetype-gen +kubetype-gen:groupVersion=networking.istio.io/v1alpha3 +genclient +k8s:deepcopy-gen=true -->


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hosts | []string | repeated | The hosts associated with the ServiceEntry. Could be a DNS name with wildcard prefix.<br>1. The hosts field is used to select matching hosts in VirtualServices and DestinationRules. 2. For HTTP traffic the HTTP Host/Authority header will be matched against the hosts field. 3. For HTTPs or TLS traffic containing Server Name Indication (SNI), the SNI value will be matched against the hosts field.<br>**NOTE 1:** When resolution is set to type DNS and no endpoints are specified, the host field will be used as the DNS name of the endpoint to route traffic to.<br>**NOTE 2:** If the hostname matches with the name of a service from another service registry such as Kubernetes that also supplies its own set of endpoints, the ServiceEntry will be treated as a decorator of the existing Kubernetes service. Properties in the service entry will be added to the Kubernetes service if applicable. Currently, the only the following additional properties will be considered by `istiod`:<br>1. subjectAltNames: In addition to verifying the SANs of the    service accounts associated with the pods of the service, the    SANs specified here will also be verified. |
  | addresses | []string | repeated | The virtual IP addresses associated with the service. Could be CIDR prefix. For HTTP traffic, generated route configurations will include http route domains for both the `addresses` and `hosts` field values and the destination will be identified based on the HTTP Host/Authority header. If one or more IP addresses are specified, the incoming traffic will be identified as belonging to this service if the destination IP matches the IP/CIDRs specified in the addresses field. If the Addresses field is empty, traffic will be identified solely based on the destination port. In such scenarios, the port on which the service is being accessed must not be shared by any other service in the mesh. In other words, the sidecar will behave as a simple TCP proxy, forwarding incoming traffic on a specified port to the specified destination endpoint IP/host. Unix domain socket addresses are not supported in this field. |
  | ports | [][istio.networking.v1alpha3.Port]({{< ref "gateway.md#istio.networking.v1alpha3.Port" >}}) | repeated | The ports associated with the external service. If the Endpoints are Unix domain socket addresses, there must be exactly one port. |
  | location | [istio.networking.v1alpha3.ServiceEntry.Location]({{< ref "service_entry.md#istio.networking.v1alpha3.ServiceEntry.Location" >}}) |  | Specify whether the service should be considered external to the mesh or part of the mesh. |
  | resolution | [istio.networking.v1alpha3.ServiceEntry.Resolution]({{< ref "service_entry.md#istio.networking.v1alpha3.ServiceEntry.Resolution" >}}) |  | Service discovery mode for the hosts. Care must be taken when setting the resolution mode to NONE for a TCP port without accompanying IP addresses. In such cases, traffic to any IP on said port will be allowed (i.e. `0.0.0.0:<port>`). |
  | endpoints | [][istio.networking.v1alpha3.WorkloadEntry]({{< ref "workload_entry.md#istio.networking.v1alpha3.WorkloadEntry" >}}) | repeated | One or more endpoints associated with the service. Only one of `endpoints` or `workloadSelector` can be specified. |
  | workloadSelector | [istio.networking.v1alpha3.WorkloadSelector]({{< ref "sidecar.md#istio.networking.v1alpha3.WorkloadSelector" >}}) |  | Applicable only for MESH_INTERNAL services. Only one of `endpoints` or `workloadSelector` can be specified. Selects one or more Kubernetes pods or VM workloads (specified using `WorkloadEntry`) based on their labels. The `WorkloadEntry` object representing the VMs should be defined in the same namespace as the ServiceEntry. |
  | exportTo | []string | repeated | A list of namespaces to which this service is exported. Exporting a service allows it to be used by sidecars, gateways and virtual services defined in other namespaces. This feature provides a mechanism for service owners and mesh administrators to control the visibility of services across namespace boundaries.<br>If no namespaces are specified then the service is exported to all namespaces by default.<br>The value "." is reserved and defines an export to the same namespace that the service is declared in. Similarly the value "*" is reserved and defines an export to all namespaces.<br>For a Kubernetes Service, the equivalent effect can be achieved by setting the annotation "networking.istio.io/exportTo" to a comma-separated list of namespace names.<br>NOTE: in the current release, the `exportTo` value is restricted to "." or "*" (i.e., the current namespace or all namespaces). |
  | subjectAltNames | []string | repeated | If specified, the proxy will verify that the server certificate's subject alternate name matches one of the specified values.<br>NOTE: When using the workloadEntry with workloadSelectors, the service account specified in the workloadEntry will also be used to derive the additional subject alternate names that should be verified. |
  




 <!-- end messages -->


<a name="istio.networking.v1alpha3.ServiceEntry.Location"></a>

### ServiceEntry.Location
Location specifies whether the service is part of Istio mesh or outside the mesh.  Location determines the behavior of several features, such as service-to-service mTLS authentication, policy enforcement, etc. When communicating with services outside the mesh, Istio's mTLS authentication is disabled, and policy enforcement is performed on the client-side as opposed to server-side.

| Name | Number | Description |
| ---- | ------ | ----------- |
| MESH_EXTERNAL | 0 | Signifies that the service is external to the mesh. Typically used to indicate external services consumed through APIs. |
| MESH_INTERNAL | 1 | Signifies that the service is part of the mesh. Typically used to indicate services added explicitly as part of expanding the service mesh to include unmanaged infrastructure (e.g., VMs added to a Kubernetes based service mesh). |



<a name="istio.networking.v1alpha3.ServiceEntry.Resolution"></a>

### ServiceEntry.Resolution
Resolution determines how the proxy will resolve the IP addresses of the network endpoints associated with the service, so that it can route to one of them. The resolution mode specified here has no impact on how the application resolves the IP address associated with the service. The application may still have to use DNS to resolve the service to an IP so that the outbound traffic can be captured by the Proxy. Alternatively, for HTTP services, the application could directly communicate with the proxy (e.g., by setting HTTP_PROXY) to talk to these services.

| Name | Number | Description |
| ---- | ------ | ----------- |
| NONE | 0 | Assume that incoming connections have already been resolved (to a specific destination IP address). Such connections are typically routed via the proxy using mechanisms such as IP table REDIRECT/ eBPF. After performing any routing related transformations, the proxy will forward the connection to the IP address to which the connection was bound. |
| STATIC | 1 | Use the static IP addresses specified in endpoints (see below) as the backing instances associated with the service. |
| DNS | 2 | Attempt to resolve the IP address by querying the ambient DNS, during request processing. If no endpoints are specified, the proxy will resolve the DNS address specified in the hosts field, if wildcards are not used. If endpoints are specified, the DNS addresses specified in the endpoints will be resolved to determine the destination IP address.  DNS resolution cannot be used with Unix domain socket endpoints. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

