
---

title: "workload.proto"

---

## Package : `discovery.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for workload.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## workload.proto


## Table of Contents
  - [WorkloadSpec](#discovery.mesh.gloo.solo.io.WorkloadSpec)
  - [WorkloadSpec.AppMesh](#discovery.mesh.gloo.solo.io.WorkloadSpec.AppMesh)
  - [WorkloadSpec.AppMesh.ContainerPort](#discovery.mesh.gloo.solo.io.WorkloadSpec.AppMesh.ContainerPort)
  - [WorkloadSpec.KubernetesWorkload](#discovery.mesh.gloo.solo.io.WorkloadSpec.KubernetesWorkload)
  - [WorkloadSpec.KubernetesWorkload.PodLabelsEntry](#discovery.mesh.gloo.solo.io.WorkloadSpec.KubernetesWorkload.PodLabelsEntry)
  - [WorkloadStatus](#discovery.mesh.gloo.solo.io.WorkloadStatus)
  - [WorkloadStatus.AppliedAccessLogRecord](#discovery.mesh.gloo.solo.io.WorkloadStatus.AppliedAccessLogRecord)
  - [WorkloadStatus.AppliedWasmDeployment](#discovery.mesh.gloo.solo.io.WorkloadStatus.AppliedWasmDeployment)







<a name="discovery.mesh.gloo.solo.io.WorkloadSpec"></a>

### WorkloadSpec
Describes a workload controlled by a discovered service mesh.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kubernetes | [discovery.mesh.gloo.solo.io.WorkloadSpec.KubernetesWorkload]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.workload#discovery.mesh.gloo.solo.io.WorkloadSpec.KubernetesWorkload" >}}) |  | Information describing workloads backed by Kubernetes Pods. |
  | mesh | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | The Mesh with which this Workload is associated. |
  | appMesh | [discovery.mesh.gloo.solo.io.WorkloadSpec.AppMesh]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.workload#discovery.mesh.gloo.solo.io.WorkloadSpec.AppMesh" >}}) |  | Metadata specific to an App Mesh controlled workload. |
  





<a name="discovery.mesh.gloo.solo.io.WorkloadSpec.AppMesh"></a>

### WorkloadSpec.AppMesh
Metadata specific to an App Mesh controlled workload.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| virtualNodeName | string |  | The value of the env var APPMESH_VIRTUAL_NODE_NAME on the App Mesh envoy proxy container. |
  | ports | [][discovery.mesh.gloo.solo.io.WorkloadSpec.AppMesh.ContainerPort]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.workload#discovery.mesh.gloo.solo.io.WorkloadSpec.AppMesh.ContainerPort" >}}) | repeated | Ports exposed by this workload. Needed for declaring App Mesh VirtualNode listeners. |
  





<a name="discovery.mesh.gloo.solo.io.WorkloadSpec.AppMesh.ContainerPort"></a>

### WorkloadSpec.AppMesh.ContainerPort
Kubernetes application container ports.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | uint32 |  |  |
  | protocol | string |  |  |
  





<a name="discovery.mesh.gloo.solo.io.WorkloadSpec.KubernetesWorkload"></a>

### WorkloadSpec.KubernetesWorkload
Describes a Kubernetes workload (e.g. a Deployment or DaemonSet).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| controller | [core.skv2.solo.io.ClusterObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ClusterObjectRef" >}}) |  | Resource reference to the Kubernetes Pod controller (i.e. Deployment, ReplicaSet, DaemonSet) for this Workload.. |
  | podLabels | [][discovery.mesh.gloo.solo.io.WorkloadSpec.KubernetesWorkload.PodLabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.workload#discovery.mesh.gloo.solo.io.WorkloadSpec.KubernetesWorkload.PodLabelsEntry" >}}) | repeated | Labels on the Pod itself (read from `metadata.labels`), which are used to determine which Services front this workload. |
  | serviceAccountName | string |  | Service account associated with the Pods owned by this controller. |
  





<a name="discovery.mesh.gloo.solo.io.WorkloadSpec.KubernetesWorkload.PodLabelsEntry"></a>

### WorkloadSpec.KubernetesWorkload.PodLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="discovery.mesh.gloo.solo.io.WorkloadStatus"></a>

### WorkloadStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The observed generation of the Workload. When this matches the Workload's `metadata.generation` it indicates that Gloo Mesh has processed the latest version of the Workload. |
  | appliedAccessLogRecords | [][discovery.mesh.gloo.solo.io.WorkloadStatus.AppliedAccessLogRecord]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.workload#discovery.mesh.gloo.solo.io.WorkloadStatus.AppliedAccessLogRecord" >}}) | repeated | The set of AccessLogRecords that have been applied to this Workload. |
  | appliedWasmDeployments | [][discovery.mesh.gloo.solo.io.WorkloadStatus.AppliedWasmDeployment]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.discovery.v1.workload#discovery.mesh.gloo.solo.io.WorkloadStatus.AppliedWasmDeployment" >}}) | repeated | The set of WasmDeployments that have been applied to this Workload. |
  





<a name="discovery.mesh.gloo.solo.io.WorkloadStatus.AppliedAccessLogRecord"></a>

### WorkloadStatus.AppliedAccessLogRecord
Describes an [AccessLogRecord]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.observability.v1alpha1.access_logging/" >}}) that applies to this Workload.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | Reference to the AccessLogRecord object. |
  | observedGeneration | int64 |  | The observed generation of the accepted AccessLogRecord. |
  | errors | []string | repeated | Any errors encountered while processing the AccessLogRecord object |
  





<a name="discovery.mesh.gloo.solo.io.WorkloadStatus.AppliedWasmDeployment"></a>

### WorkloadStatus.AppliedWasmDeployment
Describes a [WasmDeployment]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1alpha1.wasm_deployment/" >}}) that applies to this Workload.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | Reference to the WasmDeployment object. |
  | observedGeneration | int64 |  | The observed generation of the WasmDeployment. |
  | errors | []string | repeated | Any errors encountered while processing the WasmDeployment object. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

