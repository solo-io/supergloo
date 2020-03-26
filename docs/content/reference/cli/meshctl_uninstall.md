---
title: "meshctl uninstall"
weight: 5
---
## meshctl uninstall

Completely uninstall Service Mesh Hub and remove associated CRDs

### Synopsis

Completely uninstall Service Mesh Hub and remove associated CRDs

```
meshctl uninstall [flags]
```

### Options

```
  -h, --help                  help for uninstall
      --release-name string   Helm release name (default "service-mesh-hub")
      --remove-namespace      Remove the Service Mesh Hub namespace specified with -n
```

### Options inherited from parent commands

```
      --context string          Specify which context from the kubeconfig should be used; uses current context if none is specified
      --kube-timeout duration   Specify the default timeout for requests to kubernetes API servers. (default 5s)
      --kubeconfig string       Specify the kubeconfig for the current command
  -n, --namespace string        Specify the namespace where Service Mesh Hub resources should be written (default "service-mesh-hub")
  -v, --verbose                 Enable verbose mode, which outputs additional execution details that may be helpful for debugging
```

### SEE ALSO

* [meshctl](../meshctl)	 - CLI for Service Mesh Hub

