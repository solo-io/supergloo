---
title: "Concepts"
menuTitle: Concepts
description: Understanding Service Mesh Hub Concepts
weight: 20
---

Service Mesh Hub is a management plane that simplifies operations and workflows of service mesh installations across multiple clusters and deployment footprints. With Service Mesh Hub, you can install, discover, and operate a service-mesh deployment across your enterprise, deployed on premises, or in the cloud, even across heterogeneous service-mesh implementations.

![Service Mesh Hub Architecture]({{% versioned_link_path fromRoot="/img/smh-3clusters.png" %}})

## Why Service Mesh Hub

A service mesh provides powerful service-to-service communication capabilities like request routing, traffic observability, transport security, policy enforcement and more. A service mesh brings its highest value when many services (as many as possible) take advantage of its traffic control capabilities. In today's modern cloud deployments, both on-premises and in a public cloud, the expectation is that more, smaller clusters or deployment targets are preferred for the following reasons:

* Fault tolerance
* Compliance and data access
* Disaster recovery
* Scaling needs
* Geographic needs

Managing a service mesh deployment that is consistent and secure across multiple clusters is tedious and error prone at best. Service Mesh Hub provides a single unified pane of glass driven by a declarative API (CRDs in Kubernetes) that orchestrates and simplifies the operation of a multi-cluster service mesh including the following concerns:

* Unifying/federating trust domains
* Achieving a single pane of glass for operational observability
* Multi-cluster routing
* Access policy 

### A multi-cluster management plane

Service Mesh Hub consists of a set of components that run on a single cluster, often referred to as your *management plane cluster*. The management plane components are stateless and rely exclusively on declarative CRDs.  Each service mesh installation that spans a deployment footprint often has its own control plane. You can think of Service Mesh Hub as a management plane for multiple control planes.

![Service Mesh Hub Architecture]({{% versioned_link_path fromRoot="/img/concepts-smh-components.png" %}})

Once a cluster is registered with Service Mesh Hub, it can start managing that cluster - discovering workloads, pushing out configurations, unifying the trust model, scraping metrics, and more. 

In this document, we take a look at the concepts and components that comprise Service Mesh Hub.

## Concepts

### Virtual Meshes

To enable multi-cluster configuration, users will group multiple meshes together into an object called a [`VirtualMesh`]({{% versioned_link_path fromRoot="/reference/api/virtual_mesh/" %}}). The virtual mesh exposes configuration to facilitate cross-cluster communications. 

For a virtual mesh to be considered valid, Service Mesh Hub will first try to establish trust based on the [trust model](https://spiffe.io/spiffe/concepts/#trust-domain) defined by the user -- is there complete shared trust and a common root and identity? Or is there limited trust between clusters and traffic is gated by egress and ingress gateways? Service Mesh Hub ships with an agent that helps facilitate cross-cluster certificate signing requests safely, to minimize the operational burden around managing certificates. 

Once trust has been established, Service Mesh Hub will start federating services so that they are accessible across clusters. Behind the scenes, Service Mesh Hub will handle the networking -- possibly through egress and ingress gateways, and possibly affected by user-defined traffic and access policies -- and ensure requests to the service will resolve and be routed to the right destination. Networking configuration is push from Service Mesh Hub to the managed service meshes that are part of the virtual mesh. Users can fine-tune which services are federated where by editing the virtual mesh. 

### Traffic and Access Policies

Service Mesh Hub enables users to write simple configuration objects to the management plane to enact traffic and access policies between services. It was designed to be translated into the underlying mesh config, while abstracting away the mesh-specific complexity from the user. 

A [`TrafficPolicy`]({{% versioned_link_path fromRoot="/reference/api/traffic_policy/" %}}) applies between a set of sources (workloads) and destinations (traffic targets), and is used to describe rules like "when A sends POST requests to B, add a header and set the timeout to 10 seconds". Or "for every request to services on cluster C, increase the timeout and add retries". As of this release, traffic policies support timeouts, retries, CORS, traffic shifting, header manipulation, fault injection, subset routing, weighted destinations, and more. Note that some meshes don’t support all of these features; Service Mesh Hub will translate as best it can into the underlying mesh configuration, or report an error back to the user. 

An [`AccessPolicy`]({{% versioned_link_path fromRoot="/reference/api/access_policy/" %}}) also applies between sources (this time representing identities) and destinations, and is used to finely control which services are allowed to communicate. On the virtual mesh, a user can specify a global policy to restrict access, and require users to specify access policies in order to enable communication to services. 

With traffic and access policies, Service Mesh Hub gives users a powerful language to dictate how services should communicate, even within complex multi-cluster, multi-mesh applications. 

### CLI Tooling

Service Mesh Hub is tackling really hard problems related to multi-cluster networking and configuration, so to speed up your learning it comes with a command line tool called [`meshctl`]({{% versioned_link_path fromRoot="/reference/cli/meshctl/" %}}). This tool provides interactive commands to make it easier to author your first virtual mesh, register a cluster, or create a traffic or access policy. Once you’ve authored a config, it also has a `describe` command to help understand how your workloads and services are affected by your policies. 

## Components

The components that comprise Service Mesh Hub can be categorized as `mesh-discovery` and `mesh-networking` components. These components work together to discover meshes, traffic targets, and unify them using a `VirtualMesh` (see above). These components will write all of the necessary service-mesh-specific resources to the various clusters/meshes under management. For example, Service Mesh Hub would write all of the `ServiceEntry`, `VirtualService` and `DestinationRule` resources if managing Istio. Let's take a closer look at the components.
 
##### Mesh Discovery

A cluster is registered using the CLI `meshctl cluster register` command. During registration, Service Mesh Hub authenticates to the target cluster using the user-provided kubeconfig credentials, creates a `ServiceAccount` on that cluster for Service Mesh Hub and builds a kubeconfig granting access to the target cluster which is stored as a `secret` on the management-plane cluster. This kubeconfig is then used by Service Mesh Hub for all communication to that target cluster. For instance, discovery uses this kubeconfig to connect to the target cluster and start discovery. 

The discovery process is initiated by Service Mesh Hub running on the mgmt-cluster, pulling and translating information from the registered target clusters. Discovery of a target cluster is performed when the cluster is first registered and then periodically to discover and translate any newly added meshes, workloads, and traffic targets.

The first task of discovery is to create an `Input Snapshot` and begin the tranlation of the service meshes and services on the cluster to create an `Output Snapshot` that creates custom resources to the management plane cluster. 

`MeshTranslator` looks for installed service mesh control planes and then will add them to the `Output Snapshot` for a   [`Mesh`]({{% versioned_link_path fromRoot="/reference/api/mesh/" %}}) resource to be written to the management plane cluster, linked to the `KubernetesCluster` resource that was written during cluster registration. Currently, Service Mesh Hub discovers and manages both [Istio](https://istio.io) and [Open Service Mesh](https://openservicemesh.io/) meshes, with plans to support more in the near future.

`WorkloadTranslator` then looks for workloads that are associated with the mesh, such as a deployment that has created a pod injected with the sidecar proxy for that mesh. It adds them to the `Output Snapshot` which will write a [`Workload`]({{% versioned_link_path fromRoot="/reference/api/workload/" %}}) resource to the management plane cluster representing this workload. 

Finally, `TrafficTargetTranslator`  looks for services exposing the workloads of a mesh and adds them to the `Output Snapshot` which then writes a [`TrafficTarget`]({{% versioned_link_path fromRoot="/reference/api/mesh_service/" %}}) resource to the management plane cluster. 

![Service Mesh Hub Architecture]({{% versioned_link_path fromRoot="/img/concepts-smh-discovery.png" %}})

At this point, the management plane has a complete view of the meshes, workloads, and traffic targets across your multi-cluster, multi-mesh environment. 

##### Mesh Networking

While the `mesh-discovery` components discover the resources in registered clusters, the `mesh-networking` components make decisions about federating the various clusters and meshes. The `VirtualMesh` concept helps to drive the federation behavior. When you create a `VirtualMesh` resource, you define what federation and trust model to use. `Mesh-Networking` performs translation at the mesh level for a group of meshes with the`VirtulMeshTranslator` and at the service level with `TrafficTargetTranslator` which will then decide which services federate to which workloads and handle building the correct service-discovery mechanisms. For example, with Istio, Service Mesh Hub will create the appropriate `ServiceEntry` and `DestinationRule` resources to enable cross-cluster/mesh communication.

![Service Mesh Hub Architecture]({{% versioned_link_path fromRoot="/img/concepts-smh-networking.png" %}})

The `FederationTranslator`, `FailoverTranslator`, and `mTLSTranslator` are grouped within the `VirtulMeshTranslator` to handle mesh level configuration and networking.
 * `FederationTranslator` handles global DNS resolution and routing for services in remote clusters for each federated traffic target.
 * `FailoverTranslator` re-routes traffic to targets in remote clusters when local targets are unhealthy.
 * `mTLSTranslator` controls mesh-level settings for performing mTLS, including issuing & rotating root certificates used by each mesh.

The `TrafficTargetTranslator` handles the translation and policy configuration at the `TrafficTarget` (service) level. 
 * `TrafficPolicyTranslator` creates the `VirtualService` and `DestinationRule` for routing.
 * `AccessPolicyTranslator` creates the `AuthPolicy` for access control.

If users create a `TrafficPolicy` or `AccessPolicy` for Service Mesh Hub, the `mesh-networking` component will automatically translate those to the underlying mesh-specific resources. Again, for Istio, this would be `VirtualService`, `DesttinationRule`, and `AuthorizationPolicy` resources.

As opposed to the pull model of the mesh-discovery components, the mesh-networking components of Service Mesh Hub push changes down to the managed service meshes in each registered cluster. The configuration used by the control plane of each service mesh will be updated in real-time as changes are pushed from Service Mesh Hub. Many service mesh proxies, like Envoy, rely on a polling mechanism between the control plane and the proxy instances. Therefore, any changes pushed from Service Mesh Hub will be contingent on the polling cycle for the service mesh proxy instances.

## Next steps

Now that you've got an understanding of the concepts and components, check out our [Setup Guide]({{% versioned_link_path fromRoot="/setup/" %}}) to get it installed in your environment. Otherwise, please check out our [Guides]({{% versioned_link_path fromRoot="/guides/" %}}) to explore the power of Service Mesh Hub.
