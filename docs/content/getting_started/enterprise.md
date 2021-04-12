---
title: "Getting Started with Gloo Mesh Enterprise"
menuTitle: Enterprise
description: How to get started using Gloo Mesh Enterprise
weight: 10
---

The following guide describes how to get started with Gloo Mesh Enterprise on a managed Kubernetes environment such as GKE or EKS.

## Before we begin

- meshctl
- 3 clusters with contexts MGMT_CONTEXT, REMOTE_CONTEXT1, REMOTE_CONTEXT2, where MGMT_CONTEXT points to the cluster that will
run the Gloo Mesh management plane and REMOTE_CONTEXT1/2 point to clusters that are running Istio and application workloads. If
desired, the management cluster can also run a service mesh and workloads to be discovered and managed by Gloo Mesh.
- license key at GLOO_MESH_LICENSE_KEY

TODO joekelley diagram

## TODO joekelley install istio

## Installing the management components

Installing Gloo Mesh Enterprise with `meshctl` is a simple process. You will use the command `meshctl install enterprise` and supply the license key, as well as any chart values you want to update, and arguments pointing to the cluster where Gloo Mesh Enterprise will be installed. For our example, we are going to install Gloo Mesh Enterprise on the cluster `mgmt-cluster`. First, let's set a variable for the license key.

```shell
GLOO_MESH_LICENSE_KEY=<your_key_here> # You'll need to supply your own key
```

If you are running Gloo Mesh Enterprise's management plane on a cluster you intend to register (i.e. also run a service mesh), set the `enterprise-networking.cluster` value to the cluster name you intend to set for the management cluster at registration time.

TODO joekelley change default svc type to load balancer

```shell
meshctl install enterprise --license $GLOO_MESH_LICENSE_KEY
```

You should see the following output from the command:

```shell
Installing Helm chart
Finished installing chart 'gloo-mesh-enterprise' as release gloo-mesh:gloo-mesh
```

The installer has created the namespace `gloo-mesh` and installed Gloo Mesh Enterprise into the namespace using a Helm chart with default values.

### Verify install
Once you've installed Gloo Mesh, verify that the following components were installed:

```shell
kubectl get pods -n gloo-mesh
```

```shell
NAME                                     READY   STATUS    RESTARTS   AGE
dashboard-6d6b944cdb-jcpvl               3/3     Running   0          4m2s
enterprise-networking-84fc9fd6f5-rrbnq   1/1     Running   0          4m2s
rbac-webhook-84865cb7dd-sbwp7            1/1     Running   0          4m2s
```

Running the check command from meshctl will also verify everything was installed correctly:

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

## Register your remote clusters

In order to register your remote clusters with the Gloo Mesh management plane, you'll need to know the external address
of the `enterprise-networking` service. ** TODO joekelley link to relay architecture description. ** Because the service
is of type LoadBalancer by default, your cloud provider will expose the service outside the cluster. You can determine
the public address of the service with the following:

```shell
ENTERPRISE_NETWORKING_IP=$(kubectl  -n gloo-mesh get service enterprise-networking -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
ENTERPRISE_NETOWRKING_PORT=$(kubectl -n gloo-mesh get service enterprise-networking -o jsonpath='{.spec.ports[?(@.name=="grpc")].port}')
ENTERPRISE_NETOWRKING_ADDRESS=${ENTERPRISE_NETWORKING_IP}:${ENTERPRISE_NETOWRKING_PORT}
```

To register each cluster, run:

```shell
meshctl cluster register enterprise \
  --remote-context=$REMOTE_CONTEXT1 \
  --relay-server-address $ENTERPRISE_NETWORKING_ADDRESS \
  cluster1

meshctl cluster register enterprise \
  --remote-context=$REMOTE_CONTEXT2 \
  --relay-server-address $ENTERPRISE_NETWORKING_ADDRESS \
  cluster2
```

{{% notice note %}}Ensure that the `gloo-mesh` namespace in each remote cluster is not being injected by Istio.{{% /notice %}}

For each cluster, you should see the following:

```shell
Registering cluster
📃 Copying root CA relay-root-tls-secret.gloo-mesh to remote cluster from management cluster
📃 Copying bootstrap token relay-identity-token-secret.gloo-mesh to remote cluster from management cluster
💻 Installing relay agent in the remote cluster
Finished installing chart 'enterprise-agent' as release gloo-mesh:enterprise-agent
📃 Creating remote-cluster KubernetesCluster CRD in management cluster
⌚ Waiting for relay agent to have a client certificate
         Checking...
         Checking...
🗑 Removing bootstrap token
✅ Done registering cluster!
```



