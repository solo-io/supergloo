---
title: "Gloo Mesh Setup Using Kind"
menuTitle: Using Kind
description: Deploying clusters with Kind for a Gloo Mesh Setup
weight: 1
---

The installation of Gloo Mesh assumes that you have access to at least one Kubernetes cluster, and preferably two for following the multi-cluster guides. If you do not have access to remote clusters, you can instead spin up two local clusters using Kubernetes in Docker (Kind). Be aware that this will use a significant amount of RAM when both clusters are running with Istio and Gloo Mesh installed. We recommend a workstation with a minimum of 16GB, preferably 32GB if possible.

This guide will walk you through deploying two Kind clusters referred to as the `mgmt-cluster` and `remote-cluster`. The mgmt-cluster will host the Gloo Mesh installation as well as a deployment of Istio. The remote-cluster will run Istio only, and will be used in later guides to demonstrate the multi-cluster management capabilities of Gloo Mesh.

### Create your Kind Clusters

Before you run Kind on your local workstation, you will need the following pre-requisites:

* [Docker Desktop](https://www.docker.com/products/docker-desktop) or Docker running as a service
* Kind installed using `go get` or a [stable binary for your platform](https://kind.sigs.k8s.io/docs/user/quick-start)

Once you have those pieces in place, you will simply run the following commands to create the mgmt-cluster and remote-cluster clusters.

```bash
# Create the mgmt-cluster
# Set version, cluster name, and port
kindImage=kindest/node:v1.17.5
cluster=mgmt-cluster
port=32001

cat <<EOF | kind create cluster --name "${cluster}" --image $kindImage --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: ${port}
    hostPort: ${port}
    protocol: TCP
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        # Populate nodes with region/zone info, which are used by VirtualDestination locality-based failover (Enterprise-only)
        node-labels: "ingress-ready=true,topology.kubernetes.io/region=us-east-1,topology.kubernetes.io/zone=us-east-1b"
kubeadmConfigPatches:
- |
  kind: InitConfiguration
  nodeRegistration:
    kubeletExtraArgs:
      authorization-mode: "AlwaysAllow"
EOF

# Create the remote cluster
cluster=remote-cluster
port=32000

cat <<EOF | kind create cluster --name "${cluster}" --image $kindImage --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: ${port}
    hostPort: ${port}
    protocol: TCP
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        # Populate nodes with region/zone info, which are used by VirtualDestination locality-based failover (Enterprise-only)
        node-labels: "ingress-ready=true,topology.kubernetes.io/region=us-east-1,topology.kubernetes.io/zone=us-east-1c"
kubeadmConfigPatches:
- |
  kind: InitConfiguration
  nodeRegistration:
    kubeletExtraArgs:
      authorization-mode: "AlwaysAllow"
EOF

# Switch to the mgmt-cluster context
kubectl config use-context kind-mgmt-cluster
```

Set your environment variables:

You should now have the following environment variables set:
```shell script
echo $MGMT_CONTEXT
echo $REMOTE_CONTEXT
```


## Install Istio

[Deploy Istio]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}) or
[Gloo Mesh Istio]({{% versioned_link_path fromRoot="/setup/gloo_mesh_istio/" %}})
on both clusters.

## Install Gloo Mesh

##### OSS (community version)

1) Follow the [Gloo Mesh installation instructions]({{% versioned_link_path fromRoot="/setup/installation/community_installation" %}}).

2) [Register your clusters]({{% versioned_link_path fromRoot="/setup/cluster_registration/community_cluster_registration" %}}).

##### Enterprise

1) Go through the [Gloo Mesh Enterprise Prerequisites]({{% versioned_link_path fromRoot="/setup/prerequisites/enterprise_prerequisites" %}}).

2) [Register your enterprise clusters]({{% versioned_link_path fromRoot="/setup/cluster_registration/enterprise_cluster_registration" %}}).

## Follow our Guides

3) Install the [bookinfo example deployment]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployment" %}})

4) Check out some of our other [guides]({{% versioned_link_path fromRoot="/guides" %}})!
