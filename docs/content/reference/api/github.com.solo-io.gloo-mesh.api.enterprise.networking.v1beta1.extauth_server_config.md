
---

title: "extauth_server_config.proto"

---

## Package : `networking.enterprise.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for extauth_server_config.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## extauth_server_config.proto


## Table of Contents
  - [ExtauthServerConfigSpec](#networking.enterprise.mesh.gloo.solo.io.ExtauthServerConfigSpec)
  - [ExtauthServerConfigStatus](#networking.enterprise.mesh.gloo.solo.io.ExtauthServerConfigStatus)







<a name="networking.enterprise.mesh.gloo.solo.io.ExtauthServerConfigSpec"></a>

### ExtauthServerConfigSpec
ExtauthConfig contains the configuration for the Gloo Extauth server, the external extauth server used by mesh proxies to authenticate HTTP requests. One or more extauth servers may be deployed in order to authenticate traffic across East-West and North-South routes. The ExtauthConfig allows users to map a single extauth configuration to multiple extauth server instances, deployed across managed clusters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| serverConfigRefs | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated | The per-server extauth config objects will be generated from the given config for each provided ref. Each extauth server must be configured to read its server configuration from one of these refs. |
  | extauthConfig | [extauth.api.solo.io.ExtAuthConfigSpec]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.solo-apis.api.extauth.v1alpha1.extauth#extauth.api.solo.io.ExtAuthConfigSpec" >}}) |  | the configuration which will be deployed to the selected extauth limit servers. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.ExtauthServerConfigStatus"></a>

### ExtauthServerConfigStatus
The current status of the `ExtauthConfig`.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the ExtauthServerConfig metadata. If the `observedGeneration` does not match `metadata.generation`, Gloo Mesh has not processed the most recent version of this resource. |
  | errors | []string | repeated | Any errors found while processing this generation of the resource. |
  | warnings | []string | repeated | Any warnings found while processing this generation of the resource. |
  | configuredServers | [][core.skv2.solo.io.ClusterObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ClusterObjectRef" >}}) | repeated | a list of extauth limit server workloads which have been configured with this ExtauthConfig |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

