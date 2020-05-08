# Discovery for API based Service Mesh

*Date:* 4/16/2020

*Engineer*: Harvey Xia

## Problem Summary

Amazon's Service Mesh, called AppMesh, is hosted in AWS and must be interacted
with through their [REST API](https://docs.aws.amazon.com/app-mesh/latest/APIReference/Welcome.html).
This is a departure from the current state of SMH, which assumes that all managed
service meshes run on Kubernetes clusters that can be interacted with through Kubernetes config.

Fortunately the abstractions/interfaces that currently exist in SMH seem to roughly
 capture the required semantics for running discovery on API based meshes—a user registers
 a k8s cluster with SMH which grants SMH access to that cluster. These credentials
 are persisted as a k8s Secret, the creation of which is watched by a Secret EventWatcher 
 (skv2 terminology), which then triggers initialization of discovery on that cluster.

 This flow maps nicely onto an API based mesh, discussed below.

## Proposed Design

For an API based mesh, the following steps would occur:

1. User registers the service mesh REST API with SMH

2. SMH persists the credentials as a k8s Secret

    - when writing the secret, we need a field to distinguish the type of service
    mesh runtime environment, i.e. whether it's a k8s cluster or a REST API,
    so SMH knows what type of watcher to initialize (a k8s controller or a REST API watcher)

3. SMH records the existence of the REST API with a MeshAPI CRD (the exact name is subject to change,
but something to replace the existing `KubernetesCluster` CRD)

4. The existing `multiClusterHandler`, an implementation of SecretEventWatcher,
will receive a Secret create event. The Secret will indicate that the newly registered API
is a REST API, so SMH will spin up a REST API watcher accordingly.

5. Discovery for the REST API functions similar to for a k8s cluster, but instead of
running EventWatchers on k8s resources that indicate existence of Meshes, Workloads,
and Services, SMH will run a component that polls the REST API periodically for updates.

## Expected Concerns

Our terminology and in-code names should be updated to capture the fact that SMH
can integrate with service meshes through a REST API. The naming for this seems difficult. 
Here are some off-the-cuff candidates to get the creative juices flowing:

- MeshAPIServer
- APIServer
- MeshRuntime
- MeshEnvironment
