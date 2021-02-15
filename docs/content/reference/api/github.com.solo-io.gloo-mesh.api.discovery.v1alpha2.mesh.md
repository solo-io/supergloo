
---

title: "mesh.proto"

---

## Package : `discovery.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for mesh.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## mesh.proto


## Table of Contents
  - [MeshSpec](#discovery.mesh.gloo.solo.io.MeshSpec)
  - [MeshSpec.AgentInfo](#discovery.mesh.gloo.solo.io.MeshSpec.AgentInfo)
  - [MeshSpec.AwsAppMesh](#discovery.mesh.gloo.solo.io.MeshSpec.AwsAppMesh)
  - [MeshSpec.ConsulConnectMesh](#discovery.mesh.gloo.solo.io.MeshSpec.ConsulConnectMesh)
  - [MeshSpec.Istio](#discovery.mesh.gloo.solo.io.MeshSpec.Istio)
  - [MeshSpec.Istio.CitadelInfo](#discovery.mesh.gloo.solo.io.MeshSpec.Istio.CitadelInfo)
  - [MeshSpec.Istio.IngressGatewayInfo](#discovery.mesh.gloo.solo.io.MeshSpec.Istio.IngressGatewayInfo)
  - [MeshSpec.Istio.IngressGatewayInfo.WorkloadLabelsEntry](#discovery.mesh.gloo.solo.io.MeshSpec.Istio.IngressGatewayInfo.WorkloadLabelsEntry)
  - [MeshSpec.LinkerdMesh](#discovery.mesh.gloo.solo.io.MeshSpec.LinkerdMesh)
  - [MeshSpec.MeshInstallation](#discovery.mesh.gloo.solo.io.MeshSpec.MeshInstallation)
  - [MeshSpec.MeshInstallation.PodLabelsEntry](#discovery.mesh.gloo.solo.io.MeshSpec.MeshInstallation.PodLabelsEntry)
  - [MeshSpec.OSM](#discovery.mesh.gloo.solo.io.MeshSpec.OSM)
  - [MeshStatus](#discovery.mesh.gloo.solo.io.MeshStatus)
  - [MeshStatus.AppliedFailoverService](#discovery.mesh.gloo.solo.io.MeshStatus.AppliedFailoverService)
  - [MeshStatus.AppliedGlobalService](#discovery.mesh.gloo.solo.io.MeshStatus.AppliedGlobalService)
  - [MeshStatus.AppliedVirtualMesh](#discovery.mesh.gloo.solo.io.MeshStatus.AppliedVirtualMesh)







<a name="discovery.mesh.gloo.solo.io.MeshSpec"></a>

### MeshSpec
Meshes represent a currently registered service mesh.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| istio | [discovery.mesh.gloo.solo.io.MeshSpec.Istio]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.mesh#discovery.mesh.gloo.solo.io.MeshSpec.Istio" >}}) |  |  |
  | awsAppMesh | [discovery.mesh.gloo.solo.io.MeshSpec.AwsAppMesh]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.mesh#discovery.mesh.gloo.solo.io.MeshSpec.AwsAppMesh" >}}) |  |  |
  | linkerd | [discovery.mesh.gloo.solo.io.MeshSpec.LinkerdMesh]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.mesh#discovery.mesh.gloo.solo.io.MeshSpec.LinkerdMesh" >}}) |  |  |
  | consulConnect | [discovery.mesh.gloo.solo.io.MeshSpec.ConsulConnectMesh]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.mesh#discovery.mesh.gloo.solo.io.MeshSpec.ConsulConnectMesh" >}}) |  |  |
  | osm | [discovery.mesh.gloo.solo.io.MeshSpec.OSM]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.mesh#discovery.mesh.gloo.solo.io.MeshSpec.OSM" >}}) |  |  |
  | agentInfo | [discovery.mesh.gloo.solo.io.MeshSpec.AgentInfo]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.mesh#discovery.mesh.gloo.solo.io.MeshSpec.AgentInfo" >}}) |  | Information about the Gloo Mesh certificate agent if it has been installed to the remote cluster. |
  





<a name="discovery.mesh.gloo.solo.io.MeshSpec.AgentInfo"></a>

### MeshSpec.AgentInfo
information about the Gloo Mesh Cert-Agent which may be installed to the remote cluster which contains the Mesh control plane


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| agentNamespace | string |  | the namespace in which the agent is installed |
  





<a name="discovery.mesh.gloo.solo.io.MeshSpec.AwsAppMesh"></a>

### MeshSpec.AwsAppMesh
Mesh object representing AWS AppMesh.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| awsName | string |  | AWS name for the AppMesh instance, must be unique across the AWS account. |
  | region | string |  | The AWS region the AWS App Mesh control plane resources exist in. |
  | awsAccountId | string |  | The AWS Account ID associated with the Mesh. Populated at REST API registration time. |
  | arn | string |  | The unique AWS ARN associated with the Mesh. |
  | clusters | []string | repeated | The k8s clusters on which sidecars for this AppMesh instance have been discovered. |
  





<a name="discovery.mesh.gloo.solo.io.MeshSpec.ConsulConnectMesh"></a>

### MeshSpec.ConsulConnectMesh
Mesh object representing an installed ConsulConnect control plane.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installation | [discovery.mesh.gloo.solo.io.MeshSpec.MeshInstallation]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.mesh#discovery.mesh.gloo.solo.io.MeshSpec.MeshInstallation" >}}) |  |  |
  





<a name="discovery.mesh.gloo.solo.io.MeshSpec.Istio"></a>

### MeshSpec.Istio
Mesh object representing an installed Istio control plane


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installation | [discovery.mesh.gloo.solo.io.MeshSpec.MeshInstallation]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.mesh#discovery.mesh.gloo.solo.io.MeshSpec.MeshInstallation" >}}) |  | Configuration metadata about the istio control plane installation. |
  | citadelInfo | [discovery.mesh.gloo.solo.io.MeshSpec.Istio.CitadelInfo]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.mesh#discovery.mesh.gloo.solo.io.MeshSpec.Istio.CitadelInfo" >}}) |  | Configuration metadata for Istio Citadel (Istio's security component). |
  | ingressGateways | [][discovery.mesh.gloo.solo.io.MeshSpec.Istio.IngressGatewayInfo]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.mesh#discovery.mesh.gloo.solo.io.MeshSpec.Istio.IngressGatewayInfo" >}}) | repeated | Configuration metadata for Istio IngressGateway (the Istio Ingress). |
  | smartDnsProxyingEnabled | bool |  | True if smart DNS proxying is enabled, which allows for arbitrary DNS domains. |
  





<a name="discovery.mesh.gloo.solo.io.MeshSpec.Istio.CitadelInfo"></a>

### MeshSpec.Istio.CitadelInfo
Configuration metadata for Istio Citadel (Istio's security component).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| trustDomain | string |  | Istio trust domain used for https/spiffe identity. https://spiffe.io/spiffe/concepts/#trust-domain https://istio.io/docs/reference/glossary/#identity<br>If empty will default to "cluster.local". |
  | citadelServiceAccount | string |  | istio-citadel service account, used to determine identity for the Istio CA cert. If empty will default to "istio-citadel". |
  





<a name="discovery.mesh.gloo.solo.io.MeshSpec.Istio.IngressGatewayInfo"></a>

### MeshSpec.Istio.IngressGatewayInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workloadLabels | [][discovery.mesh.gloo.solo.io.MeshSpec.Istio.IngressGatewayInfo.WorkloadLabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.mesh#discovery.mesh.gloo.solo.io.MeshSpec.Istio.IngressGatewayInfo.WorkloadLabelsEntry" >}}) | repeated | Labels matching the workload which backs the gateway, defaults to `{"istio": "ingressgateway"}`. |
  | externalAddress | string |  | The externally-reachable address on which the gateway is listening for TLS connections. This will be the address used for cross-cluster connectivity. Defaults to the LoadBalancer Address (or NodeIP) of the Kubernetes Service (depending on its type). |
  | externalTlsPort | uint32 |  | The externally-reachable port on which the gateway is listening for TLS connections. This will be the port used for cross-cluster connectivity. List of common ports: https://istio.io/latest/docs/ops/deployment/requirements/#ports-used-by-istio. Defaults to 15443 (or the NodePort) of the Kubernetes Service (depending on its type). |
  | tlsContainerPort | uint32 |  | Container port on which the gateway is listening for TLS connections. Defaults to 15443. |
  





<a name="discovery.mesh.gloo.solo.io.MeshSpec.Istio.IngressGatewayInfo.WorkloadLabelsEntry"></a>

### MeshSpec.Istio.IngressGatewayInfo.WorkloadLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="discovery.mesh.gloo.solo.io.MeshSpec.LinkerdMesh"></a>

### MeshSpec.LinkerdMesh
Mesh object representing an installed Linkerd control plane.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installation | [discovery.mesh.gloo.solo.io.MeshSpec.MeshInstallation]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.mesh#discovery.mesh.gloo.solo.io.MeshSpec.MeshInstallation" >}}) |  |  |
  | clusterDomain | string |  | The cluster domain suffix this Linkerd mesh is configured with. See https://linkerd.io/2/tasks/using-custom-domain/ for info. |
  





<a name="discovery.mesh.gloo.solo.io.MeshSpec.MeshInstallation"></a>

### MeshSpec.MeshInstallation
The cluster on which the control plane for this mesh is deployed. Not all MeshTypes have a MeshInstallation. Only self-hosted control planes such as Istio and Linkerd will have installation metadata.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| namespace | string |  | Namespace in which the control plane has been installed. |
  | cluster | string |  | Cluster in which the control plane has been installed. |
  | podLabels | [][discovery.mesh.gloo.solo.io.MeshSpec.MeshInstallation.PodLabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.mesh#discovery.mesh.gloo.solo.io.MeshSpec.MeshInstallation.PodLabelsEntry" >}}) | repeated | the labels on the control plane pods (read from the deployment) |
  | version | string |  | Version of the Mesh that has been installed. Determined using the image tag on the Mesh's primary control plane image (e.g. the istio-pilot image tag). |
  | region | string |  | The region of the cluster in which the control plane has been installed. |
  | subLocalities | [][discovery.mesh.gloo.solo.io.SubLocalitySubset]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.traffic_target#discovery.mesh.gloo.solo.io.SubLocalitySubset" >}}) | repeated | List of zone+sub_zone pairs which this mesh is a part of |
  





<a name="discovery.mesh.gloo.solo.io.MeshSpec.MeshInstallation.PodLabelsEntry"></a>

### MeshSpec.MeshInstallation.PodLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="discovery.mesh.gloo.solo.io.MeshSpec.OSM"></a>

### MeshSpec.OSM
https://github.com/openservicemesh/osm


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installation | [discovery.mesh.gloo.solo.io.MeshSpec.MeshInstallation]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.mesh#discovery.mesh.gloo.solo.io.MeshSpec.MeshInstallation" >}}) |  | Information about where OSM is installed in a managed Kubernetes Cluster. |
  





<a name="discovery.mesh.gloo.solo.io.MeshStatus"></a>

### MeshStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The observed generation of the Mesh. When this matches the Mesh's metadata.generation, it indicates that mesh-networking has reconciled the latest version of the Mesh. |
  | appliedVirtualMesh | [discovery.mesh.gloo.solo.io.MeshStatus.AppliedVirtualMesh]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.mesh#discovery.mesh.gloo.solo.io.MeshStatus.AppliedVirtualMesh" >}}) |  | The VirtualMesh, if any, which contains this mesh. |
  | appliedFailoverServices | [][discovery.mesh.gloo.solo.io.MeshStatus.AppliedFailoverService]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.mesh#discovery.mesh.gloo.solo.io.MeshStatus.AppliedFailoverService" >}}) | repeated | The FailoverServices, if any, which applies to this mesh. |
  | appliedGlobalServices | [][discovery.mesh.gloo.solo.io.MeshStatus.AppliedGlobalService]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1alpha2.mesh#discovery.mesh.gloo.solo.io.MeshStatus.AppliedGlobalService" >}}) | repeated | The FailoverServices, if any, which applies to this mesh. |
  





<a name="discovery.mesh.gloo.solo.io.MeshStatus.AppliedFailoverService"></a>

### MeshStatus.AppliedFailoverService
AppliedFailoverService represents a FailoverService that has been applied to this Mesh. If an existing FailoverService becomes invalid the last applied FailoverService will be used.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | Reference to the FailoverService. |
  | observedGeneration | int64 |  | The observed generation of the accepted FailoverService. |
  | spec | [networking.mesh.gloo.solo.io.FailoverServiceSpec]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.failover_service#networking.mesh.gloo.solo.io.FailoverServiceSpec" >}}) |  | The last known valid spec of the FailoverService. |
  





<a name="discovery.mesh.gloo.solo.io.MeshStatus.AppliedGlobalService"></a>

### MeshStatus.AppliedGlobalService
AppliedGlobalService represents an GlobalService that has been applied to this Mesh.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | reference to the GlobalService object. |
  | observedGeneration | int64 |  | the observed generation of the accepted GlobalService. |
  | errors | []string | repeated | any errors encountered while processing the referenced GlobalService object |
  





<a name="discovery.mesh.gloo.solo.io.MeshStatus.AppliedVirtualMesh"></a>

### MeshStatus.AppliedVirtualMesh
AppliedVirtualMesh represents a VirtualMesh that has been applied to this Mesh. If an existing VirtualMesh becomes invalid, the last applied VirtualMesh will be used.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | Reference to the VirtualMesh. |
  | observedGeneration | int64 |  | The observed generation of the accepted VirtualMesh. |
  | spec | [networking.mesh.gloo.solo.io.VirtualMeshSpec]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.virtual_mesh#networking.mesh.gloo.solo.io.VirtualMeshSpec" >}}) |  | The last known valid spec of the VirtualMesh. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

