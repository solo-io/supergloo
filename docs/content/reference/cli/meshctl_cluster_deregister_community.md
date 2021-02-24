---
title: "meshctl cluster deregister community"
weight: 5
---
## meshctl cluster deregister community

Remove the community certificate agent

```
meshctl cluster deregister community [flags]
```

### Options

```
      --api-server-address string     Swap out the address of the remote cluster's k8s API server for the value of this flag. Set this flag when the address of the cluster domain used by the Gloo Mesh is different than that specified in the local kubeconfig.
      --cluster-name string           name of the cluster to deregister
      --federation-namespace string   namespace of the Gloo Mesh control plane in which the secret for the deregistered cluster will be created (default "gloo-mesh")
  -h, --help                          help for community
      --kubeconfig string             path to the kubeconfig from which the registered cluster will be accessed
      --mgmt-context string           name of the kubeconfig context to use for the management cluster
      --remote-context string         name of the kubeconfig context to use for the remote cluster
      --remote-namespace string       namespace in the target cluster where a service account enabling remote access will be created. If the namespace does not exist it will be created. (default "gloo-mesh")
```

### Options inherited from parent commands

```
  -v, --verbose   enable verbose logging
```

### SEE ALSO

* [meshctl cluster deregister](../meshctl_cluster_deregister)	 - Deregister a Kubernetes cluster from Gloo Mesh, cleaning up any associated resources

