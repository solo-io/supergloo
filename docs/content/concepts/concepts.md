---
title: "Core Concepts"
menuTitle: Core Concepts
description: Understanding Gloo Mesh Core Concepts
weight: 10
---

In this document, we'll take a look at some of the core concepts that underpin Gloo Mesh. These concepts, like *Virtual Meshes* and *Access Policies*, are critical in understanding how Gloo Mesh can manage multiple service meshes and act as a unified control plane.

## Virtual Meshes

To enable multi-cluster configuration, users will group multiple meshes together into an object called a [`VirtualMesh`]({{% versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.virtual_mesh/" %}}). The virtual mesh exposes configuration to facilitate cross-cluster communications.

For a virtual mesh to be considered valid, Gloo Mesh will first try to establish trust based on the [trust model](https://spiffe.io/spiffe/concepts/#trust-domain) defined by the user -- is there complete shared trust and a common root and identity? Or is there limited trust between clusters and traffic is gated by egress and ingress gateways? Gloo Mesh ships with an agent that helps facilitate cross-cluster certificate signing requests safely, to minimize the operational burden around managing certificates. 

Once trust has been established, Gloo Mesh will start federating services so that they are accessible across clusters. Behind the scenes, Gloo Mesh will handle the networking -- possibly through egress and ingress gateways, and possibly affected by user-defined traffic and access policies -- and ensure requests to the service will resolve and be routed to the right destination. Networking configuration is pushed from Gloo Mesh to the managed service meshes that are part of the virtual mesh. Users can fine-tune which services are federated where by editing the virtual mesh. 

## Traffic and Access Policies

Gloo Mesh enables users to write simple configuration objects to the management plane to enact traffic and access policies between services. It was designed to be translated into the underlying mesh config, while abstracting away the mesh-specific complexity from the user. 

A [`TrafficPolicy`]({{% versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.traffic_policy/" %}}) applies between a set of sources (workloads) and destinations, and is used to describe rules like "when A sends POST requests to B, add a header and set the timeout to 10 seconds". Or "for every request to services on cluster C, increase the timeout and add retries". As of this release, traffic policies support timeouts, retries, CORS, traffic shifting, header manipulation, fault injection, subset routing, weighted destinations, and more. Note that some meshes don’t support all of these features; Gloo Mesh will translate as best it can into the underlying mesh configuration, or report an error back to the user.

An [`AccessPolicy`]({{% versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1alpha2.access_policy/" %}}) also applies between sources (this time representing identities) and destinations, and is used to finely control which services are allowed to communicate. On the virtual mesh, a user can specify a global policy to restrict access, and require users to specify access policies in order to enable communication to services.

With traffic and access policies, Gloo Mesh gives users a powerful language to dictate how services should communicate, even within complex multi-cluster, multi-mesh applications. 

## CLI Tooling

Gloo Mesh is tackling really hard problems related to multi-cluster networking and configuration, so to speed up your learning it comes with a command line tool called [`meshctl`]({{% versioned_link_path fromRoot="/reference/cli/meshctl/" %}}). This tool provides interactive commands to make it easier to author your first virtual mesh, register a cluster, or create a traffic or access policy. Once you’ve authored a config, it also has a `describe` command to help understand how your workloads and services are affected by your policies. 

## Next Steps

With a firm grounding in the core concepts, you're now ready to learn more about [Gloo Mesh's architecture]({{% versioned_link_path fromRoot="/concepts/architecture" %}}) or dive into our [Getting Started guide]({{% versioned_link_path fromRoot="/getting_started/" %}}).
