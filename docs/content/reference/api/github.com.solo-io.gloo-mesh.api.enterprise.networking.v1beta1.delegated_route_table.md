
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







<a name="networking.enterprise.mesh.gloo.solo.io.DelegatedRouteTableSpec"></a>

### DelegatedRouteTableSpec
DelegatedRouteTable is a resource which can be referenced either from the top level RouteTable resource, or from other DelegatedRouteTables. It's primary use is to organizationally and logically separate the configuration of routes, so that the responsbilities of route configuration and maintenance can be divided between teams where appropriate.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| routes | [][common.mesh.gloo.solo.io.Route]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.route#common.mesh.gloo.solo.io.Route" >}}) | repeated | The list of HTTP routes define routing actions to be taken for incoming HTTP requests whose host header matches this virtual host. If the request matches more than one route in the list, the first route matched will be selected. If the list of routes is empty, the virtual host will be ignored by Gloo. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.DelegatedRouteTableStatus"></a>

### DelegatedRouteTableStatus
TODO: Fill in status





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

