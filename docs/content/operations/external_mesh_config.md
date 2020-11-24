---
title: Externally Provided Mesh Configuration
weight: 50
---

Gloo Mesh offers an API for configuring specific service mesh deployments. This 
is performed by translating Gloo Mesh configuration into service mesh specific 
configuration. While our recommendation is to utilize the Gloo Mesh API for all
configuration needs, there are certain situations in which the user may desire
to directly provide configuration objects for managed service meshes (e.g. to use
service mesh features that are not exposed through the Gloo Mesh API, migrating 
 from unmanaged service mesh deployments to Gloo Mesh, etc.) This document 
 describes what the user should be aware of when directly providing mesh-specific
 configuration objects.

## Gloo Mesh Object Naming Conventions

All translated objects contain an ownership label key of the form "owner.{group}"
(e.g. "owner.networking.mesh.gloo.solo.io" ) where group refers to the k8s group of the "parent object", i.e. the Gloo Mesh 
object responsible for their translation.

#### Istio

Generally speaking, Istio VirtualServices, DestinationRules, and AuthorizationPolicies 
translated by Gloo Mesh **assume the name and namespace of the traffic target being configured**.
For instance, the following TrafficPolicy:

{{< highlight yaml "hl_lines=11-13" >}}
apiVersion: networking.mesh.gloo.solo.io/v1alpha2
kind: TrafficPolicy
metadata:
  name: traffic-shift
  namespace: foobar
spec:
  destinationSelector:
  - kubeServiceRefs:
      services:
      - clusterName: remote-cluster
        name: reviews
        namespace: bookinfo
{{< /highlight >}}
 
will yield a VirtualService with the following object meta:

{{< highlight yaml "hl_lines=8-9" >}}
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  labels:
    cluster.multicluster.solo.io: remote-cluster
    owner.networking.mesh.gloo.solo.io: gloo-mesh
  name: reviews
  namespace: bookinfo
{{< /highlight >}}

When translating Gloo Mesh configuration for *federated* traffic targets (i.e.
 traffic targets that belong in service mesh deployments that are grouped in a 
 VirtualMesh), translated VirtualServices and DestinationRules assume a name of
 the form "{name}-{namespace}-{clusterName}" of the parent object, and are placed
 in the installation namespace of the service mesh.
 
The example TrafficPolicy above would yield the following VirtualService on all 
clusters that the "reviews" service is federated to:

{{< highlight yaml "hl_lines=8-9" >}}
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  labels:
    cluster.multicluster.solo.io: mgmt-cluster
    owner.networking.mesh.gloo.solo.io: gloo-mesh
  name: reviews-bookinfo-remote-cluster
  namespace: istio-system
{{< /highlight >}}

## Compatibility With Externally Provided Service Mesh Configuration

#### Istio

Istio supports disparate configuration objects that apply to the same configuration
target. We recommend getting familiar with their merge semantics for VirtualServices,
DestinationRules, and AuthorizationPolicies in order to understand the behavior 
in scenarios when both Gloo Mesh and an external source (e.g. users) are supplying
configuration for the same host or workload.

The [Istio traffic management guide](https://istio.io/latest/docs/ops/best-practices/traffic-management/#split-virtual-services)
describes the merge semantics for VirtualServices and DestinationRules. The
[Istio AuthorizationPolicy reference](https://istio.io/latest/docs/reference/config/security/authorization-policy/)
describes the merge semantics for AuthorizationPolicies.
