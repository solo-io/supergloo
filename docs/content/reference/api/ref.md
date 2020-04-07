
---
title: "core.zephyr.solo.iogithub.com/solo-io/service-mesh-hub/api/core/v1alpha1/ref.proto"
---

## Package : `core.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for github.com/solo-io/service-mesh-hub/api/core/v1alpha1/ref.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## github.com/solo-io/service-mesh-hub/api/core/v1alpha1/ref.proto


## Table of Contents
  - [ResourceRef](#core.zephyr.solo.io.ResourceRef)







<a name="core.zephyr.solo.io.ResourceRef"></a>

### ResourceRef
reference object for kubernetes objects, support multi cluster


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | The name of the object |
| namespace | [string](#string) |  | The namespace in which the object lives |
| cluster | [string](#string) |  | The cluster on which the object exists. This should be the name under which the cluster was registered in Service Mesh Hub, e.g. through `meshctl cluster register`. That name also will correspond to a KubernetesCluster resource in the management plane cluster.<br>This field not being set will result in an error if the object in question lives on a remote cluster (i.e., not a Service Mesh Hub resource). |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

