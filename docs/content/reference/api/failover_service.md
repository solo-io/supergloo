
---
title: "failover_service.proto"
---

## Package : `networking.smh.solo.io`



<a name="top"></a>

<a name="API Reference for failover_service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## failover_service.proto


## Table of Contents
  - [FailoverService](#networking.smh.solo.io.FailoverService)







<a name="networking.smh.solo.io.FailoverService"></a>

### FailoverService
A service composed of the referenced workloads with failover capabilities. The failover order is determined by the order of the declared workloads, i.e. an unhealthy workloads[0] will cause failover to workloads[1], etc.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hostname | [string](#string) |  | The DNS name of the failover service. |
| cluster | [string](#string) |  | The cluster that the failover service resides (the cluster name registered with Service Mesh Hub). |
| workloads | [][core.smh.solo.io.ResourceRef](#core.smh.solo.io.ResourceRef) | repeated | A list of workloads ordered by decreasing priority for failover. All workloads must be part of the same VirtualMesh. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

