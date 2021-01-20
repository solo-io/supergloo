
---

---

## Package : `envoy.type.metadata.v3`



<a name="top"></a>

<a name="API Reference for metadata.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## metadata.proto


## Table of Contents
  - [MetadataKey](#envoy.type.metadata.v3.MetadataKey)
  - [MetadataKey.PathSegment](#envoy.type.metadata.v3.MetadataKey.PathSegment)
  - [MetadataKind](#envoy.type.metadata.v3.MetadataKind)
  - [MetadataKind.Cluster](#envoy.type.metadata.v3.MetadataKind.Cluster)
  - [MetadataKind.Host](#envoy.type.metadata.v3.MetadataKind.Host)
  - [MetadataKind.Request](#envoy.type.metadata.v3.MetadataKind.Request)
  - [MetadataKind.Route](#envoy.type.metadata.v3.MetadataKind.Route)







<a name="envoy.type.metadata.v3.MetadataKey"></a>

### MetadataKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  | The key name of Metadata to retrieve the Struct from the metadata. Typically, it represents a builtin subsystem or custom extension. |
  | path | [][envoy.type.metadata.v3.MetadataKey.PathSegment]({{< versioned_link_path fromRoot="/reference/api/envoy.type.metadata.v3.metadata#envoy.type.metadata.v3.MetadataKey.PathSegment" >}}) | repeated | The path to retrieve the Value from the Struct. It can be a prefix or a full path, e.g. ``[prop, xyz]`` for a struct or ``[prop, foo]`` for a string in the example, which depends on the particular scenario.<br>Note: Due to that only the key type segment is supported, the path can not specify a list unless the list is the last segment. |
  





<a name="envoy.type.metadata.v3.MetadataKey.PathSegment"></a>

### MetadataKey.PathSegment



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  | If specified, use the key to retrieve the value in a Struct. |
  





<a name="envoy.type.metadata.v3.MetadataKind"></a>

### MetadataKind



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request | [envoy.type.metadata.v3.MetadataKind.Request]({{< versioned_link_path fromRoot="/reference/api/envoy.type.metadata.v3.metadata#envoy.type.metadata.v3.MetadataKind.Request" >}}) |  | Request kind of metadata. |
  | route | [envoy.type.metadata.v3.MetadataKind.Route]({{< versioned_link_path fromRoot="/reference/api/envoy.type.metadata.v3.metadata#envoy.type.metadata.v3.MetadataKind.Route" >}}) |  | Route kind of metadata. |
  | cluster | [envoy.type.metadata.v3.MetadataKind.Cluster]({{< versioned_link_path fromRoot="/reference/api/envoy.type.metadata.v3.metadata#envoy.type.metadata.v3.MetadataKind.Cluster" >}}) |  | Cluster kind of metadata. |
  | host | [envoy.type.metadata.v3.MetadataKind.Host]({{< versioned_link_path fromRoot="/reference/api/envoy.type.metadata.v3.metadata#envoy.type.metadata.v3.MetadataKind.Host" >}}) |  | Host kind of metadata. |
  





<a name="envoy.type.metadata.v3.MetadataKind.Cluster"></a>

### MetadataKind.Cluster







<a name="envoy.type.metadata.v3.MetadataKind.Host"></a>

### MetadataKind.Host







<a name="envoy.type.metadata.v3.MetadataKind.Request"></a>

### MetadataKind.Request







<a name="envoy.type.metadata.v3.MetadataKind.Route"></a>

### MetadataKind.Route






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

