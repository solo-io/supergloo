---
title: Service Dependencies
menuTitle: Service Dependencies
weight: 30
---

{{% notice note %}} Gloo Mesh Enterprise is required for this feature. {{% /notice %}}

## Before you begin
To illustrate these concepts, we will assume that:

* Gloo Mesh is [installed and running on the `mgmt-cluster`]({{% versioned_link_path fromRoot="/setup/#install-gloo-mesh" %}})
* Istio is [installed on both `mgmt-cluster` and `remote-cluster`]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}) clusters
* Both `mgmt-cluster` and `remote-cluster` clusters are [registered with Gloo Mesh]({{% versioned_link_path fromRoot="/guides/#two-registered-clusters" %}})
* The `bookinfo` app is [installed into the two clusters]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployed-on-two-clusters" %}})


{{% notice note %}}
Be sure to review the assumptions and satisfy the pre-requisites from the [Guides]({{% versioned_link_path fromRoot="/guides" %}}) top-level document.
{{% /notice %}}

## Declaring Service Dependencies

#### Motivation

Istio's [default behavior](https://istio.io/latest/docs/reference/config/networking/sidecar/)
is to configure all sidecar proxies in the mesh with all
necessary information required to send traffic to any destination in the mesh. While useful
for day one operations, this comes with a performance tradeoff that scales with the number of workloads and destinations—
proxies may be maintaining unnecessary information in memory, which in aggregate may lead
to a large memory footprint for the entire data plane.

If the operator of a service mesh has a priori knowledge of which destinations a particular workload
needs to reach, this information can be provided to Istio as a way to prune away unneeded information,
thereby alleviating memory consumption of the data plane.

#### ServiceDependency CRD

The [ServiceDependency CRD]({{% versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.service_dependency/" %}}) facilitates management of workload-to-destination dependencies, which translate to
[Istio Sidecar CRs](https://istio.io/latest/docs/reference/config/networking/sidecar/).

Consider the following example:

```yaml
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: ServiceDependency
metadata:
  name: productpage-deps
  namespace: gloo-mesh
spec:
  sourceSelectors:
  - kubeWorkloadMatcher:
      labels:
        app: productpage
      namespaces:
      - bookinfo
  destinationSelectors:
  - kubeServiceRefs:
      services:
      - name: reviews
        namespace: bookinfo
        clusterName: remote-cluster
      - name: reviews
        namespace: bookinfo
        clusterName: mgmt-cluster
```

Semantically, this object declares that Kubernetes workloads with label `app: productpage` in the `bookinfo` namespace, on any cluster,
depend on the Kubernetes service with name `reviews` in namespace `bookinfo` on both the `mgmt-cluster` and `remote-cluster`.

Assuming that the environment has a `productpage-v1` workload deployed to the `mgmt-cluster`,
Gloo Mesh will translate this object into the following Istio Sidecar object on the `mgmt-cluster`:

```yaml
apiVersion: networking.istio.io/v1beta1
kind: Sidecar
metadata:
  name: productpage-v1
  namespace: bookinfo
spec:
  egress:
  - hosts:
    - '*/reviews.bookinfo.svc.cluster.local' # the local reviews service
    - '*/reviews.bookinfo.svc.remote-cluster.soloio' # the federated remote reviews service
  workloadSelector:
    labels:
      app: productpage
      version: v1
```

#### Semantics

1. An important property to note about the ServiceDependency API is that objects are *additive*. 
   If multiple ServiceDependencies target the same workload, that workload's egress dependencies will 
   consist of the set union of all destinations declared on ServiceDependencies.

2. Depending on the underlying mesh (such as Istio), declaring a ServiceDependency for a 
   workload may prevent that workload from sending traffic to any destination other than those 
   explicitly declared. Keep this in mind as you create these objects.
   
3. The ServiceDependency API *does not provide security guarantees*. Even if a destination
   is not declared as a dependency for a given workload, dependending on the behavior of the underyling service mesh,
   that workload might still be able to communicate with that destination.
   [This holds for Istio](https://istio.io/latest/blog/2019/monitoring-external-service-traffic/#what-are-blackhole-and-passthrough-clusters)
   if `global.outboundTrafficPolicy.mode` is set to `ALLOW_ANY`.
