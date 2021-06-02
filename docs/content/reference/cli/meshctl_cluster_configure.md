---
title: "meshctl cluster configure"
weight: 5
---
## meshctl cluster configure

Configure Kubernetes Clusters registered with Gloo Mesh.

### Synopsis

Create a mapping of clusters to kubeconfig entries in ${HOME}/.gloo-mesh/meshctl-config.yaml.

```
meshctl cluster configure [flags]
```

### Options

```
      --cluster-name string                                        data plane cluster name (leave empty if this is the management cluster)
      --disable-prompt                                             Disable the interactive prompt. Use this to configure the meshctl config file with flags instead.
  -h, --help                                                       help for configure
      --kubeContext string                                         name of the kubernetes context
      --kubeconfig string                                          path to the kubeconfig file
  -f, --meshctl-config-file $HOME/.gloo-mesh/meshctl-config.yaml   path to the meshctl config file. defaults to $HOME/.gloo-mesh/meshctl-config.yaml
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging
```

### SEE ALSO

* [meshctl cluster](../meshctl_cluster)	 - Manage Kubernetes clusters registered to Gloo Mesh

