
---
title: "workload_entry.proto"
---

## Package : `istio.networking.v1alpha3`



<a name="top"></a>

<a name="API Reference for workload_entry.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## workload_entry.proto


## Table of Contents
  - [WorkloadEntry](#istio.networking.v1alpha3.WorkloadEntry)
  - [WorkloadEntry.LabelsEntry](#istio.networking.v1alpha3.WorkloadEntry.LabelsEntry)
  - [WorkloadEntry.PortsEntry](#istio.networking.v1alpha3.WorkloadEntry.PortsEntry)







<a name="istio.networking.v1alpha3.WorkloadEntry"></a>

### WorkloadEntry
WorkloadEntry enables specifying the properties of a single non-Kubernetes workload such a VM or a bare metal services that can be referred to by service entries.<br><!-- crd generation tags +cue-gen:WorkloadEntry:groupName:networking.istio.io +cue-gen:WorkloadEntry:version:v1alpha3 +cue-gen:WorkloadEntry:storageVersion +cue-gen:WorkloadEntry:annotations:helm.sh/resource-policy=keep +cue-gen:WorkloadEntry:labels:app=istio-pilot,chart=istio,heritage=Tiller,release=istio +cue-gen:WorkloadEntry:subresource:status +cue-gen:WorkloadEntry:scope:Namespaced +cue-gen:WorkloadEntry:resource:categories=istio-io,networking-istio-io,shortNames=we,plural=workloadentries +cue-gen:WorkloadEntry:printerColumn:name=Age,type=date,JSONPath=.metadata.creationTimestamp,description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC. Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata" +cue-gen:WorkloadEntry:printerColumn:name=Address,type=string,JSONPath=.spec.address,description="Address associated with the network endpoint." --><br><!-- go code generation tags +kubetype-gen +kubetype-gen:groupVersion=networking.istio.io/v1alpha3 +genclient +k8s:deepcopy-gen=true -->


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| address | string |  | Address associated with the network endpoint without the port.  Domain names can be used if and only if the resolution is set to DNS, and must be fully-qualified without wildcards. Use the form unix:///absolute/path/to/socket for Unix domain socket endpoints. |
  | ports | [][istio.networking.v1alpha3.WorkloadEntry.PortsEntry]({{< ref "workload_entry.md#istio.networking.v1alpha3.WorkloadEntry.PortsEntry" >}}) | repeated | Set of ports associated with the endpoint. If the port map is specified, it must be a map of servicePortName to this endpoint's port, such that traffic to the service port will be forwarded to the endpoint port that maps to the service's portName. If omitted, and the targetPort is specified as part of the service's port specification, traffic to the service port will be forwarded to one of the endpoints on the specified `targetPort`. If both the targetPort and endpoint's port map are not specified, traffic to a service port will be forwarded to one of the endpoints on the same port.<br>**NOTE 1:** Do not use for `unix://` addresses.<br>**NOTE 2:** endpoint port map takes precedence over targetPort. |
  | labels | [][istio.networking.v1alpha3.WorkloadEntry.LabelsEntry]({{< ref "workload_entry.md#istio.networking.v1alpha3.WorkloadEntry.LabelsEntry" >}}) | repeated | One or more labels associated with the endpoint. |
  | network | string |  | Network enables Istio to group endpoints resident in the same L3 domain/network. All endpoints in the same network are assumed to be directly reachable from one another. When endpoints in different networks cannot reach each other directly, an Istio Gateway can be used to establish connectivity (usually using the `AUTO_PASSTHROUGH` mode in a Gateway Server). This is an advanced configuration used typically for spanning an Istio mesh over multiple clusters. |
  | locality | string |  | The locality associated with the endpoint. A locality corresponds to a failure domain (e.g., country/region/zone). Arbitrary failure domain hierarchies can be represented by separating each encapsulating failure domain by /. For example, the locality of an an endpoint in US, in US-East-1 region, within availability zone az-1, in data center rack r11 can be represented as us/us-east-1/az-1/r11. Istio will configure the sidecar to route to endpoints within the same locality as the sidecar. If none of the endpoints in the locality are available, endpoints parent locality (but within the same network ID) will be chosen. For example, if there are two endpoints in same network (networkID "n1"), say e1 with locality us/us-east-1/az-1/r11 and e2 with locality us/us-east-1/az-2/r12, a sidecar from us/us-east-1/az-1/r11 locality will prefer e1 from the same locality over e2 from a different locality. Endpoint e2 could be the IP associated with a gateway (that bridges networks n1 and n2), or the IP associated with a standard service endpoint. |
  | weight | uint32 |  | The load balancing weight associated with the endpoint. Endpoints with higher weights will receive proportionally higher traffic. |
  | serviceAccount | string |  | The service account associated with the workload if a sidecar is present in the workload. The service account must be present in the same namespace as the configuration ( WorkloadEntry or a ServiceEntry) |
  





<a name="istio.networking.v1alpha3.WorkloadEntry.LabelsEntry"></a>

### WorkloadEntry.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="istio.networking.v1alpha3.WorkloadEntry.PortsEntry"></a>

### WorkloadEntry.PortsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | uint32 |  |  |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

