
---
title: "pod_bounce_directive.proto"
---

## Package : `certificates.smh.solo.io`



<a name="top"></a>

<a name="API Reference for pod_bounce_directive.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## pod_bounce_directive.proto


## Table of Contents
  - [PodBounceDirectiveSpec](#certificates.smh.solo.io.PodBounceDirectiveSpec)
  - [PodBounceDirectiveSpec.PodSelector](#certificates.smh.solo.io.PodBounceDirectiveSpec.PodSelector)
  - [PodBounceDirectiveSpec.PodSelector.LabelsEntry](#certificates.smh.solo.io.PodBounceDirectiveSpec.PodSelector.LabelsEntry)







<a name="certificates.smh.solo.io.PodBounceDirectiveSpec"></a>

### PodBounceDirectiveSpec
When certificates are issued, pods may need to be bounced (restarted) to ensure they pick up the new certificates. If so, the certificate Issuer will create a PodBounceDirective containing the namespaces and labels of the pods that need to be bounced in order to pick up the new certs.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| podsToBounce | [][PodBounceDirectiveSpec.PodSelector](#certificates.smh.solo.io.PodBounceDirectiveSpec.PodSelector) | repeated | A list of k8s pods to bounce (delete and cause a restart) when the certificate is issued. This will include the control plane pods as well as any pods which share a data plane with the target mesh. |






<a name="certificates.smh.solo.io.PodBounceDirectiveSpec.PodSelector"></a>

### PodBounceDirectiveSpec.PodSelector
Pods that will be restarted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| namespace | [string](#string) |  | The namespace in which the pods live. |
| labels | [][PodBounceDirectiveSpec.PodSelector.LabelsEntry](#certificates.smh.solo.io.PodBounceDirectiveSpec.PodSelector.LabelsEntry) | repeated | Any labels shared by the pods. |






<a name="certificates.smh.solo.io.PodBounceDirectiveSpec.PodSelector.LabelsEntry"></a>

### PodBounceDirectiveSpec.PodSelector.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

