
---
title: "core.zephyr.solo.iogithub.com/solo-io/mesh-projects/api/core/v1alpha1/identity_selector.proto"
---

## Package : `core.zephyr.solo.io`



<a name="top"></a>

<a name="API Reference for github.com/solo-io/mesh-projects/api/core/v1alpha1/identity_selector.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## github.com/solo-io/mesh-projects/api/core/v1alpha1/identity_selector.proto


## Table of Contents
  - [IdentitySelector](#core.zephyr.solo.io.IdentitySelector)
  - [IdentitySelector.Matcher](#core.zephyr.solo.io.IdentitySelector.Matcher)
  - [IdentitySelector.ServiceAccountRefs](#core.zephyr.solo.io.IdentitySelector.ServiceAccountRefs)







<a name="core.zephyr.solo.io.IdentitySelector"></a>

### IdentitySelector
Selector capable of selecting specific service identities. Useful for binding policy rules. Either (namespaces, cluster, service_account_names) or service_accounts can be specified. If all fields are omitted, any source identity is permitted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| matcher | [IdentitySelector.Matcher](#core.zephyr.solo.io.IdentitySelector.Matcher) |  |  |
| serviceAccountRefs | [IdentitySelector.ServiceAccountRefs](#core.zephyr.solo.io.IdentitySelector.ServiceAccountRefs) |  |  |






<a name="core.zephyr.solo.io.IdentitySelector.Matcher"></a>

### IdentitySelector.Matcher



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| namespaces | [][string](#string) | repeated | Namespaces to allow. If not set, any namespace is allowed. |
| clusters | [][string](#string) | repeated | Cluster to allow. If not set, any cluster is allowed. |






<a name="core.zephyr.solo.io.IdentitySelector.ServiceAccountRefs"></a>

### IdentitySelector.ServiceAccountRefs



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| serviceAccounts | [][ResourceRef](#core.zephyr.solo.io.ResourceRef) | repeated | List of ServiceAccounts to allow. If not set, any ServiceAccount is allowed. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

