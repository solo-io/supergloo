---
title: Using the Role-based API
menuTitle: Using Role-based API
weight: 100
---

{{< notice note >}}
This feature is available in Gloo Mesh Enterprise only. If you are using the open source version of Gloo Mesh, this guide will not work.
{{< /notice >}}

In the role-based API concepts document, we review the functionality of the role-based API and the core components that comprise a role. Users, Groups, and Service Accounts are supported as role binding subjects. Now let's actually get some roles deployed and bound to subjects. 

This guide will have you create two example roles and bind them to users.

## Before you begin
To illustrate these concepts, we will assume that:

* There are two clusters managed by Gloo Mesh named `cluster-1` and `cluster-2`. 
* Gloo Mesh is [installed and running on `cluster-1`]({{% versioned_link_path fromRoot="/setup/#install-gloo-mesh" %}})
* Istio is [installed on both client clusters]({{% versioned_link_path fromRoot="/guides/installing_istio" %}})
* The `bookinfo` app is [installed across the two clusters]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployed-on-two-clusters" %}})


{{% notice note %}}
Be sure to review the assumptions and satisfy the pre-requisites from the [Guides]({{% versioned_link_path fromRoot="/guides" %}}) top-level document.
{{% /notice %}}

Ensure you have the correct context names set in your environment:

```shell
MGMT_CONTEXT=your_management_plane_context (in this case cluster-1's context)
```

## Role-based API

The role-based API in Gloo Mesh Enterprise uses a `Role` Custom Resource Definition to create Custom Resources that represent roles you would like to define. The roles are then bound to users with a `RoleBinding` CRD. The Roles can be accessed via the `gmrole` alias and the RoleBindings can be accessed via the `gmrolebinding` alias.

The Roles are used to target some combination of *Workloads*, *Destinations*, *Meshes*, and *Virtual Meshes* and define actions the role is allowed to perform on the targets.

When you install Gloo Mesh Enterprise with the default settings, the role-based API is disabled by default.
If you enable it, there will be an implicit **deny** on all operations that are not explicitly allowed by a Role and RoleBinding.

Enforcement of the role-based API is accomplished through the RBAC webhook. If you would like to allow all actions, you can update the RBAC webhook by configuring the following setting in the Helm chart and updating the installation:

```yaml
rbacWebhook:
  enabled: true
  env:
    - name: RBAC_PERMISSIVE_MODE
      value: "true"
```

The permissive mode might be good for testing, but certainly shouldn't be done in a production environment. The alternative is to create and admin role that has permissions to perform all actions, and binding it to admin users who need that level of access.

Let's try and create a network policy on our Gloo Mesh Enterprise deployment without first creating a role and binding.

{{< tabs >}}
{{< tab name="YAML file" codelang="yaml">}}
apiVersion: networking.mesh.gloo.solo.io/v1
kind: TrafficPolicy
metadata:
  namespace: gloo-mesh
  name: petstore
spec:
  destinationSelector:
  - kubeServiceRefs:
      services:
        - clusterName: cluster-1
          name: petstore
          namespace: default
  policy:
    requestTimeout: 100ms
    retries:
      attempts: 5
      perTryTimeout: 5ms
{{< /tab >}}
{{< tab name="CLI inline" codelang="shell" >}}
kubectl apply --context $MGMT_CONTEXT -f - << EOF
apiVersion: networking.mesh.gloo.solo.io/v1
kind: TrafficPolicy
metadata:
  namespace: gloo-mesh
  name: petstore
spec:
  destinationSelector:
  - kubeServiceRefs:
      services:
        - clusterName: cluster-1
          name: petstore
          namespace: default
  policy:
    requestTimeout: 100ms
    retries:
      attempts: 5
      perTryTimeout: 5ms
EOF
{{< /tab >}}
{{< /tabs >}}

```shell
Error from server (User kubernetes-admin does not have the permissions necessary to perform this action.): error when creating "STDIN": admission webhook "rbac-webhook.gloo-mesh.svc" denied the request: User kubernetes-admin does not have the permissions necessary to perform this action.
```

That's precisely what we should expect. Let's start by granting the `kubernetes-admin` user permissions to create a resources.

Let's dig into some example roles starting with the admin role referenced above.

### Admin Role

Gloo Mesh's RBAC webhook component ships with an admin role that grants permissions to perform
all actions on any object. Run the following command to view the admin role object:

```shell
# add the Helm repo containing the rbac-webhook
helm repo add rbac-webhook https://storage.googleapis.com/gloo-mesh-enterprise/rbac-webhook

# show the admin-role object
helm template rbac-webhook/rbac-webhook -s templates/admin-role.yaml
```

Kubernetes Users and Groups can assume this role at install time by 
specifying the `adminSubjects` Helm value. You can view this Helm value with the following commands:

```shell
# show the available Helm values
helm show values rbac-webhook/rbac-webhook
```

You should see the following stanza whose values you can override:

```yaml
adminSubjects:
- kind: User
  name: kubernetes-admin
createAdminRole: true
```

Assuming that you've granted yourself admin permissions, if we try to create a Gloo Mesh resource again,
the result should be successful:

{{< tabs >}}
{{< tab name="YAML file" codelang="yaml">}}
apiVersion: networking.mesh.gloo.solo.io/v1
kind: TrafficPolicy
metadata:
  namespace: gloo-mesh
  name: petstore
spec:
  destinationSelector:
  - kubeServiceRefs:
      services:
        - clusterName: cluster-1
          name: petstore
          namespace: default
  policy:
    requestTimeout: 100ms
    retries:
      attempts: 5
      perTryTimeout: 5ms
{{< /tab >}}
{{< tab name="CLI inline" codelang="shell" >}}
kubectl apply --context $MGMT_CONTEXT -f - << EOF
apiVersion: networking.mesh.gloo.solo.io/v1
kind: TrafficPolicy
metadata:
  namespace: gloo-mesh
  name: petstore
spec:
  destinationSelector:
  - kubeServiceRefs:
      services:
        - clusterName: cluster-1
          name: petstore
          namespace: default
  policy:
    requestTimeout: 100ms
    retries:
      attempts: 5
      perTryTimeout: 5ms
EOF
{{< /tab >}}
{{< /tabs >}}

```shell
trafficpolicy.networking.mesh.gloo.solo.io/petstore created
```

Excellent! Let's take a look at another potential role and try binding it to a different user.

### Destination Consumer Role

If you've been following the guides, you should already have the bookstore app deployed. The role assignment defined below allows the assignee permissions to operate workloads that originate requests to a set of Destinations. They are granted permissions for configuring client-side networking policies affecting the route between their workload(s) and the relevant Destinations.

Specifically, the role allows the TrafficPolicyActions: `RETRIES`, `REQUEST_TIMEOUT`, and `FAULT_INJECTION`. We are also using DestinationSelectors to select the `ratings` service running in the `bookinfo` namespace on any registered cluster. The WorkloadSelectors configuration selects version `v1` of the app `productpage` in the namespace `bookinfo` running on any registered cluster.

The Destination Consumer Role also creates an AccessPolicyScope defining where access policies could be created, restricted by identity and Destination. Using the Identity selector, we are restricting identities to the `productpage` service account on the `bookinfo` namespace on any registered cluster. Using the DestinationSelector we are restricting Destinations to the `ratings` service running in the `bookinfo` namespace on any registered cluster

The end result is that the role allows the management of AccessPolicies for a specific identity and Destination, and allows a small number of TrafficPolicyActions on a specific service and workload. This type of fine-grained permissions could then be bound to a developer or operator responsible for managing traffic.

You can deploy the following role:

{{< tabs >}}
{{< tab name="YAML file" codelang="yaml">}}
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
  # An empty failoverServiceScopes field means that no failover services can be applied by this role bearer
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
{{< /tab >}}
{{< tab name="CLI inline" codelang="shell" >}}
kubectl apply --context $MGMT_CONTEXT -f - << EOF
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
  # An empty failoverServiceScopes field means that no failover services can be applied by this role bearer
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
EOF
{{< /tab >}}
{{< /tabs >}}

Now we can bind the role to our junior operator, Gloo.

You can now apply the following binding:

{{< tabs >}}
{{< tab name="YAML file" codelang="yaml">}}
apiVersion: rbac.enterprise.mesh.gloo.solo.io/v1
kind: RoleBinding
metadata:
  labels:
    app: gloo-mesh
  name: destination-consumer-role-binding
  namespace: gloo-mesh
spec:
  roleRef:
    name: destination-consumer-role
    namespace: gloo-mesh
  subjects:
    - kind: User
      name: Gloo
{{< /tab >}}
{{< tab name="CLI inline" codelang="shell" >}}
kubectl apply --context $MGMT_CONTEXT -f - << EOF
apiVersion: rbac.enterprise.mesh.gloo.solo.io/v1
kind: RoleBinding
metadata:
  labels:
    app: gloo-mesh
  name: destination-consumer-role-binding
  namespace: gloo-mesh
spec:
  roleRef:
    name: destination-consumer-role
    namespace: gloo-mesh
  subjects:
    - kind: User
      name: Gloo
EOF
{{< /tab >}}
{{< /tabs >}}

## Summary and Next Steps

In this guide you created and assigned two roles using the Gloo Mesh role-based API. You can read more about the role, and see additional role examples in the [concepts section]({{% versioned_link_path fromRoot="/concepts/role_based_api" %}}).
