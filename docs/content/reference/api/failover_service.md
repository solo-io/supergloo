
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
  - [FailoverServiceSpec.Port](#networking.smh.solo.io.FailoverServiceSpec.Port)
  - [FailoverServiceStatus](#networking.smh.solo.io.FailoverServiceStatus)
  - [FailoverServiceStatus.TranslatorError](#networking.smh.solo.io.FailoverServiceStatus.TranslatorError)







<a name="networking.smh.solo.io.FailoverServiceSpec"></a>

### FailoverServiceSpec
A FailoverService creates a new hostname to which services can send requests. Requests will be routed based on a list of backing services ordered by decreasing priority. When outlier detection detects that a service in the list is in an unhealthy state, requests sent to the FailoverService will be routed to the next healthy service in the list. For each service referenced in the failover services list, outlier detection must be configured using a TrafficPolicy.<br>Currently this feature only supports Services backed by Istio.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hostname | [string](#string) |  | The DNS name of the failover service. Must be unique within the service mesh instance since it is used as the hostname with which clients communicate. |
| namespace | [string](#string) |  | The namespace to locate the translated service. |
| port | [FailoverServiceSpec.Port](#networking.smh.solo.io.FailoverServiceSpec.Port) |  | The ports from which to expose this service. |
| meshes | [][core.skv2.solo.io.ObjectRef](#core.skv2.solo.io.ObjectRef) | repeated | The meshes that this failover service will be visible to. |
| failoverServices | [][core.skv2.solo.io.ClusterObjectRef](#core.skv2.solo.io.ClusterObjectRef) | repeated | A list of services ordered by decreasing priority for failover. All services must be backed by either the same service mesh instance or backed by service meshes that are grouped under a common VirtualMesh. |






<a name="networking.smh.solo.io.FailoverServiceSpec.Port"></a>

### FailoverServiceSpec.Port
The port from which to expose the failover service.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | [uint32](#uint32) |  | Port number. |
| name | [string](#string) |  | Label for the port. |
| protocol | [string](#string) |  | Protocol. |






<a name="networking.smh.solo.io.FailoverServiceStatus"></a>

### FailoverServiceStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | [int64](#int64) |  | The generation the validation_status was observed on. |
| translationStatus | [core.smh.solo.io.Status](#core.smh.solo.io.Status) |  | Whether or not the resource has been successfully translated into concrete, mesh-specific routing configuration. |
| translatorErrors | [][FailoverServiceStatus.TranslatorError](#networking.smh.solo.io.FailoverServiceStatus.TranslatorError) | repeated | Provides details on any translation errors that occurred. If any errors exist, this FailoverService has not been translated into mesh-specific config. |
| validationStatus | [core.smh.solo.io.Status](#core.smh.solo.io.Status) |  | Whether or not this resource has passed validation. This is a required step before it can be translated into concrete, mesh-specific failover configuration. |






<a name="networking.smh.solo.io.FailoverServiceStatus.TranslatorError"></a>

### FailoverServiceStatus.TranslatorError
An error pertaining to translation of the FailoverService to mesh-specific configuration.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| translatorId | [string](#string) |  | ID representing a translator that translates FailoverService to Mesh-specific config. |
| errorMessage | [string](#string) |  | Message describing the error(s). |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

