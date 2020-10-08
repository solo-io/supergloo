
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
  - [SettingsSpec.MTLS](#settings.smh.solo.io.SettingsSpec.MTLS)
  - [SettingsStatus](#settings.smh.solo.io.SettingsStatus)







<a name="settings.smh.solo.io.SettingsSpec"></a>

### SettingsSpec
Configure global settings and defaults.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| mtls | [SettingsSpec.MTLS](#settings.smh.solo.io.SettingsSpec.MTLS) |  | Configure mTLS settings for TrafficTargets. |






<a name="settings.smh.solo.io.SettingsSpec.MTLS"></a>

### SettingsSpec.MTLS
mTLS settings.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| defaultMtls | [bool](#bool) |  | If true, by default enable mTLS for all TrafficTargets unless explicitly disabled by TrafficPolicies. |






<a name="settings.smh.solo.io.SettingsStatus"></a>

### SettingsStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observedGeneration | [int64](#int64) |  | The most recent generation observed in the the Settings metadata. If the observedGeneration does not match generation, the controller has not processed the most recent version of this resource. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

