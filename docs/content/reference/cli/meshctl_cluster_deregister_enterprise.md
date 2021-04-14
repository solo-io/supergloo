---
title: "meshctl cluster deregister enterprise"
weight: 5
---
## meshctl cluster deregister enterprise

Remove the enterprise agent

### Synopsis

Deregister the remote cluster, which includes uninstalling the
enterprise agent and removing the cluster definition from the management cluster.

The name of the remote context must be passed in via the --remote-context flag
if it is different from the context currently selected in the kubernetes config.

```
meshctl cluster deregister enterprise [cluster name] [flags]
```

### Examples

```
  # These examples assume that the currently selected context
  # is the management context.

  # Deregister the currently selected cluster
  # In this situation it is both the management and a remote cluster
  meshctl cluster deregister enterprise mgmt-cluster

  # Deregister a different remote cluster separate from the current one
  meshctl cluster deregister enterprise remote-cluster --remote-context my-remote-ctx
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
      --mgmt-namespace string     namespace of the Gloo Mesh control plane in which the secret for the deregistered cluster will be created (default "gloo-mesh")
      --remote-context string     name of the kubeconfig context to use for the remote cluster
      --remote-namespace string   namespace in the target cluster where a service account enabling remote access will be created. If the namespace does not exist it will be created. (default "gloo-mesh")
  -v, --verbose                   Enable verbose logging
```

### SEE ALSO

* [meshctl cluster deregister](../meshctl_cluster_deregister)	 - Deregister a Kubernetes cluster from Gloo Mesh, cleaning up any associated resources

