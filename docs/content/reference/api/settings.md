
---
title: "settings.proto"
---

## Package : `core.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for settings.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## settings.proto


## Table of Contents
  - [SettingsSpec](#core.zephyr.solo.io.SettingsSpec)
  - [SettingsSpec.Aws](#core.zephyr.solo.io.SettingsSpec.Aws)
  - [SettingsSpec.AwsAccount](#core.zephyr.solo.io.SettingsSpec.AwsAccount)
  - [SettingsSpec.AwsAccount.DiscoverySelector](#core.zephyr.solo.io.SettingsSpec.AwsAccount.DiscoverySelector)
  - [SettingsSpec.AwsAccount.ResourceSelector](#core.zephyr.solo.io.SettingsSpec.AwsAccount.ResourceSelector)
  - [SettingsSpec.AwsAccount.ResourceSelector.Matcher](#core.zephyr.solo.io.SettingsSpec.AwsAccount.ResourceSelector.Matcher)
  - [SettingsSpec.AwsAccount.ResourceSelector.Matcher.TagsEntry](#core.zephyr.solo.io.SettingsSpec.AwsAccount.ResourceSelector.Matcher.TagsEntry)
  - [SettingsStatus](#core.zephyr.solo.io.SettingsStatus)







<a name="core.zephyr.solo.io.SettingsSpec"></a>

### SettingsSpec
Top level SMH user configuration object.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| aws | [SettingsSpec.Aws](#core.zephyr.solo.io.SettingsSpec.Aws) |  |  |






<a name="core.zephyr.solo.io.SettingsSpec.Aws"></a>

### SettingsSpec.Aws



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| disabled | [bool](#bool) |  | If true, disable integration with AWS. |
| accounts | [][SettingsSpec.AwsAccount](#core.zephyr.solo.io.SettingsSpec.AwsAccount) | repeated | Per-account AWS settings. |






<a name="core.zephyr.solo.io.SettingsSpec.AwsAccount"></a>

### SettingsSpec.AwsAccount



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| accountId | [string](#string) |  | AWS account ID. |
| meshDiscovery | [SettingsSpec.AwsAccount.DiscoverySelector](#core.zephyr.solo.io.SettingsSpec.AwsAccount.DiscoverySelector) |  | If unspecified, by default discovery will run for AppMesh in all regions. |
| eksDiscovery | [SettingsSpec.AwsAccount.DiscoverySelector](#core.zephyr.solo.io.SettingsSpec.AwsAccount.DiscoverySelector) |  | If unspecified, by default discovery will run for EKS clusters in all regions. |






<a name="core.zephyr.solo.io.SettingsSpec.AwsAccount.DiscoverySelector"></a>

### SettingsSpec.AwsAccount.DiscoverySelector
Configure which AWS resources should be discovered by SMH. An AWS resource will be selected if any of the resource_selectors apply.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| disabled | [bool](#bool) |  | If true, disable discovery. |
| resourceSelectors | [][SettingsSpec.AwsAccount.ResourceSelector](#core.zephyr.solo.io.SettingsSpec.AwsAccount.ResourceSelector) | repeated |  |






<a name="core.zephyr.solo.io.SettingsSpec.AwsAccount.ResourceSelector"></a>

### SettingsSpec.AwsAccount.ResourceSelector
For a given resource_selector to apply to a resource, the resource must match all of the resource_selector's match criteria.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| arn | [string](#string) |  | AWS resource ARN that directly references a resource. |
| matcher | [SettingsSpec.AwsAccount.ResourceSelector.Matcher](#core.zephyr.solo.io.SettingsSpec.AwsAccount.ResourceSelector.Matcher) |  |  |






<a name="core.zephyr.solo.io.SettingsSpec.AwsAccount.ResourceSelector.Matcher"></a>

### SettingsSpec.AwsAccount.ResourceSelector.Matcher
Selects all resources that exist in the specified AWS region and possess the specified tags.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| regions | [][string](#string) | repeated | AWS regions, e.g. us-east-2. If unspecified, select across all regions. |
| tags | [][SettingsSpec.AwsAccount.ResourceSelector.Matcher.TagsEntry](#core.zephyr.solo.io.SettingsSpec.AwsAccount.ResourceSelector.Matcher.TagsEntry) | repeated | AWS resource tags. If unspecified, match any tags. |






<a name="core.zephyr.solo.io.SettingsSpec.AwsAccount.ResourceSelector.Matcher.TagsEntry"></a>

### SettingsSpec.AwsAccount.ResourceSelector.Matcher.TagsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="core.zephyr.solo.io.SettingsStatus"></a>

### SettingsStatus






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

