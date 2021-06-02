---
title: "meshctl cluster list enterprise"
weight: 5
---
## meshctl cluster list enterprise

List registered clusters for Gloo Mesh enterprise edition

```
meshctl cluster list enterprise [flags]
```

### Examples

```
 meshctl cluster list enterprise
```

### Options

```
  -h, --help                 help for enterprise
      --kubeconfig string    Path to the kubeconfig from which the management cluster will be accessed
      --kubecontext string   Name of the kubeconfig context to use for the management cluster
      --namespace string     namespace that Gloo Mesh is installed in (default "gloo-mesh")
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging
```

### SEE ALSO

* [meshctl cluster list](../meshctl_cluster_list)	 - List all Kubernetes cluster registered with Gloo Mesh

