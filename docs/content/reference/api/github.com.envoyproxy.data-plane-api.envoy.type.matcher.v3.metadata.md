
---

---

## Package : `envoy.type.matcher.v3`



<a name="top"></a>

<a name="API Reference for metadata.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## metadata.proto


## Table of Contents
  - [MetadataMatcher](#envoy.type.matcher.v3.MetadataMatcher)
  - [MetadataMatcher.PathSegment](#envoy.type.matcher.v3.MetadataMatcher.PathSegment)







<a name="envoy.type.matcher.v3.MetadataMatcher"></a>

### MetadataMatcher



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| filter | string |  | The filter name to retrieve the Struct from the Metadata. |
  | path | [][envoy.type.matcher.v3.MetadataMatcher.PathSegment]({{< versioned_link_path fromRoot="/reference/api/github.com.envoyproxy.data-plane-api.envoy.type.matcher.v3.metadata#envoy.type.matcher.v3.MetadataMatcher.PathSegment" >}}) | repeated | The path to retrieve the Value from the Struct. |
  | value | [envoy.type.matcher.v3.ValueMatcher]({{< versioned_link_path fromRoot="/reference/api/envoy.type.matcher.v3.value#envoy.type.matcher.v3.ValueMatcher" >}}) |  | The MetadataMatcher is matched if the value retrieved by path is matched to this value. |
  





<a name="envoy.type.matcher.v3.MetadataMatcher.PathSegment"></a>

### MetadataMatcher.PathSegment



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  | If specified, use the key to retrieve the value in a Struct. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

