
---

---

## Package : `core.skv2.solo.io`



<a name="top"></a>

<a name="API Reference for core.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## core.proto


## Table of Contents
  - [ClusterObjectRef](#core.skv2.solo.io.ClusterObjectRef)
  - [ObjectRef](#core.skv2.solo.io.ObjectRef)
  - [ObjectRefList](#core.skv2.solo.io.ObjectRefList)
  - [ObjectSelector](#core.skv2.solo.io.ObjectSelector)
  - [ObjectSelector.Expression](#core.skv2.solo.io.ObjectSelector.Expression)
  - [ObjectSelector.LabelsEntry](#core.skv2.solo.io.ObjectSelector.LabelsEntry)
  - [Status](#core.skv2.solo.io.Status)
  - [TypedClusterObjectRef](#core.skv2.solo.io.TypedClusterObjectRef)
  - [TypedObjectRef](#core.skv2.solo.io.TypedObjectRef)

  - [ObjectSelector.Expression.Operator](#core.skv2.solo.io.ObjectSelector.Expression.Operator)
  - [Status.State](#core.skv2.solo.io.Status.State)






<a name="core.skv2.solo.io.ClusterObjectRef"></a>

### ClusterObjectRef



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | name of the resource being referenced |
  | namespace | string |  | namespace of the resource being referenced |
  | clusterName | string |  | name of the cluster in which the resource exists |
  





<a name="core.skv2.solo.io.ObjectRef"></a>

### ObjectRef



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | name of the resource being referenced |
  | namespace | string |  | namespace of the resource being referenced |
  





<a name="core.skv2.solo.io.ObjectRefList"></a>

### ObjectRefList



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| refs | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated |  |
  





<a name="core.skv2.solo.io.ObjectSelector"></a>

### ObjectSelector



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| namespaces | []string | repeated | Select Objects in these namespaces. If omitted, Gloo Mesh will only select Objects in the same namespace as the parent resource (e.g. VirtualGateway) that owns this selector. The reserved value "*" can be used to select objects in all namespaces watched by Gloo Mesh. |
  | labels | [][core.skv2.solo.io.ObjectSelector.LabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectSelector.LabelsEntry" >}}) | repeated | Select objects whose labels match the ones specified here. |
  | expressions | [][core.skv2.solo.io.ObjectSelector.Expression]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectSelector.Expression" >}}) | repeated | Expressions allow for more flexible object label matching, such as equality-based requirements, set-based requirements, or a combination of both. https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#equality-based-requirement |
  





<a name="core.skv2.solo.io.ObjectSelector.Expression"></a>

### ObjectSelector.Expression



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  | Kubernetes label key, must conform to Kubernetes syntax requirements https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set |
  | operator | [core.skv2.solo.io.ObjectSelector.Expression.Operator]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectSelector.Expression.Operator" >}}) |  | The operator can only be in, notin, =, ==, !=, exists, ! (DoesNotExist), gt (GreaterThan), lt (LessThan). |
  | values | []string | repeated |  |
  





<a name="core.skv2.solo.io.ObjectSelector.LabelsEntry"></a>

### ObjectSelector.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="core.skv2.solo.io.Status"></a>

### Status



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| state | [core.skv2.solo.io.Status.State]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.Status.State" >}}) |  | The current state of the resource |
  | message | string |  | A human readable message about the current state of the object |
  | observedGeneration | int64 |  | The most recently observed generation of the resource. This value corresponds to the `metadata.generation` of a kubernetes resource |
  | processingTime | [google.protobuf.Timestamp]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.timestamp#google.protobuf.Timestamp" >}}) |  | The time at which this status was recorded |
  | owner | [google.protobuf.StringValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.StringValue" >}}) |  | (optional) The owner of the status, this value can be used to identify the entity which wrote this status. This is useful in situations where a given resource may have multiple owners. |
  





<a name="core.skv2.solo.io.TypedClusterObjectRef"></a>

### TypedClusterObjectRef



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| apiGroup | [google.protobuf.StringValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.StringValue" >}}) |  | API group of the resource being referenced |
  | kind | [google.protobuf.StringValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.StringValue" >}}) |  | Kind of the resource being referenced |
  | name | string |  | name of the resource being referenced |
  | namespace | string |  | namespace of the resource being referenced |
  | clusterName | string |  | name of the cluster in which the resource exists |
  





<a name="core.skv2.solo.io.TypedObjectRef"></a>

### TypedObjectRef



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| apiGroup | [google.protobuf.StringValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.StringValue" >}}) |  | API group of the resource being referenced |
  | kind | [google.protobuf.StringValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.StringValue" >}}) |  | Kind of the resource being referenced |
  | name | string |  | name of the resource being referenced |
  | namespace | string |  | namespace of the resource being referenced |
  




 <!-- end messages -->


<a name="core.skv2.solo.io.ObjectSelector.Expression.Operator"></a>

### ObjectSelector.Expression.Operator


| Name | Number | Description |
| ---- | ------ | ----------- |
| Equals | 0 | = |
| DoubleEquals | 1 | == |
| NotEquals | 2 | != |
| In | 3 | in |
| NotIn | 4 | notin |
| Exists | 5 | exists |
| DoesNotExist | 6 | ! |
| GreaterThan | 7 | gt |
| LessThan | 8 | lt |



<a name="core.skv2.solo.io.Status.State"></a>

### Status.State


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

