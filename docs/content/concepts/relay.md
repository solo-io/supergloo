---
title: "Gloo Mesh Relay Architecture"
menuTitle: Relay Architecture
description: Understanding Gloo Mesh Relay Architecture
weight: 20
---

{{% notice note %}} Gloo Mesh Enterprise is required for this feature. {{% /notice %}}

## Overview

Gloo Mesh Relay is an alternative mode of deploying Gloo Mesh that provides several important advantages making it particularly suited to address enterprise concerns. This article discusses the high-level architecture of relay and the benefits it offers.

### Architecture

In relay mode, a Gloo Mesh agent is deployed to each managed cluster when the cluster is registered. These agents communicate with the central management server via gRPC to achieve cross-cluster control.

The agents perform discovery only for the cluster in which they are deployed and push discovered entities to the management server. The agents pull configuration updates from the management server in order to enforce declared state on their cluster.

Similar to the regular deployment mode, the management server watches for configuration updates in the management cluster. Reconciliation of actual vs. declared state is triggered upon receiving discovery updates from any agent.

### Security Model

The regular deployment mode requires the user to grant the Gloo Mesh management plane credentials to the Kubernetes API server for all managed clusters. This is undesirable from a security standpoint. Not only does the user have to provision credentials for all managed clusters, the Gloo Mesh management plane becomes a single point of compromise over a broad surface area.

Relay mode's distributed push model for monitoring the state of managed clusters obviates the need to grant the management server direct access
to Kubernetes API servers on managed clusters. The management server only requires a secure gRPC communication channel with its agents, the details of which are discussed below.

### Server-Agent Security

gRPC communication between the server and agents is secured through mTLS. Initial trust is established with a user-provided token that is known to the server. At registration time, the agent presents the token to the server, which if valid, will be exchanged for a TLS certificate that the agent uses for all communication to the server.

### Components

The deployment of the relay model will require a management cluster and one or more managed clusters. The management cluster will have a deployment called `enterprise-networking` running the relay server. The relay server is exposed via the `enterprise-networking` service on a default port of `9900/TCP`. The management cluster also has a properly configured ingress point allowing managed clusters to communicate with the relay server using a VirtualService listening on port 443. 

Managed clusters will have the deployment `enterprise-agent` running the relay agent, which establishes a communication channel with the relay server using gRPC. The relay agent is exposed via the `enterprise-agent` service on the default ports of `9988` and `9977`.  Because all communication is outbound from the managed clusters, there does not need to be a ingress point configured for the relay agent.

Before establishing secure communication between the relay agent and server, the relay server must issue a certificate to the agent. The management cluster will have Kubernetes secrets containing certificates for use by the server. The relay server must be able to issue certificates to the relay agent, and the relay agent must trust the certificate chain being used by the relay server. You can use self-signed certificates or certificates from your PKI.

Initial communication between the relay agent and server uses TLS provided by the relay server certificate. The agent will transmit a token value - defined in `relay-identity-token-secret` on the agent's cluster - to the relay server to validate authenticity. The token must match the value stored in `relay-identity-token-secret` on the management cluster, which is created during deployment of the relay server. Once the token is validated, the relay server will generate a certificate for the relay agent, enabling mTLS for all future communication.