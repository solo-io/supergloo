---
title: "meshctl cluster configure"
weight: 5
---
## meshctl cluster configure

Configure Kubernetes Clusters registered with Gloo Mesh.

### Synopsis

Create a mapping of clusters to kubeconfig entries in ${HOME}/.gloo-mesh/meshctl-config.yaml.

There are two modes for this - interactive and non-interactive. Each data plane cluster should be configured with
a cluster name. Note that if a cluster is both a management and data plane cluster, it will need to be configured twice.

```
meshctl cluster configure [flags]
```

### Examples

```
 meshctl cluster configure --disable-prompt --kubeconfig ${HOME}/.kube/config --kubecontext cluster1 ## Registers a management plane cluster
 meshctl cluster configure --disable-prompt --cluster-name cluster2 --kubeconfig ${HOME}/.kube/config --kubecontext cluster2 ## Registers a data plane cluster
```

### Options

```
      --cluster-name string                                        data plane cluster name (ignored if this is the management cluster)
      --disable-prompt                                             Disable the interactive prompt. Use this to configure the meshctl config file with flags instead.
  -h, --help                                                       help for configure
      --is-mgmt-cluster                                            this is the management cluster (default true)
      --kubeconfig string                                          Path to the kubeconfig from which the cluster will be accessed
      --kubecontext string                                         Name of the kubeconfig context to use for the cluster
  -c, --meshctl-config-file $HOME/.gloo-mesh/meshctl-config.yaml   path to the meshctl config file. defaults to $HOME/.gloo-mesh/meshctl-config.yaml
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging
```

### SEE ALSO

* [meshctl cluster](../meshctl_cluster)	 - Manage Kubernetes clusters registered to Gloo Mesh

