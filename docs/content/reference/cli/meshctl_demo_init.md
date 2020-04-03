---
title: "meshctl demo init"
weight: 5
---
## meshctl demo init

Bootstrap a new local Service Mesh Hub demo setup

### Synopsis

Running the Service Mesh Hub demo setup locally requires 4 tools to be installed, and accessible via the PATH. meshctl, kubectl, docker, and kind. This command will bootstrap 2 clusters, one of which will run the Service Mesh Hub management-plane as well as Istio, and the other will just run Istio.

```
meshctl demo init [flags]
```

### Options

```
  -h, --help   help for init
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

* [meshctl demo](../meshctl_demo)	 - Command line utilities for running/interacting with Service Mesh Hub demos

