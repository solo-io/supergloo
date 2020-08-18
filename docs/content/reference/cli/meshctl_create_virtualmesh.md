---
title: "meshctl create virtualmesh"
weight: 5
---
## meshctl create virtualmesh

Create a VirtualMesh resource

### Synopsis

Create a VirtualMesh resource

```
meshctl create virtualmesh [flags]
```

### Options

```
  -h, --help   help for virtualmesh
```

### Options inherited from parent commands

```
      --context string          Specify which context from the kubeconfig should be used; uses current context if none is specified
      --dry-run                 Set true to output generated resource without writing to k8s
      --kube-timeout duration   Specify the default timeout for requests to kubernetes API servers (default 5s)
      --kubeconfig string       Specify the kubeconfig for the current command
  -n, --namespace string        Specify the namespace where Service Mesh Hub resources should be written (default "service-mesh-hub")
  -o, --output string           Output format for the report. Valid values: [json, yaml] (default "yaml")
  -v, --verbose                 Enable verbose mode, which outputs additional execution details that may be helpful for debugging
```

### SEE ALSO

* [meshctl create](../meshctl_create)	 - Create a Service Mesh Hub custom resource

