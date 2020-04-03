---
title: "meshctl demo"
weight: 5
---
## meshctl demo

Command line utilities for running/interacting with Service Mesh Hub demos

### Synopsis

Command line utilities for running/interacting with Service Mesh Hub demos

```
meshctl demo [flags]
```

### Options

```
  -h, --help   help for demo
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
* [meshctl demo cleanup](../meshctl_demo_cleanup)	 - Delete the local Service Mesh Hub demo setup
* [meshctl demo init](../meshctl_demo_init)	 - Bootstrap a new local Service Mesh Hub demo setup

