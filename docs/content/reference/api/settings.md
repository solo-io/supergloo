
---
title: "settings.proto"
---

## Package : `settings.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for settings.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## settings.proto


## Table of Contents
  - [SettingsSpec](#settings.zephyr.solo.io.SettingsSpec)
  - [SettingsSpec.AwsAccount](#settings.zephyr.solo.io.SettingsSpec.AwsAccount)
  - [SettingsSpec.AwsAccount.DiscoverySelector](#settings.zephyr.solo.io.SettingsSpec.AwsAccount.DiscoverySelector)
  - [SettingsSpec.AwsAccount.ResourceSelector](#settings.zephyr.solo.io.SettingsSpec.AwsAccount.ResourceSelector)
  - [SettingsSpec.AwsAccount.ResourceSelector.Matcher](#settings.zephyr.solo.io.SettingsSpec.AwsAccount.ResourceSelector.Matcher)
  - [SettingsSpec.AwsAccount.ResourceSelector.Matcher.TagsEntry](#settings.zephyr.solo.io.SettingsSpec.AwsAccount.ResourceSelector.Matcher.TagsEntry)
  - [SettingsStatus](#settings.zephyr.solo.io.SettingsStatus)







<a name="settings.zephyr.solo.io.SettingsSpec"></a>

### SettingsSpec
Top level SMH user configuration object.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| aws | [][SettingsSpec.AwsAccount](#settings.zephyr.solo.io.SettingsSpec.AwsAccount) | repeated |  |






<a name="settings.zephyr.solo.io.SettingsSpec.AwsAccount"></a>

### SettingsSpec.AwsAccount



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| accountId | [string](#string) |  | AWS account ID. |
| appmeshDiscovery | [SettingsSpec.AwsAccount.DiscoverySelector](#settings.zephyr.solo.io.SettingsSpec.AwsAccount.DiscoverySelector) |  | For unspecified or null fields, discovery will not run for the corresponding AWS resource type. |
| eksDiscovery | [SettingsSpec.AwsAccount.DiscoverySelector](#settings.zephyr.solo.io.SettingsSpec.AwsAccount.DiscoverySelector) |  |  |






<a name="settings.zephyr.solo.io.SettingsSpec.AwsAccount.DiscoverySelector"></a>

### SettingsSpec.AwsAccount.DiscoverySelector
Configure which AWS resources should be discovered by SMH. An AWS resource will be selected if any of the resource_selectors apply.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resourceSelectors | [][SettingsSpec.AwsAccount.ResourceSelector](#settings.zephyr.solo.io.SettingsSpec.AwsAccount.ResourceSelector) | repeated |  |






<a name="settings.zephyr.solo.io.SettingsSpec.AwsAccount.ResourceSelector"></a>

### SettingsSpec.AwsAccount.ResourceSelector
For a given resource_selector to apply to a resource, the resource must match all of the resource_selector's match criteria.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| arn | [string](#string) |  | AWS resource ARN that directly references a resource. |
| matcher | [SettingsSpec.AwsAccount.ResourceSelector.Matcher](#settings.zephyr.solo.io.SettingsSpec.AwsAccount.ResourceSelector.Matcher) |  |  |






<a name="settings.zephyr.solo.io.SettingsSpec.AwsAccount.ResourceSelector.Matcher"></a>

### SettingsSpec.AwsAccount.ResourceSelector.Matcher
Selects all resources that exist in the specified AWS region and possess the specified tags.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| regions | [][string](#string) | repeated | AWS regions, e.g. us-east-2. If unspecified, select across all regions. |
| tags | [][SettingsSpec.AwsAccount.ResourceSelector.Matcher.TagsEntry](#settings.zephyr.solo.io.SettingsSpec.AwsAccount.ResourceSelector.Matcher.TagsEntry) | repeated | AWS resource tags. If unspecified, match any tags. |






<a name="settings.zephyr.solo.io.SettingsSpec.AwsAccount.ResourceSelector.Matcher.TagsEntry"></a>

### SettingsSpec.AwsAccount.ResourceSelector.Matcher.TagsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="settings.zephyr.solo.io.SettingsStatus"></a>

### SettingsStatus






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

