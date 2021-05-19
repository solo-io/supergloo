
---

title: "ca_options.proto"

---

## Package : `certificates.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for ca_options.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## ca_options.proto


## Table of Contents
  - [CommonCertOptions](#certificates.mesh.gloo.solo.io.CommonCertOptions)
  - [IntermediateCertificateAuthority](#certificates.mesh.gloo.solo.io.IntermediateCertificateAuthority)







<a name="certificates.mesh.gloo.solo.io.CommonCertOptions"></a>

### CommonCertOptions
Configuration for generating a self-signed root certificate. Uses the X.509 format, RFC5280.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ttlDays | uint32 |  | Number of days before root cert expires. Defaults to 365. |
  | rsaKeySizeBytes | uint32 |  | Size in bytes of the root cert's private key. Defaults to 4096. |
  | orgName | string |  | Root cert organization name. Defaults to "gloo-mesh". |
  





<a name="certificates.mesh.gloo.solo.io.IntermediateCertificateAuthority"></a>

### IntermediateCertificateAuthority
Specify parameters for configuring the root certificate authority for a VirtualMesh.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| vault | [certificates.mesh.gloo.solo.io.VaultCA]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.certificates.v1.vault_ca#certificates.mesh.gloo.solo.io.VaultCA" >}}) |  | Use vault as the intermediate CA source |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

