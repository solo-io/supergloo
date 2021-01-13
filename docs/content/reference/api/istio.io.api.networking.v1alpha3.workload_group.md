
---

---

## Package : `istio.networking.v1alpha3`



<a name="top"></a>

<a name="API Reference for workload_group.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## workload_group.proto


## Table of Contents
  - [ExecHealthCheckConfig](#istio.networking.v1alpha3.ExecHealthCheckConfig)
  - [HTTPHeader](#istio.networking.v1alpha3.HTTPHeader)
  - [HTTPHealthCheckConfig](#istio.networking.v1alpha3.HTTPHealthCheckConfig)
  - [ReadinessProbe](#istio.networking.v1alpha3.ReadinessProbe)
  - [TCPHealthCheckConfig](#istio.networking.v1alpha3.TCPHealthCheckConfig)
  - [WorkloadGroup](#istio.networking.v1alpha3.WorkloadGroup)
  - [WorkloadGroup.ObjectMeta](#istio.networking.v1alpha3.WorkloadGroup.ObjectMeta)
  - [WorkloadGroup.ObjectMeta.AnnotationsEntry](#istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.AnnotationsEntry)
  - [WorkloadGroup.ObjectMeta.LabelsEntry](#istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.LabelsEntry)







<a name="istio.networking.v1alpha3.ExecHealthCheckConfig"></a>

### ExecHealthCheckConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| command | []string | repeated | Command to run. Exit status of 0 is treated as live/healthy and non-zero is unhealthy. |
  





<a name="istio.networking.v1alpha3.HTTPHeader"></a>

### HTTPHeader



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The header field name |
  | value | string |  | The header field value |
  





<a name="istio.networking.v1alpha3.HTTPHealthCheckConfig"></a>

### HTTPHealthCheckConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | string |  | Path to access on the HTTP server. |
  | port | uint32 |  | Port on which the endpoint lives. |
  | host | string |  | Host name to connect to, defaults to the pod IP. You probably want to set "Host" in httpHeaders instead. |
  | scheme | string |  | HTTP or HTTPS, defaults to HTTP |
  | httpHeaders | [][istio.networking.v1alpha3.HTTPHeader]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.workload_group#istio.networking.v1alpha3.HTTPHeader" >}}) | repeated | Headers the proxy will pass on to make the request. Allows repeated headers. |
  





<a name="istio.networking.v1alpha3.ReadinessProbe"></a>

### ReadinessProbe



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| initialDelaySeconds | int32 |  | Number of seconds after the container has started before readiness probes are initiated. |
  | timeoutSeconds | int32 |  | Number of seconds after which the probe times out. Defaults to 1 second. Minimum value is 1 second. |
  | periodSeconds | int32 |  | How often (in seconds) to perform the probe. Default to 10 seconds. Minimum value is 1 second. |
  | successThreshold | int32 |  | Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1 second. |
  | failureThreshold | int32 |  | Minimum consecutive failures for the probe to be considered failed after having succeeded. Defaults to 3 seconds. |
  | httpGet | [istio.networking.v1alpha3.HTTPHealthCheckConfig]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.workload_group#istio.networking.v1alpha3.HTTPHealthCheckConfig" >}}) |  | `httpGet` is performed to a given endpoint and the status/able to connect determines health. |
  | tcpSocket | [istio.networking.v1alpha3.TCPHealthCheckConfig]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.workload_group#istio.networking.v1alpha3.TCPHealthCheckConfig" >}}) |  | Health is determined by if the proxy is able to connect. |
  | exec | [istio.networking.v1alpha3.ExecHealthCheckConfig]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.workload_group#istio.networking.v1alpha3.ExecHealthCheckConfig" >}}) |  | Health is determined by how the command that is executed exited. |
  





<a name="istio.networking.v1alpha3.TCPHealthCheckConfig"></a>

### TCPHealthCheckConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| host | string |  | Host to connect to, defaults to localhost |
  | port | uint32 |  | Port of host |
  





<a name="istio.networking.v1alpha3.WorkloadGroup"></a>

### WorkloadGroup



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metadata | [istio.networking.v1alpha3.WorkloadGroup.ObjectMeta]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.workload_group#istio.networking.v1alpha3.WorkloadGroup.ObjectMeta" >}}) |  | Metadata that will be used for all corresponding `WorkloadEntries`. User labels for a workload group should be set here in `metadata` rather than in `template`. |
  | template | [istio.networking.v1alpha3.WorkloadEntry]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.workload_entry#istio.networking.v1alpha3.WorkloadEntry" >}}) |  | Template to be used for the generation of `WorkloadEntry` resources that belong to this `WorkloadGroup`. Please note that `address` and `labels` fields should not be set in the template, and an empty `serviceAccount` should default to `default`. The workload identities (mTLS certificates) will be bootstrapped using the specified service account's token. Workload entries in this group will be in the same namespace as the workload group, and inherit the labels and annotations from the above `metadata` field. |
  | probe | [istio.networking.v1alpha3.ReadinessProbe]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.workload_group#istio.networking.v1alpha3.ReadinessProbe" >}}) |  | `ReadinessProbe` describes the configuration the user must provide for healthchecking on their workload. This configuration mirrors K8S in both syntax and logic for the most part. |
  





<a name="istio.networking.v1alpha3.WorkloadGroup.ObjectMeta"></a>

### WorkloadGroup.ObjectMeta



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| labels | [][istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.LabelsEntry]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.workload_group#istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.LabelsEntry" >}}) | repeated | Labels to attach |
  | annotations | [][istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.AnnotationsEntry]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.workload_group#istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.AnnotationsEntry" >}}) | repeated | Annotations to attach |
  





<a name="istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.AnnotationsEntry"></a>

### WorkloadGroup.ObjectMeta.AnnotationsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="istio.networking.v1alpha3.WorkloadGroup.ObjectMeta.LabelsEntry"></a>

### WorkloadGroup.ObjectMeta.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

