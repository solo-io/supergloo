
---

title: "route_table.proto"

---

## Package : `networking.enterprise.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for route_table.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## route_table.proto


## Table of Contents
  - [RouteTableSpec](#networking.enterprise.mesh.gloo.solo.io.RouteTableSpec)
  - [RouteTableStatus](#networking.enterprise.mesh.gloo.solo.io.RouteTableStatus)
  - [SelectedBy](#networking.enterprise.mesh.gloo.solo.io.SelectedBy)







<a name="networking.enterprise.mesh.gloo.solo.io.RouteTableSpec"></a>

### RouteTableSpec
RouteTable is a resource which can be referenced either from the top level VirtualHost resource, or from other RouteTables. It's primary use is to organizationally and logically separate the configuration of Routes, so that the responsibilities of Route configuration and maintenance can be divided between teams where appropriate.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| routes | [][networking.enterprise.mesh.gloo.solo.io.Route]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.route#networking.enterprise.mesh.gloo.solo.io.Route" >}}) | repeated | The list of HTTP Routes define routing actions to be taken for incoming HTTP requests whose host header matches this virtual host. If the request matches more than one Route in the list, the first Route matched will be selected. If the list of Routes is empty, the virtual host will be ignored by Gloo. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.RouteTableStatus"></a>

### RouteTableStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the RouteTable metadata. If the `observedGeneration` does not match `metadata.generation`, Gloo Mesh has not processed the most recent version of this resource. |
  | errors | []string | repeated | Any errors found while processing this generation of the resource. |
  | warnings | []string | repeated | Any warnings found while processing this generation of the resource. |
  | selectedBy | [][networking.enterprise.mesh.gloo.solo.io.SelectedBy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.route_table#networking.enterprise.mesh.gloo.solo.io.SelectedBy" >}}) | repeated | List of resources which have selected this RouteTable. Can be VirtualHosts or other RouteTables |
  | selectedRouteTables | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated | List of child RouteTables that this RouteTable delegates to |
  | appliedTrafficPolicies | [][networking.mesh.gloo.solo.io.AppliedTrafficPolicy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.applied_policies#networking.mesh.gloo.solo.io.AppliedTrafficPolicy" >}}) | repeated | The set of TrafficPolicies that have been applied to this Destination. {{/* Note: validation of this field disabled because it slows down cue tremendously*/}} |
  





<a name="networking.enterprise.mesh.gloo.solo.io.SelectedBy"></a>

### SelectedBy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Name of resource selecting this RouteTable |
  | namespace | string |  | Namespace of resource selecting this RouteTable |
  | type | string |  | Type of resource selecting this RouteTable. Can be VirtualGateway, VirtualHost, or RouteTable. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

