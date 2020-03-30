
---
title: "core.zephyr.solo.iogithub.com/solo-io/mesh-projects/api/core/v1alpha1/selector.proto"
---

## Package : `core.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for github.com/solo-io/mesh-projects/api/core/v1alpha1/selector.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## github.com/solo-io/mesh-projects/api/core/v1alpha1/selector.proto


## Table of Contents
  - [IdentitySelector](#core.zephyr.solo.io.IdentitySelector)
  - [IdentitySelector.Matcher](#core.zephyr.solo.io.IdentitySelector.Matcher)
  - [IdentitySelector.ServiceAccountRefs](#core.zephyr.solo.io.IdentitySelector.ServiceAccountRefs)
  - [ServiceSelector](#core.zephyr.solo.io.ServiceSelector)
  - [ServiceSelector.Matcher](#core.zephyr.solo.io.ServiceSelector.Matcher)
  - [ServiceSelector.Matcher.LabelsEntry](#core.zephyr.solo.io.ServiceSelector.Matcher.LabelsEntry)
  - [ServiceSelector.ServiceRefs](#core.zephyr.solo.io.ServiceSelector.ServiceRefs)
  - [WorkloadSelector](#core.zephyr.solo.io.WorkloadSelector)
  - [WorkloadSelector.LabelsEntry](#core.zephyr.solo.io.WorkloadSelector.LabelsEntry)







<a name="core.zephyr.solo.io.IdentitySelector"></a>

### IdentitySelector
Selector capable of selecting specific service identities. Useful for binding policy rules. Either (namespaces, cluster, service_account_names) or service_accounts can be specified. If all fields are omitted, any source identity is permitted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| matcher | [IdentitySelector.Matcher](#core.zephyr.solo.io.IdentitySelector.Matcher) |  |  |
| serviceAccountRefs | [IdentitySelector.ServiceAccountRefs](#core.zephyr.solo.io.IdentitySelector.ServiceAccountRefs) |  |  |






<a name="core.zephyr.solo.io.IdentitySelector.Matcher"></a>

### IdentitySelector.Matcher



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| namespaces | [][string](#string) | repeated | Namespaces to allow. If not set, any namespace is allowed. |
| clusters | [][string](#string) | repeated | Cluster to allow. If not set, any cluster is allowed. |






<a name="core.zephyr.solo.io.IdentitySelector.ServiceAccountRefs"></a>

### IdentitySelector.ServiceAccountRefs



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| serviceAccounts | [][ResourceRef](#core.zephyr.solo.io.ResourceRef) | repeated | List of ServiceAccounts to allow. If not set, any ServiceAccount is allowed. |






<a name="core.zephyr.solo.io.ServiceSelector"></a>

### ServiceSelector
Select Kubernetes services<br>Only one of (labels + namespaces + cluster) or (resource refs) may be provided. If all four are provided, it will be considered an error, and the Status of the top level resource will be updated to reflect an IllegalSelection.<br>Valid: 1. selector: labels: foo: bar hello: world namespaces: - default cluster: "cluster-name" 2. selector: refs: - name: foo namespace: bar<br>Invalid: 1. selector: labels: foo: bar hello: world namespaces: - default cluster: "cluster-name" refs: - name: foo namespace: bar<br>By default labels will select across all namespaces, unless a list of namespaces is provided, in which case it will only select from those. An empty list is equal to AllNamespaces.<br>If no labels are given, and only namespaces, all resources from the namespaces will be selected.<br>The following selector will select all resources with the following labels in every namespace, in the local cluster:<br>selector: matcher: labels: foo: bar hello: world<br>Whereas the next selector will only select from the specified namespaces (foo, bar), in the local cluster:<br>selector: matcher: labels: foo: bar hello: world namespaces - foo - bar<br>This final selector will select all resources of a given type in the target namespace (foo), in the local cluster:<br>selector matcher: namespaces - foo - bar labels: hello: world


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| matcher | [ServiceSelector.Matcher](#core.zephyr.solo.io.ServiceSelector.Matcher) |  |  |
| serviceRefs | [ServiceSelector.ServiceRefs](#core.zephyr.solo.io.ServiceSelector.ServiceRefs) |  |  |






<a name="core.zephyr.solo.io.ServiceSelector.Matcher"></a>

### ServiceSelector.Matcher



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| labels | [][ServiceSelector.Matcher.LabelsEntry](#core.zephyr.solo.io.ServiceSelector.Matcher.LabelsEntry) | repeated | If specified, all labels must exist on k8s Service, else match on any labels. |
| namespaces | [][string](#string) | repeated | If specified, match k8s Services if they exist in one of the specified namespaces. If not specified, match on any namespace. |
| clusters | [][string](#string) | repeated | If specified, match k8s Services if they exist in one of the specified clusters. If not specified, match on any cluster. |






<a name="core.zephyr.solo.io.ServiceSelector.Matcher.LabelsEntry"></a>

### ServiceSelector.Matcher.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="core.zephyr.solo.io.ServiceSelector.ServiceRefs"></a>

### ServiceSelector.ServiceRefs



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| services | [][ResourceRef](#core.zephyr.solo.io.ResourceRef) | repeated | Match k8s Services by direct reference. |






<a name="core.zephyr.solo.io.WorkloadSelector"></a>

### WorkloadSelector
Select Kubernetes workloads directly using label and/or namespace criteria. See comments on the fields for detailed semantics.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| labels | [][WorkloadSelector.LabelsEntry](#core.zephyr.solo.io.WorkloadSelector.LabelsEntry) | repeated | If specified, all labels must exist on workloads, else match on any labels. |
| namespaces | [][string](#string) | repeated | If specified, match workloads if they exist in one of the specified namespaces. If not specified, match on any namespace. |






<a name="core.zephyr.solo.io.WorkloadSelector.LabelsEntry"></a>

### WorkloadSelector.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

