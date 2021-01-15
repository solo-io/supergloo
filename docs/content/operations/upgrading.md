---
title: Upgrading Gloo Mesh
menuTitle: Upgrading Gloo Mesh
weight: 100
description: Best practices for upgrading Gloo Mesh
---

Gloo Mesh is pre-1.0 and under active development. There may be breaking API changes with every minor version bump.
An "API break" is defined as any change in the structure or behavior of a Gloo Mesh CRD that requires user intervention
at upgrade time to avoid putting Gloo Mesh into a persistent error state.

The following steps describe a general approach to navigating the Gloo Mesh upgrade process for versions of Gloo Mesh
< 1.0.0. This process involves scaling down the Gloo Mesh control plane while various components and CRDs are updated,
but should not introduce downtime in service-to-service communication or policy enforcement because Gloo Mesh does not
operate in the request path and the underlying service meshes (e.g. Istio) are not impacted by this process. That said,
while Gloo Mesh is in its initial development ahead of general availability (i.e. pre-1.0.0), there is **no guarantee**
that these steps will provide a comprehensive, zero-downtime upgrade experience. To reiterate:

{{< notice warning >}}
This upgrade process is **not guaranteed** to provide a comprehensive, zero-downtime upgrade experience.

While Gloo Mesh is under initial development and in a pre-1.0.0 state, the recommended way to access new functionality
is via a fresh install of Gloo Mesh.
{{< /notice >}}

1\. Following the steps outlined by the [Getting Started Guide]({{% versioned_link_path fromRoot="/getting_started/" %}}),
download the version of `meshctl` corresponding to the version of Gloo Mesh to which you would like to upgrade.

2\. Scale down both the `networking` and `discovery` deployments on the Gloo Mesh management cluster to zero replicas.
This will prevent each component from processing resources structured in a way not expected by the new versions of each
component.

```shell
kubectl scale deployment -n gloo-mesh networking --replicas 0
```

```shell
kubectl scale deployment -n gloo-mesh discovery --replicas 0
```

3\. Delete all resources in the `discovery.gloo.mesh.gloo.solo.io` API group such as `meshes`, `traffictargets`, and `workloads`.
These resources will be recreated when `discovery` is scaled back up later in the upgrade process. Deleting these
resources ensures that the latest discovery implementation will recreate them with the structure it expects. 

4\. Manually update the Gloo Mesh management plane CRDs in the management cluster. These include both `discovery` and
`networking` resources. Because of the way [Helm handles CRDs at upgrade time](https://helm.sh/docs/chart_best_practices/custom_resource_definitions/#some-caveats-and-explanations),
this step is required to ensure the management plane CRDs have the validation schemas expected by the version of Gloo
Mesh you are upgrading to. This will prevent errors in the management plane and ensure that configuration leveraging
the latest APIs on each Gloo Mesh are not rejected.

Gloo Mesh CRDs can be updated by extracting the contents of the Helm chart, for example [https://storage.googleapis.com/gloo-mesh/gloo-mesh/gloo-mesh-0.11.3] (https://storage.googleapis.com/gloo-mesh/gloo-mesh/gloo-mesh-0.11.3)
where `0.11.3` is replaced by the version of Gloo Mesh to which you would like to upgrade. Alternatively, CRDs can be found
on GitHub at [https://github.com/solo-io/gloo-mesh/tree/main/install/helm/gloo-mesh/crds](https://github.com/solo-io/gloo-mesh/tree/main/install/helm/gloo-mesh/crds).
From there, you can either download the repository and apply the CRDs at the tag of your choosing, or apply the CRDs
directly from GitHub with:

```shell
# Add your desired version here
UPGRADE_VERSION=v0.11.3

kubectl apply -f https://raw.githubusercontent.com/solo-io/gloo-mesh/$UPGRADE_VERSION/install/helm/gloo-mesh/crds/discovery.mesh.gloo.solo.io_v1alpha2_crds.yaml
kubectl apply -f https://raw.githubusercontent.com/solo-io/gloo-mesh/$UPGRADE_VERSION/install/helm/gloo-mesh/crds/multicluster.solo.io_v1alpha1_crds.yaml
kubectl apply -f https://raw.githubusercontent.com/solo-io/gloo-mesh/$UPGRADE_VERSION/install/helm/gloo-mesh/crds/networking.enterprise.mesh.gloo.solo.io_v1alpha1_crds.yaml
kubectl apply -f https://raw.githubusercontent.com/solo-io/gloo-mesh/$UPGRADE_VERSION/install/helm/gloo-mesh/crds/networking.mesh.gloo.solo.io_v1alpha2_crds.yaml
kubectl apply -f https://raw.githubusercontent.com/solo-io/gloo-mesh/$UPGRADE_VERSION/install/helm/gloo-mesh/crds/rbac.enterprise.mesh.gloo.solo.io_v1alpha1_crds.yaml
kubectl apply -f https://raw.githubusercontent.com/solo-io/gloo-mesh/$UPGRADE_VERSION/install/helm/gloo-mesh/crds/settings.mesh.gloo.solo.io_v1alpha2_crds.yaml
```

5\. Ensure that all `networking` resources such as `trafficpolicies`, `virtualmeshes`, `failoverservices`, or `accesspolicies`
conform to the latest APIs. Breaking changes will be highlighted in the [changelog]({{% versioned_link_path fromRoot="/reference/changelog" %}}).

6\. Upgrade the Gloo Mesh management plane components using either Helm or `meshctl`, following the steps described in
the [Gloo Mesh installation guide]({{% versioned_link_path fromRoot="/setup/install_gloo_mesh" %}}). Note that you
should use `helm upgrade` rather than `helm install` if using Helm, and that `meshctl install` can be used for both
install and upgrade operations.

7\. Re-register all clusters registered to the Gloo Mesh management plane. This will ensure that all remote cluster agents
and CRDs are updated to a version compatible with the version of Gloo Mesh you are upgrading to. Refer to the 
[setup guide on registering a cluster]({{% versioned_link_path fromRoot="/setup/register_cluster" %}})
and be sure to use the same cluster names and contexts that were used at the initial cluster registration time.

8\. Scale the `discovery` deployment to one replica, and wait for all `discovery` resources such as `meshes`, `traffictargets`,
and `workloads` to be written to the management cluster. This may take a few minutes, and will ensure that the `networking`
component has access to all the data it needs to continue processing user-provided network configuration. Discovery is
complete when the pod no longer outputs a steady stream of logs or when all expected resources can be found on the cluster.

```shell
kubectl scale deployment -n gloo-mesh discovery --replicas 1
```

9\. Scale the `networking` deployment to one replica. Errors may be propagated to various `networking` resources such as
`trafficpolicies`, `virtualmeshes`, `failoverservices`, and `accesspolicies` as it starts watches on remote clusters,
but it will reach a steady state after a few moments.

```shell
kubectl scale deployment -n gloo-mesh networking --replicas 1
```

10\. Run `meshctl check` to verify that all resources are in a healthy state. If not, check the logs on the `discovery`
and `networking` pods as well as the `status` fields on unhealthy resources to begin debugging. Refer to our 
[Troubleshooting Guide]({{% versioned_link_path fromRoot="/operations/troubleshooting" %}}) for more details.
