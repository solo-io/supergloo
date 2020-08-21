---
title: "meshctl uninstall"
weight: 5
---
## meshctl uninstall

Uninstall Service Mesh Hub from the referenced cluster

### Synopsis

Uninstall Service Mesh Hub from the referenced cluster

```
meshctl uninstall [flags]
```

### Options

```
  -d, --dry-run               Output installation manifest
  -h, --help                  help for uninstall
      --kubeconfig string     path to the kubeconfig from which the management cluster will be accessed
      --kubecontext string    name of the kubeconfig context to use for the management cluster
      --namespace string      namespace in which to install Service Mesh Hub (default "service-mesh-hub")
      --release-name string   Helm release name (default "service-mesh-hub")
  -v, --verbose               Enable verbose output
```

### SEE ALSO

* [meshctl](../meshctl)	 - The Command Line Interface for managing Service Mesh Hub.

