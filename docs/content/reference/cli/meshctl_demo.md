---
title: "meshctl demo"
weight: 5
---
## meshctl demo

Command line utilities for running/interacting with Service Mesh Hub demos

### Synopsis

Command line utilities for running/interacting with Service Mesh Hub demos

```
meshctl demo [flags]
```

### Options

```
      --cluster-name string   Registers the cluster under this name
      --demo-label string     Optionally label namespaces that are created during the demo with 'solo.io/hub-demo:$DEMO-LABEL' so they can be cleaned up later
  -h, --help                  help for demo
      --use-kind              If this is set, use KinD (Kubernetes in Docker) to stand up local clusters; can not set if --context
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
* [meshctl demo basic-hub-install](../meshctl_demo_basic-hub-install)	 - Get Service Mesh Hub installed to a cluster, and register that cluster for management through the Hub
* [meshctl demo istio](../meshctl_demo_istio)	 - Install Service Mesh Hub to a cluster, and install/register Istio to that same cluster
* [meshctl demo istio-multicluster](../meshctl_demo_istio-multicluster)	 - Install Service Mesh Hub to a cluster, then install/register two Istio installations: one to the management plane cluster, and one to a separate cluster
* [meshctl demo linkerd](../meshctl_demo_linkerd)	 - Install Service Mesh Hub to a cluster, then install/register a Linkerd mesh on that cluster

