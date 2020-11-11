---
title: Role-based API
menuTitle: Role-based API
weight: 75
---

{{% notice note %}}
Service Mesh Hub Enterprise is required for this feature.
{{% /notice %}}

At a high level, Service Mesh Hub manages control planes by translating networking policy configuration (`networking.smh.solo.io` CRD objects)
into service-mesh-specific configuration. Each networking policy object _targets_ specified discovered mesh entities (represented by `discovery.smh.solo.io` CRD objects).

Service Mesh Hub's role-based API allows organizations to restrict access to policy configuration (i.e. creation, updating, and deletion of policy configuration objects)
based on the roles of individual users, represented by a `Role` CRD. Similar to the Kubernetes RBAC model, Service Mesh Hub users are bound to one or more roles. A user may
create, update, or delete a networking policy if they are bound to at least one role that permits access for that policy.

## Roles

The `Role` CRD structure allows for fine-grained permission definition. Because Service Mesh Hub's networking policies target different discovery entities 
(some combination of workloads, traffic targets, meshes, and virtual meshes), permissions for each networking policy CRD are represented by separate _scopes_.
A bound user is permitted to configure a policy if and only if all actions (if applicable) and scopes present on the policy are permitted by the role. The semantics
for whether actions or scopes are permitted are described below.

## Scopes

Permission scopes are defined against networking policy selectors. These selectors, in the context of networking policies, control the mesh entities that
are affected by the policy. Each networking policy CRD has a different combination of selectors depending on which mesh entities it can affect.
We detail the different scopes and their associated selectors below.

**TrafficPolicyScope:** [TrafficPolicies]({{% versioned_link_path fromRoot="/reference/api/traffic_policy/" %}}) operate on routes between workloads and traffic targets.
Thus, the TrafficPolicyScope defines the set of permitted [WorkloadSelectors]({{% versioned_link_path fromRoot="/reference/api/selectors/#networking.smh.solo.io.WorkloadSelector" %}}) 
and [TrafficTargetSelectors]({{% versioned_link_path fromRoot="/reference/api/selectors/#networking.smh.solo.io.TrafficTargetSelector" %}}). In other words,
a TrafficPolicyScope represents permission for creating TrafficPolicies that configure a specific set of workloads and associated traffic targets.

**AccessPolicyScope:** [AccessPolicies]({{% versioned_link_path fromRoot="/reference/api/access_policy/" %}}) operate on routes between identities (which represent a set of workloads) and traffic targets.
Thus, the AccessPolicyScope defines the set of permitted [IdentitySelectors]({{% versioned_link_path fromRoot="/reference/api/selectors/#networking.smh.solo.io.IdentitySelector" %}}) 
and [TrafficTargetSelectors]({{% versioned_link_path fromRoot="/reference/api/selectors/#networking.smh.solo.io.TrafficTargetSelector" %}}). In other words,
an AccessPolicyScope represents permission for creating AccessPolicies that configure a specific set of identities and associated traffic targets.

**VirtualMeshScope:** [VirtualMeshes]({{% versioned_link_path fromRoot="/reference/api/virtual_mesh/" %}}) operate on discovered meshes.
Thus, the VirtualMeshScope defines the set of permitted meshes by reference (name and namespace of the corresponding Mesh object). 
In other words, a VirtualMeshScope represents permission for creating VirtualMeshes that group a set of discovered meshes.

**FailoverServiceScope:** [FailoverServices]({{% versioned_link_path fromRoot="/reference/api/failover_service/" %}}) operate on discovered meshes and traffic targets.
Thus, the FailoverServiceScope defines the set of permitted meshes by reference (name and namespace of the corresponding Mesh object) and
[backing traffic targets]({{% versioned_link_path fromRoot="/reference/api/failover_service/#networking.smh.solo.io.FailoverServiceSpec.BackingService" %}}).
In other words, a FailoverServiceScope represents permission for creating FailoverServices that creates a failover service visible on a specified
set of meshes, consisting of a set of backing traffic targets.

#### Scope Action Semantics

The set of actions defined on a scope (if applicable to the policy type) describes the set of actions bound users are permitted to configure.
Actions are enums which map onto the correspondingly named policy field (i.e. for TrafficPolicyScope, `TRAFFIC_SHIFT` maps onto the `traffic_shift` field).

#### Scope Selector Semantics:

The set of selectors (WorkloadSelector, TrafficTargetSelector, or IdentitySelector) defined on a scope describes, verbatim, the selectors that the bound users
are permitted to use, **with the important exception of omitted fields**. Omitted/empty selector fields carry wildcard semantics, allowing any value for that field.

The following TrafficTargetSelector allows only the verbatim list of selectors:

```yaml
trafficTargetSelectors:
- kubeServiceMatcher:
    name: foobar
    namespace: foobar-namespace
    clusters:
    - remote-cluster
```

However, if we modify the selector to the following:

```yaml
trafficTargetSelectors:
- kubeServiceMatcher:
    namespace: foobar-namespace
    clusters:
      - remote-cluster
```

It permits TrafficTargetSelectors to select any specific name or to omit the name altogether. The following TrafficTargetSelectors would be permitted
 for a TrafficPolicy:
 
```yaml
destinationSelector:
- kubeServiceMatcher:
    name: foobar
    namespace: foobar-namespace
    clusters:
    - remote-cluster
  kubeServiceMatcher:
    namespace: foobar-namespace
    clusters:
    - remote-cluster
```

## Example Roles / Personas

The following examples demonstrate common personas one might encounter in organizations and how they can be represented using Service Mesh Hub Roles.

**System Administrator**

System administrators are responsible for operating and maintaining infrastructure across the entirety of an organization. They require the broadest possible
permission set.

```yaml
apiVersion: rbac.smh.solo.io/v1alpha1
kind: Role
metadata:
  name: admin-role
  namespace: service-mesh-hub
spec:
  trafficPolicyScopes:
    - trafficPolicyActions:
        - ALL
      trafficTargetSelectors:
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
        - labels:
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
    - trafficTargetSelectors:
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
  failoverServiceScopes:
    - meshRefs:
        - name: "*"
          namespace: "*"
      backingServices:
        - kubeService:
            name: "*"
            namespace: "*"
            clusterName: "*"
```

This role allows complete access for configuring any aspect of any service mesh entity.

**Mesh Owner**

Mesh owners are administrators responsible for operating and maintaining a service mesh instance. They require complete permissions for a set of specified service
meshes.

```yaml
apiVersion: rbac.smh.solo.io/v1alpha1
kind: Role
metadata:
  name: mesh-owner-role
  namespace: service-mesh-hub
spec:
  trafficPolicyScopes:
    - trafficPolicyActions:
        - ALL
      trafficTargetSelectors:
        - kubeServiceMatcher:
            labels:
              "*": "*"
            namespaces:
              - "*"
            clusters:
              - "mgmt-cluster"
        - kubeServiceRefs:
            services:
              - name: "*"
                namespace: "*"
                clusterName: "mgmt-cluster"
      workloadSelectors:
        - labels:
            "*": "*"
          namespaces:
            - "*"
          clusters:
            - "mgmt-cluster"
  # An empty virtualMeshScopes field means that no virtual mesh actions are allowed
  accessPolicyScopes:
    - identitySelectors:
        - kubeIdentityMatcher:
            namespaces:
              - "*"
            clusters:
              - "mgmt-cluster"
          kubeServiceAccountRefs:
            serviceAccounts:
              - name: "*"
                namespace: "*"
                clusterName: "mgmt-cluster"
    - trafficTargetSelectors:
        - kubeServiceMatcher:
            labels:
              "*": "*"
            namespaces:
              - "*"
            clusters:
              - "mgmt-cluster"
        - kubeServiceRefs:
            services:
              - name: "*"
                namespace: "*"
                clusterName: "mgmt-cluster"
  failoverServiceScopes:
    - meshRefs:
        - name: istiod-istio-system-mgmt-cluster
          namespace: service-mesh-hub
      backingServices:
        - kubeService:
            name: "*"
            namespace: "*"
            clusterName: "*"
```

Assuming that the `istiod-istio-system-mgmt-cluster` mesh is the only control plane present on the `mgmt-cluster` cluster, this role allows access 
for configuring TrafficPolicies and AccessPolicies that affect only traffic targets controlled by the 
`istiod-istio-system-mgmt-cluster` mesh, and only FailoverServices that exist on that mesh.

**Traffic Target Publisher**

Traffic target publishers own and operate traffic targets (also referred to as services or microservices) that may be consumed by other workloads
 throughout the organization. They require permissions for configuring server-side networking policies for their traffic targets, regardless of 
 the origin of incoming traffic.
 
```yaml
apiVersion: rbac.smh.solo.io/v1alpha1
kind: Role
metadata:
  name: traffic-target-owner-role
  namespace: service-mesh-hub
spec:
  trafficPolicyScopes:
    - trafficPolicyActions:
        - ALL
      trafficTargetSelectors:
        # The absence of kubeServiceMatcher disallows selecting traffic targets by kubeServiceMatcher
        - kubeServiceRefs:
            services:
              - name: ratings
                namespace: service-mesh-hub
                clusterName: mgmt-cluster
              - name: ratings
                namespace: service-mesh-hub
                clusterName: remote-cluster
      workloadSelectors:
        - labels:
            "*": "*"
          namespaces:
            - "*"
          clusters:
            - "*"
  # An empty virtualMeshScopes field means that no virtual mesh actions are allowed
  # An empty failoverServiceScopes field means that no failover services can be applied by this role bearer
  accessPolicyScopes:
    - identitySelectors:
        - kubeServiceAccountRefs:
            serviceAccounts:
              - name: ratings
                namespace: bookinfo
                clusterName: "*"
    - trafficTargetSelectors:
        # The absence of kubeServiceMatcher disallows selecting traffic targets by kubeServiceMatcher
        - kubeServiceRefs:
            services:
              - name: ratings
                namespace: bookinfo
                clusterName: "*"
```

This role allows configuration of TrafficPolicies and AccessPolicies that affect the `ratings` traffic target in both the `mgmt-cluster` and `remote-cluster`,
with no restrictions on actions.

**Traffic Target Consumer**

Traffic target consumers own and operate workloads that originate requests to a set of traffic targets. They require permissions for configuring
client-side networking policies affecting the route between their workload(s) and the relevant traffic targets.

```yaml
apiVersion: rbac.smh.solo.io/v1alpha1
kind: Role
metadata:
  name: traffic-target-consumer-role
  namespace: service-mesh-hub
spec:
  # A traffic target consumer has the ability to configure policies that affect the network edge between
  # a specific workload and an upstream traffic target.
  trafficPolicyScopes:
    - trafficPolicyActions:
        - RETRIES
        - REQUEST_TIMEOUT
        - FAULT_INJECTION
      trafficTargetSelectors:
        # The absence of kubeServiceMatcher disallows selecting traffic targets by kubeServiceMatcher
        - kubeServiceRefs:
            services:
              - name: ratings
                namespace: bookinfo
                clusterName: "*"
      workloadSelectors:
        - labels:
            app: productpage
            version: v1
          namespaces:
            - bookinfo
          clusters:
            - "*"
  # An empty virtualMeshScopes field means that no virtual mesh actions are allowed
  # An empty failoverServiceScopes field means that no failover services can be applied by this role bearer
  accessPolicyScopes:
    - identitySelectors:
        - kubeServiceAccountRefs:
            serviceAccounts:
              - name: "productpage"
                namespace: "bookinfo"
                clusterName: "*"
    - trafficTargetSelectors:
        # The absence of kubeServiceMatcher disallows selecting traffic targets by kubeServiceMatcher
        - kubeServiceRefs:
            services:
              - name: ratings
                namespace: bookinfo
                clusterName: "*"
```

This role allows configuration of TrafficPolicies (only request timeouts, retries, and fault injections) and AccessPolicies that affect traffic originating
 from the `productpage` workload to the `ratings` traffic target on both `mgmt-cluster` and `remote-cluster`. We assume that the `productpage-service-account` 
 service account identity applies only to the `productpage` workload (and no other workloads). 
