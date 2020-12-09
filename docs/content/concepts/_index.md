---
title: "Concepts"
menuTitle: Concepts
description: Understanding Gloo Mesh Concepts
weight: 20
---

Gloo Mesh is a management plane that simplifies operations and workflows of service mesh installations across multiple clusters and deployment footprints. With Gloo Mesh, you can install, discover, and operate a service-mesh deployment across your enterprise, deployed on premises, or in the cloud, even across heterogeneous service-mesh implementations.

![Gloo Mesh Architecture]({{% versioned_link_path fromRoot="/img/gloomesh-3clusters.png" %}})

## Why Gloo Mesh

A service mesh provides powerful service-to-service communication capabilities like request routing, traffic observability, transport security, policy enforcement and more. A service mesh brings its highest value when many services (as many as possible) take advantage of its traffic control capabilities. In today's modern cloud deployments, both on-premises and in a public cloud, the expectation is that more, smaller clusters or deployment targets are preferred for the following reasons:

* Fault tolerance
* Compliance and data access
* Disaster recovery
* Scaling needs
* Geographic needs

Managing a service mesh deployment that is consistent and secure across multiple clusters is tedious and error prone at best. Gloo Mesh provides a single unified pane of glass driven by a declarative API (CRDs in Kubernetes) that orchestrates and simplifies the operation of a multi-cluster service mesh including the following concerns:

* Unifying/federating trust domains
* Achieving a single pane of glass for operational observability
* Multi-cluster routing
* Access policy 

### A multi-cluster management plane

Gloo Mesh consists of a set of components that run on a single cluster, often referred to as your *management plane cluster*. The management plane components are stateless and rely exclusively on declarative CRDs.  Each service mesh installation that spans a deployment footprint often has its own control plane. You can think of Gloo Mesh as a management plane for multiple control planes.

![Gloo Mesh Architecture]({{% versioned_link_path fromRoot="/img/concepts-gloomesh-components.png" %}})

Once a cluster is registered with Gloo Mesh, it can start managing that cluster - discovering workloads, pushing out configurations, unifying the trust model, scraping metrics, and more. 

In this section, we take a look at the concepts and components that comprise Gloo Mesh.

{{% children description="true" %}}
