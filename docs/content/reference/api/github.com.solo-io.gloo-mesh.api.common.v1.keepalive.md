
---

title: "keepalive.proto"

---

## Package : `common.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for keepalive.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## keepalive.proto


## Table of Contents
  - [TCPKeepalive](#common.mesh.gloo.solo.io.TCPKeepalive)







<a name="common.mesh.gloo.solo.io.TCPKeepalive"></a>

### TCPKeepalive
Configure TCP keepalive for the ingress gateways of all meshes in this VirtualMesh.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| probes | uint32 |  | Maximum number of TCP keepalive probes to send before determining that connection is dead. |
  | time | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | The time duration a connection needs to be idle before keep-alive probes start being sent. |
  | interval | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | The time duration between keep-alive probes. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

