
---
title: "workload_group.proto"
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
`WorkloadGroup` enables specifying the properties of a single workload for bootstrap and provides a template for `WorkloadEntry`, similar to how `Deployment` specifies properties of workloads via `Pod` templates. A `WorkloadGroup` can have more than one `WorkloadEntry`. `WorkloadGroup` has no relationship to resources which control service registry like `ServiceEntry`  and as such doesn't configure host name for these workloads.<br><!-- crd generation tags +cue-gen:WorkloadGroup:groupName:networking.istio.io +cue-gen:WorkloadGroup:version:v1alpha3 +cue-gen:WorkloadGroup:storageVersion +cue-gen:WorkloadGroup:subresource:status +cue-gen:WorkloadGroup:scope:Namespaced +cue-gen:WorkloadGroup:resource:categories=istio-io,networking-istio-io,shortNames=wg,plural=workloadgroups +cue-gen:WorkloadGroup:printerColumn:name=Age,type=date,JSONPath=.metadata.creationTimestamp,description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC. Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata" --><br><!-- go code generation tags +kubetype-gen +kubetype-gen:groupVersion=networking.istio.io/v1alpha3 +genclient +k8s:deepcopy-gen=true -->


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [istio.networking.v1alpha3.WorkloadGroup.ObjectMeta]({{< ref "workload_group.md#istio.networking.v1alpha3.WorkloadGroup.ObjectMeta" >}}) |  | Metadata that will be used for all corresponding `WorkloadEntries`. User labels for a workload group should be set here in `metadata` rather than in `template`. |
  | template | [istio.networking.v1alpha3.WorkloadEntry]({{< ref "workload_entry.md#istio.networking.v1alpha3.WorkloadEntry" >}}) |  | Template to be used for the generation of `WorkloadEntry` resources that belong to this `WorkloadGroup`. Please note that `address` and `labels` fields should not be set in the template, and an empty `serviceAccount` should default to `default`. The workload identities (mTLS certificates) will be bootstrapped using the specified service account's token. Workload entries in this group will be in the same namespace as the workload group, and inherit the labels and annotations from the above `metadata` field. |
  





<a name="istio.networking.v1alpha3.WorkloadGroup.ObjectMeta"></a>

### WorkloadGroup.ObjectMeta
`ObjectMeta` describes metadata that will be attached to a `WorkloadEntry`. It is a subset of the supported Kubernetes metadata.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| labels | [][istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.LabelsEntry]({{< ref "workload_group.md#istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.LabelsEntry" >}}) | repeated | Labels to attach |
  | annotations | [][istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.AnnotationsEntry]({{< ref "workload_group.md#istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.AnnotationsEntry" >}}) | repeated | Annotations to attach |
  





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

