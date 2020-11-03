
---
title: "virtual_mesh.proto"
---

## Package : `networking.smh.solo.io`



<a name="top"></a>

<a name="API Reference for virtual_mesh.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## virtual_mesh.proto


## Table of Contents
  - [VirtualMeshSpec](#networking.smh.solo.io.VirtualMeshSpec)
  - [VirtualMeshSpec.Federation](#networking.smh.solo.io.VirtualMeshSpec.Federation)
  - [VirtualMeshSpec.MTLSConfig](#networking.smh.solo.io.VirtualMeshSpec.MTLSConfig)
  - [VirtualMeshSpec.MTLSConfig.LimitedTrust](#networking.smh.solo.io.VirtualMeshSpec.MTLSConfig.LimitedTrust)
  - [VirtualMeshSpec.MTLSConfig.SharedTrust](#networking.smh.solo.io.VirtualMeshSpec.MTLSConfig.SharedTrust)
  - [VirtualMeshSpec.RootCertificateAuthority](#networking.smh.solo.io.VirtualMeshSpec.RootCertificateAuthority)
  - [VirtualMeshSpec.RootCertificateAuthority.SelfSignedCert](#networking.smh.solo.io.VirtualMeshSpec.RootCertificateAuthority.SelfSignedCert)
  - [VirtualMeshStatus](#networking.smh.solo.io.VirtualMeshStatus)
  - [VirtualMeshStatus.MeshesEntry](#networking.smh.solo.io.VirtualMeshStatus.MeshesEntry)

  - [VirtualMeshSpec.GlobalAccessPolicy](#networking.smh.solo.io.VirtualMeshSpec.GlobalAccessPolicy)






<a name="networking.smh.solo.io.VirtualMeshSpec"></a>

### VirtualMeshSpec
A VirtualMesh represents a logical grouping of meshes for shared configuration and cross-mesh interoperability.<br>VirtualMeshes are used to configure things like shared trust roots (for mTLS) and federation of traffic targets (for cross-cluster networking).<br>Currently, VirtualMeshes can only be constructed from Istio meshes.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| meshes | []core.skv2.solo.io.ObjectRef | repeated | The meshes contained in this virtual mesh. |
| mtlsConfig | networking.smh.solo.io.VirtualMeshSpec.MTLSConfig |  | Configuration options for managing Mutual-TLS mTLS in a virtual mesh.Sets a shared Certificate Authority across the defined meshes. |
| federation | networking.smh.solo.io.VirtualMeshSpec.Federation |  | Determine how to expose traffic targets to cross-mesh traffic using Service Federation. |
| globalAccessPolicy | networking.smh.solo.io.VirtualMeshSpec.GlobalAccessPolicy |  | Sets an Access Policy for the whole mesh. |






<a name="networking.smh.solo.io.VirtualMeshSpec.Federation"></a>

### VirtualMeshSpec.Federation
In Service Mesh Hub, "federation" refers to the ability to expose traffic targets with a global DNS name for traffic originating from any workload within the virtual mesh.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| permissive | google.protobuf.Empty |  | Select permissive mode to expose all traffic targets in a VirtualMesh to cross-cluster traffic from all workloads in that Virtual Mesh. |






<a name="networking.smh.solo.io.VirtualMeshSpec.MTLSConfig"></a>

### VirtualMeshSpec.MTLSConfig
Mutual TLS Config for a Virtual Mesh. This includes options for configuring Mutual TLS within an indvidual mesh, as well as enabling mTLS across Meshes by establishing cross-mesh trust.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| shared | networking.smh.solo.io.VirtualMeshSpec.MTLSConfig.SharedTrust |  | Shared trust (allow communication between any workloads and traffic targets in the grouped Meshes). |
| limited | networking.smh.solo.io.VirtualMeshSpec.MTLSConfig.LimitedTrust |  | Limited trust (selectively allow communication between workloads and traffic targets in the grouped Meshes). |
| autoRestartPods | bool |  | Allow Service Mesh Hub to restart mesh pods when certificates are rotated. If this option is not explicitly enabled, users must restart the pods manually for the new certificates to be picked up. `meshctl` provides the command `meshctl mesh restart` to simplify this process. |






<a name="networking.smh.solo.io.VirtualMeshSpec.MTLSConfig.LimitedTrust"></a>

### VirtualMeshSpec.MTLSConfig.LimitedTrust
Limited trust is a virtual mesh trust model which does not require all meshes sharing the same root certificate or identity model. But rather, the limited trust creates trust between meshes running on different clusters by connecting their ingress/egress gateways with a common cert/identity. In this model all requests between different have the following request path when communicating between clusters<br>cluster 1 MTLS               shared MTLS                  cluster 2 MTLS client/workload <-----------> egress gateway <----------> ingress gateway <--------------> server<br>This approach has the downside of not maintaining identity from client to server, but allows for ad-hoc addition of additional clusters into a virtual mesh.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rootCertificateAuthority | networking.smh.solo.io.VirtualMeshSpec.RootCertificateAuthority |  | Configure a Root Certificate Authority which will be shared by the members of the virtual mesh. If this is not provided, a self-signed certificate will be used by Service Mesh Hub to establish shared trust for the purposes of failover and federation. |






<a name="networking.smh.solo.io.VirtualMeshSpec.MTLSConfig.SharedTrust"></a>

### VirtualMeshSpec.MTLSConfig.SharedTrust
Shared trust is a virtual mesh trust model requiring a shared root certificate, as well as shared identity between all entities which wish to communicate within the virtual mesh.<br>The best current example of this would be the replicated control planes example from Istio: https://preliminary.istio.io/docs/setup/install/multicluster/gateways/


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rootCertificateAuthority | networking.smh.solo.io.VirtualMeshSpec.RootCertificateAuthority |  | Configure a Root Certificate Authority which will be shared by the members of the virtual mesh. If this is not provided, a self-signed certificate will be used by Service Mesh Hub to establish shared trust for the purposes of failover and federation. |






<a name="networking.smh.solo.io.VirtualMeshSpec.RootCertificateAuthority"></a>

### VirtualMeshSpec.RootCertificateAuthority
RootCertificateAuthority defines parameters for configuring the root CA for a Virtual Mesh.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| generated | networking.smh.solo.io.VirtualMeshSpec.RootCertificateAuthority.SelfSignedCert |  | Generate a self-signed root certificate with the given options. |
| secret | core.skv2.solo.io.ObjectRef |  | Use a root certificate provided in a Kubernetes Secret. [Secrets provided in this way must follow a specified format, documented here.]({{% versioned_link_path fromRoot="/guides/federate_identity/" %}}) |






<a name="networking.smh.solo.io.VirtualMeshSpec.RootCertificateAuthority.SelfSignedCert"></a>

### VirtualMeshSpec.RootCertificateAuthority.SelfSignedCert
Configuration for generating a self-signed root certificate. Uses the X.509 format, RFC5280.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ttlDays | uint32 |  | Number of days before root cert expires. Defaults to 365. |
| rsaKeySizeBytes | uint32 |  | Size in bytes of the root cert's private key. Defaults to 4096. |
| orgName | string |  | Root cert organization name. Defaults to "service-mesh-hub". |






<a name="networking.smh.solo.io.VirtualMeshStatus"></a>

### VirtualMeshStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the VirtualMesh metadata. If the observedGeneration does not match generation, the controller has not received the most recent version of this resource. |
| state | networking.smh.solo.io.ApprovalState |  | The state of the overall resource. It will only show accepted if it has been successfully applied to all target meshes. |
| errors | []string | repeated | Any errors found while processing this generation of the resource. |
| meshes | []networking.smh.solo.io.VirtualMeshStatus.MeshesEntry | repeated | The status of the VirtualMesh for each Mesh to which it has been applied. A VirtualMesh may be Accepted for some Meshes and rejected for others. |






<a name="networking.smh.solo.io.VirtualMeshStatus.MeshesEntry"></a>

### VirtualMeshStatus.MeshesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
| value | networking.smh.solo.io.ApprovalStatus |  |  |





 <!-- end messages -->


<a name="networking.smh.solo.io.VirtualMeshSpec.GlobalAccessPolicy"></a>

### VirtualMeshSpec.GlobalAccessPolicy
If ENABLED, by default disallow traffic to all Services in the VirtualMesh unless explicitly allowed through AccessControlPolicies. If DISABLED, by default allow traffic to all Services in the VirtualMesh. If MESH_DEFAULT, the default value depends on the type service mesh: Istio: false Appmesh: true

| Name | Number | Description |
| ---- | ------ | ----------- |
| MESH_DEFAULT | 0 |  |
| ENABLED | 1 |  |
| DISABLED | 2 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

