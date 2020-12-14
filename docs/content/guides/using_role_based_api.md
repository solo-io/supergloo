---
title: Using the Role-based API
menuTitle: Using Role-based API
weight: 100
---

{{< notice note >}}
This feature is available in Gloo Mesh Enterprise only. If you are using the open source version of Gloo Mesh, this guide will not work.
{{< /notice >}}

In the role-based API concepts document, we review the functionality of the role-based API and the core components that comprise a role. Now let's actually get some roles deployed and bound to subjects. 

This guide will have you create two example roles and bind them to users.

## Before you begin
To illustrate these concepts, we will assume that:

* Gloo Mesh is [installed and running on the `mgmt-cluster`]({{% versioned_link_path fromRoot="/setup/#install-gloo-mesh" %}})
* Istio is [installed on both `mgmt-cluster` and `remote-cluster`]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}) clusters
* Both `mgmt-cluster` and `remote-cluster` clusters are [registered with Gloo Mesh]({{% versioned_link_path fromRoot="/guides/#two-registered-clusters" %}})
* The `bookinfo` app is [installed into the two clusters]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployed-on-two-clusters" %}})


{{% notice note %}}
Be sure to review the assumptions and satisfy the pre-requisites from the [Guides]({{% versioned_link_path fromRoot="/guides" %}}) top-level document.
{{% /notice %}}

## Role-based API

The role-based API in Gloo Mesh Enterprise uses a `Role` Custom Resource Definition to create Custom Resources that represent roles you would like to define. The roles are then bound to users with a `RoleBinding` CRD. 

The Roles are used to target some combination of *Workloads*, *Traffic Targets*, *Meshes*, and *Virtual Meshes* and define actions the role is allowed to perform on the targets.

When you install Gloo Mesh Enterprise with the default settings, the role-based API is enabled by default. This comes with an implicit **deny** on all operations that are not explicitly allowed by a Role and RoleBinding.

Enforcement of the role-based API is accomplished through the RBAC webhook. If you would like to allow all actions, you can update the RBAC webhook by configuring the following setting in the Helm chart and updating the installation:

```yaml
rbacWebhook:
  env:
    - name: RBAC_PERMISSIVE_MODE
      value: "true"
```

That might be good for testing, but certainly shouldn't be done in a production environment. The alternative is to create and admin role that has permissions to perform all actions, and binding it to admin users who need that level of access.

Let's try and create a network policy on our Gloo Mesh Enterprise deployment without first creating a role and binding.

```shell
MGMT_CONTEXT=<your management plane cluster>
kubectl apply --context $MGMT_CONTEXT -f - << EOF
apiVersion: networking.mesh.gloo.solo.io/v1alpha2
kind: TrafficPolicy
metadata:
  namespace: gloo-mesh
  name: petstore
spec:
  destinationSelector:
  - kubeServiceRefs:
      services:
        - clusterName: mgmt-cluster
          name: petstore
          namespace: default
  requestTimeout: 100ms
  retries:
    attempts: 5
    perTryTimeout: 5ms
EOF
```

```shell
Error from server (User kubernetes-admin does not have the permissions necessary to perform this action.): error when creating "STDIN": admission webhook "rbac-webhook.gloo-mesh.svc" denied the request: User kubernetes-admin does not have the permissions necessary to perform this action.
```

That's precisely what we should expect. Let's start by granting the `kubernetes-admin` user permissions to create a resources.

Let's dig into some example roles starting with the admin role referenced above.

### Admin Role

The `full-admin-role` defined below is granting permissions to perform all actions on all scopes. Obviously this role should be treated with caution.

```yaml
apiVersion: rbac.enterprise.mesh.gloo.solo.io/v1alpha1
kind: Role
metadata:
  name: full-admin-role
  namespace: gloo-mesh
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
      trafficTargetSelectors:
        - kubeServiceMatcher:
            labels:
              "*": "*"
            namespaces:
              - "*"
            clusters:
              - "*"
          kubeServiceRefs:
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
  wasmDeploymentScopes:
    - workloadSelectors:
        - labels:
            "*": "*"
          namespaces:
            - "*"
          clusters:
            - "*"
```

You can save the above role to the file `full-admin-role.yaml` and deploy it with the following command:

```shell
MGMT_CONTEXT=<your management plane cluster>
kubectl --context $MGMT_CONTEXT apply -f full-admin-role.yaml
``` 

Next we will create the RoleBinding for the `kubernetes-admin` user. You may need to change the username depending on the configuration of your management cluster running Gloo Mesh Enterprise.

```yaml
apiVersion: rbac.enterprise.mesh.gloo.solo.io/v1alpha1
kind: RoleBinding
metadata:
  labels:
    app: gloo-mesh
  name: full-admin-role-binding
  namespace: gloo-mesh
spec:
  roleRef:
    name: full-admin-role
    namespace: gloo-mesh
  subjects:
    - kind: User
      name: kubernetes-admin
```

You can save the above binding to the file `full-admin-role-binding.yaml` and deploy it with the following command:

```shell
kubectl --context $MGMT_CONTEXT apply -f full-admin-role-binding.yaml
```

Now if we try to create a Gloo Mesh resource again, the result should be successful:

```shell
kubectl apply --context $MGMT_CONTEXT -f - << EOF
apiVersion: networking.mesh.gloo.solo.io/v1alpha2
kind: TrafficPolicy
metadata:
  namespace: gloo-mesh
  name: petstore
spec:
  destinationSelector:
  - kubeServiceRefs:
      services:
        - clusterName: mgmt-cluster
          name: petstore
          namespace: default
  requestTimeout: 100ms
  retries:
    attempts: 5
    perTryTimeout: 5ms
EOF
```

```shell
trafficpolicy.networking.mesh.gloo.solo.io/petstore created
```

Excellent! Let's take a look at another potential role and try binding it to a different user.

### Traffic Consumer Role

If you've been following the guides, you should already have the bookstore app deployed. The role assignment defined below allows the assignee permissions to operate workloads that originate requests to a set of traffic targets. They are granted permissions for configuring client-side networking policies affecting the route between their workload(s) and the relevant traffic targets.

Specifically, the role allows the TrafficPolicyActions: `RETRIES`, `REQUEST_TIMEOUT`, and `FAULT_INJECTION`. We are also using TrafficTargetSelectors to select the `ratings` service running in the `bookinfo` namespace on any registered cluster. The WorkloadSelectors configuration selects version `v1` of the app `productpage` in the namespace `bookinfo` running on any registered cluster.

The Traffic Consumer Role also creates an AccessPolicyScope defining where access policies could be created, restricted by identity and traffic target. Using the Identity selector, we are restricting identities to the `productpage` service account on the `bookinfo` namespace on any registered cluster. Using the TrafficTargetSelector we are restricting traffic targets to the `ratings` service running in the `bookinfo` namespace on any registered cluster

The end result is that the role allows the management of AccessPolicies for a specific identity and traffic target, and allows a small number of TrafficPolicyActions on a specific service and workload. This type of fine-grained permissions could then be bound to a developer or operator responsible for managing traffic.

```yaml
apiVersion: rbac.enterprise.mesh.gloo.solo.io/v1alpha1
kind: Role
metadata:
  name: traffic-target-consumer-role
  namespace: gloo-mesh
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

You can save the above role to the file `traffic-target-consumer.yaml` and deploy it with the following command:

```shell
kubectl --context $MGMT_CONTEXT apply -f traffic-target-consumer.yaml
```

Now we can bind the role to our junior operator, Gloo.

```yaml
apiVersion: rbac.enterprise.mesh.gloo.solo.io/v1alpha1
kind: RoleBinding
metadata:
  labels:
    app: gloo-mesh
  name: traffic-target-consumer-role-binding
  namespace: gloo-mesh
spec:
  roleRef:
    name: traffic-target-consumer-role
    namespace: gloo-mesh
  subjects:
    - kind: User
      name: Gloo
```

You can save the above binding to the file `traffic-target-consumer-binding.yaml` and deploy it with the following command:

```shell
kubectl --context $MGMT_CONTEXT apply -f traffic-target-consumer-binding.yaml
```

## Summary and Next Steps

In this guide you created and assigned two roles using the Gloo Mesh role-based API. You can read more about the role, and see additional role examples in the [concepts section]({{% versioned_link_path fromRoot="/concepts/role_based_api" %}}).