
---
title: "settings.proto"
---

## Package : `settings.smh.solo.io`



<a name="top"></a>

<a name="API Reference for settings.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## settings.proto


## Table of Contents
  - [SettingsSpec](#settings.smh.solo.io.SettingsSpec)
  - [SettingsStatus](#settings.smh.solo.io.SettingsStatus)







<a name="settings.smh.solo.io.SettingsSpec"></a>

### SettingsSpec
Configure global settings and defaults.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| mtls | networking.smh.solo.io.TrafficPolicySpec.MTLS |  | Configure default mTLS settings for TrafficTargets (MTLS declared in TrafficPolicies take precedence) |






<a name="settings.smh.solo.io.SettingsStatus"></a>

### SettingsStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | int64 |  | The most recent generation observed in the the Settings metadata. If the observedGeneration does not match generation, the controller has not processed the most recent version of this resource. |
| state | networking.smh.solo.io.ApprovalState |  | The state of the overall resource. It will only show accepted if no processing errors encountered. |
| errors | []string | repeated | Any errors encountered while processing Settings object. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

