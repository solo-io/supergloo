
---
title: "networking.zephyr.solo.iogithub.com/solo-io/mesh-projects/api/networking/v1alpha1/access_control.proto"
---

## Package : `networking.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for github.com/solo-io/mesh-projects/api/networking/v1alpha1/access_control.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## github.com/solo-io/mesh-projects/api/networking/v1alpha1/access_control.proto


## Table of Contents
  - [AccessControlPolicySpec](#networking.zephyr.solo.io.AccessControlPolicySpec)
  - [AccessControlPolicyStatus](#networking.zephyr.solo.io.AccessControlPolicyStatus)







<a name="networking.zephyr.solo.io.AccessControlPolicySpec"></a>

### AccessControlPolicySpec
access control policies apply ALLOW policies to communication in a mesh access control policies specify the following: ALLOW those requests: - originating from from **source pods** - sent to **destination pods** - matching one or more **request matcher** if no access control policies are present, all traffic in the mesh will be set to ALLOW


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sourceSelector | [core.zephyr.solo.io.IdentitySelector](#core.zephyr.solo.io.IdentitySelector) |  | requests originating from these pods will have the rule applied leave empty to have all pods in the mesh apply these policies<br>note that access control policies are mapped to source pods by their service account. if other pods share the same service account, this access control rule will apply to those pods as well.<br>for fine-grained access control policies, ensure that your service accounts properly reflect the desired boundary for your access control policies |
| destinationSelector | [core.zephyr.solo.io.Selector](#core.zephyr.solo.io.Selector) |  | requests destined for these pods will have the rule applied leave empty to apply to all destination pods in the mesh |
| allowedPaths | [][string](#string) | repeated | Optional. A list of HTTP paths or gRPC methods to allow. gRPC methods must be presented as fully-qualified name in the form of "/packageName.serviceName/methodName" and are case sensitive. Exact match, prefix match, and suffix match are supported for paths. For example, the path "/books/review" matches "/books/review" (exact match), "*books/" (suffix match), or "/books*" (prefix match),<br>If not specified, it allows to any path. |
| allowedMethods | [][string](#string) | repeated | Optional. A list of HTTP methods to allow (e.g., "GET", "POST"). It is ignored in gRPC case because the value is always "POST". If set to ["*"] or not specified, it allows to any method. |
| allowedPorts | [][string](#string) | repeated | Optional. A list of ports which to allow if not set any port is allowed |






<a name="networking.zephyr.solo.io.AccessControlPolicyStatus"></a>

### AccessControlPolicyStatus






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

