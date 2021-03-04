---
title: "Enterprise"
menuTitle: Enterprise
description: Registering a cluster with Gloo Mesh enterprise edition
weight: 30
---

{{% notice note %}} Gloo Mesh Enterprise is required for this feature. {{% /notice %}}

Relay is an alternative mode of deploying Gloo Mesh that confers several advantages discussed in [this document]({{% versioned_link_path fromRoot="/concepts/relay" %}}).
Cluster registration in relay mode simply consists of installing the relay agent.

**Register with Helm**

1. Install the Enterprise Agent from the Helm repository located at `https://storage.googleapis.com/gloo-mesh-enterprise/enterprise-agent`.
Make sure to review the Helm values options before installing. Some notable values include:

* `relay.cluster` will be the name by which the cluster is referenced in all Gloo Mesh configuration.
* `relay.serverAddress` is the address by which the Gloo Mesh management plane can be accessed. See the [Gloo Mesh Enterprise prerequisites]({{% versioned_link_path fromRoot="/setup/enterprise_prerequisites" %}}) for more details.
* `relay.authority` is the host header that will be passed to the server on the Gloo Mesh management plane.

Also note that the Enterprise Agent's version should match that of the `enterprise-networking` component running on the
management cluster. Run `meshctl version` on the management cluster to review the `enterprise-networking` version.

```bash
helm upgrade --install enterprise-agent enterprise-agent/enterprise-agent --namespace gloo-mesh \
  --set relay.serverAddress=${SERVER_ADDRESS} --set relay.authority=enterprise-networking.gloo-mesh \
  --set relay.cluster=mgmt-cluster --kube-context=kind-mgmt-cluster --version ${ENTERPRISE_NETWORKING_VERSION}
```

2. Create a `KubernetesCluster` object. The `metadata.name` of the object must
match the value passed for `relay.cluster` in the Helm chart above. The `spec.clusterDomain` must 
match the [local cluster domain](https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/) of the Kubernetes cluster.

```yaml
apiVersion: multicluster.solo.io/v1alpha1
kind: KubernetesCluster
metadata:
  name: your-cluster-name
  namespace: gloo-mesh
spec:
  clusterDomain: cluster.local
```


**Register with meshctl**

Instructions for registering a cluster to Gloo Mesh Enterprise with meshctl are coming soon.
