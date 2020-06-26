
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
  - [FailoverServiceStatus](#networking.smh.solo.io.FailoverServiceStatus)
  - [FailoverServiceStatus.TranslatorError](#networking.smh.solo.io.FailoverServiceStatus.TranslatorError)







<a name="networking.smh.solo.io.FailoverServiceSpec"></a>

### FailoverServiceSpec
This configures an existing service with failover functionality, where in the case of an unhealthy service, requests will be shifted over to other services in priority order defined in the list of failover services, i.e. an unhealthy target_service will cause failover to workloads[0], etc.<br>Currently this feature only supports Services backed by Istio.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| targetService | [core.smh.solo.io.ResourceRef](#core.smh.solo.io.ResourceRef) |  | The service for which to add failover functionality. |
| failoverServices | [][core.smh.solo.io.ResourceRef](#core.smh.solo.io.ResourceRef) | repeated | A list of services ordered by decreasing priority for failover. All services must be controlled by service meshes that are grouped under a common VirtualMesh. |






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

