---
title: "Community"
menuTitle: Community
description: Registering a cluster with Gloo Mesh community edition
weight: 30
---

Once you have Gloo Mesh installed, the next step is to register Kubernetes clusters that will have service meshes you want to manage. The registration process creates a service account, cluster-role, and cluster-role association on the target cluster granting the service account the necessary permissions to monitor and make changes to the cluster. The current cluster-role definition is documented in the [references]({{% versioned_link_path fromRoot="/reference/cluster_role" %}}) section of the documentation.

This guide will walk you through the basics of registering clusters using the `meshctl` tool. We will be using the two cluster contexts mentioned in the Gloo Mesh installation guide, `kind-cluster-1` and `kind-cluster-2`. Your cluster context names will likely differ, so please substitute the proper values.

## Register A Cluster

In order to identify a cluster as being managed by Gloo Mesh, we have to *register* it in our installation. Registration ensures we are aware of the cluster, and we have proper credentials to communicate with the Kubernetes API server in that cluster.

#### Remote Cluster

We will start by registering a remote cluster, i.e. a cluster not running the Gloo Mesh installation. We will need to tell `meshctl` which kubectl context to use. Let's start by storing the name of our context in a variable:

```shell
CONTEXT_2=your_remote_context (cluster 2 if following our examples)
```
We will register the cluster with the `meshctl cluster register community` command. We will assume that the management plane cluster (i.e. where `discovery` and `networking` are running) is your current kubeconfig context. If not, set the `--mgmt-context` flag. The kubeconfig context for the cluster we want to register to the management plane is specified with the `--remote-context` flag. If you are using Kind for your clusters, you will need to specify the `--api-server-address` flag along with the IP address and port of the Kubernetes API server. Select the Kind tab for the commands.

{{< tabs >}}
{{< tab name="Kubernetes" codelang="shell" >}}
meshctl cluster register community cluster-2 \
  --remote-context $CONTEXT_2
{{< /tab >}}
{{< tab name="Kind (MacOS)" codelang="shell" >}}
ADDRESS=$(docker inspect cluster-2-control-plane | jq -r '.[0].NetworkSettings.Networks.kind.IPAddress')

meshctl cluster register community cluster-2 \
  --remote-context $CONTEXT_2 \
  --api-server-address ${ADDRESS}:6443
{{< /tab >}}
{{< tab name="Kind (Linux)" codelang="shell" >}}
ADDRESS=$(docker exec "cluster-2-control-plane" ip addr show dev eth0 | sed -nE 's|\s*inet\s+([0-9.]+).*|\1|p')

meshctl cluster register community cluster-2 \
  --remote-context $CONTEXT_2 \
  --api-server-address ${ADDRESS}:6443
{{< /tab >}}
{{< /tabs >}}

```shell
successfully registered cluster cluster-2
```

The context we use must have adequate permissions on the target cluster to create the service account, cluster-role, and cluster-role assignment.


You can validate the registration by looking at the Custom Resource created in the management cluster:

```shell
kubectl get kubernetescluster -n gloo-mesh cluster-2
```

```shell
NAME             AGE
cluster-2        10m
```

The target cluster will also have the following elements created:

```shell
kubectl get sa --context $CONTEXT_2 -n gloo-mesh
```

```shell
NAME             SECRETS   AGE
cert-agent       1         11m
default          1         11m
cluster-2        1         11
```

```shell
kubectl get clusterrole --context $CONTEXT_2 gloomesh-remote-access
```

```shell
NAME                     AGE
gloomesh-remote-access   12m
```

```shell
kubectl get clusterrolebinding --context $CONTEXT_2 \
  cluster-2-gloomesh-remote-access-clusterrole-binding
```

```shell
NAME                                                        AGE
cluster-2-gloomesh-remote-access-clusterrole-binding        13m
```

#### Register the management cluster

First, let's store the name of the management cluster context in a variable:

```shell
MGMT_CONTEXT=your_management_plane_context
```

Select the *Kind* tab if you are running Kubernetes in Docker.

{{< tabs >}}
{{< tab name="Kubernetes" codelang="shell" >}}
meshctl cluster register community cluster-1 \
  --remote-context $MGMT_CONTEXT
{{< /tab >}}
{{< tab name="Kind (MacOS)" codelang="shell" >}}
ADDRESS=$(docker inspect cluster-1-control-plane | jq -r '.[0].NetworkSettings.Networks.kind.IPAddress')

meshctl cluster register community cluster-1 \
  --remote-context $MGMT_CONTEXT \
  --api-server-address ${ADDRESS}:6443
{{< /tab >}}
{{< tab name="Kind (Linux)" codelang="shell" >}}
ADDRESS=$(docker exec "cluster-1-control-plane" ip addr show dev eth0 | sed -nE 's|\s*inet\s+([0-9.]+).*|\1|p')

meshctl cluster register community cluster-1 \
  --remote-context $MGMT_CONTEXT \
  --api-server-address ${ADDRESS}:6443
{{< /tab >}}
{{< /tabs >}}

```
successfully registered cluster cluster-1
```

## What happened?

To go into slightly more detail about what just happened:

* A service account was created in the `gloo-mesh` namespace of the remote cluster
* That service account's auth token was stored in a secret in the management plane cluster
* The Gloo Mesh CSR agent was deployed in the remote cluster
* Future communications that Gloo Mesh does to the remote cluster's Kubernetes API server
  will be done using the service account auth token

## Next Steps

And we're done! Any meshes in that cluster will be discovered and available for configuration by Gloo Mesh. See the guide on [installing Istio]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}), to see how to easily get Istio running on that cluster.
