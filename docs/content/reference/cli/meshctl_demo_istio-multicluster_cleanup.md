---
title: "meshctl demo istio-multicluster cleanup"
weight: 5
---
## meshctl demo istio-multicluster cleanup

Cleanup bootstrapped local resources.

### Synopsis

Cleanup bootstrapped local resources.

```
meshctl demo istio-multicluster cleanup [flags]
```

### Options

```
      --aws-region string         Specify the AWS region for demo entities, defaults to us-east-2. (default "us-east-2")
      --eks-cluster-name string   Specify the name of the EKS cluster to cleanup.
  -h, --help                      help for cleanup
      --mesh-name string          Specify name of the App mesh to cleanup.
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

* [meshctl demo istio-multicluster](../meshctl_demo_istio-multicluster)	 - Demo Service Mesh Hub functionality with two Istio control planes deployed on separate clusters.

