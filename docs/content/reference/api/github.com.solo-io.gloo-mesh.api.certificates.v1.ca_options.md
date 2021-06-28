
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
  - [CommonCertOptions](#certificates.mesh.gloo.solo.io.CommonCertOptions)
  - [IntermediateCertificateAuthority](#certificates.mesh.gloo.solo.io.IntermediateCertificateAuthority)

  - [CertificateRotationState](#certificates.mesh.gloo.solo.io.CertificateRotationState)






<a name="certificates.mesh.gloo.solo.io.CertificateRotationCondition"></a>

### CertificateRotationCondition
CertificateRotationCondition represents a timesptamped snapshot of the certificate rotation workflow. This is used to keep track of the steps which have been completed thus far.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| timestamp | string |  | The time at which this condition was recorded |
  | state | [certificates.mesh.gloo.solo.io.CertificateRotationState]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.certificates.v1.ca_options#certificates.mesh.gloo.solo.io.CertificateRotationState" >}}) |  | The current state of the cert rotation |
  | message | string |  | A human readable message related to the current condition |
  | errors | []string | repeated | Any errors which occured during the current rotation stage |
  





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
| PENDING | 1 | The CertificateRotation has yet to be picked up by the management-plane. |
| ADDING_NEW_ROOT | 2 | The CertificateRotation is underway, both roots are set, and the new root is being propogated |
| PROPOGATING_NEW_INTERMEDIATE | 3 | The CertificateRotation is underway again. The initial verification is over, the traffic continues to work with both roots present. Now the old root is being removed, and the new root is being propgated alone to the data-plane clusters |
| DELETING_OLD_ROOT | 4 | The CertificateRotation is underway again. Removing the old-root from all data-plane clusters |
| VERIFYING | 5 | Verifying connectivity between workloads, the workflow will not progress until connectivity has been verified. This can either be manual or in the future automated |
| VERIFIED | 6 | The connectivity has been verified. |
| FINISHED | 7 | The rotation has finished, the new root has been propgated to all data-plane clusters, and traffic has been verified successfully. |
| FAILED | 8 | Processing the certificate rotation workflow failed. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

