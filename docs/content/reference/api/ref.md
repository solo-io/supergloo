
---
title: "core.zephyr.solo.iogithub.com/solo-io/mesh-projects/api/core/v1alpha1/ref.proto"
---

## Package : `core.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for github.com/solo-io/mesh-projects/api/core/v1alpha1/ref.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## github.com/solo-io/mesh-projects/api/core/v1alpha1/ref.proto


## Table of Contents
  - [ResourceRef](#core.zephyr.solo.io.ResourceRef)







<a name="core.zephyr.solo.io.ResourceRef"></a>

### ResourceRef
reference object for kubernetes objects, support multi cluster


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| aPIGroup | [google.protobuf.StringValue](#google.protobuf.StringValue) |  | The kubernetes kind of the object Up to the user to decide whether or not to use this field, can be implicit or explicit depending on the use case. |
| kind | [google.protobuf.StringValue](#google.protobuf.StringValue) |  |  |
| name | [string](#string) |  |  |
| namespace | [string](#string) |  | The namespace in which the object lives if left empty it will default to the namespace of the object referencing it. |
| cluster | [string](#string) |  | Required: the cluster on which the object exists. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

