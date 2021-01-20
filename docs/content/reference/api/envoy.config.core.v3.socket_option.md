
---

---

## Package : `envoy.config.core.v3`



<a name="top"></a>

<a name="API Reference for socket_option.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## socket_option.proto


## Table of Contents
  - [SocketOption](#envoy.config.core.v3.SocketOption)

  - [SocketOption.SocketState](#envoy.config.core.v3.SocketOption.SocketState)






<a name="envoy.config.core.v3.SocketOption"></a>

### SocketOption



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| description | string |  | An optional name to give this socket option for debugging, etc. Uniqueness is not required and no special meaning is assumed. |
  | level | int64 |  | Corresponding to the level value passed to setsockopt, such as IPPROTO_TCP |
  | name | int64 |  | The numeric name as passed to setsockopt |
  | intValue | int64 |  | Because many sockopts take an int value. |
  | bufValue | bytes |  | Otherwise it's a byte buffer. |
  | state | [envoy.config.core.v3.SocketOption.SocketState]({{< versioned_link_path fromRoot="/reference/api/envoy.config.core.v3.socket_option#envoy.config.core.v3.SocketOption.SocketState" >}}) |  | The state in which the option will be applied. When used in BindConfig STATE_PREBIND is currently the only valid value. |
  




 <!-- end messages -->


<a name="envoy.config.core.v3.SocketOption.SocketState"></a>

### SocketOption.SocketState


| Name | Number | Description |
| ---- | ------ | ----------- |
| STATE_PREBIND | 0 | Socket options are applied after socket creation but before binding the socket to a port |
| STATE_BOUND | 1 | Socket options are applied after binding the socket to a port but before calling listen() |
| STATE_LISTENING | 2 | Socket options are applied after calling listen() |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

