
---

title: "selectors.proto"

---

## Package : `common.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for selectors.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## selectors.proto


## Table of Contents
  - [IdentitySelector](#common.mesh.gloo.solo.io.IdentitySelector)
  - [IdentitySelector.KubeIdentityMatcher](#common.mesh.gloo.solo.io.IdentitySelector.KubeIdentityMatcher)
  - [IdentitySelector.KubeServiceAccountRefs](#common.mesh.gloo.solo.io.IdentitySelector.KubeServiceAccountRefs)
  - [TrafficTargetSelector](#common.mesh.gloo.solo.io.TrafficTargetSelector)
  - [TrafficTargetSelector.KubeServiceMatcher](#common.mesh.gloo.solo.io.TrafficTargetSelector.KubeServiceMatcher)
  - [TrafficTargetSelector.KubeServiceMatcher.LabelsEntry](#common.mesh.gloo.solo.io.TrafficTargetSelector.KubeServiceMatcher.LabelsEntry)
  - [TrafficTargetSelector.KubeServiceRefs](#common.mesh.gloo.solo.io.TrafficTargetSelector.KubeServiceRefs)
  - [WorkloadSelector](#common.mesh.gloo.solo.io.WorkloadSelector)
  - [WorkloadSelector.LabelsEntry](#common.mesh.gloo.solo.io.WorkloadSelector.LabelsEntry)







<a name="common.mesh.gloo.solo.io.IdentitySelector"></a>

### IdentitySelector
Select TrafficTarget identities using one or more platform-specific selectors.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeIdentityMatcher | [common.mesh.gloo.solo.io.IdentitySelector.KubeIdentityMatcher]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1alpha2.selectors#common.mesh.gloo.solo.io.IdentitySelector.KubeIdentityMatcher" >}}) |  | Match request identities based on the Kubernetes namespace and cluster. |
  | kubeServiceAccountRefs | [common.mesh.gloo.solo.io.IdentitySelector.KubeServiceAccountRefs]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1alpha2.selectors#common.mesh.gloo.solo.io.IdentitySelector.KubeServiceAccountRefs" >}}) |  | Match request identities based on the Kubernetes service account of the request. |
  





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
  





<a name="common.mesh.gloo.solo.io.TrafficTargetSelector"></a>

### TrafficTargetSelector
Select TrafficTargets using one or more platform-specific selectors.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeServiceMatcher | [common.mesh.gloo.solo.io.TrafficTargetSelector.KubeServiceMatcher]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1alpha2.selectors#common.mesh.gloo.solo.io.TrafficTargetSelector.KubeServiceMatcher" >}}) |  | Match Kubernetes Services by their labels, namespaces, and/or clusters. |
  | kubeServiceRefs | [common.mesh.gloo.solo.io.TrafficTargetSelector.KubeServiceRefs]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1alpha2.selectors#common.mesh.gloo.solo.io.TrafficTargetSelector.KubeServiceRefs" >}}) |  | Match Kubernetes Services by direct reference. |
  





<a name="common.mesh.gloo.solo.io.TrafficTargetSelector.KubeServiceMatcher"></a>

### TrafficTargetSelector.KubeServiceMatcher



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| labels | [][common.mesh.gloo.solo.io.TrafficTargetSelector.KubeServiceMatcher.LabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1alpha2.selectors#common.mesh.gloo.solo.io.TrafficTargetSelector.KubeServiceMatcher.LabelsEntry" >}}) | repeated | If specified, a match requires all labels to exist on a Kubernetes Service. When used in a networking policy, omission matches any labels. When used in a Gloo Mesh Role, a wildcard (`"*"`) must be specified to match any label key and/or value. |
  | namespaces | []string | repeated | If specified, match Kubernetes Services if they exist in one of the specified namespaces. When used in a networking policy, omission matches any namespace. When used in a Gloo Mesh Role, a wildcard (`"*"`) must be specified to match any namespace. |
  | clusters | []string | repeated | If specified, match Kubernetes Services if they exist in one of the specified clusters. When used in a networking policy, omission matches any cluster. When used in a Gloo Mesh Role, a wildcard (`"*"`) must be specified to match any cluster. |
  





<a name="common.mesh.gloo.solo.io.TrafficTargetSelector.KubeServiceMatcher.LabelsEntry"></a>

### TrafficTargetSelector.KubeServiceMatcher.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="common.mesh.gloo.solo.io.TrafficTargetSelector.KubeServiceRefs"></a>

### TrafficTargetSelector.KubeServiceRefs



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| services | [][core.skv2.solo.io.ClusterObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ClusterObjectRef" >}}) | repeated | Match Kubernetes Services by direct reference. When used in a networking policy, omission of any field (name, namespace, or clusterName) allows matching any value for that field. When used in a Gloo Mesh Role, a wildcard (`"*"`) must be specified to match any value for the given field. |
  





<a name="common.mesh.gloo.solo.io.WorkloadSelector"></a>

### WorkloadSelector
Select Workloads using one or more platform-specific selectors.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| labels | [][common.mesh.gloo.solo.io.WorkloadSelector.LabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1alpha2.selectors#common.mesh.gloo.solo.io.WorkloadSelector.LabelsEntry" >}}) | repeated | If specified, all labels must exist on Kubernetes workload. When used in a networking policy, omission matches any labels. When used in a Gloo Mesh Role, a wildcard (`"*"`) must be specified to match any label key and/or value. |
  | namespaces | []string | repeated | If specified, match Kubernetes workloads if they exist in one of the specified namespaces. When used in a networking policy, omission matches any namespace. When used in a Gloo Mesh Role, a wildcard (`"*"`) must be specified to match any namespace. |
  | clusters | []string | repeated | If specified, match Kubernetes workloads if they exist in one of the specified clusters. When used in a networking policy, omission matches any cluster. When used in a Gloo Mesh Role, a wildcard (`"*"`) must be specified to match any cluster. |
  





<a name="common.mesh.gloo.solo.io.WorkloadSelector.LabelsEntry"></a>

### WorkloadSelector.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

