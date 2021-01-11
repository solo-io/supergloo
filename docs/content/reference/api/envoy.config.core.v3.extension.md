
---

---

## Package : `envoy.config.core.v3`



<a name="top"></a>

<a name="API Reference for extension.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## extension.proto


## Table of Contents
  - [ExtensionConfigSource](#envoy.config.core.v3.ExtensionConfigSource)
  - [TypedExtensionConfig](#envoy.config.core.v3.TypedExtensionConfig)







<a name="envoy.config.core.v3.ExtensionConfigSource"></a>

### ExtensionConfigSource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| configSource | [envoy.config.core.v3.ConfigSource]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.config_source#envoy.config.core.v3.ConfigSource" >}}) |  |  |
  | defaultConfig | [google.protobuf.Any]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.any#google.protobuf.Any" >}}) |  | Optional default configuration to use as the initial configuration if there is a failure to receive the initial extension configuration or if `apply_default_config_without_warming` flag is set. |
  | applyDefaultConfigWithoutWarming | bool |  | Use the default config as the initial configuration without warming and waiting for the first discovery response. Requires the default configuration to be supplied. |
  | typeUrls | []string | repeated | A set of permitted extension type URLs. Extension configuration updates are rejected if they do not match any type URL in the set. |
  





<a name="envoy.config.core.v3.TypedExtensionConfig"></a>

### TypedExtensionConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The name of an extension. This is not used to select the extension, instead it serves the role of an opaque identifier. |
  | typedConfig | [google.protobuf.Any]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.any#google.protobuf.Any" >}}) |  | The typed config for the extension. The type URL will be used to identify the extension. In the case that the type URL is *udpa.type.v1.TypedStruct*, the inner type URL of *TypedStruct* will be utilized. See the :ref:`extension configuration overview <config_overview_extension_configuration>` for further details. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

