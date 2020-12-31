---
title: "Gloo Mesh Architecture"
menuTitle: Architecture
description: Understanding Gloo Mesh Architecture
weight: 20
---

This document details the architecture of Gloo Mesh, which informs how Gloo Mesh is deployed and operated.

## Components

The components that comprise Gloo Mesh can be categorized as `mesh-discovery` and `mesh-networking` components. These components work together to discover meshes, traffic targets, and unify them using a [`VirtualMesh`]({{% versioned_link_path fromRoot="/concepts/concepts" %}}). These components will write all of the necessary service-mesh-specific resources to the various clusters/meshes under management. For example, Gloo Mesh would write all of the `ServiceEntry`, `VirtualService` and `DestinationRule` resources if managing Istio. Let's take a closer look at the components.
 
##### Mesh Discovery

A cluster is registered using the CLI `meshctl cluster register` command. During registration, Gloo Mesh authenticates to the target cluster using the user-provided kubeconfig credentials, creates a `ServiceAccount` on that cluster for Gloo Mesh and builds a kubeconfig granting access to the target cluster which is stored as a `secret` on the management-plane cluster. This kubeconfig is then used by Gloo Mesh for all communication to that target cluster. For instance, discovery uses this kubeconfig to connect to the target cluster and start discovery. 

The discovery process is initiated by Gloo Mesh running on the management plane cluster, pulling and translating information from the registered target clusters. Discovery of a target cluster is performed when the cluster is first registered and then periodically to discover and translate any newly added meshes, workloads, and traffic targets.

The first task of discovery is to create an `Input Snapshot` and begin the translation of the service meshes and services on the cluster to create an `Output Snapshot` that creates custom resources to the management plane cluster. 

`MeshTranslator` looks for installed service mesh control planes and then will add them to the `Output Snapshot` for a   [`Mesh`]({{% versioned_link_path fromRoot="/reference/api/mesh/" %}}) resource to be written to the management plane cluster, linked to the `KubernetesCluster` resource that was written during cluster registration. Currently, Gloo Mesh discovers and manages both [Istio](https://istio.io) and [Open Service Mesh](https://openservicemesh.io/) meshes, with plans to support more in the near future.

`WorkloadTranslator` then looks for workloads that are associated with the mesh, such as a deployment that has created a pod injected with the sidecar proxy for that mesh. It adds them to the `Output Snapshot` which will write a [`Workload`]({{% versioned_link_path fromRoot="/reference/api/workload/" %}}) resource to the management plane cluster representing this workload. 

Finally, `TrafficTargetTranslator`  looks for services exposing the workloads of a mesh and adds them to the `Output Snapshot` which then writes a [`TrafficTarget`]({{% versioned_link_path fromRoot="/reference/api/mesh_service/" %}}) resource to the management plane cluster. 

![Gloo Mesh Architecture]({{% versioned_link_path fromRoot="/img/concepts-gloomesh-discovery.png" %}})

At this point, the management plane has a complete view of the meshes, workloads, and traffic targets across your multi-cluster, multi-mesh environment. 

##### Mesh Networking

While the `mesh-discovery` components discover the resources in registered clusters, the `mesh-networking` components make decisions about federating clusters and meshes. The `VirtualMesh` concept enables the federation of multiple meshes into a single managed construct. As part of the `VirtualMesh` resource, you will define a federation model and trust model to use. 

`Mesh-networking` is what performs translation at the mesh level for the group of meshes within the VirtualMesh. There are two components that comprise `mesh-networking`, `VirtualMeshTranslator` and `TrafficTargetTranslator`. The `VirtualMeshTranslator` handles translation of settings at the mesh level, mapping mesh configuration defined by the VirtualMesh to individual meshes.  The `TrafficTargetTranslator` handles translation at the service layer to manage the mapping of services to workloads and traffic targets. For example, with Istio, Gloo Mesh will create the appropriate `ServiceEntry` and `DestinationRule` resources to enable cross-cluster/mesh communication.

![Gloo Mesh Architecture]({{% versioned_link_path fromRoot="/img/concepts-gloomesh-networking.png" %}})

The `FederationTranslator`, `FailoverTranslator`, and `mTLSTranslator` are grouped within the `VirtualMeshTranslator` to handle mesh level configuration and networking.
 * `FederationTranslator` handles global DNS resolution and routing for services in remote clusters for each federated traffic target.
 * `FailoverTranslator` re-routes traffic to targets in remote clusters when local targets are unhealthy.
 * `mTLSTranslator` controls mesh-level settings for performing mTLS, including issuing & rotating root certificates used by each mesh.

The `TrafficTargetTranslator` handles the translation and policy configuration at the `TrafficTarget` (service) level. 
 * `TrafficPolicyTranslator` creates the `VirtualService` and `DestinationRule` for routing.
 * `AccessPolicyTranslator` creates the `AuthorizationPolicy` for access control.

If users create a `TrafficPolicy` or `AccessPolicy` for Gloo Mesh, the `mesh-networking` component will automatically translate those to the underlying mesh-specific resources. Again, for Istio, this would be `VirtualService`, `DestinationRule`, and `AuthorizationPolicy` resources.

As opposed to the pull model of the `mesh-discovery` components, the `mesh-networking` components of Gloo Mesh push changes down to the managed service meshes in each registered cluster. The configuration used by the control plane of each service mesh will be updated in real-time as changes are pushed from Gloo Mesh. Many service mesh proxies, like Envoy, rely on a polling mechanism between the control plane and the proxy instances. Therefore, any changes pushed from Gloo Mesh will be contingent on the polling cycle for the service mesh proxy instances.

## Next Steps

The best way to understand how Gloo Mesh functions is by deploying it yourself and trying some of the features. We recommend diving in with the [Installation guide]({{% versioned_link_path fromRoot="/setup/" %}}).
