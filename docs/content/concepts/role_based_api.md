---
title: "Gloo Mesh Role-based API"
menuTitle: Role-based API
description: Understanding Gloo Mesh Enterprise Role-based API
weight: 30
---

{{< notice note >}}
This feature is available in Gloo Mesh Enterprise only.
{{< /notice >}}

At a high level, Gloo Mesh manages control planes by translating networking policy configuration (`networking.mesh.gloo.solo.io` CRD objects)
into service-mesh-specific configuration. Each networking policy object _targets_ specified discovered mesh entities (represented by `discovery.mesh.gloo.solo.io` CRD objects).

Gloo Mesh's role-based API allows organizations to restrict access to policy configuration (i.e. creation, updating, and deletion of policy configuration objects)
based on the roles of individual users, represented by a `Role` CRD. Similar to the Kubernetes RBAC model, Gloo Mesh users are bound to one or more roles. A user may
create, update, or delete a networking policy if they are bound to at least one role that permits access for that policy.

## Roles

The `Role` CRD structure allows for fine-grained permission definition. Because Gloo Mesh's networking policies target different discovery entities 
(some combination of workloads, Destinations, meshes, and virtual meshes), permissions for each networking policy CRD are represented by separate _scopes_.
A bound user is permitted to configure a policy if and only if all actions (if applicable) and scopes present on the policy are permitted by the role. The semantics
for whether actions or scopes are permitted are described below.

## Scopes

Permission scopes are defined against networking policy selectors. These selectors, in the context of networking policies, control the mesh entities that
are affected by the policy. Each networking policy CRD has a different combination of selectors depending on which mesh entities it can affect.
We detail the different scopes and their associated selectors below.

**TrafficPolicyScope:** [TrafficPolicies]({{% versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy/" %}}) operate on routes between workloads and Destinations.
Thus, the TrafficPolicyScope defines the set of permitted [WorkloadSelectors]({{% versioned_link_path fromRoot="/reference/api/selectors/#networking.mesh.gloo.solo.io.WorkloadSelector" %}}) 
and [DestinationSelectors]({{% versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors.md#common.mesh.gloo.solo.io.DestinationSelector" %}}). In other words,
a TrafficPolicyScope represents permission for creating TrafficPolicies that configure a specific set of workloads and associated Destinations.

**AccessPolicyScope:** [AccessPolicies]({{% versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.access_policy/" %}}) operate on routes between identities (which represent a set of workloads) and Destinations.
Thus, the AccessPolicyScope defines the set of permitted [IdentitySelectors]({{% versioned_link_path fromRoot="/reference/api/selectors/#networking.mesh.gloo.solo.io.IdentitySelector" %}}) 
and [DestinationSelectors]({{% versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.selectors.md#common.mesh.gloo.solo.io.DestinationSelector" %}}). In other words,
an AccessPolicyScope represents permission for creating AccessPolicies that configure a specific set of identities and associated Destinations.

**VirtualMeshScope:** [VirtualMeshes]({{% versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.virtual_mesh/" %}}) operate on discovered meshes.
Thus, the VirtualMeshScope defines the set of permitted meshes by reference (name and namespace of the corresponding Mesh object). 
In other words, a VirtualMeshScope represents permission for creating VirtualMeshes that group a set of discovered meshes.

#### Scope Action Semantics

The set of actions defined on a scope (if applicable to the policy type) describes the set of actions bound users are permitted to configure.
Actions are enums which map onto the correspondingly named policy field (i.e. for TrafficPolicyScope, `TRAFFIC_SHIFT` maps onto the `traffic_shift` field).

#### Scope Selector Semantics:

The set of selectors (WorkloadSelector, DestinationSelector, or IdentitySelector) defined on a scope describes, verbatim, the selectors that the bound users
are permitted to use, **with the important exception of omitted fields**. Omitted/empty selector fields carry wildcard semantics, allowing any value for that field.

The following DestinationSelector allows only the verbatim list of selectors:

```yaml
destinationSelectors:
- kubeServiceMatcher:
    name: foobar
    namespace: foobar-namespace
    clusters:
    - cluster-2
```

However, if we modify the selector to the following:

```yaml
destinationSelectors:
- kubeServiceMatcher:
    namespace: foobar-namespace
    clusters:
      - cluster-2
```

It permits DestinationSelectors to select any specific name or to omit the name altogether. The following DestinationSelectors would be permitted
 for a TrafficPolicy:
 
```yaml
destinationSelector:
- kubeServiceMatcher:
    name: foobar
    namespace: foobar-namespace
    clusters:
    - cluster-2
  kubeServiceMatcher:
    namespace: foobar-namespace
    clusters:
    - cluster-2
```

## Example Roles / Personas

The following examples demonstrate common personas one might encounter in organizations and how they can be represented using Gloo Mesh Roles.

**System Administrator**

System administrators are responsible for operating and maintaining infrastructure across the entirety of an organization. They require the broadest possible
permission set.

```yaml
apiVersion: rbac.enterprise.mesh.gloo.solo.io/v1
kind: Role
metadata:
  name: admin-role
  namespace: gloo-mesh
spec:
  trafficPolicyScopes:
    - trafficPolicyActions:
        - ALL
      destinationSelectors:
        - kubeServiceMatcher:
            labels:
              "*": "*"
            namespaces:
              - "*"
            clusters:
              - "*"
        - kubeServiceRefs:
            services:
              - name: "*"
                namespace: "*"
                clusterName: "*"
      workloadSelectors:
        - kubeWorkloadMatcher:
            labels:
              "*": "*"
            namespaces:
              - "*"
            clusters:
              - "*"
  virtualMeshScopes:
    - virtualMeshActions:
        - ALL
      meshRefs:
        - name: "*"
          namespace: "*"
  accessPolicyScopes:
    - identitySelectors:
        - kubeIdentityMatcher:
            namespaces:
              - "*"
            clusters:
              - "*"
          kubeServiceAccountRefs:
            serviceAccounts:
              - name: "*"
                namespace: "*"
                clusterName: "*"
    - destinationSelectors:
        - kubeServiceMatcher:
            labels:
              "*": "*"
            namespaces:
              - "*"
            clusters:
              - "*"
        - kubeServiceRefs:
            services:
              - name: "*"
                namespace: "*"
                clusterName: "*"
```

This role allows complete access for configuring any aspect of any service mesh entity.

**Mesh Owner**

Mesh owners are administrators responsible for operating and maintaining a service mesh instance. They require complete permissions for a set of specified service
meshes.

```yaml
apiVersion: rbac.enterprise.mesh.gloo.solo.io/v1
kind: Role
metadata:
  name: mesh-owner-role
  namespace: gloo-mesh
spec:
  trafficPolicyScopes:
    - trafficPolicyActions:
        - ALL
      destinationSelectors:
        - kubeServiceMatcher:
            labels:
              "*": "*"
            namespaces:
              - "*"
            clusters:
              - "cluster-1"
        - kubeServiceRefs:
            services:
              - name: "*"
                namespace: "*"
                clusterName: "cluster-1"
      workloadSelectors:
        - kubeWorkloadMatcher:
            labels:
              "*": "*"
            namespaces:
              - "*"
            clusters:
              - "cluster-1"
  # An empty virtualMeshScopes field means that no virtual mesh actions are allowed
  accessPolicyScopes:
    - identitySelectors:
        - kubeIdentityMatcher:
            namespaces:
              - "*"
            clusters:
              - "cluster-1"
          kubeServiceAccountRefs:
            serviceAccounts:
              - name: "*"
                namespace: "*"
                clusterName: "cluster-1"
    - destinationSelectors:
        - kubeServiceMatcher:
            labels:
              "*": "*"
            namespaces:
              - "*"
            clusters:
              - "cluster-1"
        - kubeServiceRefs:
            services:
              - name: "*"
                namespace: "*"
                clusterName: "cluster-1"
```

Assuming that the `istiod-istio-system-cluster-1` mesh is the only control plane present on the `cluster-1` cluster, this role allows access 
for configuring TrafficPolicies and AccessPolicies that affect only Destinations controlled by the 
`istiod-istio-system-cluster-1` mesh.

**Destination Publisher**

Destination publishers own and operate Destinations (also referred to as services or microservices) that may be consumed by other workloads
 throughout the organization. They require permissions for configuring server-side networking policies for their Destinations, regardless of 
 the origin of incoming traffic.
 
```yaml
apiVersion: rbac.enterprise.mesh.gloo.solo.io/v1
kind: Role
metadata:
  name: destination-owner-role
  namespace: gloo-mesh
spec:
  trafficPolicyScopes:
    - trafficPolicyActions:
        - ALL
      destinationSelectors:
        # The absence of kubeServiceMatcher disallows selecting Destinations by kubeServiceMatcher
        - kubeServiceRefs:
            services:
              - name: ratings
                namespace: gloo-mesh
                clusterName: cluster-1
              - name: ratings
                namespace: gloo-mesh
                clusterName: cluster-2
      workloadSelectors:
        - kubeWorkloadMatcher:
            labels:
              "*": "*"
            namespaces:
              - "*"
            clusters:
              - "*"
  # An empty virtualMeshScopes field means that no virtual mesh actions are allowed
  accessPolicyScopes:
    - identitySelectors:
        - kubeServiceAccountRefs:
            serviceAccounts:
              - name: ratings
                namespace: bookinfo
                clusterName: "*"
    - destinationSelectors:
        # The absence of kubeServiceMatcher disallows selecting Destinations by kubeServiceMatcher
        - kubeServiceRefs:
            services:
              - name: ratings
                namespace: bookinfo
                clusterName: "*"
```

This role allows configuration of TrafficPolicies and AccessPolicies that affect the `ratings` Destination in both the `cluster-1` and `cluster-2`,
with no restrictions on actions.

**Destination Consumer**

Destination consumers own and operate workloads that originate requests to a set of Destinations. They require permissions for configuring
client-side networking policies affecting the route between their workload(s) and the relevant Destinations.

```yaml
apiVersion: rbac.enterprise.mesh.gloo.solo.io/v1
kind: Role
metadata:
  name: destination-consumer-role
  namespace: gloo-mesh
spec:
  # A Destination consumer has the ability to configure policies that affect the network edge between
  # a specific workload and an upstream Destination.
  trafficPolicyScopes:
    - trafficPolicyActions:
        - RETRIES
        - REQUEST_TIMEOUT
        - FAULT_INJECTION
      destinationSelectors:
        # The absence of kubeServiceMatcher disallows selecting Destinations by kubeServiceMatcher
        - kubeServiceRefs:
            services:
              - name: ratings
                namespace: bookinfo
                clusterName: "*"
      workloadSelectors:
        - kubeWorkloadMatcher:
            labels:
              app: productpage
              version: v1
            namespaces:
              - bookinfo
            clusters:
              - "*"
  # An empty virtualMeshScopes field means that no virtual mesh actions are allowed
  accessPolicyScopes:
    - identitySelectors:
        - kubeServiceAccountRefs:
            serviceAccounts:
              - name: "productpage"
                namespace: "bookinfo"
                clusterName: "*"
    - destinationSelectors:
        # The absence of kubeServiceMatcher disallows selecting Destinations by kubeServiceMatcher
        - kubeServiceRefs:
            services:
              - name: ratings
                namespace: bookinfo
                clusterName: "*"
```

This role allows configuration of TrafficPolicies (only request timeouts, retries, and fault injections) and AccessPolicies that affect traffic originating
 from the `productpage` workload to the `ratings` Destination on both `cluster-1` and `cluster-2`. We assume that the `productpage-service-account` 
 service account identity applies only to the `productpage` workload (and no other workloads). 

## Next Steps

To test drive the Role-based API for yourself, we recommend [following our guide]({{% versioned_link_path fromRoot="/guides/using_role_based_api" %}}) on the topic.
