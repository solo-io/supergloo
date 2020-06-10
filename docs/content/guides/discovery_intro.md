---
title: Mesh Discovery
menuTitle: Mesh Discovery
weight: 20
---

Service Mesh Hub, when appropriate clusters are registered, can automatically discover service mesh installations (based on control plane and sidecar discovery), as well as 

In this guide we will learn about the four main discovery capabilities:

1. **Kubernetes Clusters**
    - Representation of a cluster that Service Mesh Hub is aware of and is authorized to
talk to its Kubernetes API server
    - *note*: this resource is created by `meshctl` at cluster registration time
2. **Meshes**
    - Representation of a service mesh control plane that has been discovered 
3. **Mesh Workloads**
    - Representation of a pod that is a member of a service mesh; this is often determined by the presence of
an injected proxy sidecar
4. **Mesh Services**
    - Representation of a Kubernetes service that is backed by Mesh Workload pods, e.g.
pods that are a member of the service mesh


## Before you begin
To illustrate these concepts, we will assume that:

* Service Mesh Hub is [installed and running on the `management-plane-context`]({{% versioned_link_path fromRoot="/setup/#install-service-mesh-hub" %}})
* Istio [installed on both `management-plane-context` and `remote-cluster-context`]({{% versioned_link_path fromRoot="/guides/installing_istio" %}})
* Both `management-plane-context` and `remote-cluster-context` clusters [registered with Service Mesh Hub]({{% versioned_link_path fromRoot="/guides/#two-registered-clusters" %}})
* The `bookinfo` app [installed into two Istio cluster]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployed-on-two-clusters" %}})


{{% notice note %}}
Be sure to review the assumptions and satisfy the pre-requisites from the [Guides]({{% versioned_link_path fromRoot="/guides" %}}) top-level document.
{{% /notice %}}

### Discover Kubernetes Clusters

Ensure that your `kubeconfig` has the correct context set as its `currentContext`:

```shell
kubectl config use-context management-plane-context
```

Check to see that we wrote config when we registered our remote cluster:

```shell
meshctl get kubernetescluster
```

```shell
Kubernetes Clusters:
+--------------------+------------------+-------------------------------------+
|        NAME        | WRITE NAMESPACE  |             SECRET REF              |
+--------------------+------------------+-------------------------------------+
| management-plane   | service-mesh-hub | management-plane.service-mesh-hub   |
+--------------------+------------------+-------------------------------------+
| new-remote-cluster | service-mesh-hub | new-remote-cluster.service-mesh-hub |
+--------------------+------------------+-------------------------------------+
```

{{% notice note %}}
Note, you can use `kubectl` directly instead of `meshctl` for this or any command as the information stored in Service Mesh Hub is all in CRDs. For example, you can run:

```shell
kubectl get kubernetesclusters -n service-mesh-hub
```

{{% /notice %}}

### Discover Meshes

Check to see that Istio has been discovered:

```shell
meshctl get meshes
```

```
Istio:
+---------------------------------------+--------------------+--------------+---------+-------------------------------------------------+
|                 NAME                  |      CLUSTER       |  NAMESPACE   | VERSION |                   CITADELINFO                   |
+---------------------------------------+--------------------+--------------+---------+-------------------------------------------------+
| istio-istio-system-management-plane   | management-plane   | istio-system | 1.5.1   | Trust Domain: cluster.local                     |
|                                       |                    |              |         | Citadel Namespace: istio-system                 |
|                                       |                    |              |         | Citadel Service Account: istiod-service-account |
+---------------------------------------+--------------------+--------------+---------+-------------------------------------------------+
| istio-istio-system-new-remote-cluster | new-remote-cluster | istio-system | 1.5.1   | Trust Domain: cluster.local                     |
|                                       |                    |              |         | Citadel Namespace: istio-system                 |
|                                       |                    |              |         | Citadel Service Account: istiod-service-account |
+---------------------------------------+--------------------+--------------+---------+-------------------------------------------------+
```

We can print it in YAML form to see all the information we discovered:

```shell
kubectl -n service-mesh-hub get mesh istio-istio-system-new-remote-cluster -oyaml
```

(snipped for brevity)

{{< highlight yaml >}}
apiVersion: discovery.smh.solo.io/v1alpha1
kind: Mesh
metadata:  
  ... 
  labels:
    discovered_by: istio-mesh-discovery
  name: istio-istio-system-new-remote-cluster
  namespace: service-mesh-hub  
  ... 
spec:
  cluster:
    name: new-remote-cluster
    namespace: service-mesh-hub
  istio:
    citadelInfo:
      citadelNamespace: istio-system
      citadelServiceAccount: istiod-service-account
      trustDomain: cluster.local
    installation:
      installationNamespace: istio-system
      version: 1.5.1

{{< /highlight >}}

### Discover Mesh Workloads

Check to see that the `bookinfo` pods have been correctly identified as Mesh Workloads:

```shell
kubectl -n service-mesh-hub get meshworkloads
```

```
NAME                                                         AGE
istio-details-v1-default-management-plane                    2m58s
istio-istio-ingressgateway-istio-system-management-plane     3m12s
istio-istio-ingressgateway-istio-system-new-remote-cluster   3m5s
istio-productpage-v1-default-management-plane                2m58s
istio-ratings-v1-default-management-plane                    2m58s
istio-ratings-v1-default-new-remote-cluster                  2m28s
istio-reviews-v1-default-management-plane                    2m58s
istio-reviews-v2-default-management-plane                    2m58s
istio-reviews-v3-default-new-remote-cluster                  2m28s
```

### Discover Mesh Services

Similarly for the `bookinfo` services:

```shell
kubectl -n service-mesh-hub get meshservices
```

```
NAME                                                   AGE
details-default-management-plane                       3m16s
istio-ingressgateway-istio-system-management-plane     3m30s
istio-ingressgateway-istio-system-new-remote-cluster   3m23s
productpage-default-management-plane                   3m16s
ratings-default-management-plane                       3m16s
ratings-default-new-remote-cluster                     2m46s
reviews-default-management-plane                       3m16s
reviews-default-new-remote-cluster                     2m45s
```

## See it in action

Check out "Part One" of the ["Dive into Service Mesh Hub" video series](https://www.youtube.com/watch?v=4sWikVELr5M&list=PLBOtlFtGznBjr4E9xYHH9eVyiOwnk1ciK):

<iframe width="560" height="315" src="https://www.youtube.com/embed/4sWikVELr5M" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>


## Next Steps

Now that we have Istio installed, and we've seen Service Mesh Hub discover the meshes across different clusters, we can now unify them into a single [Virtual Mesh]({{% versioned_link_path fromRoot="/reference/api/virtual_mesh/" %}}). See the guide on [establishing shared trust domain for multiple meshes]({{% versioned_link_path fromRoot="/guides/federate_identity" %}}).
