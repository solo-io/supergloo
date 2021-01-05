
---

---

## Package : `istio.networking.v1alpha3`



<a name="top"></a>

<a name="API Reference for envoy_filter.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## envoy_filter.proto


## Table of Contents
  - [EnvoyFilter](#istio.networking.v1alpha3.EnvoyFilter)
  - [EnvoyFilter.ClusterMatch](#istio.networking.v1alpha3.EnvoyFilter.ClusterMatch)
  - [EnvoyFilter.EnvoyConfigObjectMatch](#istio.networking.v1alpha3.EnvoyFilter.EnvoyConfigObjectMatch)
  - [EnvoyFilter.EnvoyConfigObjectPatch](#istio.networking.v1alpha3.EnvoyFilter.EnvoyConfigObjectPatch)
  - [EnvoyFilter.ListenerMatch](#istio.networking.v1alpha3.EnvoyFilter.ListenerMatch)
  - [EnvoyFilter.ListenerMatch.FilterChainMatch](#istio.networking.v1alpha3.EnvoyFilter.ListenerMatch.FilterChainMatch)
  - [EnvoyFilter.ListenerMatch.FilterMatch](#istio.networking.v1alpha3.EnvoyFilter.ListenerMatch.FilterMatch)
  - [EnvoyFilter.ListenerMatch.SubFilterMatch](#istio.networking.v1alpha3.EnvoyFilter.ListenerMatch.SubFilterMatch)
  - [EnvoyFilter.Patch](#istio.networking.v1alpha3.EnvoyFilter.Patch)
  - [EnvoyFilter.ProxyMatch](#istio.networking.v1alpha3.EnvoyFilter.ProxyMatch)
  - [EnvoyFilter.ProxyMatch.MetadataEntry](#istio.networking.v1alpha3.EnvoyFilter.ProxyMatch.MetadataEntry)
  - [EnvoyFilter.RouteConfigurationMatch](#istio.networking.v1alpha3.EnvoyFilter.RouteConfigurationMatch)
  - [EnvoyFilter.RouteConfigurationMatch.RouteMatch](#istio.networking.v1alpha3.EnvoyFilter.RouteConfigurationMatch.RouteMatch)
  - [EnvoyFilter.RouteConfigurationMatch.VirtualHostMatch](#istio.networking.v1alpha3.EnvoyFilter.RouteConfigurationMatch.VirtualHostMatch)

  - [EnvoyFilter.ApplyTo](#istio.networking.v1alpha3.EnvoyFilter.ApplyTo)
  - [EnvoyFilter.Patch.FilterClass](#istio.networking.v1alpha3.EnvoyFilter.Patch.FilterClass)
  - [EnvoyFilter.Patch.Operation](#istio.networking.v1alpha3.EnvoyFilter.Patch.Operation)
  - [EnvoyFilter.PatchContext](#istio.networking.v1alpha3.EnvoyFilter.PatchContext)
  - [EnvoyFilter.RouteConfigurationMatch.RouteMatch.Action](#istio.networking.v1alpha3.EnvoyFilter.RouteConfigurationMatch.RouteMatch.Action)






<a name="istio.networking.v1alpha3.EnvoyFilter"></a>

### EnvoyFilter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workloadSelector | [istio.networking.v1alpha3.WorkloadSelector]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.sidecar#istio.networking.v1alpha3.WorkloadSelector" >}}) |  | Criteria used to select the specific set of pods/VMs on which this patch configuration should be applied. If omitted, the set of patches in this configuration will be applied to all workload instances in the same namespace.  If omitted, the EnvoyFilter patches will be applied to all workloads in the same namespace. If the EnvoyFilter is present in the config root namespace, it will be applied to all applicable workloads in any namespace. |
  | configPatches | [][istio.networking.v1alpha3.EnvoyFilter.EnvoyConfigObjectPatch]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.envoy_filter#istio.networking.v1alpha3.EnvoyFilter.EnvoyConfigObjectPatch" >}}) | repeated | One or more patches with match conditions. |
  





<a name="istio.networking.v1alpha3.EnvoyFilter.ClusterMatch"></a>

### EnvoyFilter.ClusterMatch



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| portNumber | uint32 |  | The service port for which this cluster was generated.  If omitted, applies to clusters for any port. |
  | service | string |  | The fully qualified service name for this cluster. If omitted, applies to clusters for any service. For services defined through service entries, the service name is same as the hosts defined in the service entry. |
  | subset | string |  | The subset associated with the service. If omitted, applies to clusters for any subset of a service. |
  | name | string |  | The exact name of the cluster to match. To match a specific cluster by name, such as the internally generated "Passthrough" cluster, leave all fields in clusterMatch empty, except the name. |
  





<a name="istio.networking.v1alpha3.EnvoyFilter.EnvoyConfigObjectMatch"></a>

### EnvoyFilter.EnvoyConfigObjectMatch



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| context | [istio.networking.v1alpha3.EnvoyFilter.PatchContext]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.envoy_filter#istio.networking.v1alpha3.EnvoyFilter.PatchContext" >}}) |  | The specific config generation context to match on. Istio Pilot generates envoy configuration in the context of a gateway, inbound traffic to sidecar and outbound traffic from sidecar. |
  | proxy | [istio.networking.v1alpha3.EnvoyFilter.ProxyMatch]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.envoy_filter#istio.networking.v1alpha3.EnvoyFilter.ProxyMatch" >}}) |  | Match on properties associated with a proxy. |
  | listener | [istio.networking.v1alpha3.EnvoyFilter.ListenerMatch]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.envoy_filter#istio.networking.v1alpha3.EnvoyFilter.ListenerMatch" >}}) |  | Match on envoy listener attributes. |
  | routeConfiguration | [istio.networking.v1alpha3.EnvoyFilter.RouteConfigurationMatch]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.envoy_filter#istio.networking.v1alpha3.EnvoyFilter.RouteConfigurationMatch" >}}) |  | Match on envoy HTTP route configuration attributes. |
  | cluster | [istio.networking.v1alpha3.EnvoyFilter.ClusterMatch]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.envoy_filter#istio.networking.v1alpha3.EnvoyFilter.ClusterMatch" >}}) |  | Match on envoy cluster attributes. |
  





<a name="istio.networking.v1alpha3.EnvoyFilter.EnvoyConfigObjectPatch"></a>

### EnvoyFilter.EnvoyConfigObjectPatch



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| applyTo | [istio.networking.v1alpha3.EnvoyFilter.ApplyTo]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.envoy_filter#istio.networking.v1alpha3.EnvoyFilter.ApplyTo" >}}) |  | Specifies where in the Envoy configuration, the patch should be applied.  The match is expected to select the appropriate object based on applyTo.  For example, an applyTo with HTTP_FILTER is expected to have a match condition on the listeners, with a network filter selection on envoy.filters.network.http_connection_manager and a sub filter selection on the HTTP filter relative to which the insertion should be performed. Similarly, an applyTo on CLUSTER should have a match (if provided) on the cluster and not on a listener. |
  | match | [istio.networking.v1alpha3.EnvoyFilter.EnvoyConfigObjectMatch]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.envoy_filter#istio.networking.v1alpha3.EnvoyFilter.EnvoyConfigObjectMatch" >}}) |  | Match on listener/route configuration/cluster. |
  | patch | [istio.networking.v1alpha3.EnvoyFilter.Patch]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.envoy_filter#istio.networking.v1alpha3.EnvoyFilter.Patch" >}}) |  | The patch to apply along with the operation. |
  





<a name="istio.networking.v1alpha3.EnvoyFilter.ListenerMatch"></a>

### EnvoyFilter.ListenerMatch



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| portNumber | uint32 |  | The service port/gateway port to which traffic is being sent/received. If not specified, matches all listeners. Even though inbound listeners are generated for the instance/pod ports, only service ports should be used to match listeners. |
  | portName | string |  | Instead of using specific port numbers, a set of ports matching a given service's port name can be selected. Matching is case insensitive. Not implemented. $hide_from_docs |
  | filterChain | [istio.networking.v1alpha3.EnvoyFilter.ListenerMatch.FilterChainMatch]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.envoy_filter#istio.networking.v1alpha3.EnvoyFilter.ListenerMatch.FilterChainMatch" >}}) |  | Match a specific filter chain in a listener. If specified, the patch will be applied to the filter chain (and a specific filter if specified) and not to other filter chains in the listener. |
  | name | string |  | Match a specific listener by its name. The listeners generated by Pilot are typically named as IP:Port. |
  





<a name="istio.networking.v1alpha3.EnvoyFilter.ListenerMatch.FilterChainMatch"></a>

### EnvoyFilter.ListenerMatch.FilterChainMatch



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The name assigned to the filter chain. |
  | sni | string |  | The SNI value used by a filter chain's match condition.  This condition will evaluate to false if the filter chain has no sni match. |
  | transportProtocol | string |  | Applies only to SIDECAR_INBOUND context. If non-empty, a transport protocol to consider when determining a filter chain match.  This value will be compared against the transport protocol of a new connection, when it's detected by the tls_inspector listener filter.<br>Accepted values include:<br>* `raw_buffer` - default, used when no transport protocol is detected. * `tls` - set when TLS protocol is detected by the TLS inspector. |
  | applicationProtocols | string |  | Applies only to sidecars. If non-empty, a comma separated set of application protocols to consider when determining a filter chain match.  This value will be compared against the application protocols of a new connection, when it's detected by one of the listener filters such as the http_inspector.<br>Accepted values include: h2,http/1.1,http/1.0 |
  | filter | [istio.networking.v1alpha3.EnvoyFilter.ListenerMatch.FilterMatch]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.envoy_filter#istio.networking.v1alpha3.EnvoyFilter.ListenerMatch.FilterMatch" >}}) |  | The name of a specific filter to apply the patch to. Set this to envoy.filters.network.http_connection_manager to add a filter or apply a patch to the HTTP connection manager. |
  





<a name="istio.networking.v1alpha3.EnvoyFilter.ListenerMatch.FilterMatch"></a>

### EnvoyFilter.ListenerMatch.FilterMatch



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The filter name to match on. For standard Envoy filters, canonical filter names should be used. Refer to https://www.envoyproxy.io/docs/envoy/latest/version_history/v1.14.0#deprecated for canonical names. |
  | subFilter | [istio.networking.v1alpha3.EnvoyFilter.ListenerMatch.SubFilterMatch]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.envoy_filter#istio.networking.v1alpha3.EnvoyFilter.ListenerMatch.SubFilterMatch" >}}) |  | The next level filter within this filter to match upon. Typically used for HTTP Connection Manager filters and Thrift filters. |
  





<a name="istio.networking.v1alpha3.EnvoyFilter.ListenerMatch.SubFilterMatch"></a>

### EnvoyFilter.ListenerMatch.SubFilterMatch



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The filter name to match on. |
  





<a name="istio.networking.v1alpha3.EnvoyFilter.Patch"></a>

### EnvoyFilter.Patch



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| operation | [istio.networking.v1alpha3.EnvoyFilter.Patch.Operation]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.envoy_filter#istio.networking.v1alpha3.EnvoyFilter.Patch.Operation" >}}) |  | Determines how the patch should be applied. |
  | value | [google.protobuf.Struct]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.struct#google.protobuf.Struct" >}}) |  | The JSON config of the object being patched. This will be merged using proto merge semantics with the existing proto in the path. |
  | filterClass | [istio.networking.v1alpha3.EnvoyFilter.Patch.FilterClass]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.envoy_filter#istio.networking.v1alpha3.EnvoyFilter.Patch.FilterClass" >}}) |  | Determines the filter insertion order. |
  





<a name="istio.networking.v1alpha3.EnvoyFilter.ProxyMatch"></a>

### EnvoyFilter.ProxyMatch



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| proxyVersion | string |  | A regular expression in golang regex format (RE2) that can be used to select proxies using a specific version of istio proxy. The Istio version for a given proxy is obtained from the node metadata field ISTIO_VERSION supplied by the proxy when connecting to Pilot. This value is embedded as an environment variable (ISTIO_META_ISTIO_VERSION) in the Istio proxy docker image. Custom proxy implementations should provide this metadata variable to take advantage of the Istio version check option. |
  | metadata | [][istio.networking.v1alpha3.EnvoyFilter.ProxyMatch.MetadataEntry]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.envoy_filter#istio.networking.v1alpha3.EnvoyFilter.ProxyMatch.MetadataEntry" >}}) | repeated | Match on the node metadata supplied by a proxy when connecting to Istio Pilot. Note that while Envoy's node metadata is of type Struct, only string key-value pairs are processed by Pilot. All keys specified in the metadata must match with exact values. The match will fail if any of the specified keys are absent or the values fail to match. |
  





<a name="istio.networking.v1alpha3.EnvoyFilter.ProxyMatch.MetadataEntry"></a>

### EnvoyFilter.ProxyMatch.MetadataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="istio.networking.v1alpha3.EnvoyFilter.RouteConfigurationMatch"></a>

### EnvoyFilter.RouteConfigurationMatch



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| portNumber | uint32 |  | The service port number or gateway server port number for which this route configuration was generated. If omitted, applies to route configurations for all ports. |
  | portName | string |  | Applicable only for GATEWAY context. The gateway server port name for which this route configuration was generated. |
  | gateway | string |  | The Istio gateway config's namespace/name for which this route configuration was generated. Applies only if the context is GATEWAY. Should be in the namespace/name format. Use this field in conjunction with the portNumber and portName to accurately select the Envoy route configuration for a specific HTTPS server within a gateway config object. |
  | vhost | [istio.networking.v1alpha3.EnvoyFilter.RouteConfigurationMatch.VirtualHostMatch]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.envoy_filter#istio.networking.v1alpha3.EnvoyFilter.RouteConfigurationMatch.VirtualHostMatch" >}}) |  | Match a specific virtual host in a route configuration and apply the patch to the virtual host. |
  | name | string |  | Route configuration name to match on. Can be used to match a specific route configuration by name, such as the internally generated "http_proxy" route configuration for all sidecars. |
  





<a name="istio.networking.v1alpha3.EnvoyFilter.RouteConfigurationMatch.RouteMatch"></a>

### EnvoyFilter.RouteConfigurationMatch.RouteMatch



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The Route objects generated by default are named as "default".  Route objects generated using a virtual service will carry the name used in the virtual service's HTTP routes. |
  | action | [istio.networking.v1alpha3.EnvoyFilter.RouteConfigurationMatch.RouteMatch.Action]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.envoy_filter#istio.networking.v1alpha3.EnvoyFilter.RouteConfigurationMatch.RouteMatch.Action" >}}) |  | Match a route with specific action type. |
  





<a name="istio.networking.v1alpha3.EnvoyFilter.RouteConfigurationMatch.VirtualHostMatch"></a>

### EnvoyFilter.RouteConfigurationMatch.VirtualHostMatch



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The VirtualHosts objects generated by Istio are named as host:port, where the host typically corresponds to the VirtualService's host field or the hostname of a service in the registry. |
  | route | [istio.networking.v1alpha3.EnvoyFilter.RouteConfigurationMatch.RouteMatch]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.envoy_filter#istio.networking.v1alpha3.EnvoyFilter.RouteConfigurationMatch.RouteMatch" >}}) |  | Match a specific route within the virtual host. |
  




 <!-- end messages -->


<a name="istio.networking.v1alpha3.EnvoyFilter.ApplyTo"></a>

### EnvoyFilter.ApplyTo


| Name | Number | Description |
| ---- | ------ | ----------- |
| INVALID | 0 |  |
| LISTENER | 1 | Applies the patch to the listener. |
| FILTER_CHAIN | 2 | Applies the patch to the filter chain. |
| NETWORK_FILTER | 3 | Applies the patch to the network filter chain, to modify an existing filter or add a new filter. |
| HTTP_FILTER | 4 | Applies the patch to the HTTP filter chain in the http connection manager, to modify an existing filter or add a new filter. |
| ROUTE_CONFIGURATION | 5 | Applies the patch to the Route configuration (rds output) inside a HTTP connection manager. This does not apply to the virtual host. Currently, only MERGE operation is allowed on the route configuration objects. |
| VIRTUAL_HOST | 6 | Applies the patch to a virtual host inside a route configuration. |
| HTTP_ROUTE | 7 | Applies the patch to a route object inside the matched virtual host in a route configuration. Currently, only MERGE operation is allowed on the route objects. |
| CLUSTER | 8 | Applies the patch to a cluster in a CDS output. Also used to add new clusters. |



<a name="istio.networking.v1alpha3.EnvoyFilter.Patch.FilterClass"></a>

### EnvoyFilter.Patch.FilterClass


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNSPECIFIED | 0 | Control plane decides where to insert the filter. Do not specify FilterClass if the filter is independent of others. |
| AUTHN | 1 | Insert filter after Istio authentication filters. |
| AUTHZ | 2 | Insert filter after Istio authorization filters. |
| STATS | 3 | Insert filter before Istio stats filters. |



<a name="istio.networking.v1alpha3.EnvoyFilter.Patch.Operation"></a>

### EnvoyFilter.Patch.Operation


| Name | Number | Description |
| ---- | ------ | ----------- |
| INVALID | 0 |  |
| MERGE | 1 | Merge the provided config with the generated config using proto merge semantics. If you are specifying config in its entirity, use REPLACE instead. |
| ADD | 2 | Add the provided config to an existing list (of listeners, clusters, virtual hosts, network filters, or http filters). This operation will be ignored when applyTo is set to ROUTE_CONFIGURATION, or HTTP_ROUTE. |
| REMOVE | 3 | Remove the selected object from the list (of listeners, clusters, virtual hosts, network filters, or http filters). Does not require a value to be specified. This operation will be ignored when applyTo is set to ROUTE_CONFIGURATION, or HTTP_ROUTE. |
| INSERT_BEFORE | 4 | Insert operation on an array of named objects. This operation is typically useful only in the context of filters, where the order of filters matter. For clusters and virtual hosts, order of the element in the array does not matter. Insert before the selected filter or sub filter. If no filter is selected, the specified filter will be inserted at the front of the list. |
| INSERT_AFTER | 5 | Insert operation on an array of named objects. This operation is typically useful only in the context of filters, where the order of filters matter. For clusters and virtual hosts, order of the element in the array does not matter. Insert after the selected filter or sub filter. If no filter is selected, the specified filter will be inserted at the end of the list. |
| INSERT_FIRST | 6 | Insert operation on an array of named objects. This operation is typically useful only in the context of filters, where the order of filters matter. For clusters and virtual hosts, order of the element in the array does not matter. Insert first in the list based on the presence of selected filter or not. This is specifically useful when you want your filter first in the list based on a match condition specified in Match clause. |
| REPLACE | 7 | Replace contents of a named filter with new contents. REPLACE operation is only valid for HTTP_FILTER and NETWORK_FILTER. If the named filter is not found, this operation has no effect. |



<a name="istio.networking.v1alpha3.EnvoyFilter.PatchContext"></a>

### EnvoyFilter.PatchContext


| Name | Number | Description |
| ---- | ------ | ----------- |
| ANY | 0 | All listeners/routes/clusters in both sidecars and gateways. |
| SIDECAR_INBOUND | 1 | Inbound listener/route/cluster in sidecar. |
| SIDECAR_OUTBOUND | 2 | Outbound listener/route/cluster in sidecar. |
| GATEWAY | 3 | Gateway listener/route/cluster. |



<a name="istio.networking.v1alpha3.EnvoyFilter.RouteConfigurationMatch.RouteMatch.Action"></a>

### EnvoyFilter.RouteConfigurationMatch.RouteMatch.Action


| Name | Number | Description |
| ---- | ------ | ----------- |
| ANY | 0 | All three route actions |
| ROUTE | 1 | Route traffic to a cluster / weighted clusters. |
| REDIRECT | 2 | Redirect request. |
| DIRECT_RESPONSE | 3 | directly respond to a request with specific payload. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

