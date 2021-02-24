---
title: "meshctl install community"
weight: 5
---
## meshctl install community

Install the community certificate agent

```
meshctl install community [flags]
```

### Options

```
      --agent-crds-chart-file string     Path to a local Helm chart for installing CRDs needed by remote agents. If unset, this command will install the agent CRDs from the publicly released Helm chart.
      --api-server-address string        Swap out the address of the remote cluster's k8s API server for the value of this flag. Set this flag when the address of the cluster domain used by the Gloo Mesh is different than that specified in the local kubeconfig.
      --cert-agent-chart-file string     Path to a local Helm chart for installing the Certificate Agent. If unset, this command will install the Certificate Agent from the publicly released Helm chart.
      --cert-agent-chart-values string   Path to a Helm values.yaml file for customizing the installation of the Certificate Agent. If unset, this command will install the Certificate Agent with default Helm values.
      --chart-file string                Path to a local Helm chart for installing Gloo Mesh. If unset, this command will install Gloo Mesh from the publicly released Helm chart.
      --chart-values-file string         File containing value overrides for the Gloo Mesh Helm chart
      --cluster-domain string            The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. Defaults to 'cluster.local'. Read more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/
      --cluster-name string              Name with which to register the cluster running Gloo Mesh, only applies if --register is also set (default "mgmt-cluster")
  -d, --dry-run                          Output installation manifest
  -h, --help                             help for community
      --kubeconfig string                path to the kubeconfig from which the management cluster will be accessed
      --kubecontext string               name of the kubeconfig context to use for the management cluster
      --namespace string                 namespace in which to install Gloo Mesh (default "gloo-mesh")
  -r, --register                         Register the cluster running Gloo Mesh
      --release-name string              Helm release name (default "gloo-mesh")
      --version string                   Version to install, defaults to latest if omitted
```

### Options inherited from parent commands

```
  -v, --verbose   enable verbose logging
```

### SEE ALSO

* [meshctl install](../meshctl_install)	 - Install Gloo Mesh

