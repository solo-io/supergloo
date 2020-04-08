---
title: "meshctl mesh install"
weight: 5
---
## meshctl mesh install

Install meshes using meshctl

### Synopsis

Install meshes using meshctl

```
meshctl mesh install (istio) [flags]
```

### Options

```
      --create-operator-namespace   Create the namespace specified by --operator-namespace (default true)
      --dry-run                     Dump the manifest that would be used to install the operator to stdout rather than apply it
  -h, --help                        help for install
      --operator-namespace string   Namespace in which to install the Mesh operator (default "istio-operator")
      --operator-spec string        Optional path to a YAML file containing an installation spec ('-' for stdin)
      --profile string              optional profile
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

* [meshctl mesh](../meshctl_mesh)	 - Manage service meshes

