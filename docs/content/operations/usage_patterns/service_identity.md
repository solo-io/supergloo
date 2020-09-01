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

The Identity Selector is used by Access Policies to identify the source for a communication. The source will be a service account 

## Workload Selector

## Service Selector

