---
title: "meshctl demo appmesh-eks init"
weight: 5
---
## meshctl demo appmesh-eks init

Bootstrap an AWS App mesh and EKS cluster demo with Service Mesh Hub

### Synopsis


Prerequisites:
	1. meshctl
	2. eksctl (https://github.com/weaveworks/eksctl)
	3. Helm (https://helm.sh/docs/intro/install/)
	4. AWS API credentials must be configured, either through the "~/.aws/credentials" file or environment variables. See these references for more information:
         a. https://docs.aws.amazon.com/cli/latest/userguide/cli-config-files.html
         b. https://docs.aws.amazon.com/cli/latest/userguide/cli-environment.html


```
meshctl demo appmesh-eks init [flags]
```

### Options

```
      --aws-region string   Specify the AWS region for demo entities, defaults to us-east-2. (default "us-east-2")
  -h, --help                help for init
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

* [meshctl demo appmesh-eks](../meshctl_demo_appmesh-eks)	 - Demo Service Mesh Hub functionality with Appmesh and EKS

