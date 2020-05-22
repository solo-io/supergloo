
---
title: "virtual_mesh.proto"
---

## Package : `networking.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for virtual_mesh.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## virtual_mesh.proto


## Table of Contents
  - [VirtualMeshSpec](#networking.zephyr.solo.io.VirtualMeshSpec)
  - [VirtualMeshSpec.CertificateAuthority](#networking.zephyr.solo.io.VirtualMeshSpec.CertificateAuthority)
  - [VirtualMeshSpec.CertificateAuthority.Builtin](#networking.zephyr.solo.io.VirtualMeshSpec.CertificateAuthority.Builtin)
  - [VirtualMeshSpec.CertificateAuthority.Provided](#networking.zephyr.solo.io.VirtualMeshSpec.CertificateAuthority.Provided)
  - [VirtualMeshSpec.Federation](#networking.zephyr.solo.io.VirtualMeshSpec.Federation)
  - [VirtualMeshSpec.LimitedTrust](#networking.zephyr.solo.io.VirtualMeshSpec.LimitedTrust)
  - [VirtualMeshSpec.SharedTrust](#networking.zephyr.solo.io.VirtualMeshSpec.SharedTrust)
  - [VirtualMeshStatus](#networking.zephyr.solo.io.VirtualMeshStatus)

  - [VirtualMeshSpec.EnforcementPolicy](#networking.zephyr.solo.io.VirtualMeshSpec.EnforcementPolicy)
  - [VirtualMeshSpec.Federation.Mode](#networking.zephyr.solo.io.VirtualMeshSpec.Federation.Mode)






<a name="networking.zephyr.solo.io.VirtualMeshSpec"></a>

### VirtualMeshSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| displayName | [string](#string) |  | User-provided display name for the virtual mesh. |
| meshes | [][core.zephyr.solo.io.ResourceRef](#core.zephyr.solo.io.ResourceRef) | repeated | The meshes contained in this virtual mesh. |
| certificateAuthority | [VirtualMeshSpec.CertificateAuthority](#networking.zephyr.solo.io.VirtualMeshSpec.CertificateAuthority) |  |  |
| federation | [VirtualMeshSpec.Federation](#networking.zephyr.solo.io.VirtualMeshSpec.Federation) |  |  |
| shared | [VirtualMeshSpec.SharedTrust](#networking.zephyr.solo.io.VirtualMeshSpec.SharedTrust) |  |  |
| limited | [VirtualMeshSpec.LimitedTrust](#networking.zephyr.solo.io.VirtualMeshSpec.LimitedTrust) |  |  |
| enforceAccessControl | [VirtualMeshSpec.EnforcementPolicy](#networking.zephyr.solo.io.VirtualMeshSpec.EnforcementPolicy) |  |  |






<a name="networking.zephyr.solo.io.VirtualMeshSpec.CertificateAuthority"></a>

### VirtualMeshSpec.CertificateAuthority



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| builtin | [VirtualMeshSpec.CertificateAuthority.Builtin](#networking.zephyr.solo.io.VirtualMeshSpec.CertificateAuthority.Builtin) |  | Use auto-generated root certificate. |
| provided | [VirtualMeshSpec.CertificateAuthority.Provided](#networking.zephyr.solo.io.VirtualMeshSpec.CertificateAuthority.Provided) |  | Use user-provided root certificate. |






<a name="networking.zephyr.solo.io.VirtualMeshSpec.CertificateAuthority.Builtin"></a>

### VirtualMeshSpec.CertificateAuthority.Builtin
Configuration for auto-generated root certificate unique to the VirtualMesh Uses the X.509 format, RFC5280


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ttlDays | [uint32](#uint32) |  | Number of days before root cert expires. Defaults to 365. |
| rsaKeySizeBytes | [uint32](#uint32) |  | Size in bytes of the root cert's private key. Defaults to 4096 |
| orgName | [string](#string) |  | Root cert organization name. Defaults to "service-mesh-hub" |






<a name="networking.zephyr.solo.io.VirtualMeshSpec.CertificateAuthority.Provided"></a>

### VirtualMeshSpec.CertificateAuthority.Provided
Configuration for user-provided root certificate.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| certificate | [core.zephyr.solo.io.ResourceRef](#core.zephyr.solo.io.ResourceRef) |  | Reference to a Secret object containing the root certificate. |






<a name="networking.zephyr.solo.io.VirtualMeshSpec.Federation"></a>

### VirtualMeshSpec.Federation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| mode | [VirtualMeshSpec.Federation.Mode](#networking.zephyr.solo.io.VirtualMeshSpec.Federation.Mode) |  |  |






<a name="networking.zephyr.solo.io.VirtualMeshSpec.LimitedTrust"></a>

### VirtualMeshSpec.LimitedTrust
Limited trust is a virtual mesh trust model which does not require all meshes sharing the same root certificate or identity model. But rather, the limited trust creates trust between meshes running on different clusters by connecting their ingress/egress gateways with a common cert/identity. In this model all requests between different have the following request path when communicating between clusters<br>cluster 1 MTLS               shared MTLS                  cluster 2 MTLS client/workload <-----------> egress gateway <----------> ingress gateway <--------------> server<br>This approach has the downside of not maintaining identity from client to server, but allows for ad-hoc addition of additional clusters into a virtual mesh.






<a name="networking.zephyr.solo.io.VirtualMeshSpec.SharedTrust"></a>

### VirtualMeshSpec.SharedTrust
Shared trust is a virtual mesh trust model requiring a shared root certificate, as well as shared identity between all entities which wish to communicate within the virtual mesh.<br>The best current example of this would be the replicated control planes example from Istio: https://preliminary.istio.io/docs/setup/install/multicluster/gateways/






<a name="networking.zephyr.solo.io.VirtualMeshStatus"></a>

### VirtualMeshStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| federationStatus | [core.zephyr.solo.io.Status](#core.zephyr.solo.io.Status) |  | Status of the process writing federation decision metadata onto MeshServices. |
| certificateStatus | [core.zephyr.solo.io.Status](#core.zephyr.solo.io.Status) |  | Status of the process signing CSRs. |
| configStatus | [core.zephyr.solo.io.Status](#core.zephyr.solo.io.Status) |  | Overall validation status of this VirtualMesh. |
| accessControlEnforcementStatus | [core.zephyr.solo.io.Status](#core.zephyr.solo.io.Status) |  | Status of ensuring that access control is enforced within this VirtualMesh. |





 <!-- end messages -->


<a name="networking.zephyr.solo.io.VirtualMeshSpec.EnforcementPolicy"></a>

### VirtualMeshSpec.EnforcementPolicy
If ENABLED, by default disallow traffic to all Services in the VirtualMesh unless explicitly allowed through AccessControlPolicies. If DISABLED, by default allow traffic to all Services in the VirtualMesh. If MESH_DEFAULT, the default value depends on the type service mesh: Istio: false Appmesh: true

| Name | Number | Description |
| ---- | ------ | ----------- |
| MESH_DEFAULT | 0 |  |
| ENABLED | 1 |  |
| DISABLED | 2 |  |



<a name="networking.zephyr.solo.io.VirtualMeshSpec.Federation.Mode"></a>

### VirtualMeshSpec.Federation.Mode


| Name | Number | Description |
| ---- | ------ | ----------- |
| PERMISSIVE | 0 | All services in a VirtualMesh will be federated to all workloads in that Virtual Mesh. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

