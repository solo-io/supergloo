---
title: "meshctl create"
weight: 5
---
## meshctl create

Create a Service Mesh Hub custom resource

### Synopsis

Utility for creating Service Mesh Hub custom resources

```
meshctl create [flags]
```

### Options

```
      --dry-run         Set true to output generated resource without writing to k8s
  -h, --help            help for create
  -o, --output string   Output format for the report. Valid values: [json, yaml] (default "yaml")
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
* [meshctl create accesscontrolpolicy](../meshctl_create_accesscontrolpolicy)	 - Create an AccessControlPolicy resource
* [meshctl create trafficpolicy](../meshctl_create_trafficpolicy)	 - Create a TrafficPolicy resource
* [meshctl create virtualmesh](../meshctl_create_virtualmesh)	 - Create a VirtualMesh resource

