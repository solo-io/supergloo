---
title: "meshctl istio"
weight: 5
---
## meshctl istio

Manage installations of Istio

### Synopsis

Manage installations of Istio

```
meshctl istio [flags]
```

### Options

```
  -h, --help   help for istio
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

* [meshctl](../meshctl)	 - CLI for Service Mesh Hub
* [meshctl istio install](../meshctl_istio_install)	 - Install Istio on the indicated cluster using the Istio installation operator (https://preliminary.istio.io/docs/setup/install/standalone-operator/)

