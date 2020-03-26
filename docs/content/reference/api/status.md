
---
title: "core.zephyr.solo.iogithub.com/solo-io/mesh-projects/api/core/v1alpha1/status.proto"
---

## Package : `core.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for github.com/solo-io/mesh-projects/api/core/v1alpha1/status.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## github.com/solo-io/mesh-projects/api/core/v1alpha1/status.proto


## Table of Contents
  - [ComputedStatus](#core.zephyr.solo.io.ComputedStatus)

  - [ComputedStatus.Status](#core.zephyr.solo.io.ComputedStatus.Status)






<a name="core.zephyr.solo.io.ComputedStatus"></a>

### ComputedStatus
a status set by Service Mesh Hub while processing a resource


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| status | [ComputedStatus.Status](#core.zephyr.solo.io.ComputedStatus.Status) |  |  |
| message | [string](#string) |  | human-readable message to be surfaced to the user |





 <!-- end messages -->


<a name="core.zephyr.solo.io.ComputedStatus.Status"></a>

### ComputedStatus.Status


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN | 0 |  |
| ACCEPTED | 1 |  |
| INVALID | 2 |  |
| PROCESSING_ERROR | 3 |  |
| CONFLICT | 4 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

