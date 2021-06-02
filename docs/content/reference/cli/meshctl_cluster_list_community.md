---
title: "meshctl cluster list community"
weight: 5
---
## meshctl cluster list community

List registered clusters for Gloo Mesh community edition

```
meshctl cluster list community [flags]
```

### Examples

```
  meshctl cluster list community
```

### Options

```
  -h, --help                 help for community
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

