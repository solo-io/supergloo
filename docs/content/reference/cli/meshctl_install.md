---
title: "meshctl install"
weight: 5
---
## meshctl install

Install Gloo Mesh

### Synopsis

Install the Gloo Mesh management plan to a Kubernetes cluster.

Go to https://www.solo.io/products/gloo-mesh/ to learn more about the
difference between the editions.


### Options

```
      --chart-file string          Path to a local Helm chart for installing Gloo Mesh.
                                   If unset, this command will install Gloo Mesh from the publicly released Helm chart.
      --chart-values-file string   File containing value overrides for the Gloo Mesh Helm chart
      --cluster-domain string      The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. Defaults to 'cluster.local'.
                                   Read more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/
      --cluster-name string        Name with which to register the cluster running Gloo Mesh, only applies if --register is also set (default "mgmt-cluster")
  -d, --dry-run                    Output installation manifest
  -h, --help                       help for install
      --kubeconfig string          Path to the kubeconfig from which the management cluster will be accessed
      --kubecontext string         Name of the kubeconfig context to use for the management cluster
      --namespace string           namespace in which to install Gloo Mesh (default "gloo-mesh")
  -r, --register                   Also register the cluster
      --version string             Version to install.
                                   Community defaults to meshctl version, enterprise defaults to latest
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging
```

### SEE ALSO

* [meshctl](../meshctl)	 - The Command Line Interface for managing Gloo Mesh.
* [meshctl install community](../meshctl_install_community)	 - Install Gloo Mesh Community
* [meshctl install enterprise](../meshctl_install_enterprise)	 - Install Gloo Mesh Enterprise (requires a license)

