
---
title: "access_control_policy.proto"
---

## Package : `networking.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for access_control_policy.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## access_control_policy.proto


## Table of Contents
  - [AccessControlPolicySpec](#networking.zephyr.solo.io.AccessControlPolicySpec)
  - [AccessControlPolicyStatus](#networking.zephyr.solo.io.AccessControlPolicyStatus)
  - [AccessControlPolicyStatus.TranslatorError](#networking.zephyr.solo.io.AccessControlPolicyStatus.TranslatorError)







<a name="networking.zephyr.solo.io.AccessControlPolicySpec"></a>

### AccessControlPolicySpec
access control policies apply ALLOW policies to communication in a mesh access control policies specify the following: ALLOW those requests: - originating from from **source pods** - sent to **destination pods** - matching the indicated request criteria (allowed_paths, allowed_methods, allowed_ports) if no access control policies are present, all traffic in the mesh will be set to ALLOW


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sourceSelector | [core.zephyr.solo.io.IdentitySelector](#core.zephyr.solo.io.IdentitySelector) |  | requests originating from these pods will have the rule applied leave empty to have all pods in the mesh apply these policies<br>note that access control policies are mapped to source pods by their service account. if other pods share the same service account, this access control rule will apply to those pods as well.<br>for fine-grained access control policies, ensure that your service accounts properly reflect the desired boundary for your access control policies |
| destinationSelector | [core.zephyr.solo.io.ServiceSelector](#core.zephyr.solo.io.ServiceSelector) |  | requests destined for these pods will have the rule applied leave empty to apply to all destination pods in the mesh |
| allowedPaths | [][string](#string) | repeated | Optional. A list of HTTP paths or gRPC methods to allow. gRPC methods must be presented as fully-qualified name in the form of "/packageName.serviceName/methodName" and are case sensitive. Exact match, prefix match, and suffix match are supported for paths. For example, the path "/books/review" matches "/books/review" (exact match), "*books/" (suffix match), or "/books*" (prefix match),<br>If not specified, it allows to any path. |
| allowedMethods | [][core.zephyr.solo.io.HttpMethodValue](#core.zephyr.solo.io.HttpMethodValue) | repeated | Optional. A list of HTTP methods to allow (e.g., "GET", "POST"). It is ignored in gRPC case because the value is always "POST". If not specified, allows any method. |
| allowedPorts | [][uint32](#uint32) | repeated | Optional. A list of ports which to allow if not set any port is allowed |






<a name="networking.zephyr.solo.io.AccessControlPolicyStatus"></a>

### AccessControlPolicyStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| translationStatus | [core.zephyr.solo.io.Status](#core.zephyr.solo.io.Status) |  | The status reported by the process translating this resource into mesh-specific resource(s). |
| translatorErrors | [][AccessControlPolicyStatus.TranslatorError](#networking.zephyr.solo.io.AccessControlPolicyStatus.TranslatorError) | repeated | More detailed errors than the base status provided by `translation_status`. |






<a name="networking.zephyr.solo.io.AccessControlPolicyStatus.TranslatorError"></a>

### AccessControlPolicyStatus.TranslatorError
TODO use a shared Status message with TrafficPolicy once autopilot allows for it


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| translatorId | [string](#string) |  | ID representing a translator that translates TrafficPolicy to Mesh-specific config |
| errorMessage | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

