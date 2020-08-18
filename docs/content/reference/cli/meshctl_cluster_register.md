---
title: "meshctl cluster register"
weight: 5
---
## meshctl cluster register

Register a new cluster by creating a service account token in that cluster through which to authorize Service Mesh Hub

### Synopsis

In order to specify the remote cluster against which to perform this operation, one or both of the --remote-kubeconfig or --remote-context flags must be set. The former selects the kubeconfig file, and the latter selects which context should be used from that kubeconfig file

```
meshctl cluster register [flags]
```

### Options

```
  -h, --help                            help for register
      --overwrite                       Overwrite any cluster registered with the cluster name provided
      --remote-cluster-name string      Name of the cluster to be operated upon
      --remote-context string           Set the context you would like to use for the remote cluster
      --remote-kubeconfig string        Set the path to the kubeconfig you would like to use for the remote cluster. Leave empty to use the default
      --remote-write-namespace string   Namespace in the remote cluster in which to write resources (default "service-mesh-hub")
      --values strings                  List of files with value overrides for the csr-agent Helm chart, (e.g. --values file1,file2 or --values file1 --values file2)
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

* [meshctl cluster](../meshctl_cluster)	 - Register and perform operations on clusters

