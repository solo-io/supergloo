
---

title: "access_logging.proto"

---

## Package : `observability.enterprise.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for access_logging.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## access_logging.proto


## Table of Contents
  - [AccessLogRecordSpec](#observability.enterprise.mesh.gloo.solo.io.AccessLogRecordSpec)
  - [AccessLogRecordSpec.Filter](#observability.enterprise.mesh.gloo.solo.io.AccessLogRecordSpec.Filter)
  - [AccessLogRecordStatus](#observability.enterprise.mesh.gloo.solo.io.AccessLogRecordStatus)







<a name="observability.enterprise.mesh.gloo.solo.io.AccessLogRecordSpec"></a>

### AccessLogRecordSpec
Describes a record of access logs sourced from a set of workloads and optionally filtered based on request criteria.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workloadSelectors | [][common.mesh.gloo.solo.io.WorkloadSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.WorkloadSelector" >}}) | repeated | Select the workloads to be configured to emit access logs. Leave empty to apply to all workloads managed by Gloo Mesh. |
  | filters | [][observability.enterprise.mesh.gloo.solo.io.AccessLogRecordSpec.Filter]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.observability.v1.access_logging#observability.enterprise.mesh.gloo.solo.io.AccessLogRecordSpec.Filter" >}}) | repeated | Configure criteria for determining which access logs will be recorded. The list is disjunctive, a request will be recorded if it matches any filter. Leave empty to emit all access logs. |
  | includedRequestHeaders | []string | repeated | Specify request headers to include in access logs. |
  | includedResponseHeaders | []string | repeated | Specify response headers to include in access logs. |
  | includedResponseTrailers | []string | repeated | Specify response trailers to include in access logs. |
  | includedFilterStateObjects | []string | repeated | Specify filter state objects to include in access logs |
  





<a name="observability.enterprise.mesh.gloo.solo.io.AccessLogRecordSpec.Filter"></a>

### AccessLogRecordSpec.Filter
Specify criteria for recording access logs. A request must match all specified criteria to be recorded.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| statusCodeMatcher | [common.mesh.gloo.solo.io.StatusCodeMatcher]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.request_matchers#common.mesh.gloo.solo.io.StatusCodeMatcher" >}}) |  | Matches against a response status code. Omit to match any status code. |
  | headerMatcher | [common.mesh.gloo.solo.io.HeaderMatcher]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.request_matchers#common.mesh.gloo.solo.io.HeaderMatcher" >}}) |  | Matches against a request or response header. Omit to match any headers. |
  





<a name="observability.enterprise.mesh.gloo.solo.io.AccessLogRecordStatus"></a>

### AccessLogRecordStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the AccessLogRecord metadata. If the `observedGeneration` does not match `metadata.generation`, Gloo Mesh has not processed the most recent version of this resource. |
  | state | [common.mesh.gloo.solo.io.ApprovalState]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.validation_state#common.mesh.gloo.solo.io.ApprovalState" >}}) |  | The state of the overall resource, will only show accepted if it has been successfully applied to all target workloads. |
  | errors | []string | repeated | Any errors encountered during processing. Also reported to any Workloads that this object applies to. |
  | workloads | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated | references to workloads that this AccessLogRecord applies to |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

