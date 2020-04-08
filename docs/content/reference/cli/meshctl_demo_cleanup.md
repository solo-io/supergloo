---
title: "meshctl demo cleanup"
weight: 5
---
## meshctl demo cleanup

Delete the local Service Mesh Hub demo setup

### Synopsis

This will delete all kind clusters running locally, so make sure to only run this script if the only kind clusters running are those created by mesctl demo init.

```
meshctl demo cleanup [flags]
```

### Options

```
  -h, --help   help for cleanup
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

* [meshctl demo](../meshctl_demo)	 - Command line utilities for running/interacting with Service Mesh Hub demos

