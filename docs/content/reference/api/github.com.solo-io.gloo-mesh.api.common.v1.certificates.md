
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
| caPath | string |  | ca_path is the mount path of the Vault PKI backend's `sign` endpoint, e.g: "my_pki_mount/sign/my-role-name". |
  | csrPath | string |  | ca_path is the mount path of the Vault PKI backend's `generate` endpoint, e.g: "my_pki_mount/intermediate/generate/exported". exported is necessary here as istio needs access to the private key |
  | server | string |  | Server is the connection address for the Vault server, e.g: "https://vault.example.com:8200". |
  | caBundle | bytes |  | PEM encoded CA bundle used to validate Vault server certificate. Only used if the Server URL is using HTTPS protocol. This parameter is ignored for plain HTTP protocol connection. If not set the system root certificates are used to validate the TLS connection. |
  | namespace | string |  | Name of the vault namespace. Namespaces is a set of features within Vault Enterprise that allows Vault environments to support Secure Multi-tenancy. e.g: "ns1" More about namespaces can be found here https://www.vaultproject.io/docs/enterprise/namespaces |
  | tokenSecretRef | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | TokenSecretRef authenticates with Vault by presenting a token. |
  | kubernetesAuth | [common.mesh.gloo.solo.io.VaultCA.Kubernetes]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.certificates#common.mesh.gloo.solo.io.VaultCA.Kubernetes" >}}) |  | Kubernetes authenticates with Vault by passing the ServiceAccount token stored in the named Secret resource to the Vault server. |
  





<a name="common.mesh.gloo.solo.io.VaultCA.Kubernetes"></a>

### VaultCA.Kubernetes



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | string |  |  |
  | secretRef | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  |  |
  | secretTokenKey | string |  |  |
  | role | string |  |  |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

