
---
title: "aws.proto"
---

## Package : `config.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for aws.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## aws.proto


## Table of Contents
  - [AwsAccountConfig](#config.zephyr.solo.io.AwsAccountConfig)
  - [AwsConfig](#config.zephyr.solo.io.AwsConfig)
  - [DiscoveryConfig](#config.zephyr.solo.io.DiscoveryConfig)
  - [DiscoverySelectors](#config.zephyr.solo.io.DiscoverySelectors)
  - [ResourceSelector](#config.zephyr.solo.io.ResourceSelector)
  - [ResourceSelector.Matcher](#config.zephyr.solo.io.ResourceSelector.Matcher)
  - [ResourceSelector.Matcher.TagsEntry](#config.zephyr.solo.io.ResourceSelector.Matcher.TagsEntry)







<a name="config.zephyr.solo.io.AwsAccountConfig"></a>

### AwsAccountConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| accountId | [string](#string) |  | AWS account ID. |
| discoveryConfig | [DiscoveryConfig](#config.zephyr.solo.io.DiscoveryConfig) |  |  |






<a name="config.zephyr.solo.io.AwsConfig"></a>

### AwsConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| awsAccountConfigs | [][AwsAccountConfig](#config.zephyr.solo.io.AwsAccountConfig) | repeated |  |






<a name="config.zephyr.solo.io.DiscoveryConfig"></a>

### DiscoveryConfig
Configure which AWS resources should be discovered by SMH. For unspecified or null fields, discovery will not run for the corresponding AWS resource type.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| appmesh | [DiscoverySelectors](#config.zephyr.solo.io.DiscoverySelectors) |  |  |
| eks | [DiscoverySelectors](#config.zephyr.solo.io.DiscoverySelectors) |  |  |






<a name="config.zephyr.solo.io.DiscoverySelectors"></a>

### DiscoverySelectors
An AWS resource will be selected if *any* of the resource_selectors apply. For a given resource_selector to apply to a resource, the resource must match *all* of the resource_selector's match criteria.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resourceSelectors | [][ResourceSelector](#config.zephyr.solo.io.ResourceSelector) | repeated |  |






<a name="config.zephyr.solo.io.ResourceSelector"></a>

### ResourceSelector



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| arn | [string](#string) |  | AWS resource ARN that directly references a resource. |
| matcher | [ResourceSelector.Matcher](#config.zephyr.solo.io.ResourceSelector.Matcher) |  |  |






<a name="config.zephyr.solo.io.ResourceSelector.Matcher"></a>

### ResourceSelector.Matcher
Selects all resources that exist in the specified AWS region and possess the specified tags.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| regions | [][string](#string) | repeated | AWS regions, e.g. us-east-2. If unspecified, select across all regions. |
| tags | [][ResourceSelector.Matcher.TagsEntry](#config.zephyr.solo.io.ResourceSelector.Matcher.TagsEntry) | repeated | AWS resource tags. If unspecified, match any tags. |






<a name="config.zephyr.solo.io.ResourceSelector.Matcher.TagsEntry"></a>

### ResourceSelector.Matcher.TagsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

