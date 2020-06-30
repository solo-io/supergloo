
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
A service composed of the referenced workloads with failover capabilities. The failover order is determined by the order of the declared workloads, i.e. an unhealthy workloads[0] will cause failover to workloads[1], etc. Currently this feature only supports Services backed by Istio.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hostname | [string](#string) |  | The DNS name of the failover service. |
| namespace | [string](#string) |  | The namespace to locate the translated service. |
| port | [FailoverServiceSpec.Port](#networking.smh.solo.io.FailoverServiceSpec.Port) |  | The ports from which to expose this service. |
| cluster | [string](#string) |  | The cluster that the failover service resides (the cluster name registered with Service Mesh Hub). |
| failoverServices | [][core.smh.solo.io.ResourceRef](#core.smh.solo.io.ResourceRef) | repeated | A list of services ordered by decreasing priority for failover. All services must be controlled by service meshes that are grouped under a common VirtualMesh. |






<a name="networking.smh.solo.io.FailoverServiceSpec.Port"></a>

### FailoverServiceSpec.Port



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | [uint32](#uint32) |  |  |
| name | [string](#string) |  |  |
| protocol | [string](#string) |  |  |






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



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| translatorId | [string](#string) |  | ID representing a translator that translates FailoverService to Mesh-specific config. |
| errorMessage | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

