
---
title: "core.proto"
---

## Package : `core.skv2.solo.io`



<a name="top"></a>

<a name="API Reference for core.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## core.proto


## Table of Contents
  - [ClusterObjectRef](#core.skv2.solo.io.ClusterObjectRef)
  - [ObjectRef](#core.skv2.solo.io.ObjectRef)
  - [Status](#core.skv2.solo.io.Status)
  - [TypedClusterObjectRef](#core.skv2.solo.io.TypedClusterObjectRef)
  - [TypedObjectRef](#core.skv2.solo.io.TypedObjectRef)

  - [Status.State](#core.skv2.solo.io.Status.State)






<a name="core.skv2.solo.io.ClusterObjectRef"></a>

### ClusterObjectRef
Resource reference for a cross-cluster-scoped object


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | name of the resource being referenced |
  | namespace | string |  | namespace of the resource being referenced |
  | clusterName | string |  | name of the cluster in which the resource exists |
  





<a name="core.skv2.solo.io.ObjectRef"></a>

### ObjectRef
Resource reference for an object


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | name of the resource being referenced |
  | namespace | string |  | namespace of the resource being referenced |
  





<a name="core.skv2.solo.io.Status"></a>

### Status
A generic status


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| state | [core.skv2.solo.io.Status.State]({{< ref "core.md#core.skv2.solo.io.Status.State" >}}) |  | The current state of the resource |
  | message | string |  | A human readable message about the current state of the object |
  | observedGeneration | int64 |  | The most recently observed generation of the resource. This value corresponds to the `metadata.generation` of a kubernetes resource |
  | processingTime | [google.protobuf.Timestamp]({{< ref "timestamp.md#google.protobuf.Timestamp" >}}) |  | The time at which this status was recorded |
  | owner | [google.protobuf.StringValue]({{< ref "wrappers.md#google.protobuf.StringValue" >}}) |  | (optional) The owner of the status, this value can be used to identify the entity which wrote this status. This is useful in situations where a given resource may have multiple owners. |
  





<a name="core.skv2.solo.io.TypedClusterObjectRef"></a>

### TypedClusterObjectRef
Resource reference for a typed, cross-cluster-scoped object


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| apiGroup | [google.protobuf.StringValue]({{< ref "wrappers.md#google.protobuf.StringValue" >}}) |  | API group of the resource being referenced |
  | kind | [google.protobuf.StringValue]({{< ref "wrappers.md#google.protobuf.StringValue" >}}) |  | Kind of the resource being referenced |
  | name | string |  | name of the resource being referenced |
  | namespace | string |  | namespace of the resource being referenced |
  | clusterName | string |  | name of the cluster in which the resource exists |
  





<a name="core.skv2.solo.io.TypedObjectRef"></a>

### TypedObjectRef
Resource reference for a typed object


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| apiGroup | [google.protobuf.StringValue]({{< ref "wrappers.md#google.protobuf.StringValue" >}}) |  | API group of the resource being referenced |
  | kind | [google.protobuf.StringValue]({{< ref "wrappers.md#google.protobuf.StringValue" >}}) |  | Kind of the resource being referenced |
  | name | string |  | name of the resource being referenced |
  | namespace | string |  | namespace of the resource being referenced |
  




 <!-- end messages -->


<a name="core.skv2.solo.io.Status.State"></a>

### Status.State
The State of a reconciled object

| Name | Number | Description |
| ---- | ------ | ----------- |
| PENDING | 0 | Waiting to be processed. |
| PROCESSING | 1 | Currently processing. |
| INVALID | 2 | Invalid parameters supplied, will not continue. |
| FAILED | 3 | Failed during processing. |
| ACCEPTED | 4 | Finished processing successfully. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

