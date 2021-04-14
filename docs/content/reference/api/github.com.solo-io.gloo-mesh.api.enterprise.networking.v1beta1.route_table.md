
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
  - [RouteTableSpec.RouteTableOptions](#networking.enterprise.mesh.gloo.solo.io.RouteTableSpec.RouteTableOptions)
  - [RouteTableStatus](#networking.enterprise.mesh.gloo.solo.io.RouteTableStatus)







<a name="networking.enterprise.mesh.gloo.solo.io.RouteTableSpec"></a>

### RouteTableSpec
A `RouteTable` is used to configure routes. It is selected by a `FederatedGateway`, and may be attached to more than one gateway. The `RouteTable` contains the top-level configuration and route options, such as domains to match against, and any options to be shared by its routes. Routes can send traffic directly to a service, or can delegate to a `DelegatedRouteTable` to perform further routing decisions.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| domains | []string | repeated | The list of domains (i.e.: matching the `Host` header of a request) that belong to this virtual host. Note that the wildcard will not match the empty string. e.g. “*-bar.foo.com” will match “baz-bar.foo.com” but not “-bar.foo.com”. Additionally, a special entry “*” is allowed which will match any host/authority header. Only a single virtual host on a gateway can match on “*”. A domain must be unique across all virtual hosts on a gateway or the config will be invalidated by Gloo Domains on virtual hosts obey the same rules as [Envoy Virtual Hosts](https://github.com/envoyproxy/envoy/blob/master/api/envoy/api/v2/route/route.proto) |
  | routes | [][common.mesh.gloo.solo.io.Route]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.route#common.mesh.gloo.solo.io.Route" >}}) | repeated | The list of HTTP routes define routing actions to be taken for incoming HTTP requests whose host header matches this virtual host. If the request matches more than one route in the list, the first route matched will be selected. If the list of routes is empty, the virtual host will be ignored by Gloo. |
  | routeTableOptions | [networking.enterprise.mesh.gloo.solo.io.RouteTableSpec.RouteTableOptions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.route_table#networking.enterprise.mesh.gloo.solo.io.RouteTableSpec.RouteTableOptions" >}}) |  | Route table options contain additional configuration to be applied to all traffic served by the route table. Some configuration here can be overridden by Route Options. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.RouteTableSpec.RouteTableOptions"></a>

### RouteTableSpec.RouteTableOptions
TODO: Fill / maybe replace with traffic policy?<br>see message VirtualHostOptions in options.proto


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| todoAddOptions | string |  |  |
  





<a name="networking.enterprise.mesh.gloo.solo.io.RouteTableStatus"></a>

### RouteTableStatus
TODO: Fill in status





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

