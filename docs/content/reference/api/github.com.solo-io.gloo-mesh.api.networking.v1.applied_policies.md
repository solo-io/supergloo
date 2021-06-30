
---

title: "applied_policies.proto"

---

## Package : `networking.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for applied_policies.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## applied_policies.proto


## Table of Contents
  - [AppliedTrafficPolicy](#networking.mesh.gloo.solo.io.AppliedTrafficPolicy)







<a name="networking.mesh.gloo.solo.io.AppliedTrafficPolicy"></a>

### AppliedTrafficPolicy
Describes a [TrafficPolicy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy" >}}) that applies to the Destination. If an existing TrafficPolicy becomes invalid, the last valid applied TrafficPolicy will be used.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | Reference to the TrafficPolicy object. |
  | observedGeneration | int64 |  | The observed generation of the accepted TrafficPolicy. |
  | spec | [networking.mesh.gloo.solo.io.TrafficPolicySpec]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec" >}}) |  | The spec of the last known valid TrafficPolicy. |
  | routes | []string | repeated | The list of routes to which the TrafficPolicy applies, as selected by their labels. Value is "*" if the TrafficPolicy applies to all routes. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

