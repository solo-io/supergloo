
---

title: "access_logging.proto"

---

## Package : `observability.enterprise.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for access_logging.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## access_logging.proto


## Table of Contents
  - [AccessLogCollectionSpec](#observability.enterprise.mesh.gloo.solo.io.AccessLogCollectionSpec)
  - [AccessLogCollectionStatus](#observability.enterprise.mesh.gloo.solo.io.AccessLogCollectionStatus)







<a name="observability.enterprise.mesh.gloo.solo.io.AccessLogCollectionSpec"></a>

### AccessLogCollectionSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workloadSelectors | [][networking.mesh.gloo.solo.io.WorkloadSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.selectors#networking.mesh.gloo.solo.io.WorkloadSelector" >}}) | repeated | Select the workloads to be configured to emit access logs. Leave empty to apply to all workloads managed by Gloo Mesh. |
  | accessLogFilters | [][envoy.config.accesslog.v3.AccessLogFilter]({{< versioned_link_path fromRoot="/reference/api/envoy.config.accesslog.v3.accesslog#envoy.config.accesslog.v3.AccessLogFilter" >}}) | repeated | Configure the criteria for determining which access logs will be emitted. Leave empty to emit all access logs. |
  





<a name="observability.enterprise.mesh.gloo.solo.io.AccessLogCollectionStatus"></a>

### AccessLogCollectionStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the AccessLogCollection metadata. If the observedGeneration does not match generation, the controller has not processed the most recent version of this resource. |
  | errors | []string | repeated | Any errors encountered during processing. Also reported to any Workloads that this object applies to. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

