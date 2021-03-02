---
title: "meshctl cluster deregister"
weight: 5
---
## meshctl cluster deregister

Deregister a Kubernetes cluster from Gloo Mesh, cleaning up any associated resources

### Synopsis

Deregister a Kubernetes cluster from Gloo Mesh, cleaning up any associated resources

The edition deregistered must match the edition that was originally registered.

### Options

```
      --cluster-name string       name of the cluster to deregister
  -h, --help                      help for deregister
      --kubeconfig string         path to the kubeconfig from which the registered cluster will be accessed
      --mgmt-context string       name of the kubeconfig context to use for the management cluster
      --mgmt-namespace string     namespace of the Gloo Mesh control plane in which the secret for the deregistered cluster will be created (default "gloo-mesh")
      --remote-context string     name of the kubeconfig context to use for the remote cluster
      --remote-namespace string   namespace in the target cluster where a service account enabling remote access will be created. If the namespace does not exist it will be created. (default "gloo-mesh")
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging
```

### SEE ALSO

* [meshctl cluster](../meshctl_cluster)	 - Interacting with remote Kubernetes clusters registered to Gloo Mesh
* [meshctl cluster deregister community](../meshctl_cluster_deregister_community)	 - Remove the community certificate agent
* [meshctl cluster deregister enterprise](../meshctl_cluster_deregister_enterprise)	 - Remove the enterprise agent

