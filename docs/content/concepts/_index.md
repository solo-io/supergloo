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

* fault tolerance
* compliance and data access
* disaster recovery
* scaling needs
* geographic needs

Managing a service mesh deployment that is consistent and secure across multiple clusters is tedious and error prone at best. Service Mesh Hub provides a single unified pane of glass driven by a declarative API (CRDs in Kubernetes) that orchestrates and simplifies the operation of a multi-cluster service mesh including the following concerns:

* Unifying/federating trust domains
* Achieving a single pane of glass for operational observability
* Multi-cluster routing
* Access policy 

### A multi-cluster management plane

Service Mesh Hub consists of a set of components that run on a single cluster, often referred to as your *management plane cluster*. The management plane components are stateless and rely exclusively on declarative CRDs.  Each service-mesh installation that spans a deployment footprint often has its own control plane. You can think of Service Mesh Hub as a management plane for multiple control planes.

Once a cluster is registered with Service Mesh Hub, it can start managing that cluster - discovering service mesh workloads, pushing out configurations, unifying the trust model, scraping metrics, and more. 

Let's take a closer look at the Service Mesh Hub Concepts

## Concepts

#### Discovery

When a cluster is registered, Service Mesh Hub starts discovery. The first task of discovery is to find any service meshes that are installed on the cluster. When it finds the control plane for a service mesh, discovery will write a `Mesh` resource to the management plane cluster, linked to the `KubernetesCluster` resource that was written during cluster registration. Currently, Service Mesh Hub discovers and manages [Istio](https://istio.io) and [Linkerd](https://linkerd.io) meshes, with plans to support more in the near future.

`Discovery` then looks for workloads that are associated with the mesh, such as a deployment that has created a pod that is injected with the sidecar proxy for that mesh. It will write a MeshWorkload resource to the management plane cluster representing this workload. 

Finally, discovery also looks for services that are exposing the workloads of a mesh, and as before, writes a `MeshService` resource to the management plane cluster. 

At this point, the management plane has a complete view of the meshes, services, and workloads across your multi-cluster, multi-mesh environment. 

#### Virtual Meshes

In order to enable multi-cluster configuration, users will group multiple meshes together into an object called a `VirtualMesh`. The virtual mesh exposes configuration to facilitate cross-cluster communications. 

In order for a virtual mesh to be considered valid, Service Mesh Hub will first try to establish trust based on the [trust model](https://spiffe.io/spiffe/concepts/#trust-domain) defined by the user -- is there complete shared trust and a common root and identity? Or is there limited trust between clusters and traffic is gated by egress and ingress gateways? Service Mesh Hub ships with an agent that helps facilitate cross-cluster certificate signing requests safely, to minimize the operational burden around managing certificates. 

Once trust has been established, Service Mesh Hub will start federating services so that they are accessible across clusters. Behind the scenes, Service Mesh Hub will handle the networking -- possibly through egress and ingress gateways, and possibly affected by user-defined traffic and access policies -- and ensure requests to the service will resolve and be routed to the right destination. Users can fine-tune which services are federated where by editing the virtual mesh. 

#### Traffic and Access Policies

Service Mesh Hub enables users to write simple configuration objects to the management plane to enact traffic and access policies between services. It was designed to be translated into the underlying mesh config, while abstracting away the mesh-specific complexity from the user. 

A `TrafficPolicy` applies between a set of sources (mesh workloads) and destinations (mesh services), and is used to describe rules like “when A sends POST requests to B, add a header and set the timeout to 10 seconds”. Or “for every request to services on cluster C, increase the timeout and add retries”. As of this release, traffic policies support timeouts, retries, cors, traffic shifting, header manipulation, fault injection, subset routing, weighted destinations, and more. Note that some meshes don’t support all of these features; Service Mesh Hub will translate as best it can into the underlying mesh configuration, or report an error back to the user. 

An `AccessControlPolicy` also applies between sources (this time representing identities) and destinations, and are used to finely control which services are allowed to communicate. On the virtual mesh, a user can specify a global policy to restrict access, and require users to specify access policies in order to enable communication to services. 

With traffic and access policies, Service Mesh Hub gives users a powerful language to dictate how services should communicate, even within complex multi-cluster, multi-mesh applications. 

#### CLI Tooling

Service Mesh Hub is tackling really hard problems related to multi-cluster networking and configuration, so to speed up your learning it comes with a command line tool called `meshctl`. This tool provides interactive commands to make it easier to author your first virtual mesh, register a cluster, or create a traffic or access policy. Once you’ve authored config, it also has a `describe` command to help understand how your workloads and services are affected by your policies. 

