---
title: Mesh Discovery
menuTitle: Mesh Discovery
weight: 20
---

Gloo Mesh can automatically discover service mesh installations on registered clusters using control plane and sidecar discovery, as well as workloads and services exposed through the service mesh.

In this guide we will learn about the four main discovery capabilities in the context of Kubernetes as the compute platform:

1. **Kubernetes Clusters**
    - Representation of a cluster that Gloo Mesh is aware of and is authorized to talk to its Kubernetes API server
    - *note*: this resource is created by `meshctl` at cluster registration time
2. **Meshes**
    - Representation of a service mesh control plane that has been discovered 
3. **Workloads**
    - Representation of a pod that is a member of a service mesh; this is often determined by the presence of an injected proxy sidecar
4. **Destinations**
    - Representation of a Kubernetes service that is backed by Workload pods, e.g. pods that are a member of the service mesh


## Before you begin
To illustrate these concepts, we will assume that:

* There are two clusters managed by Gloo Mesh named `cluster-1` and `cluster-2`. 
* Istio is [installed on both client clusters]({{% versioned_link_path fromRoot="/guides/installing_istio" %}})
* The `bookinfo` app is [installed across the two clusters]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployed-on-two-clusters" %}})


{{% notice note %}}
Be sure to review the assumptions and satisfy the pre-requisites from the [Guides]({{% versioned_link_path fromRoot="/guides" %}}) top-level document.
{{% /notice %}}

### Discover Kubernetes Clusters

Ensure that your `kubeconfig` has the correct context set as its `currentContext`:

```shell
CONTEXT_1=your_first_context
CONTEXT_2=your_second_context
kubectl config use-context $CONTEXT_1
```

Validate that the cluster have been registered by checking for `KubernetesClusters` custom resources:

```shell
kubectl get kubernetesclusters -n gloo-mesh
```

```shell
NAME                 AGE
cluster-1            23h
cluster-2            23h
```

### Discover Meshes

Check to see that Istio has been discovered:

```shell
meshctl describe mesh
```

```
+-----------------------------+----------------+-------------------+
|           METADATA          | VIRTUAL MESHES | FAILOVER SERVICES |
+-----------------------------+----------------+-------------------+
| Namespace: istio-system     |                |                   |
| Cluster: cluster-1          |                |                   |
| Type: istio                 |                |                   |
| Version: 1.8.1              |                |                   |
|                             |                |                   |
+-----------------------------+----------------+-------------------+
| Namespace: istio-system     |                |                   |
| Cluster: cluster-2          |                |                   |
| Type: istio                 |                |                   |
| Version: 1.8.1              |                |                   |
|                             |                |                   |
+-----------------------------+----------------+-------------------+
```

We can print it in YAML form to see all the information we discovered:

```shell
kubectl -n gloo-mesh get mesh istiod-istio-system-cluster-2 -oyaml
```

(snipped for brevity)

{{< highlight yaml >}}
apiVersion: discovery.mesh.gloo.solo.io/v1
kind: Mesh
metadata:
  annotations:
    deployment.kubernetes.io/revision: "1"
    [...]
  generation: 2
  labels:
    cluster.discovery.mesh.gloo.solo.io: cluster-2
    cluster.multicluster.solo.io: ""
    owner.discovery.mesh.gloo.solo.io: gloo-mesh
  name: istiod-istio-system-cluster-2
  namespace: gloo-mesh
  resourceVersion: "3218"
  selfLink: /apis/discovery.mesh.gloo.solo.io/v1/namespaces/gloo-mesh/meshes/istiod-istio-system-cluster-2
  uid: 7c079983-3ece-4aed-b71a-bf56c8cd6267
spec:
  agentInfo:
    agentNamespace: gloo-mesh
  istio:
    citadelInfo:
      IstiodServiceAccount: istiod-service-account
      trustDomain: cluster.local
    ingressGateways:
    - externalAddress: 172.20.0.3
      externalTlsPort: 32000
      tlsContainerPort: 15443
      workloadLabels:
        istio: ingressgateway
    installation:
      cluster: cluster-2
      namespace: istio-system
      podLabels:
        istio: pilot
      version: 1.8.1
status:
  observedGeneration: 2

{{< /highlight >}}

### Discover Workloads

Check to see that the `bookinfo` pods have been correctly identified as Workloads:

```shell
kubectl -n gloo-mesh get workloads
```

```
NAME                                                            AGE
details-v1-bookinfo-cluster-1-deployment                        3m54s
istio-ingressgateway-istio-system-cluster-1-deployment          23h
istio-ingressgateway-istio-system-cluster-2-deployment          23h
productpage-v1-bookinfo-cluster-1-deployment                    3m54s
ratings-v1-bookinfo-cluster-1-deployment                        3m53s
ratings-v1-bookinfo-cluster-2-deployment                        3m25s
reviews-v1-bookinfo-cluster-1-deployment                        3m53s
reviews-v2-bookinfo-cluster-1-deployment                        3m53s
reviews-v3-bookinfo-cluster-2-deployment                        2m
```

### Discover Destinations

Similarly for the `bookinfo` services:

```shell
kubectl -n gloo-mesh get destinations
```

```
NAME                                                 AGE
details-bookinfo-cluster-1                           4m23s
istio-ingressgateway-istio-system-cluster-1          23h
istio-ingressgateway-istio-system-cluster-2          23h
productpage-bookinfo-cluster-1                       4m23s
ratings-bookinfo-cluster-1                           4m22s
ratings-bookinfo-cluster-2                           3m54s
reviews-bookinfo-cluster-1                           4m22s
reviews-bookinfo-cluster-2                           2m29s
```

## See it in action

Check out "Part One" of the ["Dive into Gloo Mesh" video series](https://www.youtube.com/watch?v=4sWikVELr5M&list=PLBOtlFtGznBjr4E9xYHH9eVyiOwnk1ciK)
(note that the video content reflects Gloo Mesh <b>v0.6.1</b>):

<iframe width="560" height="315" src="https://www.youtube.com/embed/4sWikVELr5M" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## Next Steps

Now that we have Istio installed, and we've seen Gloo Mesh discover the meshes across different clusters, we can now unify them into a single [Virtual Mesh]({{% versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.virtual_mesh/" %}}). See the guide on [establishing shared trust domain for multiple meshes]({{% versioned_link_path fromRoot="/guides/federate_identity" %}}).
