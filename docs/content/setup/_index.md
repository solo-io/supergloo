---
title: "Setup"
menuTitle: Setup
description: Setting up service mesh hub
weight: 30
---

In this guide we will accomplish two tasks:

1. [Install Service Mesh Hub](#install-service-mesh-hub)
2. [Register A Cluster](#register-a-cluster)
 

We use a Kubernetes cluster to host the management plane (Service Mesh Hub) while each service mesh can run on its own independent cluster. If you don't have access to multiple clusters, see the [Getting Started Guide]({{% versioned_link_path fromRoot="/getting_started/" %}}) to get started with Kubernetes in Docker. 

![Service Mesh Hub Architecture]({{% versioned_link_path fromRoot="/img/smh-3clusters.png" %}})

You can install Service Mesh Hub onto its own cluster and register remote clusters, or you can co-locate Service Mesh Hub onto a cluster with a service mesh. The former (its own cluster) is the preferred deployment pattern, but for getting started, exploring, or to save resources, you can use the co-located deployment approach.

![Service Mesh Hub Architecture]({{% versioned_link_path fromRoot="/img/smh-2clusters.png" %}})

We will assume in this guide that we have access to two clusters and the following two contexts available in our `kubeconfig` file. 

Your actual context names will likely be different.

* `management-plane-context`
    - kubeconfig context pointing to a cluster where we will install and operate Service Mesh Hub
* `remote-cluster-context`
    - kubeconfig context pointing to a cluster where we will install and manage a service mesh using Service Mesh Hub 

To verify you're running the following commands in the correct context, run:

```shell
kubectl config use-context management-plane-context
```

If you're using only one cluster, see below.

## Install Service Mesh Hub

{{% notice note %}}
Note that these contexts need not be different; you may install and manage a service mesh in the same cluster as Service Mesh Hub. For the purposes of this guide, though, we will assume they are different.
{{% /notice %}}

### Installing with `meshctl`

`meshctl` is a CLI tool that helps bootstrap Service Mesh Hub, register clusters, install meshes, and more. Get the latest `meshctl` from the [releases page on solo-io/service-mesh-hub](https://github.com/solo-io/service-mesh-hub/releases).

You can also quickly install like this:

```shell
curl -sL https://run.solo.io/meshctl/install | sh
```

Once you have the `meshctl` tool, you can install Service Mesh Hub onto a cluster acting as the `management-plane-context` like this:


```shell
meshctl install
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
Note that the `--dry-run` outputs the entire `yaml`, but does not take care of proper ordering of resources. For example, there can be a *race* between CRDs being registered and any CRs being created that may appear to be an error. If that happens, just re-run the `kubectl apply`.
{{% /notice %}}

To undo the installation, run:

```shell
meshctl install --dry-run | kubectl delete -f -
```

### Install with Helm

The Helm charts for Service Mesh Hub support Helm 3. To install with Helm:

```shell
helm repo add smh https://storage.googleapis.com/service-mesh-hub/management-plane
helm repo update
```

{{% notice note %}}
Note that the location of the Service Mesh Hub Helm charts is subject to change. When it finds a more permanent home, we'll remove this message.
{{% /notice %}}

Then install Service Mesh Hub into the `service-mesh-hub` namespace:

```shell
helm install smh smh/service-mesh-hub —-namespace service-mesh-hub
```

### Verify install
Once you've installed Service Mesh Hub, verify what components got installed:

```shell
kubectl get po -n service-mesh-hub
NAME                               READY   STATUS    RESTARTS   AGE
mesh-discovery-7796c6bd6c-fwtck    1/1     Running   0          36s
mesh-networking-68fbf6c455-jrdbx   1/1     Running   0          36s
```

Running the check command will verify everything was installed correctly:

```shell
meshctl check
```

```shell
✅ Kubernetes API
-----------------
✅ Kubernetes API server is reachable
✅ running the minimum supported Kubernetes version (required: >=1.13)


✅ Service Mesh Hub Management Plane
------------------------------------
✅ installation namespace exists
✅ components are running


✅ Service Mesh Hub check found no errors
```

At this point you're ready to add clusters to the management plane, or discover existing service meshes on the cluster on which we just deployed Service Mesh Hub. 


## Register A Cluster

In order to identify a cluster as being managed by Service Mesh Hub, we have to *register* it in
our installation. This is both so that we are aware of it, and so that we have the proper credentials
to communicate with the Kubernetes API server in that cluster.

#### Remote Clusters
For remote clusters, we will register with the `meshctl cluster register` command. We register the context pointed to by our `remote-cluster-context` kubeconfig context like this:

```shell
meshctl cluster register \
  --remote-cluster-name new-remote-cluster \
  --remote-context remote-cluster-context
```

{{% notice note %}}
Note that the `--remote-cluster-name` is NOT the name of the cluster in your kubeconfig file -- it's a name given to the cluster Service Mesh Hub can refer to it in various configurations. You can pick a name for this.
{{% /notice %}}

```shell
Successfully wrote service account to remote cluster...
Successfully wrote kube config secret to master cluster...
Successfully set up CSR agent...

Cluster new-remote-cluster is now registered in your Service Mesh Hub installation
```

#### Register the management cluster

You can automatically register the cluster on which you deploy Service Mesh Hub (for example, if you have a mesh running there as well) with the `--register` CLI flag when you're first installing with `meshctl`:

```shell
meshctl install --register --context management-plane-context
```

By default, when you register like this, the cluster name will be `management-plane`. If you run the following, you should see the cluster registered:

```shell
kubectl get kubernetescluster -n service-mesh-hub

NAMESPACE          NAME               AGE
service-mesh-hub   management-plane   10s
```

## What happened?

To go into slightly more detail about what just happened:

* A service account was created in the `service-mesh-hub` namespace of the remote cluster
* That service account's auth token was stored in a secret in the management plane cluster
* The Service Mesh Hub CSR agent was deployed in the remote cluster
* Future communications that Service Mesh Hub does to the remote cluster's Kubernetes API server
 will be done using the service account auth token created in the first bullet point

And we're done! Any meshes in that cluster will be discovered and available to be configured at
this point. See the guide on [installing Istio]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}),
to see how to easily get Istio running on that cluster.
