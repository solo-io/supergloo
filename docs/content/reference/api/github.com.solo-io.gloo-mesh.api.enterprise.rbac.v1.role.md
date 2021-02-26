
---

title: "role.proto"

---

## Package : `rbac.enterprise.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for role.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## role.proto


## Table of Contents
  - [RoleBindingSpec](#rbac.enterprise.mesh.gloo.solo.io.RoleBindingSpec)
  - [RoleBindingStatus](#rbac.enterprise.mesh.gloo.solo.io.RoleBindingStatus)
  - [RoleSpec](#rbac.enterprise.mesh.gloo.solo.io.RoleSpec)
  - [RoleSpec.AccessLogRecordScope](#rbac.enterprise.mesh.gloo.solo.io.RoleSpec.AccessLogRecordScope)
  - [RoleSpec.AccessPolicyScope](#rbac.enterprise.mesh.gloo.solo.io.RoleSpec.AccessPolicyScope)
  - [RoleSpec.TrafficPolicyScope](#rbac.enterprise.mesh.gloo.solo.io.RoleSpec.TrafficPolicyScope)
  - [RoleSpec.VirtualDestinationScope](#rbac.enterprise.mesh.gloo.solo.io.RoleSpec.VirtualDestinationScope)
  - [RoleSpec.VirtualMeshScope](#rbac.enterprise.mesh.gloo.solo.io.RoleSpec.VirtualMeshScope)
  - [RoleSpec.WasmDeploymentScope](#rbac.enterprise.mesh.gloo.solo.io.RoleSpec.WasmDeploymentScope)
  - [RoleStatus](#rbac.enterprise.mesh.gloo.solo.io.RoleStatus)

  - [RoleSpec.TrafficPolicyScope.TrafficPolicyActions](#rbac.enterprise.mesh.gloo.solo.io.RoleSpec.TrafficPolicyScope.TrafficPolicyActions)
  - [RoleSpec.VirtualMeshScope.VirtualMeshActions](#rbac.enterprise.mesh.gloo.solo.io.RoleSpec.VirtualMeshScope.VirtualMeshActions)






<a name="rbac.enterprise.mesh.gloo.solo.io.RoleBindingSpec"></a>

### RoleBindingSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| subjects | [][core.skv2.solo.io.TypedObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.TypedObjectRef" >}}) | repeated | Specify by reference the Kubernetes Users or Groups the Role should apply to. |
  | roleRef | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | Specify by reference the Gloo Mesh Role to bind. |
  





<a name="rbac.enterprise.mesh.gloo.solo.io.RoleBindingStatus"></a>

### RoleBindingStatus







<a name="rbac.enterprise.mesh.gloo.solo.io.RoleSpec"></a>

### RoleSpec
A role represents a set of permissions for creating, updating, and deleting Gloo Mesh configuration objects. A role consists of a set of scopes for each policy type. Depending on the policy type, the permission granularity is defined at the field level or at the object level.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| trafficPolicyScopes | [][rbac.enterprise.mesh.gloo.solo.io.RoleSpec.TrafficPolicyScope]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.rbac.v1.role#rbac.enterprise.mesh.gloo.solo.io.RoleSpec.TrafficPolicyScope" >}}) | repeated | A set of TrafficPolicy configuration permissions. Permission granularity is defined at the field level. |
  | virtualMeshScopes | [][rbac.enterprise.mesh.gloo.solo.io.RoleSpec.VirtualMeshScope]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.rbac.v1.role#rbac.enterprise.mesh.gloo.solo.io.RoleSpec.VirtualMeshScope" >}}) | repeated | A set of VirtualMesh configuration permissions. Permission granularity is defined at the field level. |
  | accessPolicyScopes | [][rbac.enterprise.mesh.gloo.solo.io.RoleSpec.AccessPolicyScope]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.rbac.v1.role#rbac.enterprise.mesh.gloo.solo.io.RoleSpec.AccessPolicyScope" >}}) | repeated | A set of AccessPolicy configuration permissions. Permission granularity is defined at the object level. |
  | virtualDestinationScopes | [][rbac.enterprise.mesh.gloo.solo.io.RoleSpec.VirtualDestinationScope]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.rbac.v1.role#rbac.enterprise.mesh.gloo.solo.io.RoleSpec.VirtualDestinationScope" >}}) | repeated | A set of VirtualDestination configuration permissions. Permission granularity is defined at the object level. |
  | wasmDeploymentScopes | [][rbac.enterprise.mesh.gloo.solo.io.RoleSpec.WasmDeploymentScope]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.rbac.v1.role#rbac.enterprise.mesh.gloo.solo.io.RoleSpec.WasmDeploymentScope" >}}) | repeated | A set of WasmDeployment configuration permissions. Permission granularity is defined at the object level. |
  | accessLogRecordScopes | [][rbac.enterprise.mesh.gloo.solo.io.RoleSpec.AccessLogRecordScope]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.rbac.v1.role#rbac.enterprise.mesh.gloo.solo.io.RoleSpec.AccessLogRecordScope" >}}) | repeated | A set of AccessLogRecord configuration permissions. Permission granularity is defined at the object level. |
  





<a name="rbac.enterprise.mesh.gloo.solo.io.RoleSpec.AccessLogRecordScope"></a>

### RoleSpec.AccessLogRecordScope
Represents permissions for configuring AccessLogRecords.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workloadSelectors | [][common.mesh.gloo.solo.io.WorkloadSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.WorkloadSelector" >}}) | repeated | A list of permitted Workload selectors. |
  





<a name="rbac.enterprise.mesh.gloo.solo.io.RoleSpec.AccessPolicyScope"></a>

### RoleSpec.AccessPolicyScope
Represents permissions for configuring AccessPolicies.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| identitySelectors | [][common.mesh.gloo.solo.io.IdentitySelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.IdentitySelector" >}}) | repeated | A list of permitted identity selectors. |
  | destinationSelectors | [][common.mesh.gloo.solo.io.DestinationSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.DestinationSelector" >}}) | repeated | A list of permitted Destination selectors. |
  





<a name="rbac.enterprise.mesh.gloo.solo.io.RoleSpec.TrafficPolicyScope"></a>

### RoleSpec.TrafficPolicyScope
Represents permissions for configuring TrafficPolicies.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| trafficPolicyActions | [][rbac.enterprise.mesh.gloo.solo.io.RoleSpec.TrafficPolicyScope.TrafficPolicyActions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.rbac.v1.role#rbac.enterprise.mesh.gloo.solo.io.RoleSpec.TrafficPolicyScope.TrafficPolicyActions" >}}) | repeated | A list of permitted TrafficPolicy configuration actions. |
  | destinationSelectors | [][common.mesh.gloo.solo.io.DestinationSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.DestinationSelector" >}}) | repeated | A list of permitted Destination selectors. |
  | workloadSelectors | [][common.mesh.gloo.solo.io.WorkloadSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.WorkloadSelector" >}}) | repeated | A list of permitted Workload selectors. |
  





<a name="rbac.enterprise.mesh.gloo.solo.io.RoleSpec.VirtualDestinationScope"></a>

### RoleSpec.VirtualDestinationScope
Represents permissions for configuring VirtualDestinations.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| virtualMeshRefs | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated | A list of permitted virtual mesh references. |
  | meshRefs | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated | A list of permitted mesh references. |
  | destinationSelectors | [][common.mesh.gloo.solo.io.DestinationSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.DestinationSelector" >}}) | repeated | A list of permitted backing service selectors. |
  | destinations | [][networking.enterprise.mesh.gloo.solo.io.VirtualDestinationBackingDestination]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_destination#networking.enterprise.mesh.gloo.solo.io.VirtualDestinationBackingDestination" >}}) | repeated | A list of permitted backing Destinations. |
  





<a name="rbac.enterprise.mesh.gloo.solo.io.RoleSpec.VirtualMeshScope"></a>

### RoleSpec.VirtualMeshScope
Represents permissions for configuring VirtualMeshes.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| virtualMeshActions | [][rbac.enterprise.mesh.gloo.solo.io.RoleSpec.VirtualMeshScope.VirtualMeshActions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.rbac.v1.role#rbac.enterprise.mesh.gloo.solo.io.RoleSpec.VirtualMeshScope.VirtualMeshActions" >}}) | repeated | A list of permitted VirtualMesh configuration actions. |
  | meshRefs | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated | A list of permitted mesh references. |
  





<a name="rbac.enterprise.mesh.gloo.solo.io.RoleSpec.WasmDeploymentScope"></a>

### RoleSpec.WasmDeploymentScope
Represents permissions for configuring WasmDeployments.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workloadSelectors | [][common.mesh.gloo.solo.io.WorkloadSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.WorkloadSelector" >}}) | repeated | A list of permitted Workload selectors. |
  





<a name="rbac.enterprise.mesh.gloo.solo.io.RoleStatus"></a>

### RoleStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The observed generation of the Role. When this matches the Role's `metadata.generation` it indicates that Gloo Mesh has processed the latest version of the Role. |
  




 <!-- end messages -->


<a name="rbac.enterprise.mesh.gloo.solo.io.RoleSpec.TrafficPolicyScope.TrafficPolicyActions"></a>

### RoleSpec.TrafficPolicyScope.TrafficPolicyActions
Enums representing fields on the TrafficPolicy CRD.

| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN_TRAFFIC_POLICY_ACTION | 0 |  |
| ALL | 1 |  |
| TRAFFIC_SHIFT | 2 |  |
| FAULT_INJECTION | 3 |  |
| REQUEST_TIMEOUT | 4 |  |
| RETRIES | 5 |  |
| CORS_POLICY | 6 |  |
| MIRROR | 7 |  |
| HEADER_MANIPULATION | 8 |  |
| OUTLIER_DETECTION | 9 |  |
| MTLS_CONFIG | 10 |  |



<a name="rbac.enterprise.mesh.gloo.solo.io.RoleSpec.VirtualMeshScope.VirtualMeshActions"></a>

### RoleSpec.VirtualMeshScope.VirtualMeshActions
Enums representing fields on the VirtualMesh CRD.

| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN_VIRTUAL_MESH_ACTION | 0 |  |
| ALL | 1 |  |
| MTLS_CONFIG | 2 |  |
| FEDERATION | 3 |  |
| GLOBAL_ACCESS_POLICY | 4 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

