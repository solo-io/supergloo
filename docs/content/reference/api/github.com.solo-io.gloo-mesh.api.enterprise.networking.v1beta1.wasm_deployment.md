
---

title: "wasm_deployment.proto"

---

## Package : `networking.enterprise.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for wasm_deployment.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## wasm_deployment.proto


## Table of Contents
  - [ImagePullOptions](#networking.enterprise.mesh.gloo.solo.io.ImagePullOptions)
  - [UriSource](#networking.enterprise.mesh.gloo.solo.io.UriSource)
  - [WasmDeploymentSpec](#networking.enterprise.mesh.gloo.solo.io.WasmDeploymentSpec)
  - [WasmDeploymentStatus](#networking.enterprise.mesh.gloo.solo.io.WasmDeploymentStatus)
  - [WasmDeploymentStatus.WorkloadStatesEntry](#networking.enterprise.mesh.gloo.solo.io.WasmDeploymentStatus.WorkloadStatesEntry)
  - [WasmFilterSpec](#networking.enterprise.mesh.gloo.solo.io.WasmFilterSpec)
  - [WasmImageSource](#networking.enterprise.mesh.gloo.solo.io.WasmImageSource)

  - [WasmDeploymentStatus.WorkloadState](#networking.enterprise.mesh.gloo.solo.io.WasmDeploymentStatus.WorkloadState)






<a name="networking.enterprise.mesh.gloo.solo.io.ImagePullOptions"></a>

### ImagePullOptions
NOTE: ImagePullOptions are currently unsupported.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pullSecret | string |  | If a username/password is required, specify the name of a secret with keys: * username: <username> * password: <password><br>The secret must live in the Enterprises Agent namespace. |
  | insecureSkipVerify | bool |  | If true skip verifying the image server's TLS certificate. |
  | plainHttp | bool |  | If true use HTTP instead of HTTPS. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.UriSource"></a>

### UriSource
Specify options for fetching WASM Filters from an HTTP URI.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uri | string |  | The HTTP URI from which to fetch the filter. |
  | sha | string |  | Required**. The Sha256 Checksum of the filter binary (will be verified by the proxy). |
  





<a name="networking.enterprise.mesh.gloo.solo.io.WasmDeploymentSpec"></a>

### WasmDeploymentSpec
Deploys one or more WASM Envoy Filters to selected Sidecars and Gateways in a Mesh.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workloadSelector | [][common.mesh.gloo.solo.io.WorkloadSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.WorkloadSelector" >}}) | repeated | Sidecars/Gateways whose Workloads match these selectors will attach the specified WASM Filters. Leave empty to have all workloads in the mesh apply receive the WASM Filter. |
  | filters | [][networking.enterprise.mesh.gloo.solo.io.WasmFilterSpec]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.wasm_deployment#networking.enterprise.mesh.gloo.solo.io.WasmFilterSpec" >}}) | repeated | Specify WASM filter parameters. |
  | weight | uint32 |  | Weight is used to determine the order of WASM Filters when applying multiple WasmDeployments to a single workload. Deployed WASM filters will be sorted in order of highest to lowest weight. WasmDeployments with equal weights will be sorted non-deterministically. Note that all WASM Filters are currently inserted just before the Envoy router filter in the HTTP Connection Manager's HTTP Filter Chain. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.WasmDeploymentStatus"></a>

### WasmDeploymentStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the WasmDeployment metadata. if the observedGeneration does not match `metadata.generation`, Gloo Mesh has not processed the most recent version of this resource. |
  | error | string |  | Any errors encountered while processing this generation of the resource. This can include failures to pull a WASM image as well as missing or invalid fields in the spec. |
  | workloadStates | [][networking.enterprise.mesh.gloo.solo.io.WasmDeploymentStatus.WorkloadStatesEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.wasm_deployment#networking.enterprise.mesh.gloo.solo.io.WasmDeploymentStatus.WorkloadStatesEntry" >}}) | repeated | The state of the WasmDeployment as it has been applied to each individual Workload. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.WasmDeploymentStatus.WorkloadStatesEntry"></a>

### WasmDeploymentStatus.WorkloadStatesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [networking.enterprise.mesh.gloo.solo.io.WasmDeploymentStatus.WorkloadState]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.wasm_deployment#networking.enterprise.mesh.gloo.solo.io.WasmDeploymentStatus.WorkloadState" >}}) |  |  |
  





<a name="networking.enterprise.mesh.gloo.solo.io.WasmFilterSpec"></a>

### WasmFilterSpec
Specify the WASM Filter to deploy.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| localPathSource | string |  | Select `local_path_source` to deploy the filter from a file accessible to the workload proxy. Note that Gloo Mesh cannot verify whether the target workload proxy containers contain the given path. If filters do not load, please inspect the sidecar proxy logs. TODO(ilackarms): see if we can somehow verify the filter exists in the proxy container and surface that on the WasmDeployment status |
  | httpUriSource | [networking.enterprise.mesh.gloo.solo.io.UriSource]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.wasm_deployment#networking.enterprise.mesh.gloo.solo.io.UriSource" >}}) |  | Select `http_uri_source` to deploy the filter from an HTTP/S URI accessible to the workload proxy. Note that Gloo Mesh cannot verify whether the target workload proxy containers have HTTP accesss the given URI. If filters do not load, please inspect the sidecar proxy logs. TODO(ilackarms): see if we can somehow verify the filter exists in the proxy container and surface that on the WasmDeployment status TODO(ilackarms): we may need to provide options for customizing the Cluster given to envoy along with the HTTP Fetch URI. currently Gloo Mesh will create a simple plaintext HTTP cluster from the Host/Port specified in the URI. |
  | wasmImageSource | [networking.enterprise.mesh.gloo.solo.io.WasmImageSource]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.wasm_deployment#networking.enterprise.mesh.gloo.solo.io.WasmImageSource" >}}) |  | fetch the image from a [WASM OCI Registry](https://webassemblyhub.io/) Images can be built and pushed to registries using `meshctl` and `wasme`. |
  | staticFilterConfig | [google.protobuf.Any]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.any#google.protobuf.Any" >}}) |  | Provide configuration as a static `google.protobuf.Struct` is serialized as JSON before passing it to the plugin. `google.protobuf.BytesValue` and `google.protobuf.StringValue` are passed directly without the wrapper. |
  | dynamicFilterConfig | string |  | Provide configuration from a dynamic configuration source. This is used to connect proxies to a user-provided configuration server rather than using the WasmDeployment CR to update filter configuration. NOTE: Not currently implemented. This field serves as a placeholder. passing it to the plugin. `google.protobuf.BytesValue` and `google.protobuf.StringValue` are passed directly without the wrapper. TODO(ilackarms): implement with dynamic filter config source (FCDS) https://github.com/envoyproxy/envoy/issues/7867 |
  | rootId | string |  | The `root id` must match the `root id` defined inside the filter. If the user does not provide this field, Gloo Mesh will attempt to pull the image and set it from the `filter_conf` contained in the image config. Note that if the `filter_source` is not set to `wasm_image_source`, this field is required. |
  | vmId | string |  | An ID which will be used along with a hash of the wasm code (or the name of the registered Null VM plugin) to determine which VM will be used to load the WASM filter. All filters on the same proxy which use the same `vm_id` and code within will use the same VM. May be left blank. Sharing a VM between plugins can reduce memory utilization and make sharing of data easier which may have security implications. |
  | filterContext | [istio.networking.v1alpha3.EnvoyFilter.PatchContext]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.envoy_filter#istio.networking.v1alpha3.EnvoyFilter.PatchContext" >}}) |  | The specific config generation context to which to attach the filter. Istio generates envoy configuration in the context of a gateway, inbound traffic to sidecar and outbound traffic from sidecar. Uses the Istio default (`ANY`). |
  | insertBeforeFilter | string |  | The filter in the Envoy HTTP Filter Chain immediately before which the WASM filter will be inserted. Defaults to `envoy.router`. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.WasmImageSource"></a>

### WasmImageSource
Specify options for fetching WASM Filters from a [WASM-compatible OCI Registry](https://webassemblyhub.io/). Images can be built and pushed to registries using `meshctl` and `wasme`.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| wasmImageTag | string |  | The full tag of the wasm image. It must include the registry address at the beginning, e.g. "webassemblyhub.io/ilackarms/helloworld:v0.1". |
  




 <!-- end messages -->


<a name="networking.enterprise.mesh.gloo.solo.io.WasmDeploymentStatus.WorkloadState"></a>

### WasmDeploymentStatus.WorkloadState
WorkloadState is the state of the WasmDeployment resource as it has been applied to an individual Workload.

| Name | Number | Description |
| ---- | ------ | ----------- |
| DEPLOYMENT_PENDING | 0 | Indicates that filters have not yet been deployed to the target Workload. |
| FILTERS_DEPLOYED | 1 | Indicates the WASM Filters have been deployed to the target Workload (along with any cluster dependencies). |
| DEPLOYMENT_FAILED | 2 | Indicates deploying the WASM Filters to this Workload failed. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

