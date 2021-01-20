
---

title: "access_logging.proto"

---

## Package : `extensions.networking.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for access_logging.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## access_logging.proto


## Table of Contents
  - [AccessLog](#extensions.networking.mesh.gloo.solo.io.AccessLog)
  - [WatchAccessLogsRequest](#extensions.networking.mesh.gloo.solo.io.WatchAccessLogsRequest)



  - [Logging](#extensions.networking.mesh.gloo.solo.io.Logging)




<a name="extensions.networking.mesh.gloo.solo.io.AccessLog"></a>

### AccessLog



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| httpAccessLog | [envoy.data.accesslog.v3.HTTPAccessLogEntry]({{< versioned_link_path fromRoot="/reference/api/envoy.data.accesslog.v3.accesslog#envoy.data.accesslog.v3.HTTPAccessLogEntry" >}}) |  | An Envoy access log. |
  





<a name="extensions.networking.mesh.gloo.solo.io.WatchAccessLogsRequest"></a>

### WatchAccessLogsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workloadSelectors | [][networking.mesh.gloo.solo.io.WorkloadSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.selectors#networking.mesh.gloo.solo.io.WorkloadSelector" >}}) | repeated | Select the workloads whose access logs should be streamed. Leave empty to stream access logs for all workloads. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="extensions.networking.mesh.gloo.solo.io.Logging"></a>

### Logging
The logging service provides structured retrieval of event logs captured by Gloo Mesh.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| WatchAccessLogs | [WatchAccessLogsRequest](#extensions.networking.mesh.gloo.solo.io.WatchAccessLogsRequest) | [AccessLog](#extensions.networking.mesh.gloo.solo.io.AccessLog) stream | Stream Envoy access logs as they are captured. |

 <!-- end services -->

