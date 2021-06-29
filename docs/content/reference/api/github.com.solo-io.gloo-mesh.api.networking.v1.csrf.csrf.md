
---

title: "csrf.proto"

---

## Package : `csrf.networking.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for csrf.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## csrf.proto


## Table of Contents
  - [CsrfPolicy](#csrf.networking.mesh.gloo.solo.io.CsrfPolicy)







<a name="csrf.networking.mesh.gloo.solo.io.CsrfPolicy"></a>

### CsrfPolicy
CSRF filter config.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| filterEnabled | bool |  | Specifies that CSRF policies will be evaluated, tracked and enforced. |
  | shadowEnabled | bool |  | Specifies that CSRF policies will be evaluated and tracked, but not enforced.<br>This is intended to be used when ``filter_enabled`` is false and will be ignored otherwise. |
  | percentage | double |  | Specifies the % of requests for which the CSRF filter is enabled or when shadow mode is enabled the % of requests evaluated and tracked, but not enforced.<br>If filter_enabled or shadow_enabled is true. Envoy will lookup the runtime key to get the percentage of requests to filter.<br>.. note:: This field defaults to 100 |
  | additionalOrigins | [][common.mesh.gloo.solo.io.StringMatch]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.string_match#common.mesh.gloo.solo.io.StringMatch" >}}) | repeated | Specifies additional source origins that will be allowed in addition to the destination origin. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

