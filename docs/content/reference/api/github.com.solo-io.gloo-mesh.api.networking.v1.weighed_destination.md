
---

title: "weighed_destination.proto"

---

## Package : `networking.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for weighed_destination.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## weighed_destination.proto


## Table of Contents
  - [HeaderManipulation](#networking.mesh.gloo.solo.io.HeaderManipulation)
  - [HeaderManipulation.AppendRequestHeadersEntry](#networking.mesh.gloo.solo.io.HeaderManipulation.AppendRequestHeadersEntry)
  - [HeaderManipulation.AppendResponseHeadersEntry](#networking.mesh.gloo.solo.io.HeaderManipulation.AppendResponseHeadersEntry)
  - [WeightedDestination](#networking.mesh.gloo.solo.io.WeightedDestination)
  - [WeightedDestination.DestinationOptions](#networking.mesh.gloo.solo.io.WeightedDestination.DestinationOptions)
  - [WeightedDestination.KubeDestination](#networking.mesh.gloo.solo.io.WeightedDestination.KubeDestination)
  - [WeightedDestination.KubeDestination.SubsetEntry](#networking.mesh.gloo.solo.io.WeightedDestination.KubeDestination.SubsetEntry)
  - [WeightedDestination.VirtualDestination](#networking.mesh.gloo.solo.io.WeightedDestination.VirtualDestination)
  - [WeightedDestination.VirtualDestination.SubsetEntry](#networking.mesh.gloo.solo.io.WeightedDestination.VirtualDestination.SubsetEntry)







<a name="networking.mesh.gloo.solo.io.HeaderManipulation"></a>

### HeaderManipulation
Specify modifications to request and response headers.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| removeResponseHeaders | []string | repeated | HTTP headers to remove before returning a response to the caller. |
  | appendResponseHeaders | [][networking.mesh.gloo.solo.io.HeaderManipulation.AppendResponseHeadersEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.weighed_destination#networking.mesh.gloo.solo.io.HeaderManipulation.AppendResponseHeadersEntry" >}}) | repeated | Additional HTTP headers to add before returning a response to the caller. |
  | removeRequestHeaders | []string | repeated | HTTP headers to remove before forwarding a request to the destination service. |
  | appendRequestHeaders | [][networking.mesh.gloo.solo.io.HeaderManipulation.AppendRequestHeadersEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.weighed_destination#networking.mesh.gloo.solo.io.HeaderManipulation.AppendRequestHeadersEntry" >}}) | repeated | Additional HTTP headers to add before forwarding a request to the destination service. |
  





<a name="networking.mesh.gloo.solo.io.HeaderManipulation.AppendRequestHeadersEntry"></a>

### HeaderManipulation.AppendRequestHeadersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="networking.mesh.gloo.solo.io.HeaderManipulation.AppendResponseHeadersEntry"></a>

### HeaderManipulation.AppendResponseHeadersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="networking.mesh.gloo.solo.io.WeightedDestination"></a>

### WeightedDestination
Specify a traffic shift or routing destination along with a weight. Weight is only relevant when supplying multiple destinations.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| weight | uint32 |  | Specify the proportion of traffic to be forwarded to this destination. Weights across all of the `destinations` must sum to 100. |
  | kubeService | [networking.mesh.gloo.solo.io.WeightedDestination.KubeDestination]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.weighed_destination#networking.mesh.gloo.solo.io.WeightedDestination.KubeDestination" >}}) |  | Specify a Kubernetes Service. |
  | virtualDestination | [networking.mesh.gloo.solo.io.WeightedDestination.VirtualDestination]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.weighed_destination#networking.mesh.gloo.solo.io.WeightedDestination.VirtualDestination" >}}) |  | Specify a VirtualDestination. |
  | staticDestination | [core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) |  | Reference to a gloo mesh Static Destination |
  | clusterHeader | string |  | Envoy will determine the cluster to route to by reading the value of the HTTP header named by cluster_header from the request headers. If the header is not found or the referenced cluster does not exist, Envoy will return a 404 response. Avoid using this whenever possible, it does not allow for custom filter configuration based on Virtual Host. {{/* NOTE: unimplemented */}} |
  | options | [networking.mesh.gloo.solo.io.WeightedDestination.DestinationOptions]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.weighed_destination#networking.mesh.gloo.solo.io.WeightedDestination.DestinationOptions" >}}) |  | additional options / config for a route which will be applied when this destination is selected. |
  





<a name="networking.mesh.gloo.solo.io.WeightedDestination.DestinationOptions"></a>

### WeightedDestination.DestinationOptions
Specify functionality which will be applied to traffic when this particular destination is selected for routing.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headerManipulation | [networking.mesh.gloo.solo.io.HeaderManipulation]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.weighed_destination#networking.mesh.gloo.solo.io.HeaderManipulation" >}}) |  | manipualte headers on traffic sent to this destination |
  





<a name="networking.mesh.gloo.solo.io.WeightedDestination.KubeDestination"></a>

### WeightedDestination.KubeDestination
A Kubernetes destination.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The name of the service. |
  | namespace | string |  | The namespace of the service. |
  | clusterName | string |  | The Gloo Mesh cluster name (registration name) of the service. |
  | subset | [][networking.mesh.gloo.solo.io.WeightedDestination.KubeDestination.SubsetEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.weighed_destination#networking.mesh.gloo.solo.io.WeightedDestination.KubeDestination.SubsetEntry" >}}) | repeated | Specify, by labels, a subset of service instances to route to. |
  | port | uint32 |  | Port on the service to receive traffic. Required if the service exposes more than one port. |
  





<a name="networking.mesh.gloo.solo.io.WeightedDestination.KubeDestination.SubsetEntry"></a>

### WeightedDestination.KubeDestination.SubsetEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="networking.mesh.gloo.solo.io.WeightedDestination.VirtualDestination"></a>

### WeightedDestination.VirtualDestination
Specify a VirtualDestination traffic shift destination.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The name of the VirtualDestination object. |
  | namespace | string |  | The namespace of the VirtualDestination object. |
  | subset | [][networking.mesh.gloo.solo.io.WeightedDestination.VirtualDestination.SubsetEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.weighed_destination#networking.mesh.gloo.solo.io.WeightedDestination.VirtualDestination.SubsetEntry" >}}) | repeated | Specify, by labels, a subset of service instances backing the VirtualDestination to route to. |
  





<a name="networking.mesh.gloo.solo.io.WeightedDestination.VirtualDestination.SubsetEntry"></a>

### WeightedDestination.VirtualDestination.SubsetEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

