---
title: "Setup"
menuTitle: Setup
description: Setting up service mesh hub
weight: 30
---

Service Mesh Hub is a management plane that simplifies operations and workflows of service mesh installations across multiple clusters and deployment footprints. With Service Mesh Hub, you can install, discover, and operate a service-mesh deployment across your enterprise, deployed on premises, or in the cloud, even across heterogeneous service-mesh implementations.

## Assumptions

We will assume in this guide that we have access to two clusters and the following two contexts available in our kubeconfig file.



Your actual context names will likely be different.

* management-plane-context
    - kubeconfig context pointing to a cluster where we will install and operate Service Mesh Hub
* remote-cluster-context
    - kubeconfig context pointing to a cluster where we will install and manage a service mesh using Service Mesh Hub 


![Service Mesh Hub Architecture]({{% versioned_link_path fromRoot="/img/smh-2clusters.png" %}})

{{% notice note %}}
Note that these contexts need not be different; you may install and manage a service mesh in the same cluster as Service Mesh Hub. For the purposes of this guide, though, we will assume they are different.
{{% /notice %}}

## Guide

In this guide we will accomplish two tasks:

1. [Install Service Mesh Hub](#install-service-mesh-hub)
2. [Register A Cluster](#register-a-cluster)
 

### Install Service Mesh Hub

Ensure that your kubeconfig has the correct context set as its `currentContext`:

```shell
kubectl config use-context management-plane-context
```

Now we install Service Mesh Hub into our management plane cluster:

```shell
meshctl install
```

If that succeeds, you should see:

```shell
Service Mesh Hub has been installed to namespace service-mesh-hub
```

And that's it! Service Mesh Hub is ready to go. Once the pods have finished initializing, 
we can easily validate that the application is in a good state:

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

### Register A Cluster

In order to mark a cluster as being managed by Service Mesh Hub, we have to *register* it in
our installation. This is both so that we are aware of it, and so that we have the proper credentials
to communicate with the Kubernetes API server in that cluster.

We will register the context pointed to by our `remote-cluster-context` kube context:

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

To go into slightly more detail about what just happened:

* A service account was created in the `service-mesh-hub` namespace of the remote cluster
* That service account's auth token was stored in a secret in the management plane cluster
* The Service Mesh Hub CSR agent was deployed in the remote cluster
* Future communications that Service Mesh Hub does to the remote cluster's Kubernetes API server
 will be done using the service account auth token created in the first bullet point

And we're done! Any meshes in that cluster will be discovered and available to be configured at
this point. See the guide on [installing Istio]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}),
to see how to easily get Istio running on that cluster.
