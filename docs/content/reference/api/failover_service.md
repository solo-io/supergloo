
---
title: "failover_service.proto"
---

## Package : `networking.smh.solo.io`



<a name="top"></a>

<a name="API Reference for failover_service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## failover_service.proto


## Table of Contents
  - [FailoverServiceSpec](#networking.smh.solo.io.FailoverServiceSpec)
  - [FailoverServiceSpec.BackingService](#networking.smh.solo.io.FailoverServiceSpec.BackingService)
  - [FailoverServiceSpec.Port](#networking.smh.solo.io.FailoverServiceSpec.Port)
  - [FailoverServiceStatus](#networking.smh.solo.io.FailoverServiceStatus)
  - [FailoverServiceStatus.MeshesEntry](#networking.smh.solo.io.FailoverServiceStatus.MeshesEntry)







<a name="networking.smh.solo.io.FailoverServiceSpec"></a>

### FailoverServiceSpec
A FailoverService creates a new hostname to which services can send requests. Requests will be routed based on a list of backing services ordered by decreasing priority. When outlier detection detects that a service in the list is in an unhealthy state, requests sent to the FailoverService will be routed to the next healthy service in the list. For each service referenced in the failover services list, outlier detection must be configured using a TrafficPolicy.<br>Currently this feature only supports Services backed by Istio.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hostname | [string](#string) |  | The DNS name of the failover service. Must be unique within the service mesh instance since it is used as the hostname with which clients communicate. |
| port | [FailoverServiceSpec.Port](#networking.smh.solo.io.FailoverServiceSpec.Port) |  | The port on which the failover service listens. |
| meshes | [][core.skv2.solo.io.ObjectRef](#core.skv2.solo.io.ObjectRef) | repeated | The meshes that this failover service will be visible to. |
| backingServices | [][FailoverServiceSpec.BackingService](#networking.smh.solo.io.FailoverServiceSpec.BackingService) | repeated | The list of services backing the FailoverService, ordered by decreasing priority. All services must be backed by either the same service mesh instance or backed by service meshes that are grouped under a common VirtualMesh. |






<a name="networking.smh.solo.io.FailoverServiceSpec.BackingService"></a>

### FailoverServiceSpec.BackingService



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeService | [core.skv2.solo.io.ClusterObjectRef](#core.skv2.solo.io.ClusterObjectRef) |  | Name/namespace/cluster of a kubernetes service. |






<a name="networking.smh.solo.io.FailoverServiceSpec.Port"></a>

### FailoverServiceSpec.Port
The port on which the failover service listens.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| number | [uint32](#uint32) |  | Port number. |
| protocol | [string](#string) |  | Protocol of the requests sent to the failover service, must be one of HTTP, HTTPS, GRPC, HTTP2, MONGO, TCP, TLS. |






<a name="networking.smh.solo.io.FailoverServiceStatus"></a>

### FailoverServiceStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | [int64](#int64) |  | The most recent generation observed in the the FailoverService metadata. If the observedGeneration does not match generation, the controller has not received the most recent version of this resource. |
| state | [ApprovalState](#networking.smh.solo.io.ApprovalState) |  | The state of the overall resource, will only show accepted if it has been successfully applied to all target meshes. |
| meshes | [][FailoverServiceStatus.MeshesEntry](#networking.smh.solo.io.FailoverServiceStatus.MeshesEntry) | repeated | The status of the FailoverService for each Mesh to which it has been applied. |
| validationErrors | [][string](#string) | repeated | any errors observed which prevented the resource from being Accepted. |






<a name="networking.smh.solo.io.FailoverServiceStatus.MeshesEntry"></a>

### FailoverServiceStatus.MeshesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [ApprovalStatus](#networking.smh.solo.io.ApprovalStatus) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

