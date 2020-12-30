
---

---

## Package : `google.protobuf`



<a name="top"></a>

<a name="API Reference for api.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## api.proto


## Table of Contents
  - [Api](#google.protobuf.Api)
  - [Method](#google.protobuf.Method)
  - [Mixin](#google.protobuf.Mixin)







<a name="google.protobuf.Api"></a>

### Api



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The fully qualified name of this interface, including package name followed by the interface's simple name. |
  | methods | [][google.protobuf.Method](.././github.com.solo-io.protoc-gen-ext.external.google.protobuf.api#google.protobuf.Method) | repeated | The methods of this interface, in unspecified order. |
  | options | [][google.protobuf.Option](.././github.com.solo-io.protoc-gen-ext.external.google.protobuf.type#google.protobuf.Option) | repeated | Any metadata attached to the interface. |
  | version | string |  | A version string for this interface. If specified, must have the form `major-version.minor-version`, as in `1.10`. If the minor version is omitted, it defaults to zero. If the entire version field is empty, the major version is derived from the package name, as outlined below. If the field is not empty, the version in the package name will be verified to be consistent with what is provided here.<br>The versioning schema uses [semantic versioning](http://semver.org) where the major version number indicates a breaking change and the minor version an additive, non-breaking change. Both version numbers are signals to users what to expect from different versions, and should be carefully chosen based on the product plan.<br>The major version is also reflected in the package name of the interface, which must end in `v<major-version>`, as in `google.feature.v1`. For major versions 0 and 1, the suffix can be omitted. Zero major versions must only be used for experimental, non-GA interfaces. |
  | sourceContext | [google.protobuf.SourceContext](.././github.com.solo-io.protoc-gen-ext.external.google.protobuf.source_context#google.protobuf.SourceContext) |  | Source context for the protocol buffer service represented by this message. |
  | mixins | [][google.protobuf.Mixin](.././github.com.solo-io.protoc-gen-ext.external.google.protobuf.api#google.protobuf.Mixin) | repeated | Included interfaces. See [Mixin][]. |
  | syntax | [google.protobuf.Syntax](.././github.com.solo-io.protoc-gen-ext.external.google.protobuf.type#google.protobuf.Syntax) |  | The source syntax of the service. |
  





<a name="google.protobuf.Method"></a>

### Method



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The simple name of this method. |
  | requestTypeUrl | string |  | A URL of the input message type. |
  | requestStreaming | bool |  | If true, the request is streamed. |
  | responseTypeUrl | string |  | The URL of the output message type. |
  | responseStreaming | bool |  | If true, the response is streamed. |
  | options | [][google.protobuf.Option](.././github.com.solo-io.protoc-gen-ext.external.google.protobuf.type#google.protobuf.Option) | repeated | Any metadata attached to the method. |
  | syntax | [google.protobuf.Syntax](.././github.com.solo-io.protoc-gen-ext.external.google.protobuf.type#google.protobuf.Syntax) |  | The source syntax of this method. |
  





<a name="google.protobuf.Mixin"></a>

### Mixin



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The fully qualified name of the interface which is included. |
  | root | string |  | If non-empty specifies a path under which inherited HTTP paths are rooted. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

