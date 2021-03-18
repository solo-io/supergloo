---
title: "Enterprise"
menuTitle: Enterprise
description: Registering a cluster with Gloo Mesh enterprise edition
weight: 30
---

{{% notice note %}} Gloo Mesh Enterprise is required for this feature. {{% /notice %}}

Relay is an alternative mode of deploying Gloo Mesh that confers several advantages discussed in [this document]({{% versioned_link_path fromRoot="/concepts/relay" %}}).
Cluster registration in relay mode simply consists of installing the relay agent.

This guide will walk you through the basics of registering clusters using the `meshctl` tool or with Helm. We will be using the two cluster contexts mentioned in the Gloo Mesh installation guide, `kind-mgmt-cluster` and `kind-remote-cluster`. Your cluster context names may differ, so please substitute the proper values.

## Register A Cluster

In order to identify a cluster as being managed by Gloo Mesh Enterprise, we have to *register* it in our installation. Registration ensures we are aware of the cluster, and we have properly configured a remote relay agent to talk to the local relay server. In this example, we will register our remote cluster with Gloo Mesh Enterprise running on the management cluster.

### Register with `meshctl`

We can use the CLI tool `meshctl` to register our remote cluster. The command we use will be `meshctl cluster register enterprise`. This is specific to Gloo Mesh **Enterprise**, and different in nature than the `meshctl cluster register community` command.

To register our remote cluster, there are a few key pieces of information we need:

1. `**cluster name**` - The name we would like to register the cluster with.
1. `**remote-context**` - The Kubernetes context with access to the remote cluster being registered.
1. `**relay-server-address**` - The address of the relay server running on the management cluster.

Following the [Gloo Mesh Enterprise prerequisites guide]({{% versioned_link_path fromRoot="/setup/enterprise_prerequisites" %}}), you should already have a virtual service for the relay server exposed on an ingress gateway. Assuming you are using kind and Istio, you can retrieve the ingress address by running the following commands.

```shell
MGMT_CONTEXT=kind-mgmt-cluster # Update value as needed
kubectl config use-context $MGMT_CONTEXT

mgmtIngressAddress=$(kubectl get node -ojson | jq -r ".items[0].status.addresses[0].address")
mgmtIngressPort=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="https")].nodePort}')
ingressAddress=${mgmtIngressAddress}:${mgmtIngressPort}
```

Let's set variables for the remaining values:

```bash
CLUSTER_NAME=remote-cluster
REMOTE_CONTEXT=kind-remote-cluster # Update value as needed
```

Now we are ready to register the remote cluster:

```shell
meshctl cluster register enterprise \
  --remote-context=$REMOTE_CONTEXT \
  --relay-server-address $ingressAddress \
  $CLUSTER_NAME
```

You should see the following output:

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

### Register with Helm

You can also register a remote cluster using the Enterprise Agent Helm repository. The same information used for `meshctl` registration will be needed here as well. This portion assumes you have run through the enterprise prerequisites, which includes the following actions on the remote cluster:

* Creating the `gloo-mesh` namespace
* Copying over the self-signed root CA certificate from the management cluster (`relay-root-tls-secret`)
* Copy over the validation token for relay agent initialization (`relay-identity-token-secret`)

If you have not followed these steps, the relay agent deployment will fail.

You can add the repository by running the following:

```shell
helm repo add gloo-mesh-enterprise https://storage.googleapis.com/gloo-mesh-enterprise/gloo-mesh-enterprise
helm repo update
```

#### Install the Enterprise Agent

Install the Enterprise Agent from the Helm repository located at `https://storage.googleapis.com/gloo-mesh-enterprise/enterprise-agent`.
Make sure to review the Helm values options before installing. Some notable values include:

* `relay.cluster` will be the name by which the cluster is referenced in all Gloo Mesh configuration.
* `relay.serverAddress` is the address by which the Gloo Mesh management plane can be accessed.
* `relay.authority` is the host header that will be passed to the server on the Gloo Mesh management plane.

Also note that the Enterprise Agent's version should match that of the `enterprise-networking` component running on the
management cluster. Run `meshctl version` on the management cluster to review the `enterprise-networking` version.

First we will get the ingress address for the relay server.

```shell
MGMT_CONTEXT=kind-mgmt-cluster # Update value as needed
kubectl config use-context $MGMT_CONTEXT

mgmtIngressAddress=$(kubectl get node -ojson | jq -r ".items[0].status.addresses[0].address")
mgmtIngressPort=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="https")].nodePort}')
ingressAddress=${mgmtIngressAddress}:${mgmtIngressPort}
```

Then we will set our variables:

```shell
CLUSTER_NAME=remote-cluster
REMOTE_CONTEXT=kind-remote-cluster # Update value as needed
ENTERPRISE_NETWORKING_VERSION=<current version> # Update based on meshctl version output
```

And now we will deploy the relay agent in the remote cluster.


```bash
helm install enterprise-agent enterprise-agent/enterprise-agent \
  --namespace gloo-mesh \
  --set relay.serverAddress=${ingressAddress} \
  --set relay.authority=enterprise-networking.gloo-mesh \
  --set relay.cluster=${CLUSTER_NAME} \
  --kube-context=${REMOTE_CONTEXT} \
  --version ${ENTERPRISE_NETWORKING_VERSION}
```

#### Add a Kubernetes Cluster Object

We've successfully deployed the relay agent in the remote cluster. Now we need to add a `KubernetesCluster` object to make the relay server aware of the remote cluster. The `metadata.name` of the object must match the value passed for `relay.cluster` in the Helm chart above. The `spec.clusterDomain` must 
match the [local cluster domain](https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/) of the Kubernetes cluster.


```shell
kubectl apply -f- <<EOF
apiVersion: multicluster.solo.io/v1alpha1
kind: KubernetesCluster
metadata:
  name: remote-cluster # Update value as needed
  namespace: gloo-mesh
spec:
  clusterDomain: cluster.local
EOF
```


