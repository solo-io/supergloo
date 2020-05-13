
---
title: "aws.proto"
---

## Package : `settings.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for aws.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## aws.proto


## Table of Contents
  - [AwsAccountSettings](#settings.zephyr.solo.io.AwsAccountSettings)
  - [AwsSettings](#settings.zephyr.solo.io.AwsSettings)
  - [DiscoverySelectors](#settings.zephyr.solo.io.DiscoverySelectors)
  - [DiscoverySettings](#settings.zephyr.solo.io.DiscoverySettings)
  - [ResourceSelector](#settings.zephyr.solo.io.ResourceSelector)
  - [ResourceSelector.Matcher](#settings.zephyr.solo.io.ResourceSelector.Matcher)
  - [ResourceSelector.Matcher.TagsEntry](#settings.zephyr.solo.io.ResourceSelector.Matcher.TagsEntry)







<a name="settings.zephyr.solo.io.AwsAccountSettings"></a>

### AwsAccountSettings



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| accountId | [string](#string) |  | AWS account ID. |
| discoverySettings | [DiscoverySettings](#settings.zephyr.solo.io.DiscoverySettings) |  |  |






<a name="settings.zephyr.solo.io.AwsSettings"></a>

### AwsSettings



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| awsAccountSettings | [][AwsAccountSettings](#settings.zephyr.solo.io.AwsAccountSettings) | repeated |  |






<a name="settings.zephyr.solo.io.DiscoverySelectors"></a>

### DiscoverySelectors
An AWS resource will be selected if *any* of the resource_selectors apply. For a given resource_selector to apply to a resource, the resource must match *all* of the resource_selector's match criteria.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resourceSelectors | [][ResourceSelector](#settings.zephyr.solo.io.ResourceSelector) | repeated |  |






<a name="settings.zephyr.solo.io.DiscoverySettings"></a>

### DiscoverySettings
Configure which AWS resources should be discovered by SMH. For unspecified or null fields, discovery will not run for the corresponding AWS resource type.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| appmesh | [DiscoverySelectors](#settings.zephyr.solo.io.DiscoverySelectors) |  |  |
| eks | [DiscoverySelectors](#settings.zephyr.solo.io.DiscoverySelectors) |  |  |






<a name="settings.zephyr.solo.io.ResourceSelector"></a>

### ResourceSelector



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| arn | [string](#string) |  | AWS resource ARN that directly references a resource. |
| matcher | [ResourceSelector.Matcher](#settings.zephyr.solo.io.ResourceSelector.Matcher) |  |  |






<a name="settings.zephyr.solo.io.ResourceSelector.Matcher"></a>

### ResourceSelector.Matcher
Selects all resources that exist in the specified AWS region and possess the specified tags.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| regions | [][string](#string) | repeated | AWS regions, e.g. us-east-2. If unspecified, select across all regions. |
| tags | [][ResourceSelector.Matcher.TagsEntry](#settings.zephyr.solo.io.ResourceSelector.Matcher.TagsEntry) | repeated | AWS resource tags. If unspecified, match any tags. |






<a name="settings.zephyr.solo.io.ResourceSelector.Matcher.TagsEntry"></a>

### ResourceSelector.Matcher.TagsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

