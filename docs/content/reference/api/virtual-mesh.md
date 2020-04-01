
---
title: "networking.zephyr.solo.iogithub.com/solo-io/mesh-projects/api/networking/v1alpha1/virtual-mesh.proto"
---

## Package : `networking.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for github.com/solo-io/mesh-projects/api/networking/v1alpha1/virtual-mesh.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## github.com/solo-io/mesh-projects/api/networking/v1alpha1/virtual-mesh.proto


## Table of Contents
  - [CertificateAuthority](#networking.zephyr.solo.io.CertificateAuthority)
  - [CertificateAuthority.Builtin](#networking.zephyr.solo.io.CertificateAuthority.Builtin)
  - [CertificateAuthority.Provided](#networking.zephyr.solo.io.CertificateAuthority.Provided)
  - [Federation](#networking.zephyr.solo.io.Federation)
  - [VirtualMeshSpec](#networking.zephyr.solo.io.VirtualMeshSpec)
  - [VirtualMeshSpec.LimitedTrust](#networking.zephyr.solo.io.VirtualMeshSpec.LimitedTrust)
  - [VirtualMeshSpec.SharedTrust](#networking.zephyr.solo.io.VirtualMeshSpec.SharedTrust)
  - [VirtualMeshStatus](#networking.zephyr.solo.io.VirtualMeshStatus)

  - [Federation.Mode](#networking.zephyr.solo.io.Federation.Mode)






<a name="networking.zephyr.solo.io.CertificateAuthority"></a>

### CertificateAuthority



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| builtin | [CertificateAuthority.Builtin](#networking.zephyr.solo.io.CertificateAuthority.Builtin) |  | Use auto-generated root certificate |
| provided | [CertificateAuthority.Provided](#networking.zephyr.solo.io.CertificateAuthority.Provided) |  | Use user-provided root certificate |






<a name="networking.zephyr.solo.io.CertificateAuthority.Builtin"></a>

### CertificateAuthority.Builtin
Configuration for auto-generated root certificate unique to the VirtualMesh Uses the X.509 format, RFC5280


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ttlDays | [uint32](#uint32) |  | Number of days before root cert expires. Defaults to 365. |
| rsaKeySizeBytes | [uint32](#uint32) |  | Size in bytes of the root cert's private key. Defaults to 4096 |
| orgName | [string](#string) |  | Root cert organization name. Defaults to "service-mesh-hub" |






<a name="networking.zephyr.solo.io.CertificateAuthority.Provided"></a>

### CertificateAuthority.Provided
Configuration for user-provided root certificate


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| certificate | [core.zephyr.solo.io.ResourceRef](#core.zephyr.solo.io.ResourceRef) |  | Resource reference for a Secret object containing the root certificate certificate: name: cacerts namespace: service-mesh-hub (default write namespace) |






<a name="networking.zephyr.solo.io.Federation"></a>

### Federation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| mode | [Federation.Mode](#networking.zephyr.solo.io.Federation.Mode) |  |  |






<a name="networking.zephyr.solo.io.VirtualMeshSpec"></a>

### VirtualMeshSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| displayName | [string](#string) |  | User-provided display name for the virtual mesh. |
| meshes | [][core.zephyr.solo.io.ResourceRef](#core.zephyr.solo.io.ResourceRef) | repeated | The meshes contained in this virtual mesh. |
| certificateAuthority | [CertificateAuthority](#networking.zephyr.solo.io.CertificateAuthority) |  |  |
| federation | [Federation](#networking.zephyr.solo.io.Federation) |  |  |
| shared | [VirtualMeshSpec.SharedTrust](#networking.zephyr.solo.io.VirtualMeshSpec.SharedTrust) |  |  |
| limited | [VirtualMeshSpec.LimitedTrust](#networking.zephyr.solo.io.VirtualMeshSpec.LimitedTrust) |  |  |
| enforceAccessControl | [bool](#bool) |  | If true, by default disallow traffic to all Services in the VirtualMesh unless explicitly allowed through AccessControlPolicies. If false, by default allow traffic to all Services in the VirtualMesh. Defaults to false when not set. |






<a name="networking.zephyr.solo.io.VirtualMeshSpec.LimitedTrust"></a>

### VirtualMeshSpec.LimitedTrust
Limited trust is a virtual mesh trust model which does not require all meshes sharing the same root certificate or identity model. But rather, the limited trust creates trust between meshes running on different clusters by connecting their ingress/egress gateways with a common cert/identity. In this model all requests between different have the following request path when communicating between clusters<br>cluster 1 MTLS               shared MTLS                  cluster 2 MTLS client/workload <-----------> egress gateway <----------> ingress gateway <--------------> server<br>This approach has the downside of not maintaining identity from client to server, but allows for ad-hoc addition of additional clusters into a virtual mesh.






<a name="networking.zephyr.solo.io.VirtualMeshSpec.SharedTrust"></a>

### VirtualMeshSpec.SharedTrust
Shared trust is a virtual mesh trust model requiring a shared root certificate, as well as shared identity between all entities which wish to communicate within the virtual mesh.<br>The best current example of this would be the replicated control planes example from istio: https://preliminary.istio.io/docs/setup/install/multicluster/gateways/






<a name="networking.zephyr.solo.io.VirtualMeshStatus"></a>

### VirtualMeshStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| federationStatus | [core.zephyr.solo.io.ComputedStatus](#core.zephyr.solo.io.ComputedStatus) |  |  |
| certificateStatus | [core.zephyr.solo.io.ComputedStatus](#core.zephyr.solo.io.ComputedStatus) |  |  |
| configStatus | [core.zephyr.solo.io.ComputedStatus](#core.zephyr.solo.io.ComputedStatus) |  |  |
| accessControlEnforcementStatus | [core.zephyr.solo.io.ComputedStatus](#core.zephyr.solo.io.ComputedStatus) |  |  |





 <!-- end messages -->


<a name="networking.zephyr.solo.io.Federation.Mode"></a>

### Federation.Mode


| Name | Number | Description |
| ---- | ------ | ----------- |
| PERMISSIVE | 0 | federate everything to everybody |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

