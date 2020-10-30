
---
title: "selectors.proto"
---

## Package : `networking.smh.solo.io`



<a name="top"></a>

<a name="API Reference for selectors.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## selectors.proto


## Table of Contents
  - [IdentitySelector](#networking.smh.solo.io.IdentitySelector)
  - [IdentitySelector.KubeIdentityMatcher](#networking.smh.solo.io.IdentitySelector.KubeIdentityMatcher)
  - [IdentitySelector.KubeServiceAccountRefs](#networking.smh.solo.io.IdentitySelector.KubeServiceAccountRefs)
  - [TrafficTargetSelector](#networking.smh.solo.io.TrafficTargetSelector)
  - [TrafficTargetSelector.KubeServiceMatcher](#networking.smh.solo.io.TrafficTargetSelector.KubeServiceMatcher)
  - [TrafficTargetSelector.KubeServiceMatcher.LabelsEntry](#networking.smh.solo.io.TrafficTargetSelector.KubeServiceMatcher.LabelsEntry)
  - [TrafficTargetSelector.KubeServiceRefs](#networking.smh.solo.io.TrafficTargetSelector.KubeServiceRefs)
  - [WorkloadSelector](#networking.smh.solo.io.WorkloadSelector)
  - [WorkloadSelector.LabelsEntry](#networking.smh.solo.io.WorkloadSelector.LabelsEntry)







<a name="networking.smh.solo.io.IdentitySelector"></a>

### IdentitySelector
Selector capable of selecting specific service identities. Useful for binding policy rules. Either (namespaces, cluster, service_account_names) or service_accounts can be specified. If all fields are omitted, any source identity is permitted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeIdentityMatcher | networking.smh.solo.io.IdentitySelector.KubeIdentityMatcher |  | A KubeIdentityMatcher matches request identities based on the k8s namespace and cluster. |
| kubeServiceAccountRefs | networking.smh.solo.io.IdentitySelector.KubeServiceAccountRefs |  | KubeServiceAccountRefs matches request identities based on the k8s service account of request. |






<a name="networking.smh.solo.io.IdentitySelector.KubeIdentityMatcher"></a>

### IdentitySelector.KubeIdentityMatcher



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| namespaces | []string | repeated | If specified, match k8s identity if it exists in one of the specified namespaces. If not specified, match on any namespace. When used in a networking policy, omission matches any namespace. When used in a Role, a wildcard `"*"` must be explicitly used to match any namespace. |
| clusters | []string | repeated | If specified, match k8s identity if it exists in one of the specified clusters. If not specified, match on any cluster. When used in a networking policy, omission matches any cluster. When used in a Role, a wildcard `"*"` must be explicitly used to match any cluster. |






<a name="networking.smh.solo.io.IdentitySelector.KubeServiceAccountRefs"></a>

### IdentitySelector.KubeServiceAccountRefs



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| serviceAccounts | []core.skv2.solo.io.ClusterObjectRef | repeated | Match k8s ServiceAccounts by direct reference. When used in a networking policy, omission of any field (name, namespace, or clusterName) allows matching any value for that field. When used in a Role, a wildcard `"*"` must be explicitly used to match any value for the given field. |






<a name="networking.smh.solo.io.TrafficTargetSelector"></a>

### TrafficTargetSelector
Select TrafficTargets using one or more platform-specific selection objects.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeServiceMatcher | networking.smh.solo.io.TrafficTargetSelector.KubeServiceMatcher |  | A KubeServiceMatcher matches kubernetes services by their labels, namespaces, and/or clusters. |
| kubeServiceRefs | networking.smh.solo.io.TrafficTargetSelector.KubeServiceRefs |  | Match individual k8s Services by direct reference. |






<a name="networking.smh.solo.io.TrafficTargetSelector.KubeServiceMatcher"></a>

### TrafficTargetSelector.KubeServiceMatcher



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| labels | []networking.smh.solo.io.TrafficTargetSelector.KubeServiceMatcher.LabelsEntry | repeated | If specified, all labels must exist on k8s Service. When used in a networking policy, omission matches any labels. When used in a Role, a wildcard `"*"` must be explicitly used to match any label key and/or value. |
| namespaces | []string | repeated | If specified, match k8s Services if they exist in one of the specified namespaces. If not specified, match on any namespace. When used in a networking policy, omission matches any namespace. When used in a Role, a wildcard `"*"` must be explicitly used to match any namespace. |
| clusters | []string | repeated | If specified, match k8s Services if they exist in one of the specified clusters. If not specified, match on any cluster. When used in a networking policy, omission matches any cluster. When used in a Role, a wildcard `"*"` must be explicitly used to match any cluster. |






<a name="networking.smh.solo.io.TrafficTargetSelector.KubeServiceMatcher.LabelsEntry"></a>

### TrafficTargetSelector.KubeServiceMatcher.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
| value | string |  |  |






<a name="networking.smh.solo.io.TrafficTargetSelector.KubeServiceRefs"></a>

### TrafficTargetSelector.KubeServiceRefs



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| services | []core.skv2.solo.io.ClusterObjectRef | repeated | Match k8s Services by direct reference. When used in a networking policy, omission of any field (name, namespace, or clusterName) allows matching any value for that field. When used in a Role, a wildcard `"*"` must be explicitly used to match any value for the given field. |






<a name="networking.smh.solo.io.WorkloadSelector"></a>

### WorkloadSelector
Select Kubernetes workloads directly using label namespace and/or cluster criteria. See comments on the fields for detailed semantics.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| labels | []networking.smh.solo.io.WorkloadSelector.LabelsEntry | repeated | If specified, all labels must exist on k8s workload. When used in a networking policy, omission matches any labels. When used in a Role, a wildcard `"*"` must be explicitly used to match any label key and/or value. |
| namespaces | []string | repeated | If specified, match k8s workloads if they exist in one of the specified namespaces. If not specified, match on any namespace. When used in a networking policy, omission matches any namespace. When used in a Role, a wildcard `"*"` must be explicitly used to match any namespace. |
| clusters | []string | repeated | If specified, match k8s workloads if they exist in one of the specified clusters. If not specified, match on any cluster. When used in a networking policy, omission matches any cluster. When used in a Role, a wildcard `"*"` must be explicitly used to match any cluster. |






<a name="networking.smh.solo.io.WorkloadSelector.LabelsEntry"></a>

### WorkloadSelector.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
| value | string |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

