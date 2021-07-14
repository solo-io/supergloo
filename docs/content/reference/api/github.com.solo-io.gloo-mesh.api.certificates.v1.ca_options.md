
---

title: "ca_options.proto"

---

## Package : `certificates.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for ca_options.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## ca_options.proto


## Table of Contents
  - [CertificateRotationCondition](#certificates.mesh.gloo.solo.io.CertificateRotationCondition)
  - [CertificateRotationVerificationMethod](#certificates.mesh.gloo.solo.io.CertificateRotationVerificationMethod)
  - [CommonCertOptions](#certificates.mesh.gloo.solo.io.CommonCertOptions)
  - [IntermediateCertificateAuthority](#certificates.mesh.gloo.solo.io.IntermediateCertificateAuthority)

  - [CertificateRotationState](#certificates.mesh.gloo.solo.io.CertificateRotationState)
  - [CertificateRotationStrategy](#certificates.mesh.gloo.solo.io.CertificateRotationStrategy)






<a name="certificates.mesh.gloo.solo.io.CertificateRotationCondition"></a>

### CertificateRotationCondition
CertificateRotationCondition represents a timesptamped snapshot of the certificate rotation workflow. This is used to keep track of the steps which have been completed thus far.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| timestamp | string |  | The time at which this condition was recorded |
  | state | [certificates.mesh.gloo.solo.io.CertificateRotationState]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.certificates.v1.ca_options#certificates.mesh.gloo.solo.io.CertificateRotationState" >}}) |  | The current state of the cert rotation |
  | message | string |  | A human readable message related to the current condition |
  | errors | []string | repeated | Any errors which occurred during the current rotation stage |
  





<a name="certificates.mesh.gloo.solo.io.CertificateRotationVerificationMethod"></a>

### CertificateRotationVerificationMethod



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| none | [google.protobuf.Empty]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.empty#google.protobuf.Empty" >}}) |  | Verification not enabled. NOTE: This setting is only recommended for testing. When enabled rotation will continue from step to step without any kind of verification. |
  | manual | [google.protobuf.Empty]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.empty#google.protobuf.Empty" >}}) |  | Verification must be completed manually. This involves using our certificate verification  endpoint when the certificates are in a VERIFYING state |
  





<a name="certificates.mesh.gloo.solo.io.CommonCertOptions"></a>

### CommonCertOptions
Configuration for generating a self-signed root certificate. Uses the X.509 format, RFC5280.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ttlDays | uint32 |  | Number of days before root cert expires. Defaults to 365. |
  | rsaKeySizeBytes | uint32 |  | Size in bytes of the root cert's private key. Defaults to 4096. |
  | orgName | string |  | Root cert organization name. Defaults to "gloo-mesh". |
  | secretRotationGracePeriodRatio | float |  | The ratio of cert lifetime to refresh a cert. For example, at 0.10 and 1 hour TTL, we would refresh 6 minutes before expiration |
  





<a name="certificates.mesh.gloo.solo.io.IntermediateCertificateAuthority"></a>

### IntermediateCertificateAuthority
Specify parameters for configuring the root certificate authority for a VirtualMesh.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| vault | [certificates.mesh.gloo.solo.io.VaultCA]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.certificates.v1.vault_ca#certificates.mesh.gloo.solo.io.VaultCA" >}}) |  | Use vault as the intermediate CA source |
  




 <!-- end messages -->


<a name="certificates.mesh.gloo.solo.io.CertificateRotationState"></a>

### CertificateRotationState
State of Certificate Rotation Possible states in which a CertificateRotation can exist.

| Name | Number | Description |
| ---- | ------ | ----------- |
| NOT_APPLICABLE | 0 | No Certificate rotation is currently happening |
| ADDING_NEW_ROOT | 1 | The CertificateRotation is underway, both roots are set, and the new root is being propagated |
| PROPAGATING_NEW_INTERMEDIATE | 2 | The CertificateRotation is underway again. The initial verification is over, the traffic continues to work with both roots present. Now the old root is being removed, and the new root is being propagated alone to the data-plane clusters |
| DELETING_OLD_ROOT | 3 | The CertificateRotation is underway again. Removing the old-root from all data-plane clusters |
| VERIFYING | 4 | Verifying connectivity between workloads, the workflow will not progress until connectivity has been verified. This can either be manual or in the future automated |
| VERIFIED | 5 | The connectivity has been verified. |
| ROLLING_BACK | 6 | The connectivity has been deemed to not be functioning properly, rolling back to the last known good state. |
| FINISHED | 7 | The rotation has finished, the new root has been propagated to all data-plane clusters, and traffic has been verified successfully. |
| FAILED | 8 | Processing the certificate rotation workflow failed. |



<a name="certificates.mesh.gloo.solo.io.CertificateRotationStrategy"></a>

### CertificateRotationStrategy


| Name | Number | Description |
| ---- | ------ | ----------- |
| MULTI_ROOT | 0 | The default certificate rotation strategy. This strategy involves three steps which ensure that traffic in the mesh will experience no downtime. For an in depth explination of how this strategy works in Istio see the [following blog](https://blog.christianposta.com/diving-into-istio-1-6-certificate-rotation/) The steps are as follows: 1. ADDING_NEW_ROOT    During this step the new root-cert will be appended to the old root-cert, and then distributed.    The intermediate will continue to be signed by the original root. 2. PROPAGATING_NEW_INTERMEDIATE    During this step both root-certs will still be distributed. In addition the intermediate will now    be signed by the new root key. 3. DELETING_OLD_ROOT    During this step the old root is no longer included, and the intermediate will continue to be signed    by the new root key. |
| NONE | 1 | Do not use any rotation strategy. NOTE: This can lead to downtime while workloads transition from one root of trust to another |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

