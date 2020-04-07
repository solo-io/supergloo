
---
title: "discovery.zephyr.solo.iogithub.com/solo-io/service-mesh-hub/api/discovery/v1alpha1/mesh.proto"
---

## Package : `discovery.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for github.com/solo-io/service-mesh-hub/api/discovery/v1alpha1/mesh.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## github.com/solo-io/service-mesh-hub/api/discovery/v1alpha1/mesh.proto


## Table of Contents
  - [MeshSpec](#discovery.zephyr.solo.io.MeshSpec)
  - [MeshSpec.AwsAppMesh](#discovery.zephyr.solo.io.MeshSpec.AwsAppMesh)
  - [MeshSpec.ConsulConnectMesh](#discovery.zephyr.solo.io.MeshSpec.ConsulConnectMesh)
  - [MeshSpec.IstioMesh](#discovery.zephyr.solo.io.MeshSpec.IstioMesh)
  - [MeshSpec.IstioMesh.CitadelInfo](#discovery.zephyr.solo.io.MeshSpec.IstioMesh.CitadelInfo)
  - [MeshSpec.LinkerdMesh](#discovery.zephyr.solo.io.MeshSpec.LinkerdMesh)
  - [MeshSpec.MeshInstallation](#discovery.zephyr.solo.io.MeshSpec.MeshInstallation)
  - [MeshStatus](#discovery.zephyr.solo.io.MeshStatus)







<a name="discovery.zephyr.solo.io.MeshSpec"></a>

### MeshSpec
Meshes represent a currently registered service mesh.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| istio | [MeshSpec.IstioMesh](#discovery.zephyr.solo.io.MeshSpec.IstioMesh) |  |  |
| awsAppMesh | [MeshSpec.AwsAppMesh](#discovery.zephyr.solo.io.MeshSpec.AwsAppMesh) |  |  |
| linkerd | [MeshSpec.LinkerdMesh](#discovery.zephyr.solo.io.MeshSpec.LinkerdMesh) |  |  |
| consulConnect | [MeshSpec.ConsulConnectMesh](#discovery.zephyr.solo.io.MeshSpec.ConsulConnectMesh) |  |  |
| cluster | [core.zephyr.solo.io.ResourceRef](#core.zephyr.solo.io.ResourceRef) |  | The cluster on which this mesh resides. |






<a name="discovery.zephyr.solo.io.MeshSpec.AwsAppMesh"></a>

### MeshSpec.AwsAppMesh
Mesh object representing AWS App Mesh


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installation | [MeshSpec.MeshInstallation](#discovery.zephyr.solo.io.MeshSpec.MeshInstallation) |  |  |
| region | [string](#string) |  | The AWS region the AWS App Mesh control plane resources exist in. |






<a name="discovery.zephyr.solo.io.MeshSpec.ConsulConnectMesh"></a>

### MeshSpec.ConsulConnectMesh



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installation | [MeshSpec.MeshInstallation](#discovery.zephyr.solo.io.MeshSpec.MeshInstallation) |  |  |






<a name="discovery.zephyr.solo.io.MeshSpec.IstioMesh"></a>

### MeshSpec.IstioMesh
Mesh object representing an installed Istio control plane


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installation | [MeshSpec.MeshInstallation](#discovery.zephyr.solo.io.MeshSpec.MeshInstallation) |  |  |
| citadelInfo | [MeshSpec.IstioMesh.CitadelInfo](#discovery.zephyr.solo.io.MeshSpec.IstioMesh.CitadelInfo) |  |  |






<a name="discovery.zephyr.solo.io.MeshSpec.IstioMesh.CitadelInfo"></a>

### MeshSpec.IstioMesh.CitadelInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| trustDomain | [string](#string) |  | Istio trust domain used for https/spiffe identity. https://spiffe.io/spiffe/concepts/#trust-domain https://istio.io/docs/reference/glossary/#identity<br>If empty will default to "cluster.local" |
| citadelNamespace | [string](#string) |  | istio-citadel namespace, used to determine identity for the Istio CA cert. If empty will default to MeshInstallation.installation_namespace |
| citadelServiceAccount | [string](#string) |  | istio-citadel service account, used to determine identity for the Istio CA cert. If empty will default to "istio-citadel" |






<a name="discovery.zephyr.solo.io.MeshSpec.LinkerdMesh"></a>

### MeshSpec.LinkerdMesh
Mesh object representing an installed Linkerd control plane


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installation | [MeshSpec.MeshInstallation](#discovery.zephyr.solo.io.MeshSpec.MeshInstallation) |  |  |






<a name="discovery.zephyr.solo.io.MeshSpec.MeshInstallation"></a>

### MeshSpec.MeshInstallation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installationNamespace | [string](#string) |  | Namespace in which the control plane has been installed. |
| version | [string](#string) |  | version of the mesh which has been installed Note that the version may be "latest" |






<a name="discovery.zephyr.solo.io.MeshStatus"></a>

### MeshStatus






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

