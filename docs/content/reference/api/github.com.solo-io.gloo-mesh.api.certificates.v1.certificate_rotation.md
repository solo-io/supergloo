
---

title: "certificate_rotation.proto"

---

## Package : `certificates.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for certificate_rotation.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## certificate_rotation.proto


## Table of Contents
  - [CertificateRotationSpec](#certificates.mesh.gloo.solo.io.CertificateRotationSpec)
  - [CertificateRotationStatus](#certificates.mesh.gloo.solo.io.CertificateRotationStatus)
  - [CertificateRotationStatus.Condition](#certificates.mesh.gloo.solo.io.CertificateRotationStatus.Condition)

  - [CertificateRotationStatus.State](#certificates.mesh.gloo.solo.io.CertificateRotationStatus.State)






<a name="certificates.mesh.gloo.solo.io.CertificateRotationSpec"></a>

### CertificateRotationSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| virtualMesh | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | VirtualMesh which you would like to rotate the Certificate for. |
  | newRootCaSecret | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | Reference to a secret containing the new root_ca. |
  





<a name="certificates.mesh.gloo.solo.io.CertificateRotationStatus"></a>

### CertificateRotationStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the CertificateRotation metadata. If the `observedGeneration` does not match `metadata.generation`, the issuer has not processed the most recent version of this request. |
  | error | string |  | Any error observed which prevented the CertificateRotation from being processed. If the error is empty, the request has been processed successfully |
  | state | [certificates.mesh.gloo.solo.io.CertificateRotationStatus.State]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.certificates.v1.certificate_rotation#certificates.mesh.gloo.solo.io.CertificateRotationStatus.State" >}}) |  | The current state of the CertificateRotation workflow reported by the issuer. |
  | conditions | [][certificates.mesh.gloo.solo.io.CertificateRotationStatus.Condition]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.certificates.v1.certificate_rotation#certificates.mesh.gloo.solo.io.CertificateRotationStatus.Condition" >}}) | repeated | A list of all conditions which have occured Examples of reported events may be individual meshes which have been properly rotated. |
  





<a name="certificates.mesh.gloo.solo.io.CertificateRotationStatus.Condition"></a>

### CertificateRotationStatus.Condition
Condition represents a snapshot of the state of the CertificateRotation at a given point in time.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| lastTransitionTime | string |  |  |
  | lastUpdateTime | string |  |  |
  | message | string |  | Human readable message about condition |
  | state | [certificates.mesh.gloo.solo.io.CertificateRotationStatus.State]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.certificates.v1.certificate_rotation#certificates.mesh.gloo.solo.io.CertificateRotationStatus.State" >}}) |  | State observed during condition |
  




 <!-- end messages -->


<a name="certificates.mesh.gloo.solo.io.CertificateRotationStatus.State"></a>

### CertificateRotationStatus.State
Possible states in which a CertificateRotation can exist.

| Name | Number | Description |
| ---- | ------ | ----------- |
| PENDING | 0 | The CertificateRotation has yet to be picked up by the management-plane. |
| ROTATING_MULTI_ROOT | 1 | The CertificateRotation is underway, both roots are set, and the new root is being propogated |
| ROTATING_NEW_ROOT | 2 | The CertificateRotation is underway again. The initial verification is over, the traffic continues to work with both roots present. Now the old root is being removed, and the new root is being propgated alone to the data-plane clusters |
| VERIFYING | 3 | Verifying connectivity between workloads, the workflow will not progress until connectivity has been verified. This can either be manual or in the future automated |
| FINISHED | 4 | The rotation has finished, the new root has been propgated to all data-plane clusters, and traffic has been verified for a 2nd time. |
| FAILED | 5 | Processing the certificate rotation workflow failed. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

