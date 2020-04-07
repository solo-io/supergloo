
---
title: "status.proto"
---

## Package : `core.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for status.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## status.proto


## Table of Contents
  - [Status](#core.zephyr.solo.io.Status)

  - [Status.State](#core.zephyr.solo.io.Status.State)






<a name="core.zephyr.solo.io.Status"></a>

### Status
Status set by Service Mesh Hub while processing a resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| state | [Status.State](#core.zephyr.solo.io.Status.State) |  |  |
| message | [string](#string) |  | Human-readable message with details concerning the reason this state was reached. |





 <!-- end messages -->


<a name="core.zephyr.solo.io.Status.State"></a>

### Status.State


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

