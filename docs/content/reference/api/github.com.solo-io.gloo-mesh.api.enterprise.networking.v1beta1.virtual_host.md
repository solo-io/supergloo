
---

title: "virtual_host.proto"

---

## Package : `networking.enterprise.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for virtual_host.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## virtual_host.proto


## Table of Contents
  - [VirtualHostOptions](#networking.enterprise.mesh.gloo.solo.io.VirtualHostOptions)
  - [VirtualHostSpec](#networking.enterprise.mesh.gloo.solo.io.VirtualHostSpec)
  - [VirtualHostStatus](#networking.enterprise.mesh.gloo.solo.io.VirtualHostStatus)







<a name="networking.enterprise.mesh.gloo.solo.io.VirtualHostOptions"></a>

### VirtualHostOptions



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| trafficPolicy | [networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy" >}}) |  |  |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualHostSpec"></a>

### VirtualHostSpec
A `VirtualHost` is used to configure routes. It is selected by a `VirtualGateway`, and may be attached to more than one gateway. The `VirtualHost` contains the top-level configuration and route options, such as domains to match against, and any options to be shared by its routes. Routes can send traffic directly to a service, or can delegate to a `RouteTable` to perform further routing decisions.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| domains | []string | repeated | The list of domains (i.e.: matching the `Host` header of a request) that belong to this virtual host. Note that the wildcard will not match the empty string. e.g. “*-bar.foo.com” will match “baz-bar.foo.com” but not “-bar.foo.com”. Additionally, a special entry “*” is allowed which will match any host/authority header. Only a single virtual host on a gateway can match on “*”. A domain must be unique across all virtual hosts on a gateway or the config will be invalidated by Gloo Domains on virtual hosts obey the same rules as [Envoy Virtual Hosts](https://github.com/envoyproxy/envoy/blob/master/api/envoy/api/v2/route/route.proto) |
  | routes | [][networking.enterprise.mesh.gloo.solo.io.Route]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.route#networking.enterprise.mesh.gloo.solo.io.Route" >}}) | repeated | The list of HTTP routes define routing actions to be taken for incoming HTTP requests whose host header matches this virtual host. If the request matches more than one route in the list, the first route matched will be selected. If the list of routes is empty, the virtual host will be ignored by Gloo. |
  | options | [networking.enterprise.mesh.gloo.solo.io.VirtualHostOptions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.virtual_host#networking.enterprise.mesh.gloo.solo.io.VirtualHostOptions" >}}) |  | Route table options contain additional configuration to be applied to all traffic served by the route table. Some configuration here can be overridden by Route Options. OutlierDetection and TrafficShift isn't supported on the route level. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.VirtualHostStatus"></a>

### VirtualHostStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the VirtualHost metadata. If the `observedGeneration` does not match `metadata.generation`, Gloo Mesh has not processed the most recent version of this resource. |
  | errors | []string | repeated | Any errors found while processing this generation of the resource. |
  | warnings | []string | repeated | Any warnings found while processing this generation of the resource. |
  | attachedVirtualGateways | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated | List of each VirtualGateway which has selected this VirtualHost |
  | selectedRouteTables | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated | List of RouteTables that this Route table delegates to |
  | appliedTrafficPolicies | [][networking.mesh.gloo.solo.io.AppliedTrafficPolicy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.applied_policies#networking.mesh.gloo.solo.io.AppliedTrafficPolicy" >}}) | repeated | The set of TrafficPolicies that have been applied to this Destination. {{/* Note: validation of this field disabled because it slows down cue tremendously*/}} |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

