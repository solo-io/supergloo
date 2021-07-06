---
title: "Enterprise - Insecure"
menuTitle: Enterprise - Insecure
description: Registering a cluster insecurely with Gloo Mesh enterprise edition
weight: 30
---

{{% notice note %}} Gloo Mesh Enterprise version v1.1.0-beta20 or above is required for this feature. {{% /notice %}}
{{% notice warning %}} This guide configures Gloo Mesh Enterprise without secure communication. This is not advised for production usage. {{% /notice %}}

This guide will walk you through the basics of registering clusters for management by Gloo Mesh Enterprise using Helm without any secure communication. This is not fit for production, but provides an easy setup, that would allow you to quickly evaluate Gloo Mesh before investing in securing the communication.

## Install Gloo Mesh in Insecure mode

```shell
MGMT_CONTEXT=kind-cluster-1 # Update value as needed
kubectl create namespace gloo-mesh
helm install gloo-mesh-enterprise gloo-mesh-enterprise/gloo-mesh-enterprise --namespace gloo-mesh \
  --set global.insecure=true \
  --set licenseKey=${GLOO_MESH_LICENSE_KEY}
  --kube-context=${MGMT_CONTEXT} \
```

## Register A Cluster Insecurely

In order to identify a cluster as being managed by Gloo Mesh Enterprise, we have to *register* it in our installation. Registration ensures we are aware of the cluster, and we have properly configured a remote [relay]({{% versioned_link_path fromRoot="/concepts/relay" %}}) *agent* to talk to the local relay *server*. In this example, we will register our remote cluster with Gloo Mesh Enterprise running on the management cluster.

### Register with `meshctl`

We can use the CLI tool `meshctl` to register our remote cluster. The command we use will be `meshctl cluster register enterprise`. This is specific to Gloo Mesh **Enterprise**, and different in nature than the `meshctl cluster register community` command.

Registering with `meshctl` in insecure mode is very similar to the registration in secure mode [see here]({{% versioned_link_path fromRoot="/setup/cluster_registration/enterprise_cluster_registration/#register-with-meshctl" %}}). We just need to add the `--relay-server-insecure=true` flag to the command line.

For example:

```shell
meshctl cluster register enterprise \
  --remote-context=$REMOTE_CONTEXT \
  --relay-server-address $RELAY_ADDRESS \
  --relay-server-insecure=true \
  $CLUSTER_NAME
```

You should see the following output:

```
Registering cluster
💻 Installing relay agent in the remote cluster
Finished installing chart 'enterprise-agent' as release gloo-mesh:enterprise-agent
📃 Creating remote-cluster KubernetesCluster CRD in management cluster
✅ Done registering cluster!
```

The `meshctl` command accomplished the following activities:

* Created the `gloo-mesh` namespace
* Installed the relay agent in the remote cluster
* Created the KubernetesCluster CRD in the management cluster

When registering a remote cluster using Helm, you will need to run through these tasks yourself. The next section details how to accomplish those tasks and install the relay agent with Helm.

### Register with Helm Insecurely
You can also register a remote cluster using the Enterprise Agent Helm repository. The same information used for `meshctl` registration will be needed here as well. To install, you will need to create the `gloo-mesh` namespace.

#### Prerequisites

First create the namespace in the remote cluster:

```shell
CLUSTER_NAME=cluster-2 # Update value as needed
REMOTE_CONTEXT=kind-cluster-2 # Update value as needed

kubectl create ns gloo-mesh --context $REMOTE_CONTEXT
```

#### Install the Enterprise Agent

We are going to install the Enterprise Agent from the Helm repository located at `https://storage.googleapis.com/gloo-mesh-enterprise/enterprise-agent`.
Make sure to review the Helm values options before installing. Some notable values include:

* `relay.cluster` will be the name by which the cluster is referenced in all Gloo Mesh configuration.
* `relay.serverAddress` is the address by which the Gloo Mesh management plane can be accessed.

Also note that the Enterprise Agent's version should match that of the `enterprise-networking` component running on the
management cluster. Run `meshctl version` on the management cluster to review the `enterprise-networking` version.

If you haven't already, you can add the repository by running the following:

```shell
helm repo add enterprise-agent https://storage.googleapis.com/gloo-mesh-enterprise/enterprise-agent
helm repo update
```

By default, the `enterprise-networking` service is of type LoadBalancer, and the cloud provider managing your Kubernetes cluster will automatically provision a public IP for the service. Get the complete `relay-server-address` with:

{{< tabs >}}
{{< tab name="IP LoadBalancer address (GKE)" codelang="yaml">}}
MGMT_INGRESS_ADDRESS=$(kubectl get svc -n gloo-mesh enterprise-networking -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
MGMT_INGRESS_PORT=$(kubectl -n gloo-mesh get service enterprise-networking -o jsonpath='{.spec.ports[?(@.name=="grpc")].port}')
RELAY_ADDRESS=${MGMT_INGRESS_ADDRESS}:${MGMT_INGRESS_PORT}
{{< /tab >}}
{{< tab name="Hostname LoadBalancer address (EKS)" codelang="shell" >}}
MGMT_INGRESS_ADDRESS=$(kubectl get svc -n gloo-mesh enterprise-networking -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
MGMT_INGRESS_PORT=$(kubectl -n gloo-mesh get service enterprise-networking -o jsonpath='{.spec.ports[?(@.name=="grpc")].port}')
RELAY_ADDRESS=${MGMT_INGRESS_ADDRESS}:${MGMT_INGRESS_PORT}
{{< /tab >}}
{{< /tabs >}}

If the above commands left you with a `$RELAY_ADDRESS` value that is empty or incomplete, make sure the `enterprise-networking`
service is available to clients outside the cluster, perhaps through a NodePort or ingress solution, and find the address
before continuing. 


Then we will set our variables:

```shell
CLUSTER_NAME=cluster-2 # Update value as needed
REMOTE_CONTEXT=kind-cluster-2 # Update value as needed
ENTERPRISE_NETWORKING_VERSION=<current version> # Update based on meshctl version output
```

{{% notice note %}} If you have `jq` installed, you can use this command to get the correct value for `ENTERPRISE_NETWORKING_VERSION`:

```shell
ENTERPRISE_NETWORKING_VERSION=$(meshctl version | jq '.server[].components[] | select(.componentName == "enterprise-networking") | .images[] | select(.name == "enterprise-networking") | .version')
```

 {{% /notice %}}

And now we will deploy the relay agent in the remote cluster.


```bash
helm install enterprise-agent enterprise-agent/enterprise-agent \
  --namespace gloo-mesh \
  --set relay.serverAddress=${RELAY_ADDRESS} \
  --set relay.cluster=${CLUSTER_NAME} \
  --set global.insecure=true \
  --kube-context=${REMOTE_CONTEXT} \
  --version ${ENTERPRISE_NETWORKING_VERSION}
```

#### Add a Kubernetes Cluster Object

We've successfully deployed the relay agent in the remote cluster. Now we need to add a `KubernetesCluster` object to the management cluster to make the relay server aware of the remote cluster. The `metadata.name` of the object must match the value passed for `relay.cluster` in the Helm chart above. The `spec.clusterDomain` must match the [local cluster domain](https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/) of the Kubernetes cluster.


```shell
kubectl apply --context $MGMT_CONTEXT -f- <<EOF
apiVersion: multicluster.solo.io/v1alpha1
kind: KubernetesCluster
metadata:
  name: ${CLUSTER_NAME}
  namespace: gloo-mesh
spec:
  clusterDomain: cluster.local
EOF
```

#### Validate the Registration

We can validate the registration process by first checking to make sure the relay agent pod and secrets have been created on the remote cluster:

```shell
kubectl get pods -n gloo-mesh --context $REMOTE_CONTEXT
```

```shell
NAME                                READY   STATUS    RESTARTS   AGE
enterprise-agent-64fc8cc9c5-v7b97   1/1     Running   7          25m

kubectl get secrets -n gloo-mesh --context $REMOTE_CONTEXT

NAME                                     TYPE                                  DATA   AGE
default-token-fcx9w                      kubernetes.io/service-account-token   3      18h
enterprise-agent-token-55mvq             kubernetes.io/service-account-token   3      25m
relay-client-tls-secret                  Opaque                                3      6m24s
relay-identity-token-secret              Opaque                                1      29m
relay-root-tls-secret                    Opaque                                1      18h
sh.helm.release.v1.enterprise-agent.v1   helm.sh/release.v1                    1      25m
```

We can check the logs on the `enterprise-networking` pod on the management cluster for communication from the remote cluster.

```shell
kubectl -n gloo-mesh --context $MGMT_CONTEXT logs deployment/enterprise-networking | grep $CLUSTER_NAME
```

You should see messages similar to:

```shell
{"level":"debug","ts":1616160185.5505846,"logger":"pull-resource-deltas","msg":"recieved request for delta: response_nonce:\"1\"","metadata":{":authority":["enterprise-networking.gloo-mesh.svc.cluster.local:11100"],"content-type":["application/grpc"],"user-agent":["grpc-go/1.34.0"],"x-cluster-id":["remote-cluster"]},"peer":"10.244.0.17:40074"}
```

## Next Steps

And we're done! Any meshes in that cluster will be discovered and available for configuration by Gloo Mesh Enterprise. See the guide on [installing Istio]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}), to see how to easily get Istio running on that cluster.
