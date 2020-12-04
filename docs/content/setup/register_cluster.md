---
title: "Register a Cluster with Gloo Mesh"
menuTitle: Register Cluster
description: Registering a cluster with Gloo Mesh's management plane
weight: 30
---

Once you have Gloo Mesh or Gloo Mesh Enterprise installed, the next step is to register Kubernetes clusters that will have service mesh's you want to manage. The registration process creates a service account, cluster-role, and cluster-role association on the target cluster granting the service account the necessary permissions to monitor and make changes to the cluster. The current cluster-role definition is documented in the References section of the documentation.

This guide will walk you through the basics of registering clusters using the `meshctl` tool. We will be using the two cluster contexts mentioned in the Gloo Mesh installation guide, `mgmt-cluster` and `remote-cluster`. Your cluster context names will likely differ, so please substitute the proper values.

## Register A Cluster

In order to identify a cluster as being managed by Gloo Mesh, we have to *register* it in our installation. Registration ensures we are aware of the cluster, and we have proper credentials to communicate with the Kubernetes API server in that cluster.

#### Remote Cluster

We will start by registering a remote cluster, i.e. a cluster not running the Gloo Mesh installation. We will need to tell `meshctl` which kubectl context to use. Let's start by storing the name of our context in a variable:

```shell
REMOTE_CONTEXT=your_remote_context
```
We will register the cluster with the `meshctl cluster register` command. The kubeconfig context we want to use for the registration process is specified with the `-remote-context` flag as shown below.

```shell
meshctl cluster register \
  --cluster-name remote-cluster \
  --remote-context $REMOTE_CONTEXT
```

The context we use must have adequate permissions on the target cluster to create the service account, cluster-role, and cluster-role assignment.

{{% notice note %}}
Note that the `--cluster-name` is NOT the name of the cluster in your kubeconfig file -- it's a name Gloo Mesh can use to refer to the cluster in various configurations. You can pick any name you want for this.
{{% /notice %}}

```shell
INFO[0003] successfully registered cluster remote-cluster
```

{{< notice note >}}
If you are using Kind for your Kubernetes clusters, you will need to add the argument `--api-server-address` to the `meshctl register cluster` command. The value will depend on what operating system your are using. For macOS the value is `host.docker.internal`. For Linux, you can find the address by running the following: `docker exec "${cluster_name}-control-plane" ip addr show dev eth0 | sed -nE 's|\s*inet\s+([0-9.]+).*|\1|p')` and appending `:6443` for the port.
{{< /notice >}}

You can validate the registration by looking at the Custom Resource created in the management cluster:

```shell
kubectl get kubernetescluster -n gloo-mesh remote-cluster

NAME             AGE
remote-cluster   10m
```

The target cluster will also have the following elements created:

```shell
kubectl get sa --context $REMOTE_CONTEXT -n gloo-mesh

NAME             SECRETS   AGE
cert-agent       1         11m
default          1         11m
remote-cluster   1         11

kubectl get clusterrole --context $REMOTE_CONTEXT gloomesh-remote-access

NAME                     AGE
gloomesh-remote-access   12m

kubectl get clusterrolebinding --context $REMOTE_CONTEXT \
  remote-cluster-gloomesh-remote-access-clusterrole-binding

NAME                                                        AGE
remote-cluster-gloomesh-remote-access-clusterrole-binding   13m
```

#### Register the management cluster

You can automatically register the cluster on which you deploy Gloo Mesh (for example, if you have a mesh running there as well) with the `--register` CLI flag when you're first installing with `meshctl`:

```shell
meshctl install --register --context mgmt-cluster-context
```

By default, when you register like this, the cluster name will be `mgmt-cluster`. If you run the following, you should see the cluster registered:

```shell
kubectl get kubernetescluster -n gloo-mesh

NAMESPACE          NAME               AGE
gloo-mesh          mgmt-cluster       10s
```

{{< notice note >}}
If you are using Kind for your Kubernetes clusters, you will need to add the argument `--api-server-address` to the `meshctl install` command. The value will depend on what operating system your are using. For macOS the value is `host.docker.internal`. For Linux, you can find the address by running the following: `docker exec "${cluster_name}-control-plane" ip addr show dev eth0 | sed -nE 's|\s*inet\s+([0-9.]+).*|\1|p')` and appending `:6443` for the port.
{{< /notice >}}

If you choose not to register the management cluster during installation, you can always register the cluster using the `meshctl cluster register` command we used in the previous section.

```shell
MGMT_CONTEXT=your_management_plane_context

meshctl cluster register \
  --cluster-name mgmt-cluster \
  --remote-context $MGMT_CONTEXT

INFO[0003] successfully registered cluster mgmt-cluster
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