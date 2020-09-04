---
title: "Getting Started"
menuTitle: Getting Started
description: How to get started using Service Mesh Hub
weight: 10
---

Welcome to Service Mesh Hub, the open-source, multi-cluster, multi-mesh management plane. Service Mesh Hub simplifies service-mesh operations and lets you manage multiple clusters of a service mesh from a centralized management plane. Service Mesh Hub takes care of things like shared-trust/root CA federation, workload discovery, unified multi-cluster/global traffic policy, access policy, and more. 

## Getting `meshctl`

Service Mesh Hub has a CLI tool called `meshctl` that helps bootstrap Service Mesh Hub, register clusters, describe configured resources, and more. Get the latest `meshctl` from the [releases page on solo-io/service-mesh-hub](https://github.com/solo-io/service-mesh-hub/releases).

You can also quickly install like this:

```shell
curl -sL https://run.solo.io/meshctl/install | sh
```

Once you've downloaded the correct binary for your architecture, run the following to make sure it's working correctly:

```shell
meshctl version
```

You can add `meshctl` to your path for global access on the command line. See:


* [Adding to your path on Windows](https://helpdeskgeek.com/windows-10/add-windows-path-environment-variable/)
* [Adding to your path on Mac](https://osxdaily.com/2014/08/14/add-new-path-to-path-command-line/)
* [Adding to your path on Linux](https://linuxize.com/post/how-to-add-directory-to-path-in-linux/)


## Spinning up clusters with Kind (Kubernetes in Docker)

The easiest way to get started is by using Kind to run local Kubernetes clusters in Docker. There is a `demo` command in meshctl that will create a full demonstration environment on your local system. All you need is Docker, Kind, and kubectl installed. 

* [Docker](https://www.docker.com/products/docker-desktop) for desktop
* [Kind](https://kind.sigs.k8s.io) Kubernetes in Docker

If you prefer to use an existing Kubernetes cluster, check out our [Setup Guide]({{% versioned_link_path fromRoot="/setup/" %}}).

To spin up two Kubernetes clusters with Kind, run:

```shell
meshctl demo istio-multicluster init
```
This will spin up two Kubernetes clusters in Docker with Istio installed on each. Additionally, this will install Service Mesh Hub on one of the clusters. Both clusters will be **registered** with Service Mesh Hub under the names `mgmt-cluster` and `remote-cluster`, which will be used throughout the documentation.

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
export MGMT_PLANE_CTX=kind-mgmt-cluster
export REMOTE_CTX=kind-remote-cluster
```

Then you can run the following to connect to the mgmt-cluster cluster:

```shell
kubectl --context $MGMT_PLANE_CTX get po -n service-mesh-hub
```

You should see Service Mesh Hub installed:

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


You should see something similar:

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

Setting up Kind and multiple clusters on your machine isn't always the easiest, and there may be some issues/hurdles you run into, especially on "company laptops" with extra security constraints. If you ran into any issues in the previous steps, please join us on the [Solo.io slack](https://slack.solo.io) and we'll be more than happy to help troubleshoot. 

You should be ready to run the steps in the rest of the [Guides]({{% versioned_link_path fromRoot="/guides/" %}}).

### Clean up

Cleaning up this demo environment is as simple as running the following:

```shell
meshctl demo istio-multicluster cleanup
```

## Next steps

In this quick-start guide, we installed Service Mesh Hub. If these installation use cases were too simplistic or not representative of your environment, please check out our [Setup Guide]({{% versioned_link_path fromRoot="/setup/" %}}). Otherwise, please check out our [Guides]({{% versioned_link_path fromRoot="/guides/" %}}) to explore the power of Service Mesh Hub.
