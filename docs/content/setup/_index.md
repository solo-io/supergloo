---
title: "Setup"
menuTitle: Setup
description: Setting up Gloo Mesh
weight: 30
---

In this guide we will accomplish two tasks:

1. [Install Gloo Mesh](#install-gloo-mesh)
2. [Register A Cluster](#register-a-cluster)
 

We use a Kubernetes cluster to host the management plane (Gloo Mesh) while each service mesh can run on its own independent cluster. If you don't have access to multiple clusters, see the [Getting Started Guide]({{% versioned_link_path fromRoot="/getting_started/" %}}) to get started with Kubernetes in Docker. 

![Gloo Mesh Architecture]({{% versioned_link_path fromRoot="/img/gloomesh-3clusters.png" %}})

You can install Gloo Mesh onto its own cluster and register remote clusters, or you can co-locate Gloo Mesh onto a cluster with a service mesh. The former (its own cluster) is the preferred deployment pattern, but for getting started, exploring, or to save resources, you can use the co-located deployment approach.

![Gloo Mesh Architecture]({{% versioned_link_path fromRoot="/img/gloomesh-2clusters.png" %}})


## Assumptions for setup

We will assume in this guide that we have access to two clusters and the following two contexts available in our `kubeconfig` file. 

Your actual context names will likely be different.

* `mgmt-cluster-context`
    - kubeconfig context pointing to a cluster where we will install and operate Gloo Mesh
* `remote-cluster-context`
    - kubeconfig context pointing to a cluster where we will install and manage a service mesh using Gloo Mesh 

To verify you're running the following commands in the correct context, run:

```shell
MGMT_CONTEXT=your_management_plane_context
REMOTE_CONTEXT=your_remote_context

kubectl config use-context $MGMT_CONTEXT
```

### Using Kind

If you do not have access to two Kubernetes clusters, you can easily create two on your local workstation using Kind. Simply run the following commands to create the mgmt-cluster and remote-cluster clusters.

```bash
#Set version, cluster name, and port
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
        node-labels: "ingress-ready=true"
kubeadmConfigPatches:
- |
  kind: InitConfiguration
  nodeRegistration:
    kubeletExtraArgs:
      authorization-mode: "AlwaysAllow"
EOF

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
        node-labels: "ingress-ready=true"
kubeadmConfigPatches:
- |
  kind: InitConfiguration
  nodeRegistration:
    kubeletExtraArgs:
      authorization-mode: "AlwaysAllow"
EOF

#Switch to the mgmt-cluster context
kubectl config use-context kind-mgmt-cluster
```

## Install Gloo Mesh

{{% notice note %}}
Note that these contexts need not be different; you may install and manage a service mesh in the same cluster as Gloo Mesh. For the purposes of this guide, though, we will assume they are different.
{{% /notice %}}

### Installing with `meshctl`

`meshctl` is a CLI tool that helps bootstrap Gloo Mesh, register clusters, describe configured resources, and more. Get the latest `meshctl` from the [releases page on solo-io/gloo-mesh](https://github.com/solo-io/gloo-mesh/releases).

You can also quickly install like this:

```shell
curl -sL https://run.solo.io/meshctl/install | sh
```

Once you have the `meshctl` tool, you can install Gloo Mesh onto a cluster acting as the `mgmt-cluster` like this:


```shell
meshctl install
```

If you're not using the context for the `mgmt-cluster`, you can explicitly specify it like this:

```shell
meshctl install --kubecontext mgmt-cluster-context
```

You should see output similar to this:

```shell
Creating namespace gloo-mesh... Done.
Starting Gloo Mesh installation...
Gloo Mesh successfully installed!
Gloo Mesh has been installed to namespace gloo-mesh
```

To undo the installation, run `uninstall`:

```shell
meshctl uninstall
```

### Installing with `kubectl apply`

If you prefer working directly with the Kubernetes resources, (either to use `kubectl apply` or to put into CI/CD), `meshctl` can output `yaml` from the `install` (or any) command with the `--dry-run` flag:

```shell
meshctl install --dry-run
```

You can use this output to later do `kubectl apply`:

```shell
meshctl install --dry-run | kubectl --context apply -f -
```

{{% notice note %}}
Note that the `--dry-run` outputs the entire `yaml`, but does not take care of proper ordering of resources. For example, there can be a *race* between Custom Resource Definitions being registered and any Custom Resources being created that may appear to be an error. If that happens, just re-run the `kubectl apply`.
{{% /notice %}}

To undo the installation, run:

```shell
meshctl install --dry-run | kubectl delete -f -
```

### Install with Helm

The Helm charts for Gloo Mesh support Helm 3. To install with Helm:

```shell
helm repo add gloo-mesh https://storage.googleapis.com/gloo-mesh/gloo-mesh
helm repo update
```

{{% notice note %}}
Note that the location of the Gloo Mesh Helm charts is subject to change. When it finds a more permanent home, we'll remove this message.
{{% /notice %}}

Then install Gloo Mesh into the `gloo-mesh` namespace:

```shell
kubectl create ns gloo-mesh
helm install gloo-mesh gloo-mesh/gloo-mesh --namespace gloo-mesh
```

### Verify install
Once you've installed Gloo Mesh, verify what components got installed:

```shell
kubectl get po -n gloo-mesh

NAME                          READY   STATUS    RESTARTS   AGE
discovery-66675cf6fd-cdlpq    1/1     Running   0          32m
networking-6d7686564d-ngrdq   1/1     Running   0          32m
```

Running the check command will verify everything was installed correctly:

```shell
meshctl check
```

```shell
Gloo Mesh
-------------------
✅ Gloo Mesh pods are running

Management Configuration
---------------------------
✅ Gloo Mesh networking configuration resources are in a valid state
```

At this point you're ready to add clusters to the management plane, or discover existing service meshes on the cluster on which we just deployed Gloo Mesh. 


## Register A Cluster

In order to identify a cluster as being managed by Gloo Mesh, we have to *register* it in
our installation. This is both so that we are aware of it, and so that we have the proper credentials
to communicate with the Kubernetes API server in that cluster.

#### Remote Clusters
For remote clusters, we will register with the `meshctl cluster register` command. We register the context pointed to by our `remote-cluster-context` kubeconfig context like this:

```shell
meshctl cluster register \
  --remote-cluster-name remote-cluster \
  --remote-context remote-cluster-context
```

{{% notice note %}}
Note that the `--remote-cluster-name` is NOT the name of the cluster in your kubeconfig file -- it's a name given to the cluster Gloo Mesh can refer to it in various configurations. You can pick a name for this.
{{% /notice %}}

```shell
Successfully wrote service account to remote cluster...
Successfully wrote kube config secret to master cluster...
Successfully set up CSR agent...

Cluster remote-cluster is now registered in your Gloo Mesh installation
```

{{< notice note >}}
If you are using Kind for your Kubernetes clusters, you will need to add the argument `--api-server-address` to the `meshctl register cluster` command. The value will depend on what operating system your are using. For macOS the value is `host.docker.internal`. For Linux, you can find the address by running the following: `docker exec "${cluster_name}-control-plane" ip addr show dev eth0 | sed -nE 's|\s*inet\s+([0-9.]+).*|\1|p')` and appending `:6443` for the port.
{{< /notice >}}

#### Register the management cluster

You can automatically register the cluster on which you deploy Gloo Mesh (for example, if you have a mesh running there as well) with the `--register` CLI flag when you're first installing with `meshctl`:

```shell
meshctl install --register --context mgmt-cluster-context
```

By default, when you register like this, the cluster name will be `mgmt-cluster`. If you run the following, you should see the cluster registered:

```shell
kubectl get kubernetescluster -n gloo-mesh

NAMESPACE          NAME               AGE
gloo-mesh   mgmt-cluster       10s
```

{{< notice note >}}
If you are using Kind for your Kubernetes clusters, you will need to add the argument `--api-server-address` to the `meshctl install` command. The value will depend on what operating system your are using. For macOS the value is `host.docker.internal`. For Linux, you can find the address by running the following: `docker exec "${cluster_name}-control-plane" ip addr show dev eth0 | sed -nE 's|\s*inet\s+([0-9.]+).*|\1|p')` and appending `:6443` for the port.
{{< /notice >}}

## What happened?

To go into slightly more detail about what just happened:

* A service account was created in the `gloo-mesh` namespace of the remote cluster
* That service account's auth token was stored in a secret in the management plane cluster
* The Gloo Mesh CSR agent was deployed in the remote cluster
* Future communications that Gloo Mesh does to the remote cluster's Kubernetes API server
 will be done using the service account auth token created in the first bullet point

## Next Steps

And we're done! Any meshes in that cluster will be discovered and available to be configured at this point. See the guide on [installing Istio]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}), to see how to easily get Istio running on that cluster.
