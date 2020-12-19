
---
title: "virtual_service.proto"
---

## Package : `istio.networking.v1alpha3`



<a name="top"></a>

<a name="API Reference for virtual_service.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## virtual_service.proto


## Table of Contents
  - [CorsPolicy](#istio.networking.v1alpha3.CorsPolicy)
  - [Delegate](#istio.networking.v1alpha3.Delegate)
  - [Destination](#istio.networking.v1alpha3.Destination)
  - [HTTPFaultInjection](#istio.networking.v1alpha3.HTTPFaultInjection)
  - [HTTPFaultInjection.Abort](#istio.networking.v1alpha3.HTTPFaultInjection.Abort)
  - [HTTPFaultInjection.Delay](#istio.networking.v1alpha3.HTTPFaultInjection.Delay)
  - [HTTPMatchRequest](#istio.networking.v1alpha3.HTTPMatchRequest)
  - [HTTPMatchRequest.HeadersEntry](#istio.networking.v1alpha3.HTTPMatchRequest.HeadersEntry)
  - [HTTPMatchRequest.QueryParamsEntry](#istio.networking.v1alpha3.HTTPMatchRequest.QueryParamsEntry)
  - [HTTPMatchRequest.SourceLabelsEntry](#istio.networking.v1alpha3.HTTPMatchRequest.SourceLabelsEntry)
  - [HTTPMatchRequest.WithoutHeadersEntry](#istio.networking.v1alpha3.HTTPMatchRequest.WithoutHeadersEntry)
  - [HTTPRedirect](#istio.networking.v1alpha3.HTTPRedirect)
  - [HTTPRetry](#istio.networking.v1alpha3.HTTPRetry)
  - [HTTPRewrite](#istio.networking.v1alpha3.HTTPRewrite)
  - [HTTPRoute](#istio.networking.v1alpha3.HTTPRoute)
  - [HTTPRouteDestination](#istio.networking.v1alpha3.HTTPRouteDestination)
  - [Headers](#istio.networking.v1alpha3.Headers)
  - [Headers.HeaderOperations](#istio.networking.v1alpha3.Headers.HeaderOperations)
  - [Headers.HeaderOperations.AddEntry](#istio.networking.v1alpha3.Headers.HeaderOperations.AddEntry)
  - [Headers.HeaderOperations.SetEntry](#istio.networking.v1alpha3.Headers.HeaderOperations.SetEntry)
  - [L4MatchAttributes](#istio.networking.v1alpha3.L4MatchAttributes)
  - [L4MatchAttributes.SourceLabelsEntry](#istio.networking.v1alpha3.L4MatchAttributes.SourceLabelsEntry)
  - [Percent](#istio.networking.v1alpha3.Percent)
  - [PortSelector](#istio.networking.v1alpha3.PortSelector)
  - [RouteDestination](#istio.networking.v1alpha3.RouteDestination)
  - [StringMatch](#istio.networking.v1alpha3.StringMatch)
  - [TCPRoute](#istio.networking.v1alpha3.TCPRoute)
  - [TLSMatchAttributes](#istio.networking.v1alpha3.TLSMatchAttributes)
  - [TLSMatchAttributes.SourceLabelsEntry](#istio.networking.v1alpha3.TLSMatchAttributes.SourceLabelsEntry)
  - [TLSRoute](#istio.networking.v1alpha3.TLSRoute)
  - [VirtualService](#istio.networking.v1alpha3.VirtualService)







<a name="istio.networking.v1alpha3.CorsPolicy"></a>

### CorsPolicy
Describes the Cross-Origin Resource Sharing (CORS) policy, for a given service. Refer to [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/Access_control_CORS) for further details about cross origin resource sharing. For example, the following rule restricts cross origin requests to those originating from example.com domain using HTTP POST/GET, and sets the `Access-Control-Allow-Credentials` header to false. In addition, it only exposes `X-Foo-bar` header and sets an expiry period of 1 day.<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: VirtualService metadata:   name: ratings-route spec:   hosts:   - ratings.prod.svc.cluster.local   http:   - route:     - destination:         host: ratings.prod.svc.cluster.local         subset: v1     corsPolicy:       allowOrigins:       - exact: https://example.com       allowMethods:       - POST       - GET       allowCredentials: false       allowHeaders:       - X-Foo-Bar       maxAge: "24h" ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: VirtualService metadata:   name: ratings-route spec:   hosts:   - ratings.prod.svc.cluster.local   http:   - route:     - destination:         host: ratings.prod.svc.cluster.local         subset: v1     corsPolicy:       allowOrigins:       - exact: https://example.com       allowMethods:       - POST       - GET       allowCredentials: false       allowHeaders:       - X-Foo-Bar       maxAge: "24h" ``` {{</tab>}} {{</tabs>}}


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| allowOrigin | []string | repeated | The list of origins that are allowed to perform CORS requests. The content will be serialized into the Access-Control-Allow-Origin header. Wildcard * will allow all origins. $hide_from_docs |
  | allowOrigins | [][istio.networking.v1alpha3.StringMatch]({{< ref "virtual_service.md#istio.networking.v1alpha3.StringMatch" >}}) | repeated | String patterns that match allowed origins. An origin is allowed if any of the string matchers match. If a match is found, then the outgoing Access-Control-Allow-Origin would be set to the origin as provided by the client. |
  | allowMethods | []string | repeated | List of HTTP methods allowed to access the resource. The content will be serialized into the Access-Control-Allow-Methods header. |
  | allowHeaders | []string | repeated | List of HTTP headers that can be used when requesting the resource. Serialized to Access-Control-Allow-Headers header. |
  | exposeHeaders | []string | repeated | A list of HTTP headers that the browsers are allowed to access. Serialized into Access-Control-Expose-Headers header. |
  | maxAge | [google.protobuf.Duration]({{< ref "duration.md#google.protobuf.Duration" >}}) |  | Specifies how long the results of a preflight request can be cached. Translates to the `Access-Control-Max-Age` header. |
  | allowCredentials | [google.protobuf.BoolValue]({{< ref "wrappers.md#google.protobuf.BoolValue" >}}) |  | Indicates whether the caller is allowed to send the actual request (not the preflight) using credentials. Translates to `Access-Control-Allow-Credentials` header. |
  





<a name="istio.networking.v1alpha3.Delegate"></a>

### Delegate
Describes the delegate VirtualService. The following routing rules forward the traffic to `/productpage` by a delegate VirtualService named `productpage`, forward the traffic to `/reviews` by a delegate VirtualService named `reviews`.<br>```yaml apiVersion: networking.istio.io/v1alpha3 kind: VirtualService metadata:   name: bookinfo spec:   hosts:   - "bookinfo.com"   gateways:   - mygateway   http:   - match:     - uri:         prefix: "/productpage"     delegate:        name: productpage        namespace: nsA   - match:     - uri:         prefix: "/reviews"     delegate:         name: reviews         namespace: nsB ```<br>```yaml apiVersion: networking.istio.io/v1alpha3 kind: VirtualService metadata:   name: productpage   namespace: nsA spec:   http:   - match:      - uri:         prefix: "/productpage/v1/"     route:     - destination:         host: productpage-v1.nsA.svc.cluster.local   - route:     - destination:         host: productpage.nsA.svc.cluster.local ```<br>```yaml apiVersion: networking.istio.io/v1alpha3 kind: VirtualService metadata:   name: reviews   namespace: nsB spec:   http:   - route:     - destination:         host: reviews.nsB.svc.cluster.local ```


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Name specifies the name of the delegate VirtualService. |
  | namespace | string |  | Namespace specifies the namespace where the delegate VirtualService resides. By default, it is same to the root's. |
  





<a name="istio.networking.v1alpha3.Destination"></a>

### Destination
Destination indicates the network addressable service to which the request/connection will be sent after processing a routing rule. The destination.host should unambiguously refer to a service in the service registry. Istio's service registry is composed of all the services found in the platform's service registry (e.g., Kubernetes services, Consul services), as well as services declared through the [ServiceEntry](https://istio.io/docs/reference/config/networking/service-entry/#ServiceEntry) resource.<br>*Note for Kubernetes users*: When short names are used (e.g. "reviews" instead of "reviews.default.svc.cluster.local"), Istio will interpret the short name based on the namespace of the rule, not the service. A rule in the "default" namespace containing a host "reviews will be interpreted as "reviews.default.svc.cluster.local", irrespective of the actual namespace associated with the reviews service. _To avoid potential misconfigurations, it is recommended to always use fully qualified domain names over short names._<br>The following Kubernetes example routes all traffic by default to pods of the reviews service with label "version: v1" (i.e., subset v1), and some to subset v2, in a Kubernetes environment.<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: VirtualService metadata:   name: reviews-route   namespace: foo spec:   hosts:   - reviews # interpreted as reviews.foo.svc.cluster.local   http:   - match:     - uri:         prefix: "/wpcatalog"     - uri:         prefix: "/consumercatalog"     rewrite:       uri: "/newcatalog"     route:     - destination:         host: reviews # interpreted as reviews.foo.svc.cluster.local         subset: v2   - route:     - destination:         host: reviews # interpreted as reviews.foo.svc.cluster.local         subset: v1 ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: VirtualService metadata:   name: reviews-route   namespace: foo spec:   hosts:   - reviews # interpreted as reviews.foo.svc.cluster.local   http:   - match:     - uri:         prefix: "/wpcatalog"     - uri:         prefix: "/consumercatalog"     rewrite:       uri: "/newcatalog"     route:     - destination:         host: reviews # interpreted as reviews.foo.svc.cluster.local         subset: v2   - route:     - destination:         host: reviews # interpreted as reviews.foo.svc.cluster.local         subset: v1 ``` {{</tab>}} {{</tabs>}}<br>And the associated DestinationRule<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: DestinationRule metadata:   name: reviews-destination   namespace: foo spec:   host: reviews # interpreted as reviews.foo.svc.cluster.local   subsets:   - name: v1     labels:       version: v1   - name: v2     labels:       version: v2 ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: DestinationRule metadata:   name: reviews-destination   namespace: foo spec:   host: reviews # interpreted as reviews.foo.svc.cluster.local   subsets:   - name: v1     labels:       version: v1   - name: v2     labels:       version: v2 ``` {{</tab>}} {{</tabs>}}<br>The following VirtualService sets a timeout of 5s for all calls to productpage.prod.svc.cluster.local service in Kubernetes. Notice that there are no subsets defined in this rule. Istio will fetch all instances of productpage.prod.svc.cluster.local service from the service registry and populate the sidecar's load balancing pool. Also, notice that this rule is set in the istio-system namespace but uses the fully qualified domain name of the productpage service, productpage.prod.svc.cluster.local. Therefore the rule's namespace does not have an impact in resolving the name of the productpage service.<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: VirtualService metadata:   name: my-productpage-rule   namespace: istio-system spec:   hosts:   - productpage.prod.svc.cluster.local # ignores rule namespace   http:   - timeout: 5s     route:     - destination:         host: productpage.prod.svc.cluster.local ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: VirtualService metadata:   name: my-productpage-rule   namespace: istio-system spec:   hosts:   - productpage.prod.svc.cluster.local # ignores rule namespace   http:   - timeout: 5s     route:     - destination:         host: productpage.prod.svc.cluster.local ``` {{</tab>}} {{</tabs>}}<br>To control routing for traffic bound to services outside the mesh, external services must first be added to Istio's internal service registry using the ServiceEntry resource. VirtualServices can then be defined to control traffic bound to these external services. For example, the following rules define a Service for wikipedia.org and set a timeout of 5s for HTTP requests.<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: ServiceEntry metadata:   name: external-svc-wikipedia spec:   hosts:   - wikipedia.org   location: MESH_EXTERNAL   ports:   - number: 80     name: example-http     protocol: HTTP   resolution: DNS<br>apiVersion: networking.istio.io/v1alpha3 kind: VirtualService metadata:   name: my-wiki-rule spec:   hosts:   - wikipedia.org   http:   - timeout: 5s     route:     - destination:         host: wikipedia.org ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: ServiceEntry metadata:   name: external-svc-wikipedia spec:   hosts:   - wikipedia.org   location: MESH_EXTERNAL   ports:   - number: 80     name: example-http     protocol: HTTP   resolution: DNS<br>apiVersion: networking.istio.io/v1alpha3 kind: VirtualService metadata:   name: my-wiki-rule spec:   hosts:   - wikipedia.org   http:   - timeout: 5s     route:     - destination:         host: wikipedia.org ``` {{</tab>}} {{</tabs>}}


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| host | string |  | The name of a service from the service registry. Service names are looked up from the platform's service registry (e.g., Kubernetes services, Consul services, etc.) and from the hosts declared by [ServiceEntry](https://istio.io/docs/reference/config/networking/service-entry/#ServiceEntry). Traffic forwarded to destinations that are not found in either of the two, will be dropped.<br>*Note for Kubernetes users*: When short names are used (e.g. "reviews" instead of "reviews.default.svc.cluster.local"), Istio will interpret the short name based on the namespace of the rule, not the service. A rule in the "default" namespace containing a host "reviews will be interpreted as "reviews.default.svc.cluster.local", irrespective of the actual namespace associated with the reviews service. To avoid potential misconfiguration, it is recommended to always use fully qualified domain names over short names. |
  | subset | string |  | The name of a subset within the service. Applicable only to services within the mesh. The subset must be defined in a corresponding DestinationRule. |
  | port | [istio.networking.v1alpha3.PortSelector]({{< ref "virtual_service.md#istio.networking.v1alpha3.PortSelector" >}}) |  | Specifies the port on the host that is being addressed. If a service exposes only a single port it is not required to explicitly select the port. |
  





<a name="istio.networking.v1alpha3.HTTPFaultInjection"></a>

### HTTPFaultInjection
HTTPFaultInjection can be used to specify one or more faults to inject while forwarding HTTP requests to the destination specified in a route. Fault specification is part of a VirtualService rule. Faults include aborting the Http request from downstream service, and/or delaying proxying of requests. A fault rule MUST HAVE delay or abort or both.<br>*Note:* Delay and abort faults are independent of one another, even if both are specified simultaneously.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| delay | [istio.networking.v1alpha3.HTTPFaultInjection.Delay]({{< ref "virtual_service.md#istio.networking.v1alpha3.HTTPFaultInjection.Delay" >}}) |  | Delay requests before forwarding, emulating various failures such as network issues, overloaded upstream service, etc. |
  | abort | [istio.networking.v1alpha3.HTTPFaultInjection.Abort]({{< ref "virtual_service.md#istio.networking.v1alpha3.HTTPFaultInjection.Abort" >}}) |  | Abort Http request attempts and return error codes back to downstream service, giving the impression that the upstream service is faulty. |
  





<a name="istio.networking.v1alpha3.HTTPFaultInjection.Abort"></a>

### HTTPFaultInjection.Abort
Abort specification is used to prematurely abort a request with a pre-specified error code. The following example will return an HTTP 400 error code for 1 out of every 1000 requests to the "ratings" service "v1".<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: VirtualService metadata:   name: ratings-route spec:   hosts:   - ratings.prod.svc.cluster.local   http:   - route:     - destination:         host: ratings.prod.svc.cluster.local         subset: v1     fault:       abort:         percentage:           value: 0.1         httpStatus: 400 ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: VirtualService metadata:   name: ratings-route spec:   hosts:   - ratings.prod.svc.cluster.local   http:   - route:     - destination:         host: ratings.prod.svc.cluster.local         subset: v1     fault:       abort:         percentage:           value: 0.1         httpStatus: 400 ``` {{</tab>}} {{</tabs>}}<br>The _httpStatus_ field is used to indicate the HTTP status code to return to the caller. The optional _percentage_ field can be used to only abort a certain percentage of requests. If not specified, all requests are aborted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| httpStatus | int32 |  | HTTP status code to use to abort the Http request. |
  | grpcStatus | string |  | $hide_from_docs |
  | http2Error | string |  | $hide_from_docs |
  | percentage | [istio.networking.v1alpha3.Percent]({{< ref "virtual_service.md#istio.networking.v1alpha3.Percent" >}}) |  | Percentage of requests to be aborted with the error code provided. |
  





<a name="istio.networking.v1alpha3.HTTPFaultInjection.Delay"></a>

### HTTPFaultInjection.Delay
Delay specification is used to inject latency into the request forwarding path. The following example will introduce a 5 second delay in 1 out of every 1000 requests to the "v1" version of the "reviews" service from all pods with label env: prod<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: VirtualService metadata:   name: reviews-route spec:   hosts:   - reviews.prod.svc.cluster.local   http:   - match:     - sourceLabels:         env: prod     route:     - destination:         host: reviews.prod.svc.cluster.local         subset: v1     fault:       delay:         percentage:           value: 0.1         fixedDelay: 5s ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: VirtualService metadata:   name: reviews-route spec:   hosts:   - reviews.prod.svc.cluster.local   http:   - match:     - sourceLabels:         env: prod     route:     - destination:         host: reviews.prod.svc.cluster.local         subset: v1     fault:       delay:         percentage:           value: 0.1         fixedDelay: 5s ``` {{</tab>}} {{</tabs>}}<br>The _fixedDelay_ field is used to indicate the amount of delay in seconds. The optional _percentage_ field can be used to only delay a certain percentage of requests. If left unspecified, all request will be delayed.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| percent | int32 |  | Percentage of requests on which the delay will be injected (0-100). Use of integer `percent` value is deprecated. Use the double `percentage` field instead. |
  | fixedDelay | [google.protobuf.Duration]({{< ref "duration.md#google.protobuf.Duration" >}}) |  | Add a fixed delay before forwarding the request. Format: 1h/1m/1s/1ms. MUST be >=1ms. |
  | exponentialDelay | [google.protobuf.Duration]({{< ref "duration.md#google.protobuf.Duration" >}}) |  | $hide_from_docs |
  | percentage | [istio.networking.v1alpha3.Percent]({{< ref "virtual_service.md#istio.networking.v1alpha3.Percent" >}}) |  | Percentage of requests on which the delay will be injected. |
  





<a name="istio.networking.v1alpha3.HTTPMatchRequest"></a>

### HTTPMatchRequest
HttpMatchRequest specifies a set of criterion to be met in order for the rule to be applied to the HTTP request. For example, the following restricts the rule to match only requests where the URL path starts with /ratings/v2/ and the request contains a custom `end-user` header with value `jason`.<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: VirtualService metadata:   name: ratings-route spec:   hosts:   - ratings.prod.svc.cluster.local   http:   - match:     - headers:         end-user:           exact: jason       uri:         prefix: "/ratings/v2/"       ignoreUriCase: true     route:     - destination:         host: ratings.prod.svc.cluster.local ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: VirtualService metadata:   name: ratings-route spec:   hosts:   - ratings.prod.svc.cluster.local   http:   - match:     - headers:         end-user:           exact: jason       uri:         prefix: "/ratings/v2/"       ignoreUriCase: true     route:     - destination:         host: ratings.prod.svc.cluster.local ``` {{</tab>}} {{</tabs>}}<br>HTTPMatchRequest CANNOT be empty. **Note:** No regex string match can be set when delegate VirtualService is specified.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The name assigned to a match. The match's name will be concatenated with the parent route's name and will be logged in the access logs for requests matching this route. |
  | uri | [istio.networking.v1alpha3.StringMatch]({{< ref "virtual_service.md#istio.networking.v1alpha3.StringMatch" >}}) |  | URI to match values are case-sensitive and formatted as follows:<br>- `exact: "value"` for exact string match<br>- `prefix: "value"` for prefix-based match<br>- `regex: "value"` for ECMAscript style regex-based match<br>**Note:** Case-insensitive matching could be enabled via the `ignore_uri_case` flag. |
  | scheme | [istio.networking.v1alpha3.StringMatch]({{< ref "virtual_service.md#istio.networking.v1alpha3.StringMatch" >}}) |  | URI Scheme values are case-sensitive and formatted as follows:<br>- `exact: "value"` for exact string match<br>- `prefix: "value"` for prefix-based match<br>- `regex: "value"` for ECMAscript style regex-based match |
  | method | [istio.networking.v1alpha3.StringMatch]({{< ref "virtual_service.md#istio.networking.v1alpha3.StringMatch" >}}) |  | HTTP Method values are case-sensitive and formatted as follows:<br>- `exact: "value"` for exact string match<br>- `prefix: "value"` for prefix-based match<br>- `regex: "value"` for ECMAscript style regex-based match |
  | authority | [istio.networking.v1alpha3.StringMatch]({{< ref "virtual_service.md#istio.networking.v1alpha3.StringMatch" >}}) |  | HTTP Authority values are case-sensitive and formatted as follows:<br>- `exact: "value"` for exact string match<br>- `prefix: "value"` for prefix-based match<br>- `regex: "value"` for ECMAscript style regex-based match |
  | headers | [][istio.networking.v1alpha3.HTTPMatchRequest.HeadersEntry]({{< ref "virtual_service.md#istio.networking.v1alpha3.HTTPMatchRequest.HeadersEntry" >}}) | repeated | The header keys must be lowercase and use hyphen as the separator, e.g. _x-request-id_.<br>Header values are case-sensitive and formatted as follows:<br>- `exact: "value"` for exact string match<br>- `prefix: "value"` for prefix-based match<br>- `regex: "value"` for ECMAscript style regex-based match<br>If the value is empty and only the name of header is specfied, presence of the header is checked. **Note:** The keys `uri`, `scheme`, `method`, and `authority` will be ignored. |
  | port | uint32 |  | Specifies the ports on the host that is being addressed. Many services only expose a single port or label ports with the protocols they support, in these cases it is not required to explicitly select the port. |
  | sourceLabels | [][istio.networking.v1alpha3.HTTPMatchRequest.SourceLabelsEntry]({{< ref "virtual_service.md#istio.networking.v1alpha3.HTTPMatchRequest.SourceLabelsEntry" >}}) | repeated | One or more labels that constrain the applicability of a rule to workloads with the given labels. If the VirtualService has a list of gateways specified in the top-level `gateways` field, it must include the reserved gateway `mesh` for this field to be applicable. |
  | gateways | []string | repeated | Names of gateways where the rule should be applied. Gateway names in the top-level `gateways` field of the VirtualService (if any) are overridden. The gateway match is independent of sourceLabels. |
  | queryParams | [][istio.networking.v1alpha3.HTTPMatchRequest.QueryParamsEntry]({{< ref "virtual_service.md#istio.networking.v1alpha3.HTTPMatchRequest.QueryParamsEntry" >}}) | repeated | Query parameters for matching.<br>Ex: - For a query parameter like "?key=true", the map key would be "key" and   the string match could be defined as `exact: "true"`. - For a query parameter like "?key", the map key would be "key" and the   string match could be defined as `exact: ""`. - For a query parameter like "?key=123", the map key would be "key" and the   string match could be defined as `regex: "\d+$"`. Note that this   configuration will only match values like "123" but not "a123" or "123a".<br>**Note:** `prefix` matching is currently not supported. |
  | ignoreUriCase | bool |  | Flag to specify whether the URI matching should be case-insensitive.<br>**Note:** The case will be ignored only in the case of `exact` and `prefix` URI matches. |
  | withoutHeaders | [][istio.networking.v1alpha3.HTTPMatchRequest.WithoutHeadersEntry]({{< ref "virtual_service.md#istio.networking.v1alpha3.HTTPMatchRequest.WithoutHeadersEntry" >}}) | repeated | withoutHeader has the same syntax with the header, but has opposite meaning. If a header is matched with a matching rule among withoutHeader, the traffic becomes not matched one. |
  | sourceNamespace | string |  | Source namespace constraining the applicability of a rule to workloads in that namespace. If the VirtualService has a list of gateways specified in the top-level `gateways` field, it must include the reserved gateway `mesh` for this field to be applicable. |
  





<a name="istio.networking.v1alpha3.HTTPMatchRequest.HeadersEntry"></a>

### HTTPMatchRequest.HeadersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [istio.networking.v1alpha3.StringMatch]({{< ref "virtual_service.md#istio.networking.v1alpha3.StringMatch" >}}) |  |  |
  





<a name="istio.networking.v1alpha3.HTTPMatchRequest.QueryParamsEntry"></a>

### HTTPMatchRequest.QueryParamsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [istio.networking.v1alpha3.StringMatch]({{< ref "virtual_service.md#istio.networking.v1alpha3.StringMatch" >}}) |  |  |
  





<a name="istio.networking.v1alpha3.HTTPMatchRequest.SourceLabelsEntry"></a>

### HTTPMatchRequest.SourceLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="istio.networking.v1alpha3.HTTPMatchRequest.WithoutHeadersEntry"></a>

### HTTPMatchRequest.WithoutHeadersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | [istio.networking.v1alpha3.StringMatch]({{< ref "virtual_service.md#istio.networking.v1alpha3.StringMatch" >}}) |  |  |
  





<a name="istio.networking.v1alpha3.HTTPRedirect"></a>

### HTTPRedirect
HTTPRedirect can be used to send a 301 redirect response to the caller, where the Authority/Host and the URI in the response can be swapped with the specified values. For example, the following rule redirects requests for /v1/getProductRatings API on the ratings service to /v1/bookRatings provided by the bookratings service.<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: VirtualService metadata:   name: ratings-route spec:   hosts:   - ratings.prod.svc.cluster.local   http:   - match:     - uri:         exact: /v1/getProductRatings     redirect:       uri: /v1/bookRatings       authority: newratings.default.svc.cluster.local   ... ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: VirtualService metadata:   name: ratings-route spec:   hosts:   - ratings.prod.svc.cluster.local   http:   - match:     - uri:         exact: /v1/getProductRatings     redirect:       uri: /v1/bookRatings       authority: newratings.default.svc.cluster.local   ... ``` {{</tab>}} {{</tabs>}}


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uri | string |  | On a redirect, overwrite the Path portion of the URL with this value. Note that the entire path will be replaced, irrespective of the request URI being matched as an exact path or prefix. |
  | authority | string |  | On a redirect, overwrite the Authority/Host portion of the URL with this value. |
  | redirectCode | uint32 |  | On a redirect, Specifies the HTTP status code to use in the redirect response. The default response code is MOVED_PERMANENTLY (301). |
  





<a name="istio.networking.v1alpha3.HTTPRetry"></a>

### HTTPRetry
Describes the retry policy to use when a HTTP request fails. For example, the following rule sets the maximum number of retries to 3 when calling ratings:v1 service, with a 2s timeout per retry attempt.<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: VirtualService metadata:   name: ratings-route spec:   hosts:   - ratings.prod.svc.cluster.local   http:   - route:     - destination:         host: ratings.prod.svc.cluster.local         subset: v1     retries:       attempts: 3       perTryTimeout: 2s       retryOn: gateway-error,connect-failure,refused-stream ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: VirtualService metadata:   name: ratings-route spec:   hosts:   - ratings.prod.svc.cluster.local   http:   - route:     - destination:         host: ratings.prod.svc.cluster.local         subset: v1     retries:       attempts: 3       perTryTimeout: 2s       retryOn: gateway-error,connect-failure,refused-stream ``` {{</tab>}} {{</tabs>}}


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| attempts | int32 |  | Number of retries for a given request. The interval between retries will be determined automatically (25ms+). Actual number of retries attempted depends on the request `timeout` of the [HTTP route](https://istio.io/docs/reference/config/networking/virtual-service/#HTTPRoute). |
  | perTryTimeout | [google.protobuf.Duration]({{< ref "duration.md#google.protobuf.Duration" >}}) |  | Timeout per retry attempt for a given request. format: 1h/1m/1s/1ms. MUST BE >=1ms. |
  | retryOn | string |  | Specifies the conditions under which retry takes place. One or more policies can be specified using a ‘,’ delimited list. See the [retry policies](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/router_filter#x-envoy-retry-on) and [gRPC retry policies](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/router_filter#x-envoy-retry-grpc-on) for more details. |
  | retryRemoteLocalities | [google.protobuf.BoolValue]({{< ref "wrappers.md#google.protobuf.BoolValue" >}}) |  | Flag to specify whether the retries should retry to other localities. See the [retry plugin configuration](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/http/http_connection_management#retry-plugin-configuration) for more details. |
  





<a name="istio.networking.v1alpha3.HTTPRewrite"></a>

### HTTPRewrite
HTTPRewrite can be used to rewrite specific parts of a HTTP request before forwarding the request to the destination. Rewrite primitive can be used only with HTTPRouteDestination. The following example demonstrates how to rewrite the URL prefix for api call (/ratings) to ratings service before making the actual API call.<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: VirtualService metadata:   name: ratings-route spec:   hosts:   - ratings.prod.svc.cluster.local   http:   - match:     - uri:         prefix: /ratings     rewrite:       uri: /v1/bookRatings     route:     - destination:         host: ratings.prod.svc.cluster.local         subset: v1 ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: VirtualService metadata:   name: ratings-route spec:   hosts:   - ratings.prod.svc.cluster.local   http:   - match:     - uri:         prefix: /ratings     rewrite:       uri: /v1/bookRatings     route:     - destination:         host: ratings.prod.svc.cluster.local         subset: v1 ``` {{</tab>}} {{</tabs>}}


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uri | string |  | rewrite the path (or the prefix) portion of the URI with this value. If the original URI was matched based on prefix, the value provided in this field will replace the corresponding matched prefix. |
  | authority | string |  | rewrite the Authority/Host header with this value. |
  





<a name="istio.networking.v1alpha3.HTTPRoute"></a>

### HTTPRoute
Describes match conditions and actions for routing HTTP/1.1, HTTP2, and gRPC traffic. See VirtualService for usage examples.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The name assigned to the route for debugging purposes. The route's name will be concatenated with the match's name and will be logged in the access logs for requests matching this route/match. |
  | match | [][istio.networking.v1alpha3.HTTPMatchRequest]({{< ref "virtual_service.md#istio.networking.v1alpha3.HTTPMatchRequest" >}}) | repeated | Match conditions to be satisfied for the rule to be activated. All conditions inside a single match block have AND semantics, while the list of match blocks have OR semantics. The rule is matched if any one of the match blocks succeed. |
  | route | [][istio.networking.v1alpha3.HTTPRouteDestination]({{< ref "virtual_service.md#istio.networking.v1alpha3.HTTPRouteDestination" >}}) | repeated | A HTTP rule can either redirect or forward (default) traffic. The forwarding target can be one of several versions of a service (see glossary in beginning of document). Weights associated with the service version determine the proportion of traffic it receives. |
  | redirect | [istio.networking.v1alpha3.HTTPRedirect]({{< ref "virtual_service.md#istio.networking.v1alpha3.HTTPRedirect" >}}) |  | A HTTP rule can either redirect or forward (default) traffic. If traffic passthrough option is specified in the rule, route/redirect will be ignored. The redirect primitive can be used to send a HTTP 301 redirect to a different URI or Authority. |
  | delegate | [istio.networking.v1alpha3.Delegate]({{< ref "virtual_service.md#istio.networking.v1alpha3.Delegate" >}}) |  | Delegate is used to specify the particular VirtualService which can be used to define delegate HTTPRoute. It can be set only when `Route` and `Redirect` are empty, and the route rules of the delegate VirtualService will be merged with that in the current one. **NOTE**:    1. Only one level delegation is supported.    2. The delegate's HTTPMatchRequest must be a strict subset of the root's,       otherwise there is a conflict and the HTTPRoute will not take effect. |
  | rewrite | [istio.networking.v1alpha3.HTTPRewrite]({{< ref "virtual_service.md#istio.networking.v1alpha3.HTTPRewrite" >}}) |  | Rewrite HTTP URIs and Authority headers. Rewrite cannot be used with Redirect primitive. Rewrite will be performed before forwarding. |
  | timeout | [google.protobuf.Duration]({{< ref "duration.md#google.protobuf.Duration" >}}) |  | Timeout for HTTP requests. |
  | retries | [istio.networking.v1alpha3.HTTPRetry]({{< ref "virtual_service.md#istio.networking.v1alpha3.HTTPRetry" >}}) |  | Retry policy for HTTP requests. |
  | fault | [istio.networking.v1alpha3.HTTPFaultInjection]({{< ref "virtual_service.md#istio.networking.v1alpha3.HTTPFaultInjection" >}}) |  | Fault injection policy to apply on HTTP traffic at the client side. Note that timeouts or retries will not be enabled when faults are enabled on the client side. |
  | mirror | [istio.networking.v1alpha3.Destination]({{< ref "virtual_service.md#istio.networking.v1alpha3.Destination" >}}) |  | Mirror HTTP traffic to a another destination in addition to forwarding the requests to the intended destination. Mirrored traffic is on a best effort basis where the sidecar/gateway will not wait for the mirrored cluster to respond before returning the response from the original destination.  Statistics will be generated for the mirrored destination. |
  | mirrorPercent | [google.protobuf.UInt32Value]({{< ref "wrappers.md#google.protobuf.UInt32Value" >}}) |  | Percentage of the traffic to be mirrored by the `mirror` field. Use of integer `mirror_percent` value is deprecated. Use the double `mirror_percentage` field instead |
  | mirrorPercentage | [istio.networking.v1alpha3.Percent]({{< ref "virtual_service.md#istio.networking.v1alpha3.Percent" >}}) |  | Percentage of the traffic to be mirrored by the `mirror` field. If this field is absent, all the traffic (100%) will be mirrored. Max value is 100. |
  | corsPolicy | [istio.networking.v1alpha3.CorsPolicy]({{< ref "virtual_service.md#istio.networking.v1alpha3.CorsPolicy" >}}) |  | Cross-Origin Resource Sharing policy (CORS). Refer to [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) for further details about cross origin resource sharing. |
  | headers | [istio.networking.v1alpha3.Headers]({{< ref "virtual_service.md#istio.networking.v1alpha3.Headers" >}}) |  | Header manipulation rules |
  





<a name="istio.networking.v1alpha3.HTTPRouteDestination"></a>

### HTTPRouteDestination
Each routing rule is associated with one or more service versions (see glossary in beginning of document). Weights associated with the version determine the proportion of traffic it receives. For example, the following rule will route 25% of traffic for the "reviews" service to instances with the "v2" tag and the remaining traffic (i.e., 75%) to "v1".<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: VirtualService metadata:   name: reviews-route spec:   hosts:   - reviews.prod.svc.cluster.local   http:   - route:     - destination:         host: reviews.prod.svc.cluster.local         subset: v2       weight: 25     - destination:         host: reviews.prod.svc.cluster.local         subset: v1       weight: 75 ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: VirtualService metadata:   name: reviews-route spec:   hosts:   - reviews.prod.svc.cluster.local   http:   - route:     - destination:         host: reviews.prod.svc.cluster.local         subset: v2       weight: 25     - destination:         host: reviews.prod.svc.cluster.local         subset: v1       weight: 75 ``` {{</tab>}} {{</tabs>}}<br>And the associated DestinationRule<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: DestinationRule metadata:   name: reviews-destination spec:   host: reviews.prod.svc.cluster.local   subsets:   - name: v1     labels:       version: v1   - name: v2     labels:       version: v2 ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: DestinationRule metadata:   name: reviews-destination spec:   host: reviews.prod.svc.cluster.local   subsets:   - name: v1     labels:       version: v1   - name: v2     labels:       version: v2 ``` {{</tab>}} {{</tabs>}}<br>Traffic can also be split across two entirely different services without having to define new subsets. For example, the following rule forwards 25% of traffic to reviews.com to dev.reviews.com<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: VirtualService metadata:   name: reviews-route-two-domains spec:   hosts:   - reviews.com   http:   - route:     - destination:         host: dev.reviews.com       weight: 25     - destination:         host: reviews.com       weight: 75 ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: VirtualService metadata:   name: reviews-route-two-domains spec:   hosts:   - reviews.com   http:   - route:     - destination:         host: dev.reviews.com       weight: 25     - destination:         host: reviews.com       weight: 75 ``` {{</tab>}} {{</tabs>}}


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| destination | [istio.networking.v1alpha3.Destination]({{< ref "virtual_service.md#istio.networking.v1alpha3.Destination" >}}) |  | Destination uniquely identifies the instances of a service to which the request/connection should be forwarded to. |
  | weight | int32 |  | The proportion of traffic to be forwarded to the service version. (0-100). Sum of weights across destinations SHOULD BE == 100. If there is only one destination in a rule, the weight value is assumed to be 100. |
  | headers | [istio.networking.v1alpha3.Headers]({{< ref "virtual_service.md#istio.networking.v1alpha3.Headers" >}}) |  | Header manipulation rules |
  





<a name="istio.networking.v1alpha3.Headers"></a>

### Headers
Message headers can be manipulated when Envoy forwards requests to, or responses from, a destination service. Header manipulation rules can be specified for a specific route destination or for all destinations. The following VirtualService adds a `test` header with the value `true` to requests that are routed to any `reviews` service destination. It also romoves the `foo` response header, but only from responses coming from the `v1` subset (version) of the `reviews` service.<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: VirtualService metadata:   name: reviews-route spec:   hosts:   - reviews.prod.svc.cluster.local   http:   - headers:       request:         set:           test: true     route:     - destination:         host: reviews.prod.svc.cluster.local         subset: v2       weight: 25     - destination:         host: reviews.prod.svc.cluster.local         subset: v1       headers:         response:           remove:           - foo       weight: 75 ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: VirtualService metadata:   name: reviews-route spec:   hosts:   - reviews.prod.svc.cluster.local   http:   - headers:       request:         set:           test: true     route:     - destination:         host: reviews.prod.svc.cluster.local         subset: v2       weight: 25     - destination:         host: reviews.prod.svc.cluster.local         subset: v1       headers:         response:           remove:           - foo       weight: 75 ``` {{</tab>}} {{</tabs>}}


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request | [istio.networking.v1alpha3.Headers.HeaderOperations]({{< ref "virtual_service.md#istio.networking.v1alpha3.Headers.HeaderOperations" >}}) |  | Header manipulation rules to apply before forwarding a request to the destination service |
  | response | [istio.networking.v1alpha3.Headers.HeaderOperations]({{< ref "virtual_service.md#istio.networking.v1alpha3.Headers.HeaderOperations" >}}) |  | Header manipulation rules to apply before returning a response to the caller |
  





<a name="istio.networking.v1alpha3.Headers.HeaderOperations"></a>

### Headers.HeaderOperations
HeaderOperations Describes the header manipulations to apply


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| set | [][istio.networking.v1alpha3.Headers.HeaderOperations.SetEntry]({{< ref "virtual_service.md#istio.networking.v1alpha3.Headers.HeaderOperations.SetEntry" >}}) | repeated | Overwrite the headers specified by key with the given values |
  | add | [][istio.networking.v1alpha3.Headers.HeaderOperations.AddEntry]({{< ref "virtual_service.md#istio.networking.v1alpha3.Headers.HeaderOperations.AddEntry" >}}) | repeated | Append the given values to the headers specified by keys (will create a comma-separated list of values) |
  | remove | []string | repeated | Remove a the specified headers |
  





<a name="istio.networking.v1alpha3.Headers.HeaderOperations.AddEntry"></a>

### Headers.HeaderOperations.AddEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="istio.networking.v1alpha3.Headers.HeaderOperations.SetEntry"></a>

### Headers.HeaderOperations.SetEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="istio.networking.v1alpha3.L4MatchAttributes"></a>

### L4MatchAttributes
L4 connection match attributes. Note that L4 connection matching support is incomplete.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| destinationSubnets | []string | repeated | IPv4 or IPv6 ip addresses of destination with optional subnet.  E.g., a.b.c.d/xx form or just a.b.c.d. |
  | port | uint32 |  | Specifies the port on the host that is being addressed. Many services only expose a single port or label ports with the protocols they support, in these cases it is not required to explicitly select the port. |
  | sourceSubnet | string |  | IPv4 or IPv6 ip address of source with optional subnet. E.g., a.b.c.d/xx form or just a.b.c.d $hide_from_docs |
  | sourceLabels | [][istio.networking.v1alpha3.L4MatchAttributes.SourceLabelsEntry]({{< ref "virtual_service.md#istio.networking.v1alpha3.L4MatchAttributes.SourceLabelsEntry" >}}) | repeated | One or more labels that constrain the applicability of a rule to workloads with the given labels. If the VirtualService has a list of gateways specified in the top-level `gateways` field, it should include the reserved gateway `mesh` in order for this field to be applicable. |
  | gateways | []string | repeated | Names of gateways where the rule should be applied. Gateway names in the top-level `gateways` field of the VirtualService (if any) are overridden. The gateway match is independent of sourceLabels. |
  | sourceNamespace | string |  | Source namespace constraining the applicability of a rule to workloads in that namespace. If the VirtualService has a list of gateways specified in the top-level `gateways` field, it must include the reserved gateway `mesh` for this field to be applicable. |
  





<a name="istio.networking.v1alpha3.L4MatchAttributes.SourceLabelsEntry"></a>

### L4MatchAttributes.SourceLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="istio.networking.v1alpha3.Percent"></a>

### Percent
Percent specifies a percentage in the range of [0.0, 100.0].


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | double |  |  |
  





<a name="istio.networking.v1alpha3.PortSelector"></a>

### PortSelector
PortSelector specifies the number of a port to be used for matching or selection for final routing.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| number | uint32 |  | Valid port number |
  





<a name="istio.networking.v1alpha3.RouteDestination"></a>

### RouteDestination
L4 routing rule weighted destination.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| destination | [istio.networking.v1alpha3.Destination]({{< ref "virtual_service.md#istio.networking.v1alpha3.Destination" >}}) |  | Destination uniquely identifies the instances of a service to which the request/connection should be forwarded to. |
  | weight | int32 |  | The proportion of traffic to be forwarded to the service version. If there is only one destination in a rule, all traffic will be routed to it irrespective of the weight. |
  





<a name="istio.networking.v1alpha3.StringMatch"></a>

### StringMatch
Describes how to match a given string in HTTP headers. Match is case-sensitive.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| exact | string |  | exact string match |
  | prefix | string |  | prefix-based match |
  | regex | string |  | RE2 style regex-based match (https://github.com/google/re2/wiki/Syntax). |
  





<a name="istio.networking.v1alpha3.TCPRoute"></a>

### TCPRoute
Describes match conditions and actions for routing TCP traffic. The following routing rule forwards traffic arriving at port 27017 for mongo.prod.svc.cluster.local to another Mongo server on port 5555.<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: VirtualService metadata:   name: bookinfo-Mongo spec:   hosts:   - mongo.prod.svc.cluster.local   tcp:   - match:     - port: 27017     route:     - destination:         host: mongo.backup.svc.cluster.local         port:           number: 5555 ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: VirtualService metadata:   name: bookinfo-Mongo spec:   hosts:   - mongo.prod.svc.cluster.local   tcp:   - match:     - port: 27017     route:     - destination:         host: mongo.backup.svc.cluster.local         port:           number: 5555 ``` {{</tab>}} {{</tabs>}}


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| match | [][istio.networking.v1alpha3.L4MatchAttributes]({{< ref "virtual_service.md#istio.networking.v1alpha3.L4MatchAttributes" >}}) | repeated | Match conditions to be satisfied for the rule to be activated. All conditions inside a single match block have AND semantics, while the list of match blocks have OR semantics. The rule is matched if any one of the match blocks succeed. |
  | route | [][istio.networking.v1alpha3.RouteDestination]({{< ref "virtual_service.md#istio.networking.v1alpha3.RouteDestination" >}}) | repeated | The destination to which the connection should be forwarded to. |
  





<a name="istio.networking.v1alpha3.TLSMatchAttributes"></a>

### TLSMatchAttributes
TLS connection match attributes.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sniHosts | []string | repeated | SNI (server name indicator) to match on. Wildcard prefixes can be used in the SNI value, e.g., *.com will match foo.example.com as well as example.com. An SNI value must be a subset (i.e., fall within the domain) of the corresponding virtual serivce's hosts. |
  | destinationSubnets | []string | repeated | IPv4 or IPv6 ip addresses of destination with optional subnet.  E.g., a.b.c.d/xx form or just a.b.c.d. |
  | port | uint32 |  | Specifies the port on the host that is being addressed. Many services only expose a single port or label ports with the protocols they support, in these cases it is not required to explicitly select the port. |
  | sourceLabels | [][istio.networking.v1alpha3.TLSMatchAttributes.SourceLabelsEntry]({{< ref "virtual_service.md#istio.networking.v1alpha3.TLSMatchAttributes.SourceLabelsEntry" >}}) | repeated | One or more labels that constrain the applicability of a rule to workloads with the given labels. If the VirtualService has a list of gateways specified in the top-level `gateways` field, it should include the reserved gateway `mesh` in order for this field to be applicable. |
  | gateways | []string | repeated | Names of gateways where the rule should be applied. Gateway names in the top-level `gateways` field of the VirtualService (if any) are overridden. The gateway match is independent of sourceLabels. |
  | sourceNamespace | string |  | Source namespace constraining the applicability of a rule to workloads in that namespace. If the VirtualService has a list of gateways specified in the top-level `gateways` field, it must include the reserved gateway `mesh` for this field to be applicable. |
  





<a name="istio.networking.v1alpha3.TLSMatchAttributes.SourceLabelsEntry"></a>

### TLSMatchAttributes.SourceLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="istio.networking.v1alpha3.TLSRoute"></a>

### TLSRoute
Describes match conditions and actions for routing unterminated TLS traffic (TLS/HTTPS) The following routing rule forwards unterminated TLS traffic arriving at port 443 of gateway called "mygateway" to internal services in the mesh based on the SNI value.<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: VirtualService metadata:   name: bookinfo-sni spec:   hosts:   - "*.bookinfo.com"   gateways:   - mygateway   tls:   - match:     - port: 443       sniHosts:       - login.bookinfo.com     route:     - destination:         host: login.prod.svc.cluster.local   - match:     - port: 443       sniHosts:       - reviews.bookinfo.com     route:     - destination:         host: reviews.prod.svc.cluster.local ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: VirtualService metadata:   name: bookinfo-sni spec:   hosts:   - "*.bookinfo.com"   gateways:   - mygateway   tls:   - match:     - port: 443       sniHosts:       - login.bookinfo.com     route:     - destination:         host: login.prod.svc.cluster.local   - match:     - port: 443       sniHosts:       - reviews.bookinfo.com     route:     - destination:         host: reviews.prod.svc.cluster.local ``` {{</tab>}} {{</tabs>}}


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| match | [][istio.networking.v1alpha3.TLSMatchAttributes]({{< ref "virtual_service.md#istio.networking.v1alpha3.TLSMatchAttributes" >}}) | repeated | Match conditions to be satisfied for the rule to be activated. All conditions inside a single match block have AND semantics, while the list of match blocks have OR semantics. The rule is matched if any one of the match blocks succeed. |
  | route | [][istio.networking.v1alpha3.RouteDestination]({{< ref "virtual_service.md#istio.networking.v1alpha3.RouteDestination" >}}) | repeated | The destination to which the connection should be forwarded to. |
  





<a name="istio.networking.v1alpha3.VirtualService"></a>

### VirtualService
Configuration affecting traffic routing.<br><!-- crd generation tags +cue-gen:VirtualService:groupName:networking.istio.io +cue-gen:VirtualService:version:v1alpha3 +cue-gen:VirtualService:storageVersion +cue-gen:VirtualService:annotations:helm.sh/resource-policy=keep +cue-gen:VirtualService:labels:app=istio-pilot,chart=istio,heritage=Tiller,release=istio +cue-gen:VirtualService:subresource:status +cue-gen:VirtualService:scope:Namespaced +cue-gen:VirtualService:resource:categories=istio-io,networking-istio-io,shortNames=vs +cue-gen:VirtualService:printerColumn:name=Gateways,type=string,JSONPath=.spec.gateways,description="The names of gateways and sidecars  that should apply these routes" +cue-gen:VirtualService:printerColumn:name=Hosts,type=string,JSONPath=.spec.hosts,description="The destination hosts to which traffic is being sent" +cue-gen:VirtualService:printerColumn:name=Age,type=date,JSONPath=.metadata.creationTimestamp,description="CreationTimestamp is a timestamp  representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations.  Clients may not set this value. It is represented in RFC3339 form and is in UTC. Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata" --><br><!-- go code generation tags +kubetype-gen +kubetype-gen:groupVersion=networking.istio.io/v1alpha3 +genclient +k8s:deepcopy-gen=true -->


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hosts | []string | repeated | The destination hosts to which traffic is being sent. Could be a DNS name with wildcard prefix or an IP address.  Depending on the platform, short-names can also be used instead of a FQDN (i.e. has no dots in the name). In such a scenario, the FQDN of the host would be derived based on the underlying platform.<br>A single VirtualService can be used to describe all the traffic properties of the corresponding hosts, including those for multiple HTTP and TCP ports. Alternatively, the traffic properties of a host can be defined using more than one VirtualService, with certain caveats. Refer to the [Operations Guide](https://istio.io/docs/ops/best-practices/traffic-management/#split-virtual-services) for details.<br>*Note for Kubernetes users*: When short names are used (e.g. "reviews" instead of "reviews.default.svc.cluster.local"), Istio will interpret the short name based on the namespace of the rule, not the service. A rule in the "default" namespace containing a host "reviews" will be interpreted as "reviews.default.svc.cluster.local", irrespective of the actual namespace associated with the reviews service. _To avoid potential misconfigurations, it is recommended to always use fully qualified domain names over short names._<br>The hosts field applies to both HTTP and TCP services. Service inside the mesh, i.e., those found in the service registry, must always be referred to using their alphanumeric names. IP addresses are allowed only for services defined via the Gateway.<br>*Note*: It must be empty for a delegate VirtualService. |
  | gateways | []string | repeated | The names of gateways and sidecars that should apply these routes. Gateways in other namespaces may be referred to by `<gateway namespace>/<gateway name>`; specifying a gateway with no namespace qualifier is the same as specifying the VirtualService's namespace. A single VirtualService is used for sidecars inside the mesh as well as for one or more gateways. The selection condition imposed by this field can be overridden using the source field in the match conditions of protocol-specific routes. The reserved word `mesh` is used to imply all the sidecars in the mesh. When this field is omitted, the default gateway (`mesh`) will be used, which would apply the rule to all sidecars in the mesh. If a list of gateway names is provided, the rules will apply only to the gateways. To apply the rules to both gateways and sidecars, specify `mesh` as one of the gateway names. |
  | http | [][istio.networking.v1alpha3.HTTPRoute]({{< ref "virtual_service.md#istio.networking.v1alpha3.HTTPRoute" >}}) | repeated | An ordered list of route rules for HTTP traffic. HTTP routes will be applied to platform service ports named 'http-*'/'http2-*'/'grpc-*', gateway ports with protocol HTTP/HTTP2/GRPC/ TLS-terminated-HTTPS and service entry ports using HTTP/HTTP2/GRPC protocols.  The first rule matching an incoming request is used. |
  | tls | [][istio.networking.v1alpha3.TLSRoute]({{< ref "virtual_service.md#istio.networking.v1alpha3.TLSRoute" >}}) | repeated | An ordered list of route rule for non-terminated TLS & HTTPS traffic. Routing is typically performed using the SNI value presented by the ClientHello message. TLS routes will be applied to platform service ports named 'https-*', 'tls-*', unterminated gateway ports using HTTPS/TLS protocols (i.e. with "passthrough" TLS mode) and service entry ports using HTTPS/TLS protocols.  The first rule matching an incoming request is used.  NOTE: Traffic 'https-*' or 'tls-*' ports without associated virtual service will be treated as opaque TCP traffic. |
  | tcp | [][istio.networking.v1alpha3.TCPRoute]({{< ref "virtual_service.md#istio.networking.v1alpha3.TCPRoute" >}}) | repeated | An ordered list of route rules for opaque TCP traffic. TCP routes will be applied to any port that is not a HTTP or TLS port. The first rule matching an incoming request is used. |
  | exportTo | []string | repeated | A list of namespaces to which this virtual service is exported. Exporting a virtual service allows it to be used by sidecars and gateways defined in other namespaces. This feature provides a mechanism for service owners and mesh administrators to control the visibility of virtual services across namespace boundaries.<br>If no namespaces are specified then the virtual service is exported to all namespaces by default.<br>The value "." is reserved and defines an export to the same namespace that the virtual service is declared in. Similarly the value "*" is reserved and defines an export to all namespaces.<br>NOTE: in the current release, the `exportTo` value is restricted to "." or "*" (i.e., the current namespace or all namespaces). |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

