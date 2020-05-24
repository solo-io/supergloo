---
title: "meshctl demo istio-multicluster"
weight: 5
---
## meshctl demo istio-multicluster

Demo Service Mesh Hub functionality with two Istio control planes deployed on separate clusters.

### Synopsis

Demo Service Mesh Hub functionality with two Istio control planes deployed on separate clusters.

```
meshctl demo istio-multicluster [flags]
```

### Options

```
  -h, --help   help for istio-multicluster
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
* [meshctl demo istio-multicluster cleanup](../meshctl_demo_istio-multicluster_cleanup)	 - Cleanup bootstrapped resources AWS Appmesh and EKS resources
* [meshctl demo istio-multicluster init](../meshctl_demo_istio-multicluster_init)	 - Bootstrap an AWS App mesh and EKS cluster demo with Service Mesh Hub

