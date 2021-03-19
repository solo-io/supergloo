---
title: "meshctl dashboard"
weight: 5
---
## meshctl dashboard

Port forwards the Gloo Mesh Enterprise UI and opens it in a browser if available

```
meshctl dashboard [flags]
```

### Options

```
  -h, --help                 help for dashboard
      --kubeconfig string    Path to the kubeconfig from which the management cluster will be accessed
      --kubecontext string   Name of the kubeconfig context to use for the management cluster
      --namespace string     The namespace that the Gloo Mesh UI is deployed in (default "gloo-mesh")
  -p, --port uint32          The local port to forward to the dashboard (default 8090)
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging
```

### SEE ALSO

* [meshctl](../meshctl)	 - The Command Line Interface for managing Gloo Mesh.

