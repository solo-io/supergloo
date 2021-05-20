---
title: "meshctl cluster deregister"
weight: 5
---
## meshctl cluster deregister

Deregister a Kubernetes cluster from Gloo Mesh, cleaning up any associated resources

### Synopsis


Deregistering a cluster removes the installed agent from the remote cluster as
well as the other created resources such as service accounts. The edition
must match the edition that the cluster was originally registered with.

The name of the context of the cluster to dregister must be provided via the
--remote-context flag. It is important that the remote context and the name
passed as an argument are for the same cluster otherwise unexpected behavior
may occur.

If the management cluster is different than the one that the current context
points then it an be provided via the --mgmt-context flag.

### Options

```
      --cluster-name string       name of the cluster to deregister
  -h, --help                      help for deregister
      --kubeconfig string         path to the kubeconfig from which the registered cluster will be accessed
      --mgmt-context string       name of the kubeconfig context to use for the management cluster
      --mgmt-kubeconfig string    path to the kubeconfig context to use for the management cluster (defaults to ~/.kube/config)
      --mgmt-namespace string     namespace of the Gloo Mesh control plane in which the secret for the deregistered cluster will be created (default "gloo-mesh")
      --remote-context string     name of the kubeconfig context to use for the remote cluster
      --remote-namespace string   namespace in the target cluster where a service account enabling remote access will be created. If the namespace does not exist it will be created. (default "gloo-mesh")
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging
```

### SEE ALSO

* [meshctl cluster](../meshctl_cluster)	 - Manage Kubernetes clusters registered to Gloo Mesh
* [meshctl cluster deregister community](../meshctl_cluster_deregister_community)	 - Remove the community certificate agent
* [meshctl cluster deregister enterprise](../meshctl_cluster_deregister_enterprise)	 - Remove the enterprise agent

