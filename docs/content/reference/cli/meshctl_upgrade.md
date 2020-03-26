---
title: "meshctl upgrade"
weight: 5
---
## meshctl upgrade

In-place upgrade of the meshctl binary

### Synopsis

In-place upgrade of the meshctl binary

```
meshctl upgrade [flags]
```

### Options

```
  -h, --help             help for upgrade
      --path string      Desired path for your upgraded meshctl binary. Defaults to the location of your currently executing binary.
      --release string   Which meshctl release to download. Specify a tag corresponding to the desired version of meshctl or "latest". Service Mesh Hub releases can be found here: https://github.com/solo-io/service-mesh-hub/releases. Omitting this tag defaults to "latest". (default "latest")
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

