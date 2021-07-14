
---

title: "route.proto"

---

## Package : `networking.enterprise.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for route.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## route.proto


## Table of Contents
  - [DelegateAction](#networking.enterprise.mesh.gloo.solo.io.DelegateAction)
  - [DirectResponseAction](#networking.enterprise.mesh.gloo.solo.io.DirectResponseAction)
  - [RedirectAction](#networking.enterprise.mesh.gloo.solo.io.RedirectAction)
  - [Route](#networking.enterprise.mesh.gloo.solo.io.Route)
  - [Route.LabelsEntry](#networking.enterprise.mesh.gloo.solo.io.Route.LabelsEntry)
  - [Route.RouteAction](#networking.enterprise.mesh.gloo.solo.io.Route.RouteAction)

  - [RedirectAction.RedirectResponseCode](#networking.enterprise.mesh.gloo.solo.io.RedirectAction.RedirectResponseCode)






<a name="networking.enterprise.mesh.gloo.solo.io.DelegateAction"></a>

### DelegateAction
Note: This message needs to be at this level (rather than nested) due to cue restrictions. DelegateActions are used to delegate routing decisions to other resources, for example RouteTables.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| refs | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated | Delegate to the RouteTable resources with matching `name` and `namespace`. |
  | selector | [core.skv2.solo.io.ObjectSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectSelector" >}}) |  | Delegate to the RouteTables that match the given selector. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.DirectResponseAction"></a>

### DirectResponseAction
TODO: Shuold we be using v4alpha now? https://github.com/envoyproxy/envoy/blob/5a8bfa20dc3c85ecb61826d122696ecaa75dffa0/api/envoy/config/route/v4alpha/route_components.proto#L1396 Note: This message needs to be at this level (rather than nested) due to cue restrictions. DirectResponseAction is copied directly from https://github.com/envoyproxy/envoy/blob/master/api/envoy/api/v2/route/route.proto


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| status | uint32 |  | Specifies the HTTP response status to be returned. |
  | body | string |  | Specifies the content of the response body. If this setting is omitted, no body is included in the generated response.<br>Note: Headers can be specified using the Header Modification feature in the enclosing Route, ConnectionHandler, or Gateway options. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.RedirectAction"></a>

### RedirectAction
Note: This message needs to be at this level (rather than nested) due to cue restrictions. Notice: RedirectAction is copied directly from https://github.com/envoyproxy/envoy/blob/master/api/envoy/api/v2/route/route.proto


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hostRedirect | string |  | The host portion of the URL will be swapped with this value. |
  | pathRedirect | string |  | The path portion of the URL will be swapped with this value. |
  | prefixRewrite | string |  | Indicates that during redirection, the matched prefix (or path) should be swapped with this value. This option allows redirect URLs be dynamically created based on the request.<br>  Pay attention to the use of trailing slashes as mentioned in   `RouteAction`'s `prefix_rewrite`. |
  | responseCode | [networking.enterprise.mesh.gloo.solo.io.RedirectAction.RedirectResponseCode]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.route#networking.enterprise.mesh.gloo.solo.io.RedirectAction.RedirectResponseCode" >}}) |  | The HTTP status code to use in the redirect response. The default response code is MOVED_PERMANENTLY (301). |
  | httpsRedirect | bool |  | The scheme portion of the URL will be swapped with "https". |
  | stripQuery | bool |  | Indicates that during redirection, the query portion of the URL will be removed. Default value is false. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.Route"></a>

### Route
A route specifies how to match a request and what action to take when the request is matched.<br>When a request matches on a route, the route can perform one of the following actions: - *Route* the request to a destination - Reply with a *Direct Response* - Send a *Redirect* response to the client - *Delegate* the action for the request to one or more [`RouteTable`]({{< ref "/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.route_table.md" >}}) resources DelegateActions can be used to delegate the behavior for a set out routes to `RouteTable` resources.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The name provides a convenience for users to be able to refer to a route by name. It includes names of VS, Route, and RouteTable ancestors of the Route. |
  | matchers | [][networking.mesh.gloo.solo.io.HttpMatcher]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.request_matchers#networking.mesh.gloo.solo.io.HttpMatcher" >}}) | repeated | Matchers contain parameters for matching requests (i.e., based on HTTP path, headers, etc.). If empty, the route will match all requests (i.e, a single "/" path prefix matcher). For delegated routes, any parent matcher must have a `prefix` path matcher. |
  | routeAction | [networking.enterprise.mesh.gloo.solo.io.Route.RouteAction]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.route#networking.enterprise.mesh.gloo.solo.io.Route.RouteAction" >}}) |  | This action is the primary action to be selected for most routes. The RouteAction tells the proxy to route requests to an upstream. |
  | redirectAction | [networking.enterprise.mesh.gloo.solo.io.RedirectAction]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.route#networking.enterprise.mesh.gloo.solo.io.RedirectAction" >}}) |  | Redirect actions tell the proxy to return a redirect response to the downstream client. |
  | directResponseAction | [networking.enterprise.mesh.gloo.solo.io.DirectResponseAction]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.route#networking.enterprise.mesh.gloo.solo.io.DirectResponseAction" >}}) |  | Return an arbitrary HTTP response directly, without proxying. |
  | delegateAction | [networking.enterprise.mesh.gloo.solo.io.DelegateAction]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.route#networking.enterprise.mesh.gloo.solo.io.DelegateAction" >}}) |  | Delegate routing actions for the given matcher to one or more RouteTables. |
  | options | [networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy" >}}) |  | Route Options extend the behavior of routes. Route options include configuration such as retries, rate limiting, and request/response transformation. RouteOption behavior will be inherited by delegated routes which do not specify their own `options` |
  | labels | [][networking.enterprise.mesh.gloo.solo.io.Route.LabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.route#networking.enterprise.mesh.gloo.solo.io.Route.LabelsEntry" >}}) | repeated | Specify labels for this route, which are used by other resources (e.g. TrafficPolicy) to select specific routes within a given gateway object. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.Route.LabelsEntry"></a>

### Route.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="networking.enterprise.mesh.gloo.solo.io.Route.RouteAction"></a>

### Route.RouteAction
RouteActions are used to route matched requests to upstreams.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| destinations | [][networking.mesh.gloo.solo.io.WeightedDestination]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.weighed_destination#networking.mesh.gloo.solo.io.WeightedDestination" >}}) | repeated | Defines the destination upstream for routing Some destinations require additional configuration for the route (e.g. AWS upstreams require a function name to be specified). |
  




 <!-- end messages -->


<a name="networking.enterprise.mesh.gloo.solo.io.RedirectAction.RedirectResponseCode"></a>

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

