
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
| kubeIdentityMatcher | [IdentitySelector.KubeIdentityMatcher](#networking.smh.solo.io.IdentitySelector.KubeIdentityMatcher) |  | A KubeIdentityMatcher matches request identities based on the k8s namespace and cluster. |
| kubeServiceAccountRefs | [IdentitySelector.KubeServiceAccountRefs](#networking.smh.solo.io.IdentitySelector.KubeServiceAccountRefs) |  | KubeServiceAccountRefs matches request identities based on the k8s service account of request. |






<a name="networking.smh.solo.io.IdentitySelector.KubeIdentityMatcher"></a>

### IdentitySelector.KubeIdentityMatcher



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| namespaces | [][string](#string) | repeated | Namespaces to allow. If not set, any namespace is allowed. |
| clusters | [][string](#string) | repeated | Cluster to allow. If not set, any cluster is allowed. |






<a name="networking.smh.solo.io.IdentitySelector.KubeServiceAccountRefs"></a>

### IdentitySelector.KubeServiceAccountRefs



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| serviceAccounts | [][core.skv2.solo.io.ClusterObjectRef](#core.skv2.solo.io.ClusterObjectRef) | repeated | List of ServiceAccounts to allow. If not set, any ServiceAccount is allowed. |






<a name="networking.smh.solo.io.TrafficTargetSelector"></a>

### TrafficTargetSelector
Select Kubernetes services.<br>Only one of (labels + namespaces + cluster) or (resource refs) may be provided. If all four are provided, it will be considered an error, and the Status of the top level resource will be updated to reflect an IllegalSelection.<br>Valid: 1. selector: matcher: labels: foo: bar hello: world namespaces: - default cluster: "cluster-name" 2. selector: matcher: refs: - name: foo namespace: bar<br>Invalid: 1. selector: matcher: labels: foo: bar hello: world namespaces: - default cluster: "cluster-name" refs: - name: foo namespace: bar<br>By default labels will select across all namespaces, unless a list of namespaces is provided, in which case it will only select from those. An empty list is equal to AllNamespaces.<br>If no labels are given, and only namespaces, all resources from the namespaces will be selected.<br>The following selector will select all resources with the following labels in every namespace, in the local cluster:<br>selector: matcher: labels: foo: bar hello: world<br>Whereas the next selector will only select from the specified namespaces (foo, bar), in the local cluster:<br>selector: matcher: labels: foo: bar hello: world namespaces - foo - bar<br>This final selector will select all resources of a given type in the target namespace (foo), in the local cluster:<br>selector matcher: namespaces - foo - bar labels: hello: world


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubeServiceMatcher | [TrafficTargetSelector.KubeServiceMatcher](#networking.smh.solo.io.TrafficTargetSelector.KubeServiceMatcher) |  | A KubeServiceMatcher matches kubernetes services by the namespaces and clusters they belong to, as well as the provided labels. |
| kubeServiceRefs | [TrafficTargetSelector.KubeServiceRefs](#networking.smh.solo.io.TrafficTargetSelector.KubeServiceRefs) |  | Match individual k8s Services by direct reference. |






<a name="networking.smh.solo.io.TrafficTargetSelector.KubeServiceMatcher"></a>

### TrafficTargetSelector.KubeServiceMatcher



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| labels | [][TrafficTargetSelector.KubeServiceMatcher.LabelsEntry](#networking.smh.solo.io.TrafficTargetSelector.KubeServiceMatcher.LabelsEntry) | repeated | If specified, all labels must exist on k8s Service, else match on any labels. |
| namespaces | [][string](#string) | repeated | If specified, match k8s Services if they exist in one of the specified namespaces. If not specified, match on any namespace. |
| clusters | [][string](#string) | repeated | If specified, match k8s Services if they exist in one of the specified clusters. If not specified, match on any cluster. |






<a name="networking.smh.solo.io.TrafficTargetSelector.KubeServiceMatcher.LabelsEntry"></a>

### TrafficTargetSelector.KubeServiceMatcher.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="networking.smh.solo.io.TrafficTargetSelector.KubeServiceRefs"></a>

### TrafficTargetSelector.KubeServiceRefs



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| services | [][core.skv2.solo.io.ClusterObjectRef](#core.skv2.solo.io.ClusterObjectRef) | repeated | Match k8s Services by direct reference. |






<a name="networking.smh.solo.io.WorkloadSelector"></a>

### WorkloadSelector
Select Kubernetes workloads directly using label and/or namespace criteria. See comments on the fields for detailed semantics.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| labels | [][WorkloadSelector.LabelsEntry](#networking.smh.solo.io.WorkloadSelector.LabelsEntry) | repeated | If specified, all labels must exist on workloads, else match on any labels. |
| namespaces | [][string](#string) | repeated | If specified, match workloads if they exist in one of the specified namespaces. If not specified, match on any namespace. |






<a name="networking.smh.solo.io.WorkloadSelector.LabelsEntry"></a>

### WorkloadSelector.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

