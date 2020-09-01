---
title: "Multicluster Service Identity"
menuTitle: Multicluster Service Identity
description: Guidance on using Service Mesh Hub to apply consistent policies to services across clusters.
weight: 10
---

Service Mesh Hub was created to simplify the management of multiple service mesh deployments across multiple clusters. One of the key abstractions of Service Mesh Hub is the VirtualMesh, essentially the combination of multiple service mesh into a single logical entity. The VirtualMesh enables cross-cluster, secured communication with a common root certificate authority serving to provide trusted mTLS communications.

To further manage communication within and between clusters, Service Mesh Hub includes Access Policies and Traffic Policies. The Access Policies define what communication is allowed between sources (identities) and destinations. In addition to source and destination, the Access Policy can also specify paths, methods, and ports for a request. Traffic Policies define how communication between a source and destination is handled, including things like traffic shifting, fault injection, and header manipulation. 

Both the Access and Traffic Policies use `sourceSelector` for workloads/identities and `destinationSelector` for services. The selection syntax provides an opportunity to abstract a workload, identity, or service and apply consistent policies across multiple resources in different meshes and clusters.

In this document, we will walk through an example of using each selector type to implement consistent settings and policies across multiple resources.

## Identity Selector

The Identity Selector is used by Access Policies to identify the source for a communication. The source can be define using a `KubeIdentityMatcher` or `KubeServiceAccountRef`. The `KubeIdentityMatcher` includes a list of allowed namespaces and a list of allowed clusters. The `KubeServiceAccountRef` will refer to a singular service account in a namespace and cluster.

It is possible to use the `KubeIdentityMatcher` with just a list of namespaces and no cluster, which would have the effect of allowing traffic from those namespaces in any cluster to the destination. This could be useful when trying to apply an access policy to a source application that uses a common namespace across all clusters. The following code snippet illustrates such a configuration:

```yaml
kubeIdentityMatcher:
  namespaces:
    - app1
    - app2
```

In a situation where there are two clusters, `cluster-one` and `cluster-two`, the Access Policy would allow traffic from the namespace `app1` and `app2` in either cluster. More importantly, if `cluster-three` were added with the same namespaces, it would also be allowed. If we would like to be more restrictive, the clusters in question can be added to the code snippet as well:

```yaml
kubeIdentityMatcher:
  namespaces:
    - app1
    - app2
  clusters:
    - cluster-one
    - cluster-two
```

Now if `cluster-three` is added to Service Mesh Hub, the `app1` and `app2` namespaces in that cluster would not be allowed to send traffic until the Access Policy was updated.

## Workload Selector

## Service Selector

