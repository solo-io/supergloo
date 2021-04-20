---
title: "Bootstrap Gloo Mesh on Kind"
menuTitle: Enterprise
description: Get started running Gloo Mesh or Gloo Mesh Enterprise locally in Kind.
weight: 20
---

The easiest way to get started with Gloo Mesh is by using Kind to run local Kubernetes clusters in Docker. There is a `demo` command in meshctl that will create a full demonstration environment on your local system. All you need is Docker, Kind, and kubectl installed. 

* [Docker](https://www.docker.com/products/docker-desktop) for desktop, with at least 8GB of RAM allocated
* [Kind](https://kind.sigs.k8s.io) Kubernetes in Docker
* [istioctl](https://istio.io/latest/docs/setup/getting-started/#download) Command line utility for Istio

If you prefer to use an existing Kubernetes cluster, check out our [Setup Guide]({{% versioned_link_path fromRoot="/setup/" %}}).

To spin up two Kubernetes clusters with Kind, run:

{{< tabs >}}
{{< tab name="Community" codelang="shell" >}}
meshctl demo istio-multicluster init
{{< /tab >}}
{{< tab name="Enterprise" codelang="shell" >}}
GLOO_MESH_LICENSE_KEY=<your_key_here> # You'll need to supply your own key
meshctl demo istio-multicluster init --enterprise --license $GLOO_MESH_LICENSE_KEY
{{< /tab >}}
{{< /tabs >}}

The command will do the following:

* Create two kind clusters: `mgmt-cluster` and `remote-cluster`
* Install Gloo Mesh on the `mgmt-cluster`
* Install Istio on both clusters
* Register both clusters with Gloo Mesh under the names `mgmt-cluster` and `remote-cluster`
* Deploy BookInfo sample application on both clusters under the `bookinfo` namespace

```shell
Creating cluster mgmt-cluster with ingress port 32001
Creating cluster "mgmt-cluster" ...
 ✓ Ensuring node image (kindest/node:v1.17.5) �
 ✓ Preparing nodes ��
 ✓ Writing configuration ��
 ✓ Starting control-plane 🕹
 ✓ Installing CNI ��
 ✓ Installing StorageClass ��
Set kubectl context to "kind-mgmt-cluster"
You can now use your cluster with:

kubectl cluster-info --context kind-mgmt-cluster

...

Creating cluster remote-cluster with ingress port 32000
Creating cluster "remote-cluster" ...
 ✓ Ensuring node image (kindest/node:v1.17.5) �
 ✓ Preparing nodes ��
 ✓ Writing configuration ��
 ✓ Starting control-plane 🕹
 ✓ Installing CNI ��
 ✓ Installing StorageClass ��
Set kubectl context to "kind-remote-cluster"
You can now use your cluster with:

kubectl cluster-info --context kind-remote-cluster
```

To connect to each of the clusters, run the following:

```shell
export MGMT_CONTEXT=kind-mgmt-cluster
export REMOTE_CONTEXT=kind-remote-cluster
```

Then you can run the following to connect to the mgmt-cluster cluster:

```shell
kubectl --context $MGMT_CONTEXT get po -n gloo-mesh
```

You should see Gloo Mesh installed:

```shell
NAME                              READY   STATUS    RESTARTS   AGE
csr-agent-8445578f6d-6hzls        1/1     Running   0          3m28s
mesh-discovery-8657d4dd66-dlks8   1/1     Running   0          3m32s
mesh-networking-58b68b7b6-ljjcr   1/1     Running   0          3m32s
```

To verify the installation came up successfully and everything is in a good state:

```shell
meshctl check
```


You should see something similar to the following:

```shell
Gloo Mesh
-------------------
✅ Gloo Mesh pods are running

Management Configuration
---------------------------
✅ Gloo Mesh networking configuration resources are in a valid state
```

Setting up Kind and multiple clusters on your machine isn't always the easiest, and there may be some issues/hurdles you run into, especially on "company laptops" with extra security constraints. If you ran into any issues in the previous steps, please join us on the [Solo.io slack](https://slack.solo.io) and we'll be more than happy to help troubleshoot. 

## Next steps

In this quick-start guide, we installed Gloo Mesh and registered clusters. If these installation use cases were too simplistic or not representative of your environment, please check out our [Setup Guide]({{% versioned_link_path fromRoot="/setup/" %}}). Otherwise, please check out our [Guides]({{% versioned_link_path fromRoot="/guides/" %}}) to explore the power of Gloo Mesh.

### Clean up

Cleaning up this demo environment is as simple as running the following:

```shell
meshctl demo istio-multicluster cleanup
```