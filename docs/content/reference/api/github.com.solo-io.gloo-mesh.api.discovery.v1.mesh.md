
---

title: "mesh.proto"

---

## Package : `discovery.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for mesh.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## mesh.proto


## Table of Contents
  - [MeshInstallation](#discovery.mesh.gloo.solo.io.MeshInstallation)
  - [MeshInstallation.PodLabelsEntry](#discovery.mesh.gloo.solo.io.MeshInstallation.PodLabelsEntry)
  - [MeshSpec](#discovery.mesh.gloo.solo.io.MeshSpec)
  - [MeshSpec.AgentInfo](#discovery.mesh.gloo.solo.io.MeshSpec.AgentInfo)
  - [MeshSpec.AwsAppMesh](#discovery.mesh.gloo.solo.io.MeshSpec.AwsAppMesh)
  - [MeshSpec.ConsulConnectMesh](#discovery.mesh.gloo.solo.io.MeshSpec.ConsulConnectMesh)
  - [MeshSpec.Istio](#discovery.mesh.gloo.solo.io.MeshSpec.Istio)
  - [MeshSpec.Istio.IngressGatewayInfo](#discovery.mesh.gloo.solo.io.MeshSpec.Istio.IngressGatewayInfo)
  - [MeshSpec.Istio.IngressGatewayInfo.WorkloadLabelsEntry](#discovery.mesh.gloo.solo.io.MeshSpec.Istio.IngressGatewayInfo.WorkloadLabelsEntry)
  - [MeshSpec.LinkerdMesh](#discovery.mesh.gloo.solo.io.MeshSpec.LinkerdMesh)
  - [MeshSpec.OSM](#discovery.mesh.gloo.solo.io.MeshSpec.OSM)
  - [MeshStatus](#discovery.mesh.gloo.solo.io.MeshStatus)
  - [MeshStatus.AppliedIngressGateway](#discovery.mesh.gloo.solo.io.MeshStatus.AppliedIngressGateway)
  - [MeshStatus.AppliedVirtualDestination](#discovery.mesh.gloo.solo.io.MeshStatus.AppliedVirtualDestination)
  - [MeshStatus.AppliedVirtualMesh](#discovery.mesh.gloo.solo.io.MeshStatus.AppliedVirtualMesh)







<a name="discovery.mesh.gloo.solo.io.MeshInstallation"></a>

### MeshInstallation
Describes the Kubernetes cluster on which the control plane for this mesh is deployed. Only self-hosted control planes such as Istio, Linkerd, OSM, and ConsulConnect will have installation metadata.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| namespace | string |  | Namespace in which the control plane has been installed. |
  | cluster | string |  | The Gloo Mesh cluster in which the control plane has been installed. |
  | podLabels | [][discovery.mesh.gloo.solo.io.MeshInstallation.PodLabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.mesh#discovery.mesh.gloo.solo.io.MeshInstallation.PodLabelsEntry" >}}) | repeated | The labels on the control plane pods (read from the deployment). |
  | version | string |  | The version of the Mesh that has been installed, which is determined using the image tag on the mesh's primary control plane image (e.g. the istio-pilot image tag). |
  | region | string |  | The region of the cluster in which the control plane has been installed, which is determined from the value of the [Kubernetes region topology label](https://kubernetes.io/docs/reference/labels-annotations-taints/#topologykubernetesioregion) on any Kubernetes node associated with the cluster of this mesh. |
  





<a name="discovery.mesh.gloo.solo.io.MeshInstallation.PodLabelsEntry"></a>

### MeshInstallation.PodLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="discovery.mesh.gloo.solo.io.MeshSpec"></a>

### MeshSpec
Describes a service mesh control plane deployment.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| istio | [discovery.mesh.gloo.solo.io.MeshSpec.Istio]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.mesh#discovery.mesh.gloo.solo.io.MeshSpec.Istio" >}}) |  | Describes an [Istio](https://istio.io/) service mesh. |
  | awsAppMesh | [discovery.mesh.gloo.solo.io.MeshSpec.AwsAppMesh]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.mesh#discovery.mesh.gloo.solo.io.MeshSpec.AwsAppMesh" >}}) |  | Describes an [AWS App Mesh](https://aws.amazon.com/app-mesh/) service mesh. |
  | linkerd | [discovery.mesh.gloo.solo.io.MeshSpec.LinkerdMesh]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.mesh#discovery.mesh.gloo.solo.io.MeshSpec.LinkerdMesh" >}}) |  | Describes a [Linkerd](https://linkerd.io/) service mesh. |
  | consulConnect | [discovery.mesh.gloo.solo.io.MeshSpec.ConsulConnectMesh]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.mesh#discovery.mesh.gloo.solo.io.MeshSpec.ConsulConnectMesh" >}}) |  | Describes a [Consul Connect](https://www.consul.io/docs/connect) service mesh. |
  | osm | [discovery.mesh.gloo.solo.io.MeshSpec.OSM]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.mesh#discovery.mesh.gloo.solo.io.MeshSpec.OSM" >}}) |  | Describes an [Open Service Mesh](https://openservicemesh.io/) service mesh. |
  | agentInfo | [discovery.mesh.gloo.solo.io.MeshSpec.AgentInfo]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.mesh#discovery.mesh.gloo.solo.io.MeshSpec.AgentInfo" >}}) |  | Describes the Gloo Mesh agent if it has been installed to the managed cluster. |
  





<a name="discovery.mesh.gloo.solo.io.MeshSpec.AgentInfo"></a>

### MeshSpec.AgentInfo
Describes the Gloo Mesh agent which may be installed to the managed cluster containing the mesh control plane.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| agentNamespace | string |  | The namespace in which the Gloo Mesh agent is installed. |
  





<a name="discovery.mesh.gloo.solo.io.MeshSpec.AwsAppMesh"></a>

### MeshSpec.AwsAppMesh
Describes an AWS App Mesh instance.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| awsName | string |  | The AWS name for the App Mesh instance, must be unique across all AppMesh instances owned by the AWS account. |
  | region | string |  | The AWS region the App Mesh control plane resources exist in. |
  | awsAccountId | string |  | The AWS Account ID associated with the Mesh. Populated at REST API registration time. |
  | arn | string |  | The unique AWS ARN associated with the App Mesh instance. |
  | clusters | []string | repeated | The Kubernetes clusters on which sidecars for this App Mesh instance have been discovered. |
  





<a name="discovery.mesh.gloo.solo.io.MeshSpec.ConsulConnectMesh"></a>

### MeshSpec.ConsulConnectMesh
Describes a ConsulConnect deployment.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installation | [discovery.mesh.gloo.solo.io.MeshInstallation]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.mesh#discovery.mesh.gloo.solo.io.MeshInstallation" >}}) |  | Describes the ConsulConnect control plane deployment. |
  





<a name="discovery.mesh.gloo.solo.io.MeshSpec.Istio"></a>

### MeshSpec.Istio
Describes an Istio deployment.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installation | [discovery.mesh.gloo.solo.io.MeshInstallation]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.mesh#discovery.mesh.gloo.solo.io.MeshInstallation" >}}) |  | Describes the Istio control plane deployment. |
  | trustDomain | string |  | The Istio trust domain used for https/[spiffe](https://spiffe.io/spiffe/concepts/#trust-domain) [identity](https://istio.io/docs/reference/glossary/#identity). If empty will default to ["cluster.local"](https://github.com/istio/istio/blob/e768f408a7de224e64ccdfb2634442541ce08e6a/pilot/cmd/pilot-agent/main.go#L118). |
  | istiodServiceAccount | string |  | The istiod service account which determines identity for the Istio CA cert. |
  | ingressGateways | [][discovery.mesh.gloo.solo.io.MeshSpec.Istio.IngressGatewayInfo]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.mesh#discovery.mesh.gloo.solo.io.MeshSpec.Istio.IngressGatewayInfo" >}}) | repeated | DEPRECATED: external address data for an ingress gateway destination and workload live in the relevant Destination and Workload objects. Describes the ingress gateway. |
  | smartDnsProxyingEnabled | bool |  | True if smart DNS proxying is enabled, which allows for arbitrary DNS domains. |
  





<a name="discovery.mesh.gloo.solo.io.MeshSpec.Istio.IngressGatewayInfo"></a>

### MeshSpec.Istio.IngressGatewayInfo
DEPRECATED: external address data for an ingress gateway destination and workload live in the relevant Destination and Workload objects. Describes the ingress gateway.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | TODO: remove name and namespace as ingress gw info now contains a ref The name of the Ingress Gateway Service |
  | namespace | string |  | The namespace in which the ingress gateway is running. |
  | workloadLabels | [][discovery.mesh.gloo.solo.io.MeshSpec.Istio.IngressGatewayInfo.WorkloadLabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.mesh#discovery.mesh.gloo.solo.io.MeshSpec.Istio.IngressGatewayInfo.WorkloadLabelsEntry" >}}) | repeated | The ingress service selector labels for the gateway. [Defaults to](https://github.com/istio/istio/blob/ab6cc48134a698d7ad218a83390fe27e8098919f/pkg/config/constants/constants.go#L73) `{"istio": "ingressgateway"}`. |
  | externalAddress | string |  | DEPRECATED: in favor of dns_name or external_ip |
  | dnsName | string |  | Will be populated if the LoadBalancer Address is a DNS name |
  | ip | string |  | Will be populated if the LoadBalancer Address is an IP |
  | externalTlsPort | uint32 |  | The externally-reachable port on which the gateway is listening for TLS connections. This will be the port used for cross-cluster connectivity. See the list of [common ports used by Istio](https://istio.io/latest/docs/ops/deployment/requirements/#ports-used-by-istio). Defaults to 15443 (or the NodePort) of the Kubernetes service (depending on its type). |
  | tlsContainerPort | uint32 |  | Container port on which the gateway is listening for TLS connections. Defaults to 15443. |
  





<a name="discovery.mesh.gloo.solo.io.MeshSpec.Istio.IngressGatewayInfo.WorkloadLabelsEntry"></a>

### MeshSpec.Istio.IngressGatewayInfo.WorkloadLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="discovery.mesh.gloo.solo.io.MeshSpec.LinkerdMesh"></a>

### MeshSpec.LinkerdMesh
Describes a Linkerd deployment.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installation | [discovery.mesh.gloo.solo.io.MeshInstallation]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.mesh#discovery.mesh.gloo.solo.io.MeshInstallation" >}}) |  | Describes the Linkerd control plane deployment. |
  | clusterDomain | string |  | The cluster domain suffix this Linkerd mesh is configured with. See [this reference](https://linkerd.io/2/tasks/using-custom-domain/) for more info. |
  





<a name="discovery.mesh.gloo.solo.io.MeshSpec.OSM"></a>

### MeshSpec.OSM
Describes an [OSM](https://github.com/openservicemesh/osm) deployment.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installation | [discovery.mesh.gloo.solo.io.MeshInstallation]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.mesh#discovery.mesh.gloo.solo.io.MeshInstallation" >}}) |  | Describes the OSM control plane deployment. |
  





<a name="discovery.mesh.gloo.solo.io.MeshStatus"></a>

### MeshStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The observed generation of the Mesh. When this matches the Mesh's metadata.generation, it indicates that Gloo Mesh has processed the latest version of the Mesh. |
  | appliedVirtualMesh | [discovery.mesh.gloo.solo.io.MeshStatus.AppliedVirtualMesh]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.mesh#discovery.mesh.gloo.solo.io.MeshStatus.AppliedVirtualMesh" >}}) |  | The VirtualMesh, if any, which contains this Mesh. |
  | appliedVirtualDestinations | [][discovery.mesh.gloo.solo.io.MeshStatus.AppliedVirtualDestination]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.mesh#discovery.mesh.gloo.solo.io.MeshStatus.AppliedVirtualDestination" >}}) | repeated | The VirtualDestinations, if any, which apply to this Mesh. |
  | appliedEastWestIngressGateways | [][discovery.mesh.gloo.solo.io.MeshStatus.AppliedIngressGateway]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.mesh#discovery.mesh.gloo.solo.io.MeshStatus.AppliedIngressGateway" >}}) | repeated | The Destination(s) acting as ingress gateways for east west traffic. |
  





<a name="discovery.mesh.gloo.solo.io.MeshStatus.AppliedIngressGateway"></a>

### MeshStatus.AppliedIngressGateway



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| destinationRef | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | The Destination on the mesh that acts as an ingress gateway for the mesh. |
  | externalAddresses | []string | repeated | The externally accessible address(es) for this ingress gateway Destination. |
  | destinationPort | uint32 |  | The port on the ingress gateway Destination designated for receiving cross cluster traffic. |
  | containerPort | uint32 |  | The port on the ingress gateway's backing Workload(s) designated for receiving cross cluster traffic. |
  





<a name="discovery.mesh.gloo.solo.io.MeshStatus.AppliedVirtualDestination"></a>

### MeshStatus.AppliedVirtualDestination
Describes a [VirtualDestination]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.virtual_destination.md" >}}) that applies to this Mesh.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | Reference to the applied VirtualDestination object. |
  | observedGeneration | int64 |  | The observed generation of the accepted VirtualDestination. |
  | errors | []string | repeated | Any errors encountered while processing the VirtualDestination. |
  





<a name="discovery.mesh.gloo.solo.io.MeshStatus.AppliedVirtualMesh"></a>

### MeshStatus.AppliedVirtualMesh
Describes a [VirtualMesh]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.virtual_mesh" >}}) that applies to this Mesh. If an existing applied VirtualMesh becomes invalid, the last applied VirtualMesh will be used.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | Reference to the applied VirtualMesh object. |
  | observedGeneration | int64 |  | The observed generation of the accepted VirtualMesh. |
  | spec | [networking.mesh.gloo.solo.io.VirtualMeshSpec]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.virtual_mesh#networking.mesh.gloo.solo.io.VirtualMeshSpec" >}}) |  | The spec of the last known valid VirtualMesh. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

