---
title: "meshctl cluster deregister enterprise"
weight: 5
---
## meshctl cluster deregister enterprise

Remove the enterprise agent

### Synopsis

Deregister the remote cluster, which includes uninstalling the
enterprise agent and removing the cluster definition from the management cluster.

```
meshctl cluster deregister enterprise [cluster name] [flags]
```

### Examples

```
 meshctl cluster deregister enterprise remote-cluster --remote-context my-remote
```

### Options

```
  -h, --help   help for enterprise
```

### Options inherited from parent commands

```
      --cluster-name string       name of the cluster to deregister
      --kubeconfig string         path to the kubeconfig from which the registered cluster will be accessed
      --mgmt-context string       name of the kubeconfig context to use for the management cluster
      --mgmt-kubeconfig string    path to the kubeconfig file to use for the management cluster if different from control plane kubeconfig file location
      --mgmt-namespace string     namespace of the Gloo Mesh control plane in which the secret for the deregistered cluster will be created (default "gloo-mesh")
      --remote-context string     name of the kubeconfig context to use for the remote cluster
      --remote-namespace string   namespace in the target cluster where a service account enabling remote access will be created. If the namespace does not exist it will be created. (default "gloo-mesh")
  -v, --verbose                   Enable verbose logging
```

### SEE ALSO

* [meshctl cluster deregister](../meshctl_cluster_deregister)	 - Deregister a Kubernetes cluster from Gloo Mesh, cleaning up any associated resources

