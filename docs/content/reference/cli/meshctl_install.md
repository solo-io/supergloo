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
      --agent-crds-chart-file string   Path to a local Helm chart for installing CRDs needed by remote agents.
                                       If unset, this command will install the agent CRDs from the publicly released Helm chart.
      --api-server-address string      Swap out the address of the remote cluster's k8s API server for the value of this flag.
                                       Set this flag when the address of the cluster domain used by the Gloo Mesh is different than that specified in the local kubeconfig.
      --chart-file string              Path to a local Helm chart for installing Gloo Mesh.
                                       If unset, this command will install Gloo Mesh from the publicly released Helm chart.
      --chart-values-file string       File containing value overrides for the Gloo Mesh Helm chart
      --cluster-domain string          The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. Defaults to 'cluster.local'.
                                       Read more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/
      --cluster-name string            Name with which to register the cluster running Gloo Mesh, only applies if --register is also set (default "mgmt-cluster")
  -d, --dry-run                        Output installation manifest
  -h, --help                           help for install
      --kubeconfig string              Path to the kubeconfig from which the management cluster will be accessed
      --kubecontext string             Name of the kubeconfig context to use for the management cluster
      --namespace string               namespace in which to install Gloo Mesh (default "gloo-mesh")
  -r, --register                       Also register the cluster
      --version string                 Version to install, defaults to latest if omitted
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging
```

### SEE ALSO

* [meshctl](../meshctl)	 - The Command Line Interface for managing Gloo Mesh.
* [meshctl install community](../meshctl_install_community)	 - Install the open source community edition
* [meshctl install enterprise](../meshctl_install_enterprise)	 - Install the enterprise editionn (requires a license)

