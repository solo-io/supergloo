
---

title: "certificates.proto"

---

## Package : `common.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for certificates.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## certificates.proto


## Table of Contents
  - [CommonCertOptions](#common.mesh.gloo.solo.io.CommonCertOptions)
  - [VaultCA](#common.mesh.gloo.solo.io.VaultCA)
  - [VaultCA.AppRole](#common.mesh.gloo.solo.io.VaultCA.AppRole)
  - [VaultCA.Kubernetes](#common.mesh.gloo.solo.io.VaultCA.Kubernetes)







<a name="common.mesh.gloo.solo.io.CommonCertOptions"></a>

### CommonCertOptions
Configuration for generating a self-signed root certificate. Uses the X.509 format, RFC5280.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ttlDays | uint32 |  | Number of days before root cert expires. Defaults to 365. |
  | rsaKeySizeBytes | uint32 |  | Size in bytes of the root cert's private key. Defaults to 4096. |
  | orgName | string |  | Root cert organization name. Defaults to "gloo-mesh". |
  





<a name="common.mesh.gloo.solo.io.VaultCA"></a>

### VaultCA



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| caPath | string |  |  |
  | csrPath | string |  |  |
  | server | string |  |  |
  | caBundle | bytes |  |  |
  | tokenSecretRef | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  |  |
  | appRole | [common.mesh.gloo.solo.io.VaultCA.AppRole]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.certificates#common.mesh.gloo.solo.io.VaultCA.AppRole" >}}) |  |  |
  | kubernetesAuth | [common.mesh.gloo.solo.io.VaultCA.Kubernetes]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.certificates#common.mesh.gloo.solo.io.VaultCA.Kubernetes" >}}) |  |  |
  





<a name="common.mesh.gloo.solo.io.VaultCA.AppRole"></a>

### VaultCA.AppRole



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | string |  |  |
  | roleId | string |  |  |
  | secretRef | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  |  |
  





<a name="common.mesh.gloo.solo.io.VaultCA.Kubernetes"></a>

### VaultCA.Kubernetes



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | string |  |  |
  | secretRef | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  |  |
  | role | string |  |  |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

