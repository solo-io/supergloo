---
title: "meshctl istio install"
weight: 5
---
## meshctl istio install

Install Istio on the indicated cluster using the Istio installation operator (https://preliminary.istio.io/docs/setup/install/standalone-operator/)

### Synopsis

Install Istio on the indicated cluster using the Istio installation operator (https://preliminary.istio.io/docs/setup/install/standalone-operator/)

```
meshctl istio install [flags]
```

### Options

```
      --create-operator-crd         Register the IstioOperator CRD in the remote cluster (default true)
      --create-operator-namespace   Create the namespace specified by --operator-namespace (default true)
      --dry-run                     Dump the manifest that would be used to install the operator to stdout rather than apply it
  -h, --help                        help for install
      --operator-namespace string   Namespace in which to install the Istio operator (default "istio-operator")
      --operator-spec string        Optional path to a YAML file containing an IstioOperator resource
      --profile string              Install Istio in one of its pre-configured profiles; supported profiles: [default, demo, minimal, remote, sds] (https://preliminary.istio.io/docs/setup/additional-setup/config-profiles/)
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

* [meshctl istio](../meshctl_istio)	 - Manage installations of Istio

