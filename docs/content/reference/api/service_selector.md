
---
title: "service_selector.proto"
---

## Package : `core.smh.solo.io`



<a name="top"></a>

<a name="API Reference for service_selector.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## service_selector.proto


## Table of Contents
  - [ServiceSelector](#core.smh.solo.io.ServiceSelector)
  - [ServiceSelector.Matcher](#core.smh.solo.io.ServiceSelector.Matcher)
  - [ServiceSelector.Matcher.LabelsEntry](#core.smh.solo.io.ServiceSelector.Matcher.LabelsEntry)
  - [ServiceSelector.ServiceRefs](#core.smh.solo.io.ServiceSelector.ServiceRefs)







<a name="core.smh.solo.io.ServiceSelector"></a>

### ServiceSelector
Select Kubernetes services<br>Only one of (labels + namespaces + cluster) or (resource refs) may be provided. If all four are provided, it will be considered an error, and the Status of the top level resource will be updated to reflect an IllegalSelection.<br>Valid: 1. selector: matcher: labels: foo: bar hello: world namespaces: - default cluster: "cluster-name" 2. selector: matcher: refs: - name: foo namespace: bar<br>Invalid: 1. selector: matcher: labels: foo: bar hello: world namespaces: - default cluster: "cluster-name" refs: - name: foo namespace: bar<br>By default labels will select across all namespaces, unless a list of namespaces is provided, in which case it will only select from those. An empty list is equal to AllNamespaces.<br>If no labels are given, and only namespaces, all resources from the namespaces will be selected.<br>The following selector will select all resources with the following labels in every namespace, in the local cluster:<br>selector: matcher: labels: foo: bar hello: world<br>Whereas the next selector will only select from the specified namespaces (foo, bar), in the local cluster:<br>selector: matcher: labels: foo: bar hello: world namespaces - foo - bar<br>This final selector will select all resources of a given type in the target namespace (foo), in the local cluster:<br>selector matcher: namespaces - foo - bar labels: hello: world


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| matcher | [ServiceSelector.Matcher](#core.smh.solo.io.ServiceSelector.Matcher) |  |  |
| serviceRefs | [ServiceSelector.ServiceRefs](#core.smh.solo.io.ServiceSelector.ServiceRefs) |  |  |






<a name="core.smh.solo.io.ServiceSelector.Matcher"></a>

### ServiceSelector.Matcher



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| labels | [][ServiceSelector.Matcher.LabelsEntry](#core.smh.solo.io.ServiceSelector.Matcher.LabelsEntry) | repeated | If specified, all labels must exist on k8s Service, else match on any labels. |
| namespaces | [][string](#string) | repeated | If specified, match k8s Services if they exist in one of the specified namespaces. If not specified, match on any namespace. |
| clusters | [][string](#string) | repeated | If specified, match k8s Services if they exist in one of the specified clusters. If not specified, match on any cluster. |






<a name="core.smh.solo.io.ServiceSelector.Matcher.LabelsEntry"></a>

### ServiceSelector.Matcher.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="core.smh.solo.io.ServiceSelector.ServiceRefs"></a>

### ServiceSelector.ServiceRefs



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| services | [][ResourceRef](#core.smh.solo.io.ResourceRef) | repeated | Match k8s Services by direct reference. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

