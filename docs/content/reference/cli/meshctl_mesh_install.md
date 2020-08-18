---
title: "meshctl mesh install"
weight: 5
---
## meshctl mesh install

Install meshes using meshctl

### Synopsis

Install meshes using meshctl

```
meshctl mesh install [flags]
```

### Options

```
  -h, --help   help for install
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
* [meshctl mesh install istio1.5](../meshctl_mesh_install_istio1.5)	 - Install Istio version 1.5
* [meshctl mesh install istio1.6](../meshctl_mesh_install_istio1.6)	 - Install Istio version 1.6

