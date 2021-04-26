
---

title: "delegated_route_table.proto"

---

## Package : `networking.enterprise.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for delegated_route_table.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## delegated_route_table.proto


## Table of Contents
  - [DelegatedRouteTableSpec](#networking.enterprise.mesh.gloo.solo.io.DelegatedRouteTableSpec)
  - [DelegatedRouteTableStatus](#networking.enterprise.mesh.gloo.solo.io.DelegatedRouteTableStatus)
  - [selectedBy](#networking.enterprise.mesh.gloo.solo.io.selectedBy)







<a name="networking.enterprise.mesh.gloo.solo.io.DelegatedRouteTableSpec"></a>

### DelegatedRouteTableSpec
DelegatedRouteTable is a resource which can be referenced either from the top level RouteTable resource, or from other DelegatedRouteTables. It's primary use is to organizationally and logically separate the configuration of routes, so that the responsbilities of route configuration and maintenance can be divided between teams where appropriate.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| routes | [][common.mesh.gloo.solo.io.Route]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.route#common.mesh.gloo.solo.io.Route" >}}) | repeated | The list of HTTP routes define routing actions to be taken for incoming HTTP requests whose host header matches this virtual host. If the request matches more than one route in the list, the first route matched will be selected. If the list of routes is empty, the virtual host will be ignored by Gloo. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.DelegatedRouteTableStatus"></a>

### DelegatedRouteTableStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the DelegatedRouteTable metadata. If the `observedGeneration` does not match `metadata.generation`, Gloo Mesh has not processed the most recent version of this resource. |
  | errors | []string | repeated | Any errors found while processing this generation of the resource. |
  | warnings | []string | repeated | Any warnings found while processing this generation of the resource. |
  | selectedBy | [][networking.enterprise.mesh.gloo.solo.io.selectedBy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.delegated_route_table#networking.enterprise.mesh.gloo.solo.io.selectedBy" >}}) | repeated | List of resources which have selected this DelegatedRouteTable. Can be RouteTables or other DelegatedRouteTables |
  | selectedDelegatedRouteTables | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated | List of DelegatedRouteTables that this DelegatedRouteTable delegates to |
  





<a name="networking.enterprise.mesh.gloo.solo.io.selectedBy"></a>

### selectedBy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Name of resource selecting this DelegatedRouteTable |
  | namespace | string |  | Namespace of resource selecting this DelegatedRouteTable |
  | type | string |  | Type of resource selecting this DelegatedRoute Table. Can be FederatedGateway, RouteTable, or DelegatedRouteTable. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

