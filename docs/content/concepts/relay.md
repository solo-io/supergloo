---
title: "Gloo Mesh Relay Architecture"
menuTitle: Relay Architecture
description: Understanding Gloo Mesh Relay Architecture
weight: 20
---

{{% notice note %}} Gloo Mesh Enterprise is required for this feature. {{% /notice %}}

## Overview

Relay is an alternative mode of deploying Gloo Mesh that provides several important
advantages which make it particularly suited to address enterprise concerns. This article
discusses the high-level architecture of relay and the benefits it offers.

### Architecture

In relay mode, a Gloo Mesh agent is deployed to each managed cluster at cluster
registration time. These agents communicate with the central management server
via gRPC to achieve cross cluster control.

Among other functions, the agents perform discovery only for the cluster in which
they are deployed and push the discovered entities to the management server.
The agents pull configuration updates from the management server in order to
enforce declared state on their cluster.

Similar to the regular deployment mode, the management server watches for
configuration updates in the management cluster. Reconciliation of actual vs. declared
state is triggered upon receiving discovery updates from any agent.

### Security Model

The regular deployment mode requires the user to grant the Gloo Mesh management plane
credentials to the Kubernetes API server for all managed clusters. This is undesirable
from a security standpoint. Not only does the user have to provision credentials for all
managed clusters, the Gloo Mesh management plane becomes a single point of compromise
over a broad surface area.

Relay mode's distributed push model for monitoring the state of managed
clusters obviates the need for granting the management server direct access
to Kubernetes API servers. The management server only requires a secure gRPC communication
channel with its agents, the details of which are discussed below.

### Server-Agent Security

gRPC communication between the server and agents is secured through mTLS. Initial trust
is established using a user-provided token that is known to the server. At registration
time, the agent presents the token to the server, which if valid, will be exchanged for a
TLS certificate that the agent uses for all communication to the server.
