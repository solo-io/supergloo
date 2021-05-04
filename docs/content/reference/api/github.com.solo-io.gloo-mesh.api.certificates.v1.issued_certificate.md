
---

title: "issued_certificate.proto"

---

## Package : `certificates.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for issued_certificate.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## issued_certificate.proto


## Table of Contents
  - [IssuedCertificateSpec](#certificates.mesh.gloo.solo.io.IssuedCertificateSpec)
  - [IssuedCertificateStatus](#certificates.mesh.gloo.solo.io.IssuedCertificateStatus)

  - [IssuedCertificateStatus.State](#certificates.mesh.gloo.solo.io.IssuedCertificateStatus.State)






<a name="certificates.mesh.gloo.solo.io.IssuedCertificateSpec"></a>

### IssuedCertificateSpec
IssuedCertificates are used to issue SSL certificates to remote Kubernetes clusters from a central (out-of-cluster) Certificate Authority.<br>When an IssuedCertificate is created, a certificate is issued to a remote cluster by a central Certificate Authority via the following workflow:<br>1. The Certificate Issuer creates the IssuedCertificate resource on the remote cluster 2. The Certificate Signature Requesting Agent installed to the remote cluster generates a Certificate Signing Request and writes it to the status of the IssuedCertificate 3. Finally, the Certificate Issuer generates signed a certificate for the CSR and writes it back as Kubernetes Secret in the remote cluster.<br>Trust can therefore be established across clusters without requiring private keys to ever leave the node.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hosts | []string | repeated | A list of hostnames and IPs to generate a certificate for. This can also be set to the identity running the workload, e.g. a Kubernetes service account.<br>Generally for an Istio CA this will take the form `spiffe://cluster.local/ns/istio-system/sa/citadel`.<br>"cluster.local" may be replaced by the root of trust domain for the mesh. |
  | org | string |  | The organization for this certificate. Deprecated in favor of `common_cert_options.org` |
  | commonCertOptions | [common.mesh.gloo.solo.io.CommonCertOptions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.certificates#common.mesh.gloo.solo.io.CommonCertOptions" >}}) |  | Common certificate options used to generate this intermediate certificate |
  | signingCertificateSecret | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | The secret containing the root SSL certificate used to sign this IssuedCertificate (located in the certificate issuer's cluster). |
  | vaultCa | [common.mesh.gloo.solo.io.VaultCA]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.certificates#common.mesh.gloo.solo.io.VaultCA" >}}) |  | The vault_ca config which will be used to sign this intermediate cert |
  | issuedCertificateSecret | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | The secret containing the SSL certificate to be generated for this IssuedCertificate (located in the Gloo Mesh agent's cluster). |
  | podBounceDirective | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | A reference to a PodBounceDirective specifying a list of Kubernetes pods to bounce (delete and cause a restart) when the certificate is issued.<br>Istio-controlled pods require restarting in order for Envoy proxies to pick up the newly issued certificate due to [this issue](https://github.com/istio/istio/issues/22993).<br>This will include the control plane pods as well as any Pods which share a data plane with the target mesh. |
  





<a name="certificates.mesh.gloo.solo.io.IssuedCertificateStatus"></a>

### IssuedCertificateStatus
The IssuedCertificate status is written by the CertificateRequesting agent.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the IssuedCertificate metadata. If the `observedGeneration` does not match `metadata.generation`, the Gloo Mesh agent has not processed the most recent version of this IssuedCertificate. |
  | error | string |  | Any error observed which prevented the CertificateRequest from being processed. If the error is empty, the request has been processed successfully. |
  | state | [certificates.mesh.gloo.solo.io.IssuedCertificateStatus.State]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.certificates.v1.issued_certificate#certificates.mesh.gloo.solo.io.IssuedCertificateStatus.State" >}}) |  | The current state of the IssuedCertificate workflow, reported by the agent. |
  




 <!-- end messages -->


<a name="certificates.mesh.gloo.solo.io.IssuedCertificateStatus.State"></a>

### IssuedCertificateStatus.State
Possible states in which an IssuedCertificate can exist.

| Name | Number | Description |
| ---- | ------ | ----------- |
| PENDING | 0 | The IssuedCertificate has yet to be picked up by the agent. |
| REQUESTED | 1 | The agent has created a local private key and a CertificateRequest for the IssuedCertificate. In this state, the agent is waiting for the Issuer to issue certificates for the CertificateRequest before proceeding. |
| ISSUED | 2 | The certificate has been issued. Any pods that require restarting will be restarted at this point. |
| FINISHED | 3 | The reply from the Issuer has been processed and the agent has placed the final certificate secret in the target location specified by the IssuedCertificate. |
| FAILED | 4 | Processing the certificate workflow failed. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

