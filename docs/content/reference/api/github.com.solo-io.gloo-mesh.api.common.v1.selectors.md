
---

title: "selectors.proto"

---

## Package : `common.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for selectors.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## selectors.proto


## Table of Contents
  - [DestinationSelector](#common.mesh.gloo.solo.io.DestinationSelector)
  - [DestinationSelector.KubeServiceMatcher](#common.mesh.gloo.solo.io.DestinationSelector.KubeServiceMatcher)
  - [DestinationSelector.KubeServiceMatcher.LabelsEntry](#common.mesh.gloo.solo.io.DestinationSelector.KubeServiceMatcher.LabelsEntry)
  - [DestinationSelector.KubeServiceRefs](#common.mesh.gloo.solo.io.DestinationSelector.KubeServiceRefs)
  - [IdentitySelector](#common.mesh.gloo.solo.io.IdentitySelector)
  - [IdentitySelector.KubeIdentityMatcher](#common.mesh.gloo.solo.io.IdentitySelector.KubeIdentityMatcher)
  - [IdentitySelector.KubeServiceAccountRefs](#common.mesh.gloo.solo.io.IdentitySelector.KubeServiceAccountRefs)
  - [RouteTableSelector](#common.mesh.gloo.solo.io.RouteTableSelector)
  - [RouteTableSelector.Expression](#common.mesh.gloo.solo.io.RouteTableSelector.Expression)
  - [RouteTableSelector.LabelsEntry](#common.mesh.gloo.solo.io.RouteTableSelector.LabelsEntry)
  - [WorkloadSelector](#common.mesh.gloo.solo.io.WorkloadSelector)
  - [WorkloadSelector.KubeWorkloadMatcher](#common.mesh.gloo.solo.io.WorkloadSelector.KubeWorkloadMatcher)
  - [WorkloadSelector.KubeWorkloadMatcher.LabelsEntry](#common.mesh.gloo.solo.io.WorkloadSelector.KubeWorkloadMatcher.LabelsEntry)

  - [RouteTableSelector.Expression.Operator](#common.mesh.gloo.solo.io.RouteTableSelector.Expression.Operator)






<a name="common.mesh.gloo.solo.io.DestinationSelector"></a>

### DestinationSelector
Select Destinations using one or more platform-specific selectors.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeServiceMatcher | [common.mesh.gloo.solo.io.DestinationSelector.KubeServiceMatcher]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.DestinationSelector.KubeServiceMatcher" >}}) |  | Match Kubernetes Services by their labels, namespaces, and/or clusters. |
  | kubeServiceRefs | [common.mesh.gloo.solo.io.DestinationSelector.KubeServiceRefs]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.DestinationSelector.KubeServiceRefs" >}}) |  | Match Kubernetes Services by direct reference. |
  





<a name="common.mesh.gloo.solo.io.DestinationSelector.KubeServiceMatcher"></a>

### DestinationSelector.KubeServiceMatcher
Match Kubernetes Services by their labels, namespaces, and/or clusters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| labels | [][common.mesh.gloo.solo.io.DestinationSelector.KubeServiceMatcher.LabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.DestinationSelector.KubeServiceMatcher.LabelsEntry" >}}) | repeated | If specified, a match requires all labels to exist on a Kubernetes Service. When used in a networking policy, omission matches any labels. When used in a Gloo Mesh Role, a wildcard (`"*"`) must be specified to match any label key and/or value. |
  | namespaces | []string | repeated | If specified, match Kubernetes Services if they exist in one of the specified namespaces. When used in a networking policy, omission matches any namespace. When used in a Gloo Mesh Role, a wildcard (`"*"`) must be specified to match any namespace. |
  | clusters | []string | repeated | If specified, match Kubernetes Services if they exist in one of the specified clusters. When used in a networking policy, omission matches any cluster. When used in a Gloo Mesh Role, a wildcard (`"*"`) must be specified to match any cluster. |
  





<a name="common.mesh.gloo.solo.io.DestinationSelector.KubeServiceMatcher.LabelsEntry"></a>

### DestinationSelector.KubeServiceMatcher.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="common.mesh.gloo.solo.io.DestinationSelector.KubeServiceRefs"></a>

### DestinationSelector.KubeServiceRefs
Match Kubernetes Services by direct reference.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| services | [][core.skv2.solo.io.ClusterObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ClusterObjectRef" >}}) | repeated | Match Kubernetes Services by direct reference. When used in a networking policy, omission of any field (name, namespace, or clusterName) allows matching any value for that field. When used in a Gloo Mesh Role, a wildcard (`"*"`) must be specified to match any value for the given field. |
  





<a name="common.mesh.gloo.solo.io.IdentitySelector"></a>

### IdentitySelector
Select Destination identities using one or more platform-specific selectors.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeIdentityMatcher | [common.mesh.gloo.solo.io.IdentitySelector.KubeIdentityMatcher]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.IdentitySelector.KubeIdentityMatcher" >}}) |  | Match request identities based on the Kubernetes namespace and cluster. |
  | kubeServiceAccountRefs | [common.mesh.gloo.solo.io.IdentitySelector.KubeServiceAccountRefs]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.IdentitySelector.KubeServiceAccountRefs" >}}) |  | Match request identities based on the Kubernetes service account of the request. |
  





<a name="common.mesh.gloo.solo.io.IdentitySelector.KubeIdentityMatcher"></a>

### IdentitySelector.KubeIdentityMatcher



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| namespaces | []string | repeated | If specified, match a Kubernetes identity if it exists in one of the specified namespaces. When used in a networking policy, omission matches any namespace. When used in a Gloo Mesh Role, a wildcard (`"*"`) must be specified to match any namespace. |
  | clusters | []string | repeated | If specified, match a Kubernetes identity if it exists in one of the specified clusters. When used in a networking policy, omission matches any cluster. When used in a Gloo Mesh Role, a wildcard (`"*"`) must be specified to match any cluster. |
  





<a name="common.mesh.gloo.solo.io.IdentitySelector.KubeServiceAccountRefs"></a>

### IdentitySelector.KubeServiceAccountRefs



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| serviceAccounts | [][core.skv2.solo.io.ClusterObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ClusterObjectRef" >}}) | repeated | Match Kubernetes service accounts by direct reference. When used in a networking policy, omission of any field (name, namespace, or clusterName) allows matching any value for that field. When used in a Gloo Mesh Role, a wildcard (`"*"`) must be specified to match any value for the given field. |
  





<a name="common.mesh.gloo.solo.io.RouteTableSelector"></a>

### RouteTableSelector
TODO: move to package alongside RouteTable Select route tables (or delegated route tables) for delegation by namespace, labels, or both.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| namespaces | []string | repeated | Delegate to Route Tables in these namespaces. If omitted, Gloo will only select Route Tables in the same namespace as the resource (FederatedGateway, Route Table or Delegated Route Table) that owns this selector. The reserved value "*" can be used to select Route Tables in all namespaces watched by Gloo. |
  | labels | [][common.mesh.gloo.solo.io.RouteTableSelector.LabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.RouteTableSelector.LabelsEntry" >}}) | repeated | Delegate to Route Tables whose labels match the ones specified here. |
  | expressions | [][common.mesh.gloo.solo.io.RouteTableSelector.Expression]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.RouteTableSelector.Expression" >}}) | repeated | Expressions allow for more flexible Route Tables label matching, such as equality-based requirements, set-based requirements, or a combination of both. https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#equality-based-requirement |
  





<a name="common.mesh.gloo.solo.io.RouteTableSelector.Expression"></a>

### RouteTableSelector.Expression



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  | Kubernetes label key, must conform to Kubernetes syntax requirements https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set |
  | operator | [common.mesh.gloo.solo.io.RouteTableSelector.Expression.Operator]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.RouteTableSelector.Expression.Operator" >}}) |  | The operator can only be in, notin, =, ==, !=, exists, ! (DoesNotExist), gt (GreaterThan), lt (LessThan). |
  | values | []string | repeated |  |
  





<a name="common.mesh.gloo.solo.io.RouteTableSelector.LabelsEntry"></a>

### RouteTableSelector.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="common.mesh.gloo.solo.io.WorkloadSelector"></a>

### WorkloadSelector
Select Workloads using one or more platform-specific selectors.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeWorkloadMatcher | [common.mesh.gloo.solo.io.WorkloadSelector.KubeWorkloadMatcher]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.WorkloadSelector.KubeWorkloadMatcher" >}}) |  | Match Kubernetes workloads by their labels, namespaces, and/or clusters. |
  





<a name="common.mesh.gloo.solo.io.WorkloadSelector.KubeWorkloadMatcher"></a>

### WorkloadSelector.KubeWorkloadMatcher
Match Kubernetes workloads by their labels, namespaces, and/or clusters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| labels | [][common.mesh.gloo.solo.io.WorkloadSelector.KubeWorkloadMatcher.LabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.WorkloadSelector.KubeWorkloadMatcher.LabelsEntry" >}}) | repeated | If specified, all labels must exist on Kubernetes workload. When used in a networking policy, omission matches any labels. When used in a Gloo Mesh Role, a wildcard (`"*"`) must be specified to match any label key and/or value. |
  | namespaces | []string | repeated | If specified, match Kubernetes workloads if they exist in one of the specified namespaces. When used in a networking policy, omission matches any namespace. When used in a Gloo Mesh Role, a wildcard (`"*"`) must be specified to match any namespace. |
  | clusters | []string | repeated | If specified, match Kubernetes workloads if they exist in one of the specified clusters. When used in a networking policy, omission matches any cluster. When used in a Gloo Mesh Role, a wildcard (`"*"`) must be specified to match any cluster. |
  





<a name="common.mesh.gloo.solo.io.WorkloadSelector.KubeWorkloadMatcher.LabelsEntry"></a>

### WorkloadSelector.KubeWorkloadMatcher.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  




 <!-- end messages -->


<a name="common.mesh.gloo.solo.io.RouteTableSelector.Expression.Operator"></a>

### RouteTableSelector.Expression.Operator
Route Table Selector expression operator, while the set-based syntax differs from Kubernetes (kubernetes: `key: !mylabel`, gloo: `key: mylabel, operator: "!"` | kubernetes: `key: mylabel`, gloo: `key: mylabel, operator: exists`), the functionality remains the same.

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


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

