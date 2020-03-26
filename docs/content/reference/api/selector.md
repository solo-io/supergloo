
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
  - [Selector](#core.zephyr.solo.io.Selector)
  - [Selector.LabelsEntry](#core.zephyr.solo.io.Selector.LabelsEntry)







<a name="core.zephyr.solo.io.IdentitySelector"></a>

### IdentitySelector
Special selector capable of selecting specific service identities. Useful for binding policy rules.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| namespaces | [][string](#string) | repeated | list of namespaces to search |
| cluster | [google.protobuf.StringValue](#google.protobuf.StringValue) |  | If empty string or nil, select in the cluster local to the enclosing resource, else select in the referenced remote cluster |
| serviceAccounts | [][string](#string) | repeated | List of service accounts within the given namespaces to select. If empty, all service accounts will be selected. |






<a name="core.zephyr.solo.io.Selector"></a>

### Selector
Global selector used to select resources from any set of clusters, namespaces, and/or labels<br>Specifies a method by which to select pods within a mesh for the application of rules and policies.<br>Only one of (labels + namespaces + cluster) or (resource refs) may be provided. If all four are provided, it will be considered an error, and the Status of the top level resource will be updated to reflect an IllegalSelection.<br>Currently only selection on the local cluster is supported, indicated by a nil cluster field.<br>Valid: 1. selector: labels: foo: bar hello: world namespaces: - default cluster: "cluster-name" 2. selector: refs: - name: foo namespace: bar<br>Invalid: 1. selector: labels: foo: bar hello: world namespaces: - default cluster: "cluster-name" refs: - name: foo namespace: bar<br>By default labels will select across all namespaces, unless a list of namespaces is provided, in which case it will only select from those. An empty or nil list is equal to AllNamespaces.<br>If no labels are given, and only namespaces, the full list of resources from the namespace will be selected.<br>The following selector will select all resources with the following labels in every namespace, in the local cluster:<br>selector: labels: foo: bar hello: world<br>Whereas the next selector will only select from the specified namespaces (foo, bar), in the local cluster:<br>selector: labels: foo: bar hello: world namespaces - foo - bar<br>This final selector will select all resources of a given type in the target namespace (foo), in the local cluster:<br>selector namespaces - foo - bar labels: hello: world


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| labels | [][Selector.LabelsEntry](#core.zephyr.solo.io.Selector.LabelsEntry) | repeated | map of labels to match against |
| namespaces | [][string](#string) | repeated | list of namespaces to search |
| cluster | [google.protobuf.StringValue](#google.protobuf.StringValue) |  | If empty string or nil, select in the cluster local to the enclosing resource, else select in the referenced remote cluster |
| refs | [][ResourceRef](#core.zephyr.solo.io.ResourceRef) | repeated | Apply the selector to one or more services by adding their refs here. If the resources are not in the local cluster, the "cluster" field must be populated with the remote cluster name. |






<a name="core.zephyr.solo.io.Selector.LabelsEntry"></a>

### Selector.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

