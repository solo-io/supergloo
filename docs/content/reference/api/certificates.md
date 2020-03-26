
---
title: "security.zephyr.solo.iogithub.com/solo-io/mesh-projects/api/security/v1alpha1/certificates.proto"
---

## Package : `security.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for github.com/solo-io/mesh-projects/api/security/v1alpha1/certificates.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## github.com/solo-io/mesh-projects/api/security/v1alpha1/certificates.proto


## Table of Contents
  - [CertConfig](#security.zephyr.solo.io.CertConfig)
  - [ThirdPartyApprovalWorkflow](#security.zephyr.solo.io.ThirdPartyApprovalWorkflow)
  - [VirtualMeshCertificateSigningRequestSpec](#security.zephyr.solo.io.VirtualMeshCertificateSigningRequestSpec)
  - [VirtualMeshCertificateSigningRequestStatus](#security.zephyr.solo.io.VirtualMeshCertificateSigningRequestStatus)
  - [VirtualMeshCertificateSigningResponse](#security.zephyr.solo.io.VirtualMeshCertificateSigningResponse)

  - [ThirdPartyApprovalWorkflow.ApprovalStatus](#security.zephyr.solo.io.ThirdPartyApprovalWorkflow.ApprovalStatus)






<a name="security.zephyr.solo.io.CertConfig"></a>

### CertConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hosts | [][string](#string) | repeated | list of hostnames and IPs to generate a certificate for. This can also be set to the identity running the workload, like kubernetes service account.<br>Generally for an istio CA this will take the values: 1. spiffe://cluster.local/ns/istio-system/sa/citadel 2. localhost<br>The cluster.local may be replaced by the root of trust domain for the mesh |
| org | [string](#string) |  | Organization for this certificate. |
| meshType | [core.zephyr.solo.io.MeshType](#core.zephyr.solo.io.MeshType) |  | In the future, the type of mesh, and level of trust will need to be specified here, but for the time being we are only supporting shared trust in istio. |






<a name="security.zephyr.solo.io.ThirdPartyApprovalWorkflow"></a>

### ThirdPartyApprovalWorkflow



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| lastUpdatedTime | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | time when the status was last updated |
| message | [string](#string) |  | a user readable message regarding the status of the CSR |
| approvalStatus | [ThirdPartyApprovalWorkflow.ApprovalStatus](#security.zephyr.solo.io.ThirdPartyApprovalWorkflow.ApprovalStatus) |  |  |






<a name="security.zephyr.solo.io.VirtualMeshCertificateSigningRequestSpec"></a>

### VirtualMeshCertificateSigningRequestSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| csrData | [bytes](#bytes) |  | Base64-encoded PKCS#10 CSR data |
| certConfig | [CertConfig](#security.zephyr.solo.io.CertConfig) |  |  |
| virtualMeshRef | [core.zephyr.solo.io.ResourceRef](#core.zephyr.solo.io.ResourceRef) |  | Reference to the virtual mesh which this CSR corresponds to. This is important as it allows the virtual mesh operator to know which trust bundle to use when signing the new certificates.<br>When the CSR is first created by the Virtual Mesh operator, this data will be added by it. However, during a cert rotation scenario this is not possible. Therefore, the csr-agent will write this data to the secret so that it can be retrieved when the cert is going to expire. TODO: Decide how exactly we want to store this data. |






<a name="security.zephyr.solo.io.VirtualMeshCertificateSigningRequestStatus"></a>

### VirtualMeshCertificateSigningRequestStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| response | [VirtualMeshCertificateSigningResponse](#security.zephyr.solo.io.VirtualMeshCertificateSigningResponse) |  | Response from the certificate authority |
| thirdPartyApproval | [ThirdPartyApprovalWorkflow](#security.zephyr.solo.io.ThirdPartyApprovalWorkflow) |  | Workflow for approving Certificate Signing Requests |
| computedStatus | [core.zephyr.solo.io.ComputedStatus](#core.zephyr.solo.io.ComputedStatus) |  |  |






<a name="security.zephyr.solo.io.VirtualMeshCertificateSigningResponse"></a>

### VirtualMeshCertificateSigningResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| caCertificate | [bytes](#bytes) |  | If request was approved, the controller will place the issued certificate here. |
| rootCertificate | [bytes](#bytes) |  | root cert shared by all clusters, safe to send over the wire |





 <!-- end messages -->


<a name="security.zephyr.solo.io.ThirdPartyApprovalWorkflow.ApprovalStatus"></a>

### ThirdPartyApprovalWorkflow.ApprovalStatus


| Name | Number | Description |
| ---- | ------ | ----------- |
| PENDING | 0 | have a default value which represents not being set as proto enums require a default 0th value |
| APPROVED | 1 |  |
| DENIED | 2 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

