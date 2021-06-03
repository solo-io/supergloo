---
title: "meshctl cluster list"
weight: 5
---
## meshctl cluster list

List all Kubernetes cluster registered with Gloo Mesh

```
meshctl cluster list [flags]
```

### Options

```
  -h, --help                 help for list
      --kubeconfig string    Path to the kubeconfig from which the management cluster will be accessed
      --kubecontext string   Name of the kubeconfig context to use for the management cluster
      --namespace string     namespace that Gloo Mesh is installed in (default "gloo-mesh")
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging
```

### SEE ALSO

* [meshctl cluster](../meshctl_cluster)	 - Manage Kubernetes clusters registered to Gloo Mesh

