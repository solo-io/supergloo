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

1. Install the relay agent from the Helm repository located at `https://storage.googleapis.com/gloo-mesh-enterprise/enterprise-agent`.
Make sure to review the Helm values options before installing. The value of `relay.cluster` will
be the name by which the cluster is referenced in all Gloo Mesh configuration.

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

Register with `meshctl cluster register enterprise`, see details about usage and parameters [here]({{% versioned_link_path fromRoot="/reference/cli/meshctl_cluster_register_enterprise" %}}).

[comment]: <> (TODO add example command here once https://github.com/solo-io/gloo-mesh/pull/1290 is done)
