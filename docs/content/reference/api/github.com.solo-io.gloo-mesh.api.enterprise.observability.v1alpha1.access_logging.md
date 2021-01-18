
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
  - [AccessLogCollectionSpec.Filter](#observability.enterprise.mesh.gloo.solo.io.AccessLogCollectionSpec.Filter)
  - [AccessLogCollectionStatus](#observability.enterprise.mesh.gloo.solo.io.AccessLogCollectionStatus)







<a name="observability.enterprise.mesh.gloo.solo.io.AccessLogCollectionSpec"></a>

### AccessLogCollectionSpec
Describes a collection of access logs sourced from a set of workloads and optionally filtered based on request criteria.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workloadSelectors | [][networking.mesh.gloo.solo.io.WorkloadSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.selectors#networking.mesh.gloo.solo.io.WorkloadSelector" >}}) | repeated | Select the workloads to be configured to emit access logs. Leave empty to apply to all workloads managed by Gloo Mesh. |
  | accessLogFilters | [][observability.enterprise.mesh.gloo.solo.io.AccessLogCollectionSpec.Filter]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.observability.v1alpha1.access_logging#observability.enterprise.mesh.gloo.solo.io.AccessLogCollectionSpec.Filter" >}}) | repeated | Configure request criteria for determining which access logs will be collected. The list is disjunctive, a request will be collected if it matches any filter. Leave empty to emit all access logs. |
  





<a name="observability.enterprise.mesh.gloo.solo.io.AccessLogCollectionSpec.Filter"></a>

### AccessLogCollectionSpec.Filter
Specify request criteria for collecting access logs. A request must match all specified criteria to be collected.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| httpStatusCodes | []int32 | repeated | Matches a request if it contains any of the specified status codes. Omit to match any status code. |
  | headerMatchers | [][networking.mesh.gloo.solo.io.TrafficPolicySpec.HeaderMatcher]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec.HeaderMatcher" >}}) | repeated | Matches a request if it matches any of the header matchers below. Omit to match any header(s). |
  





<a name="observability.enterprise.mesh.gloo.solo.io.AccessLogCollectionStatus"></a>

### AccessLogCollectionStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the AccessLogCollection metadata. If the observedGeneration does not match generation, the controller has not processed the most recent version of this resource. |
  | state | [networking.mesh.gloo.solo.io.ApprovalState]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.validation_state#networking.mesh.gloo.solo.io.ApprovalState" >}}) |  | The state of the overall resource, will only show accepted if it has been successfully applied to all target meshes. |
  | errors | []string | repeated | Any errors encountered during processing. Also reported to any Workloads that this object applies to. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

