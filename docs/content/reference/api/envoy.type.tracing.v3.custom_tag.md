
---

---

## Package : `envoy.type.tracing.v3`



<a name="top"></a>

<a name="API Reference for custom_tag.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## custom_tag.proto


## Table of Contents
  - [CustomTag](#envoy.type.tracing.v3.CustomTag)
  - [CustomTag.Environment](#envoy.type.tracing.v3.CustomTag.Environment)
  - [CustomTag.Header](#envoy.type.tracing.v3.CustomTag.Header)
  - [CustomTag.Literal](#envoy.type.tracing.v3.CustomTag.Literal)
  - [CustomTag.Metadata](#envoy.type.tracing.v3.CustomTag.Metadata)







<a name="envoy.type.tracing.v3.CustomTag"></a>

### CustomTag



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tag | string |  | Used to populate the tag name. |
  | literal | [envoy.type.tracing.v3.CustomTag.Literal]({{< versioned_link_path fromRoot="/reference/api/envoy.type.tracing.v3.custom_tag#envoy.type.tracing.v3.CustomTag.Literal" >}}) |  | A literal custom tag. |
  | environment | [envoy.type.tracing.v3.CustomTag.Environment]({{< versioned_link_path fromRoot="/reference/api/envoy.type.tracing.v3.custom_tag#envoy.type.tracing.v3.CustomTag.Environment" >}}) |  | An environment custom tag. |
  | requestHeader | [envoy.type.tracing.v3.CustomTag.Header]({{< versioned_link_path fromRoot="/reference/api/envoy.type.tracing.v3.custom_tag#envoy.type.tracing.v3.CustomTag.Header" >}}) |  | A request header custom tag. |
  | metadata | [envoy.type.tracing.v3.CustomTag.Metadata]({{< versioned_link_path fromRoot="/reference/api/envoy.type.tracing.v3.custom_tag#envoy.type.tracing.v3.CustomTag.Metadata" >}}) |  | A custom tag to obtain tag value from the metadata. |
  





<a name="envoy.type.tracing.v3.CustomTag.Environment"></a>

### CustomTag.Environment



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Environment variable name to obtain the value to populate the tag value. |
  | defaultValue | string |  | When the environment variable is not found, the tag value will be populated with this default value if specified, otherwise no tag will be populated. |
  





<a name="envoy.type.tracing.v3.CustomTag.Header"></a>

### CustomTag.Header



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Header name to obtain the value to populate the tag value. |
  | defaultValue | string |  | When the header does not exist, the tag value will be populated with this default value if specified, otherwise no tag will be populated. |
  





<a name="envoy.type.tracing.v3.CustomTag.Literal"></a>

### CustomTag.Literal



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | string |  | Static literal value to populate the tag value. |
  





<a name="envoy.type.tracing.v3.CustomTag.Metadata"></a>

### CustomTag.Metadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [envoy.type.metadata.v3.MetadataKind]({{< versioned_link_path fromRoot="/reference/api/envoy.type.metadata.v3.metadata#envoy.type.metadata.v3.MetadataKind" >}}) |  | Specify what kind of metadata to obtain tag value from. |
  | metadataKey | [envoy.type.metadata.v3.MetadataKey]({{< versioned_link_path fromRoot="/reference/api/envoy.type.metadata.v3.metadata#envoy.type.metadata.v3.MetadataKey" >}}) |  | Metadata key to define the path to retrieve the tag value. |
  | defaultValue | string |  | When no valid metadata is found, the tag value would be populated with this default value if specified, otherwise no tag would be populated. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

