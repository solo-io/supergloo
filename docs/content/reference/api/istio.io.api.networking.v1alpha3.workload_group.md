
---

---

## Package : `istio.networking.v1alpha3`



<a name="top"></a>

<a name="API Reference for workload_group.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## workload_group.proto


## Table of Contents
  - [WorkloadGroup](#istio.networking.v1alpha3.WorkloadGroup)
  - [WorkloadGroup.ObjectMeta](#istio.networking.v1alpha3.WorkloadGroup.ObjectMeta)
  - [WorkloadGroup.ObjectMeta.AnnotationsEntry](#istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.AnnotationsEntry)
  - [WorkloadGroup.ObjectMeta.LabelsEntry](#istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.LabelsEntry)







<a name="istio.networking.v1alpha3.WorkloadGroup"></a>

### WorkloadGroup



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [istio.networking.v1alpha3.WorkloadGroup.ObjectMeta]({{< ref "istio.io.api.networking.v1alpha3.workload_group.md#istio.networking.v1alpha3.WorkloadGroup.ObjectMeta" >}}) |  | Metadata that will be used for all corresponding `WorkloadEntries`. User labels for a workload group should be set here in `metadata` rather than in `template`. |
  | template | [istio.networking.v1alpha3.WorkloadEntry]({{< ref "istio.io.api.networking.v1alpha3.workload_entry.md#istio.networking.v1alpha3.WorkloadEntry" >}}) |  | Template to be used for the generation of `WorkloadEntry` resources that belong to this `WorkloadGroup`. Please note that `address` and `labels` fields should not be set in the template, and an empty `serviceAccount` should default to `default`. The workload identities (mTLS certificates) will be bootstrapped using the specified service account's token. Workload entries in this group will be in the same namespace as the workload group, and inherit the labels and annotations from the above `metadata` field. |
  





<a name="istio.networking.v1alpha3.WorkloadGroup.ObjectMeta"></a>

### WorkloadGroup.ObjectMeta



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| labels | [][istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.LabelsEntry]({{< ref "istio.io.api.networking.v1alpha3.workload_group.md#istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.LabelsEntry" >}}) | repeated | Labels to attach |
  | annotations | [][istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.AnnotationsEntry]({{< ref "istio.io.api.networking.v1alpha3.workload_group.md#istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.AnnotationsEntry" >}}) | repeated | Annotations to attach |
  





<a name="istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.AnnotationsEntry"></a>

### WorkloadGroup.ObjectMeta.AnnotationsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.LabelsEntry"></a>

### WorkloadGroup.ObjectMeta.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

