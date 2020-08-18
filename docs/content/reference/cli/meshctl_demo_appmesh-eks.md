---
title: "meshctl demo appmesh-eks"
weight: 5
---
## meshctl demo appmesh-eks

Demo Service Mesh Hub functionality with Appmesh and EKS

### Synopsis

Demo Service Mesh Hub functionality with Appmesh and EKS

```
meshctl demo appmesh-eks [flags]
```

### Options

```
  -h, --help   help for appmesh-eks
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
* [meshctl demo appmesh-eks cleanup](../meshctl_demo_appmesh-eks_cleanup)	 - Cleanup bootstrapped resources AWS Appmesh and EKS resources
* [meshctl demo appmesh-eks init](../meshctl_demo_appmesh-eks_init)	 - Bootstrap an AWS App mesh and EKS cluster demo with Service Mesh Hub

