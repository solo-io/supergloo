# Discovery for Consul

*Date:* 5/21/2020

*Engineer*: Harvey Xia

## Problem Summary

Consul differs from other service meshes (e.g. Istio, Appmesh, Linkerd) in that its [architecture is fully decentralized](https://www.consul.io/docs/internals/architecture.html). There is no notion
of a centralized control plane. Instead, Consul is deployed as a set of [agents](https://www.consul.io/docs/agent) that run in either "server" or
 "client" mode. Among the available server agents, the [Raft consensus protocol](https://raft.github.io/) is used to elect a single leader who
 is responsible for processing all queries and transactions.
 
The set of Consul agents that constitute a single logical "mesh" (the term is used loosely here) is dynamic—Consul uses a [gossip protocol](https://www.consul.io/docs/internals/gossip)
to distribute membership information so that each agent is aware of other agents in the membership group. Agents running in distinct 
membership groups with no mutual awareness can be [joined into a single group](https://www.consul.io/docs/commands/join.html) as illustrated
 by [this tutorial](https://learn.hashicorp.com/consul/day-2-operations/servers).
 
Service Mesh Hub requires a way to identify and distinguish between different Consul membership groups. If there existed a reliable, unique 
identifier for a Consul membership group, Service Mesh Hub could simply use it to identify and distinguish between different Consul membership 
groups.

Unfortunately, upon investigation and experimentation on a local Consul deployment, I have not found any data that can act as a reliable identifier for 
a membership group. It appears that the identity of a Consul membership group consists solely in its members.

## Proposed Design

If the identity of a Consul membership group consists solely in its members, Service Mesh Hub can identify a Consul membership group as 
its set of members. More specifically, we can just use the set of server members since we expect this to have less membership churn than clients: 

> Within each datacenter, we have a mixture of clients and servers. It is expected that there be between three to five servers. This strikes
 a balance between availability in the case of failure and performance, as consensus gets progressively slower as more machines are added.
  However, there is no limit to the number of clients, and they can easily scale into the thousands or tens of thousands.
>[Reference link.](https://www.consul.io/docs/internals/architecture.html)

Mesh discovery would proceed with the usual strategy of scanning k8s pods for the presence of the `consul-connect-envoy-sidecar` container,
and if found, computes the SMH Mesh entity with the following:

1. Mesh name would be a string derived from a deterministic hash function using as input an ordered list of member server IP addresses
    - this data can be retrieved from the Consul REST API, which should be accessible from any environment running a Consul agent
2. The `ConsulConnectMesh` proto message (a field on the Mesh CRD) contains a list of member server IP addresses as metadata for downstream use

This implies that if server membership of the group changes (new server joins, an existing server leaves, existing server changes IP address
 (I'm not sure how/whether this can happen)), SMH will produce a new Mesh CR to represent the new Consul group, and the previous Mesh CR will
 be deleted. All MeshWorkloads (and thus MeshServices) will be recomputed as well because of the dependency of MeshWorkloads on the existence
  of its parent Mesh.
  
The cluster tenancy scanner would then populate this Mesh CR with all k8s clusters that are running a Consul member agent by deriving the same
Mesh CR keyed on the hashed name as computed by Mesh discovery above. The rest of discovery
proceeds the usual way.

## Expected Concerns

This approach defines Consul Mesh identity as the set of member server agents. If the membership of server agents undergoes high churn (i.e.
frequent joining and leaving of members), Service Mesh Hub discovery would need to keep up by reprocessing all Mesh/MeshWorkload/MeshServices
for each new set of members. It is assumed that the set of Consul servers in any particular membership group is relatively stable, an assumption
that rests largely on a priori intuition based on their architecture documentation (e.g. "It is expected that there be between three to five servers.")
