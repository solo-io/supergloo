
---

title: "route.proto"

---

## Package : `common.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for route.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## route.proto


## Table of Contents
  - [DelegateAction](#common.mesh.gloo.solo.io.DelegateAction)
  - [DirectResponseAction](#common.mesh.gloo.solo.io.DirectResponseAction)
  - [RedirectAction](#common.mesh.gloo.solo.io.RedirectAction)
  - [Route](#common.mesh.gloo.solo.io.Route)
  - [Route.RouteAction](#common.mesh.gloo.solo.io.Route.RouteAction)
  - [Route.RouteAction.Destinations](#common.mesh.gloo.solo.io.Route.RouteAction.Destinations)
  - [Route.RouteAction.Destinations.DestinationOptions](#common.mesh.gloo.solo.io.Route.RouteAction.Destinations.DestinationOptions)
  - [Route.RouteOptions](#common.mesh.gloo.solo.io.Route.RouteOptions)

  - [RedirectAction.RedirectResponseCode](#common.mesh.gloo.solo.io.RedirectAction.RedirectResponseCode)






<a name="common.mesh.gloo.solo.io.DelegateAction"></a>

### DelegateAction
Note: This message needs to be at this level (rather than nested) due to cue restrictions. DelegateActions are used to delegate routing decisions to other resources, for example Route Tables.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | Delegate to the Route Table resource with the given `name` and `namespace. |
  | selector | [common.mesh.gloo.solo.io.VirtualHostSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors#common.mesh.gloo.solo.io.VirtualHostSelector" >}}) |  | Delegate to the Delegated Route Tables that match the given selector. |
  | passFullPath | bool |  | If set to true, `passFullPath` will send the full path for the delegated child route to match on. When false, the part of the path matched in this resource will be removed, and the delegated child resource will only match on the remainder of the path. Default value is false. |
  





<a name="common.mesh.gloo.solo.io.DirectResponseAction"></a>

### DirectResponseAction
TODO: Shuold we be using v4alpha now? https://github.com/envoyproxy/envoy/blob/5a8bfa20dc3c85ecb61826d122696ecaa75dffa0/api/envoy/config/route/v4alpha/route_components.proto#L1396 Note: This message needs to be at this level (rather than nested) due to cue restrictions. DirectResponseAction is copied directly from https://github.com/envoyproxy/envoy/blob/master/api/envoy/api/v2/route/route.proto


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| status | uint32 |  | Specifies the HTTP response status to be returned. |
  | body | string |  | Specifies the content of the response body. If this setting is omitted, no body is included in the generated response.<br>  Note: Headers can be specified using the Header Modification feature in the enclosing   Route, ConnectionHandler, or Gateway options. |
  





<a name="common.mesh.gloo.solo.io.RedirectAction"></a>

### RedirectAction
Note: This message needs to be at this level (rather than nested) due to cue restrictions. Notice: RedirectAction is copied directly from https://github.com/envoyproxy/envoy/blob/master/api/envoy/api/v2/route/route.proto


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hostRedirect | string |  | The host portion of the URL will be swapped with this value. |
  | pathRedirect | string |  | The path portion of the URL will be swapped with this value. |
  | prefixRewrite | string |  | Indicates that during redirection, the matched prefix (or path) should be swapped with this value. This option allows redirect URLs be dynamically created based on the request.<br>  Pay attention to the use of trailing slashes as mentioned in   `RouteAction`'s `prefix_rewrite`. |
  | responseCode | [common.mesh.gloo.solo.io.RedirectAction.RedirectResponseCode]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.route#common.mesh.gloo.solo.io.RedirectAction.RedirectResponseCode" >}}) |  | The HTTP status code to use in the redirect response. The default response code is MOVED_PERMANENTLY (301). |
  | httpsRedirect | bool |  | The scheme portion of the URL will be swapped with "https". |
  | stripQuery | bool |  | Indicates that during redirection, the query portion of the URL will be removed. Default value is false. |
  





<a name="common.mesh.gloo.solo.io.Route"></a>

### Route
A route specifies how to match a request and what action to take when the request is matched.<br>When a request matches on a route, the route can perform one of the following actions: - *Route* the request to a destination - Reply with a *Direct Response* - Send a *Redirect* response to the client - *Delegate* the action for the request to one or more top-level [`VirtualHost`]({{< ref "/reference/api/github.com/solo-io/gloo/projects/gateway/api/v1/virtual_host.proto.sk.md" >}}) resources DelegateActions can be used to delegate the behavior for a set out routes with a given *prefix* to top-level `VirtualHost` resources.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The name provides a convenience for users to be able to refer to a route by name. It includes names of vs, route, and route table ancestors of the route. |
  | matchers | [][common.mesh.gloo.solo.io.HttpMatcher]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.request_matchers#common.mesh.gloo.solo.io.HttpMatcher" >}}) | repeated | Matchers contain parameters for matching requests (i.e., based on HTTP path, headers, etc.). If empty, the route will match all requests (i.e, a single "/" path prefix matcher). For delegated routes, any parent matcher must have a `prefix` path matcher. |
  | routeAction | [common.mesh.gloo.solo.io.Route.RouteAction]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.route#common.mesh.gloo.solo.io.Route.RouteAction" >}}) |  | This action is the primary action to be selected for most routes. The RouteAction tells the proxy to route requests to an upstream. |
  | redirectAction | [common.mesh.gloo.solo.io.RedirectAction]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.route#common.mesh.gloo.solo.io.RedirectAction" >}}) |  | Redirect actions tell the proxy to return a redirect response to the downstream client. |
  | directResponseAction | [common.mesh.gloo.solo.io.DirectResponseAction]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.route#common.mesh.gloo.solo.io.DirectResponseAction" >}}) |  | Return an arbitrary HTTP response directly, without proxying. |
  | delegateAction | [common.mesh.gloo.solo.io.DelegateAction]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.route#common.mesh.gloo.solo.io.DelegateAction" >}}) |  | Delegate routing actions for the given matcher to one or more RouteTables. |
  | options | [common.mesh.gloo.solo.io.Route.RouteOptions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.route#common.mesh.gloo.solo.io.Route.RouteOptions" >}}) |  | Route Options extend the behavior of routes. Route options include configuration such as retries, rate limiting, and request/response transformation. RouteOption behavior will be inherited by delegated routes which do not specify their own `options` |
  





<a name="common.mesh.gloo.solo.io.Route.RouteAction"></a>

### Route.RouteAction
RouteActions are used to route matched requests to upstreams.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| destinations | [][common.mesh.gloo.solo.io.Route.RouteAction.Destinations]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.route#common.mesh.gloo.solo.io.Route.RouteAction.Destinations" >}}) | repeated | Defines the destination upstream for routing Some destinations require additional configuration for the route (e.g. AWS upstreams require a function name to be specified). |
  





<a name="common.mesh.gloo.solo.io.Route.RouteAction.Destinations"></a>

### Route.RouteAction.Destinations



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| static | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | Reference to a gloo mesh Static Destination |
  | virtual | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | Reference to a gloo mesh VirtualDestination |
  | kube | [core.skv2.solo.io.ClusterObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ClusterObjectRef" >}}) |  | Reference to a Kubernetes Service. Note that the service must exist in the same mesh or virtual mesh (with federation enabled) as  each gateway workload which routes to this destination. |
  | clusterHeader | string |  | Envoy will determine the cluster to route to by reading the value of the HTTP header named by cluster_header from the request headers. If the header is not found or the referenced cluster does not exist, Envoy will return a 404 response. Avoid using this whenever possible, it does not allow for custom filter configuration based on Virtual Host. |
  | weight | uint32 |  | Relative weight of this destination to others in the same route. If omitted, all destinations in the route will be load balanced between evenly. |
  | destinationOptions | [common.mesh.gloo.solo.io.Route.RouteAction.Destinations.DestinationOptions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.route#common.mesh.gloo.solo.io.Route.RouteAction.Destinations.DestinationOptions" >}}) |  | options applied when this destination is selected from a list of multiple destinations |
  





<a name="common.mesh.gloo.solo.io.Route.RouteAction.Destinations.DestinationOptions"></a>

### Route.RouteAction.Destinations.DestinationOptions



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headerModification | string |  | TODO: Some subset of traffic policy (whatever Istio supports) TODO: Use correct type for header_modification |
  





<a name="common.mesh.gloo.solo.io.Route.RouteOptions"></a>

### Route.RouteOptions
TODO: Route Options





 <!-- end messages -->


<a name="common.mesh.gloo.solo.io.RedirectAction.RedirectResponseCode"></a>

### RedirectAction.RedirectResponseCode


| Name | Number | Description |
| ---- | ------ | ----------- |
| MOVED_PERMANENTLY | 0 | Moved Permanently HTTP Status Code - 301. |
| FOUND | 1 | Found HTTP Status Code - 302. |
| SEE_OTHER | 2 | See Other HTTP Status Code - 303. |
| TEMPORARY_REDIRECT | 3 | Temporary Redirect HTTP Status Code - 307. |
| PERMANENT_REDIRECT | 4 | Permanent Redirect HTTP Status Code - 308. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

