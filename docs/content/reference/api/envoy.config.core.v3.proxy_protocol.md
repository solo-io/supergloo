
---

---

## Package : `envoy.config.core.v3`



<a name="top"></a>

<a name="API Reference for proxy_protocol.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## proxy_protocol.proto


## Table of Contents
  - [ProxyProtocolConfig](#envoy.config.core.v3.ProxyProtocolConfig)

  - [ProxyProtocolConfig.Version](#envoy.config.core.v3.ProxyProtocolConfig.Version)






<a name="envoy.config.core.v3.ProxyProtocolConfig"></a>

### ProxyProtocolConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| version | [envoy.config.core.v3.ProxyProtocolConfig.Version]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.proxy_protocol#envoy.config.core.v3.ProxyProtocolConfig.Version" >}}) |  | The PROXY protocol version to use. See https://www.haproxy.org/download/2.1/doc/proxy-protocol.txt for details |
  




 <!-- end messages -->


<a name="envoy.config.core.v3.ProxyProtocolConfig.Version"></a>

### ProxyProtocolConfig.Version


| Name | Number | Description |
| ---- | ------ | ----------- |
| V1 | 0 | PROXY protocol version 1. Human readable format. |
| V2 | 1 | PROXY protocol version 2. Binary format. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

