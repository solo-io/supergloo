
---
title: "mesh.proto"
---

## Package : `discovery.smh.solo.io`



<a name="top"></a>

<a name="API Reference for mesh.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## mesh.proto


## Table of Contents
  - [MeshSpec](#discovery.smh.solo.io.MeshSpec)
  - [MeshSpec.AwsAppMesh](#discovery.smh.solo.io.MeshSpec.AwsAppMesh)
  - [MeshSpec.ConsulConnectMesh](#discovery.smh.solo.io.MeshSpec.ConsulConnectMesh)
  - [MeshSpec.Istio1_5](#discovery.smh.solo.io.MeshSpec.Istio1_5)
  - [MeshSpec.Istio1_6](#discovery.smh.solo.io.MeshSpec.Istio1_6)
  - [MeshSpec.IstioMesh](#discovery.smh.solo.io.MeshSpec.IstioMesh)
  - [MeshSpec.IstioMesh.CitadelInfo](#discovery.smh.solo.io.MeshSpec.IstioMesh.CitadelInfo)
  - [MeshSpec.LinkerdMesh](#discovery.smh.solo.io.MeshSpec.LinkerdMesh)
  - [MeshSpec.MeshInstallation](#discovery.smh.solo.io.MeshSpec.MeshInstallation)
  - [MeshStatus](#discovery.smh.solo.io.MeshStatus)







<a name="discovery.smh.solo.io.MeshSpec"></a>

### MeshSpec
Meshes represent a currently registered service mesh.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| istio15 | [MeshSpec.Istio1_5](#discovery.smh.solo.io.MeshSpec.Istio1_5) |  |  |
| istio16 | [MeshSpec.Istio1_6](#discovery.smh.solo.io.MeshSpec.Istio1_6) |  |  |
| awsAppMesh | [MeshSpec.AwsAppMesh](#discovery.smh.solo.io.MeshSpec.AwsAppMesh) |  |  |
| linkerd | [MeshSpec.LinkerdMesh](#discovery.smh.solo.io.MeshSpec.LinkerdMesh) |  |  |
| consulConnect | [MeshSpec.ConsulConnectMesh](#discovery.smh.solo.io.MeshSpec.ConsulConnectMesh) |  |  |
| cluster | [core.smh.solo.io.ResourceRef](#core.smh.solo.io.ResourceRef) |  | The cluster on which the control plane for this mesh is deployed. This field may not apply to all Mesh types, such as AppMesh, whose control planes are located externally to any user accessible compute platform. |






<a name="discovery.smh.solo.io.MeshSpec.AwsAppMesh"></a>

### MeshSpec.AwsAppMesh
Mesh object representing AWS AppMesh


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | AWS name for the AppMesh instance, must be unique across the AWS account. |
| region | [string](#string) |  | The AWS region the AWS App Mesh control plane resources exist in. |
| awsAccountId | [string](#string) |  | The AWS Account ID associated with the Mesh. Populated at REST API registration time. |
| clusters | [][string](#string) | repeated | The k8s clusters on which sidecars for this AppMesh instance have been discovered. |






<a name="discovery.smh.solo.io.MeshSpec.ConsulConnectMesh"></a>

### MeshSpec.ConsulConnectMesh



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installation | [MeshSpec.MeshInstallation](#discovery.smh.solo.io.MeshSpec.MeshInstallation) |  |  |






<a name="discovery.smh.solo.io.MeshSpec.Istio1_5"></a>

### MeshSpec.Istio1_5



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [MeshSpec.IstioMesh](#discovery.smh.solo.io.MeshSpec.IstioMesh) |  |  |






<a name="discovery.smh.solo.io.MeshSpec.Istio1_6"></a>

### MeshSpec.Istio1_6



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [MeshSpec.IstioMesh](#discovery.smh.solo.io.MeshSpec.IstioMesh) |  |  |






<a name="discovery.smh.solo.io.MeshSpec.IstioMesh"></a>

### MeshSpec.IstioMesh
Mesh object representing an installed Istio control plane


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installation | [MeshSpec.MeshInstallation](#discovery.smh.solo.io.MeshSpec.MeshInstallation) |  |  |
| citadelInfo | [MeshSpec.IstioMesh.CitadelInfo](#discovery.smh.solo.io.MeshSpec.IstioMesh.CitadelInfo) |  |  |






<a name="discovery.smh.solo.io.MeshSpec.IstioMesh.CitadelInfo"></a>

### MeshSpec.IstioMesh.CitadelInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| trustDomain | [string](#string) |  | Istio trust domain used for https/spiffe identity. https://spiffe.io/spiffe/concepts/#trust-domain https://istio.io/docs/reference/glossary/#identity<br>If empty will default to "cluster.local" |
| citadelNamespace | [string](#string) |  | istio-citadel namespace, used to determine identity for the Istio CA cert. If empty will default to MeshInstallation.installation_namespace |
| citadelServiceAccount | [string](#string) |  | istio-citadel service account, used to determine identity for the Istio CA cert. If empty will default to "istio-citadel" |






<a name="discovery.smh.solo.io.MeshSpec.LinkerdMesh"></a>

### MeshSpec.LinkerdMesh
Mesh object representing an installed Linkerd control plane


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installation | [MeshSpec.MeshInstallation](#discovery.smh.solo.io.MeshSpec.MeshInstallation) |  |  |
| clusterDomain | [string](#string) |  | The cluster domain suffix this Linkerd mesh is configured with. See https://linkerd.io/2/tasks/using-custom-domain/ for info |






<a name="discovery.smh.solo.io.MeshSpec.MeshInstallation"></a>

### MeshSpec.MeshInstallation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installationNamespace | [string](#string) |  | Namespace in which the control plane has been installed. |
| version | [string](#string) |  | version of the mesh which has been installed Note that the version may be "latest" |






<a name="discovery.smh.solo.io.MeshStatus"></a>

### MeshStatus






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

