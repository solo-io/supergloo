---
title: "meshctl check"
weight: 5
---
## meshctl check

Check the status of a Service Mesh Hub installation

### Synopsis

Check the status of a Service Mesh Hub installation

```
meshctl check [flags]
```

### Options

```
  -h, --help            help for check
  -o, --output string   Output format for the report. Valid values: [pretty, json] (default "pretty")
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

