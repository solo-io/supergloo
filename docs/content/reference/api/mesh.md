
---
title: "discovery.zephyr.solo.iogithub.com/solo-io/mesh-projects/api/discovery/v1alpha1/mesh.proto"
---

## Package : `discovery.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for github.com/solo-io/mesh-projects/api/discovery/v1alpha1/mesh.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## github.com/solo-io/mesh-projects/api/discovery/v1alpha1/mesh.proto


## Table of Contents
  - [AwsAppMesh](#discovery.zephyr.solo.io.AwsAppMesh)
  - [ConsulConnectMesh](#discovery.zephyr.solo.io.ConsulConnectMesh)
  - [IstioMesh](#discovery.zephyr.solo.io.IstioMesh)
  - [IstioMesh.CitadelInfo](#discovery.zephyr.solo.io.IstioMesh.CitadelInfo)
  - [LinkerdMesh](#discovery.zephyr.solo.io.LinkerdMesh)
  - [MeshInstallation](#discovery.zephyr.solo.io.MeshInstallation)
  - [MeshSpec](#discovery.zephyr.solo.io.MeshSpec)
  - [MeshStatus](#discovery.zephyr.solo.io.MeshStatus)







<a name="discovery.zephyr.solo.io.AwsAppMesh"></a>

### AwsAppMesh
Mesh object representing AWS App Mesh


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installation | [MeshInstallation](#discovery.zephyr.solo.io.MeshInstallation) |  |  |
| region | [string](#string) |  | The AWS region the AWS App Mesh control plane resources exist in. |






<a name="discovery.zephyr.solo.io.ConsulConnectMesh"></a>

### ConsulConnectMesh



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installation | [MeshInstallation](#discovery.zephyr.solo.io.MeshInstallation) |  |  |






<a name="discovery.zephyr.solo.io.IstioMesh"></a>

### IstioMesh
Mesh object representing an installed Istio control plane


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installation | [MeshInstallation](#discovery.zephyr.solo.io.MeshInstallation) |  |  |
| citadelInfo | [IstioMesh.CitadelInfo](#discovery.zephyr.solo.io.IstioMesh.CitadelInfo) |  |  |






<a name="discovery.zephyr.solo.io.IstioMesh.CitadelInfo"></a>

### IstioMesh.CitadelInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| trustDomain | [string](#string) |  | istio trust domain used for https/spiffe identity. https://spiffe.io/spiffe/concepts/#trust-domain https://istio.io/docs/reference/glossary/#identity<br>If empty will default to "cluster.local" |
| citadelNamespace | [string](#string) |  | istio-citadel namespace, used to determine identity for istio CA cert. If empty will default to MeshInstallation.installation_namespace |
| citadelServiceAccount | [string](#string) |  | istio-citadel service account, used to determine identity for istio CA cert. If empty will default to "istio-citadel" |






<a name="discovery.zephyr.solo.io.LinkerdMesh"></a>

### LinkerdMesh
Mesh object representing an installed Linkerd control plane


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installation | [MeshInstallation](#discovery.zephyr.solo.io.MeshInstallation) |  |  |






<a name="discovery.zephyr.solo.io.MeshInstallation"></a>

### MeshInstallation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installationNamespace | [string](#string) |  | where the control plane has been installed |
| version | [string](#string) |  | version of the mesh which has been installed Note that the version may be "latest" |






<a name="discovery.zephyr.solo.io.MeshSpec"></a>

### MeshSpec
Meshes represent a currently registered service mesh.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| istio | [IstioMesh](#discovery.zephyr.solo.io.IstioMesh) |  |  |
| awsAppMesh | [AwsAppMesh](#discovery.zephyr.solo.io.AwsAppMesh) |  |  |
| linkerd | [LinkerdMesh](#discovery.zephyr.solo.io.LinkerdMesh) |  |  |
| consulConnect | [ConsulConnectMesh](#discovery.zephyr.solo.io.ConsulConnectMesh) |  |  |
| cluster | [core.zephyr.solo.io.ResourceRef](#core.zephyr.solo.io.ResourceRef) |  |  |






<a name="discovery.zephyr.solo.io.MeshStatus"></a>

### MeshStatus






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

