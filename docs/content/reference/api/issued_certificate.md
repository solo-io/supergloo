
---
title: "issued_certificate.proto"
---

## Package : `certificates.smh.solo.io`



<a name="top"></a>

<a name="API Reference for issued_certificate.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## issued_certificate.proto


## Table of Contents
  - [IssuedCertificateSpec](#certificates.smh.solo.io.IssuedCertificateSpec)
  - [IssuedCertificateSpec.PodSelector](#certificates.smh.solo.io.IssuedCertificateSpec.PodSelector)
  - [IssuedCertificateSpec.PodSelector.LabelsEntry](#certificates.smh.solo.io.IssuedCertificateSpec.PodSelector.LabelsEntry)
  - [IssuedCertificateStatus](#certificates.smh.solo.io.IssuedCertificateStatus)

  - [IssuedCertificateStatus.State](#certificates.smh.solo.io.IssuedCertificateStatus.State)






<a name="certificates.smh.solo.io.IssuedCertificateSpec"></a>

### IssuedCertificateSpec
IssuedCertificates are used to issue SSL certificates to remote Kubernetes clusters from a central (out-of-cluster) Certificate Authority.<br>When an IssuedCertificate is created, a certificate is issued to a remote cluster by a central Certificate Authority via the following workflow: - The Certificate Issuer creates the IssuedCertificate resource on the remote cluster - The Certificate Signature Requesting Agent installed to the remote cluster generates a Certificate Signing Request and writes it to the status of the IssuedCertificate - Finally, the Certificate Issuer generates signed a certificate for the CSR and writes it back as Secret in the remote cluster.<br>Shared trust can therefore be established across clusters without requiring private keys to ever leave the node.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hosts | [][string](#string) | repeated | list of hostnames and IPs to generate a certificate for. This can also be set to the identity running the workload, like kubernetes service account.<br>Generally for an Istio CA this will take the form `spiffe://cluster.local/ns/istio-system/sa/citadel`.<br>"cluster.local" may be replaced by the root of trust domain for the mesh. |
| org | [string](#string) |  | Organization for this certificate. |
| signingCertificateSecret | [core.skv2.solo.io.ObjectRef](#core.skv2.solo.io.ObjectRef) |  | the secret containing the root SSL certificate used to sign this IssuedCertificate (located in the Certificate Issuer's cluster) |
| issuedCertificateSecret | [core.skv2.solo.io.ObjectRef](#core.skv2.solo.io.ObjectRef) |  | the secret containing the SSL certificate to be generated for this IssuedCertificate (located in the Certificate Agent's cluster) |
| podsToBounce | [][IssuedCertificateSpec.PodSelector](#certificates.smh.solo.io.IssuedCertificateSpec.PodSelector) | repeated | a list of k8s pods to bounce (delete and cause a restart) when the certificate is issued. this will include the control plane pods as well as any pods which share a data plane with the target mesh. |






<a name="certificates.smh.solo.io.IssuedCertificateSpec.PodSelector"></a>

### IssuedCertificateSpec.PodSelector
select pods for deletion


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| namespace | [string](#string) |  | namespace in which the pods live |
| labels | [][IssuedCertificateSpec.PodSelector.LabelsEntry](#certificates.smh.solo.io.IssuedCertificateSpec.PodSelector.LabelsEntry) | repeated | labels shared by the pods |






<a name="certificates.smh.solo.io.IssuedCertificateSpec.PodSelector.LabelsEntry"></a>

### IssuedCertificateSpec.PodSelector.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="certificates.smh.solo.io.IssuedCertificateStatus"></a>

### IssuedCertificateStatus
the IssuedCertificate status is written by the CertificateRequesting agent


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | [int64](#int64) |  | The most recent generation observed in the the IssuedCertificate metadata. if the observedGeneration does not match generation, the Certificate Requesting Agent has not processed the most recent version of this IssuedCertificate. |
| error | [string](#string) |  | any error observed which prevented the CertificateRequest from being processed. if the error is empty, the request has been processed successfully |
| state | [IssuedCertificateStatus.State](#certificates.smh.solo.io.IssuedCertificateStatus.State) |  | the current state of the IssuedCertificate workflow, reported by the agent. |





 <!-- end messages -->


<a name="certificates.smh.solo.io.IssuedCertificateStatus.State"></a>

### IssuedCertificateStatus.State
possible states in which an IssuedCertificate can exist

| Name | Number | Description |
| ---- | ------ | ----------- |
| PENDING | 0 | the IssuedCertificate has yet to be picked up by the agent. |
| REQUESTED | 1 | the agent has created a local private key and a CertificateRequest for the IssuedCertificate. in this state, the agent is waiting for the Issuer to issue certificates for the CertificateRequest before proceeding. |
| ISSUED | 2 | the certificate has been issued. if any pods require being restarted, they'll be restarted at this point. |
| FINISHED | 3 | the reply from the Issuer has been process and the agent has placed the final certificate secret in the target location specified by the IssuedCertificate. |
| FAILED | 4 | processing the certificate workflow failed. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

