
---

---

## Package : `envoy.config.core.v3`



<a name="top"></a>

<a name="API Reference for backoff.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## backoff.proto


## Table of Contents
  - [BackoffStrategy](#envoy.config.core.v3.BackoffStrategy)







<a name="envoy.config.core.v3.BackoffStrategy"></a>

### BackoffStrategy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| baseInterval | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | The base interval to be used for the next back off computation. It should be greater than zero and less than or equal to :ref:`max_interval <envoy_api_field_config.core.v3.BackoffStrategy.max_interval>`. |
  | maxInterval | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Specifies the maximum interval between retries. This parameter is optional, but must be greater than or equal to the :ref:`base_interval <envoy_api_field_config.core.v3.BackoffStrategy.base_interval>` if set. The default is 10 times the :ref:`base_interval <envoy_api_field_config.core.v3.BackoffStrategy.base_interval>`. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

