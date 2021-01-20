
---

---

## Package : `envoy.config.core.v3`



<a name="top"></a>

<a name="API Reference for http_uri.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## http_uri.proto


## Table of Contents
  - [HttpUri](#envoy.config.core.v3.HttpUri)







<a name="envoy.config.core.v3.HttpUri"></a>

### HttpUri



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uri | string |  | The HTTP server URI. It should be a full FQDN with protocol, host and path.<br>Example:<br>.. code-block:: yaml<br>   uri: https://www.googleapis.com/oauth2/v1/certs |
  | cluster | string |  | A cluster is created in the Envoy "cluster_manager" config section. This field specifies the cluster name.<br>Example:<br>.. code-block:: yaml<br>   cluster: jwks_cluster |
  | timeout | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Sets the maximum duration in milliseconds that a response can take to arrive upon request. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

