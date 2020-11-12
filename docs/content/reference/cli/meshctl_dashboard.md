---
title: "meshctl dashboard"
weight: 5
---
## meshctl dashboard

Port forwards the Gloo Mesh Enterprise UI and opens it in a browser if available

### Synopsis

Port forwards the Gloo Mesh Enterprise UI and opens it in a browser if available

```
meshctl dashboard [flags]
```

### Options

```
  -h, --help                     help for dashboard
      --kubeconfig string        path to the kubeconfig from which the management cluster will be accessed
      --kubecontext string       name of the kubeconfig context to use for the management cluster
      --kubes-namespace string   The namespace that the Gloo Mesh UI is deployed in (default "gloo-mesh")
  -p, --port uint32              The local port to forward to the dashboard (default 8090)
```

### SEE ALSO

* [meshctl](../meshctl)	 - The Command Line Interface for managing Gloo Mesh.

