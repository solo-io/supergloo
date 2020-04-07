
---
title: "core.zephyr.solo.iogithub.com/solo-io/service-mesh-hub/api/core/v1alpha1/workload_selector.proto"
---

## Package : `core.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for github.com/solo-io/service-mesh-hub/api/core/v1alpha1/workload_selector.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## github.com/solo-io/service-mesh-hub/api/core/v1alpha1/workload_selector.proto


## Table of Contents
  - [WorkloadSelector](#core.zephyr.solo.io.WorkloadSelector)
  - [WorkloadSelector.LabelsEntry](#core.zephyr.solo.io.WorkloadSelector.LabelsEntry)







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

