---
title: "meshctl cluster deregister"
weight: 5
---
## meshctl cluster deregister

Deregister an existing cluster

### Synopsis

Deregister an existing cluster

```
meshctl cluster deregister [flags]
```

### Options

```
  -h, --help                         help for deregister
      --remote-cluster-name string   Name of the cluster to deregister
```

### Options inherited from parent commands

```
      --context string          Specify which context from the kubeconfig should be used; uses current context if none is specified
      --kube-timeout duration   Specify the default timeout for requests to kubernetes API servers (default 5s)
      --kubeconfig string       Specify the kubeconfig for the current command
  -n, --namespace string        Specify the namespace where Service Mesh Hub resources should be written (default "service-mesh-hub")
  -v, --verbose                 Enable verbose mode, which outputs additional execution details that may be helpful for debugging
```

### SEE ALSO

* [meshctl cluster](../meshctl_cluster)	 - Register and perform operations on clusters

