---
title: "meshctl install"
weight: 5
---
## meshctl install

Install Service Mesh Hub

### Synopsis

Install Service Mesh Hub

```
meshctl install [flags]
```

### Options

```
      --api-server-address string        Swap out the address of the remote cluster's k8s API server for the value of this flag. Set this flag when the address of the cluster domain used by the Service Mesh Hub is different than that specified in the local kubeconfig.
      --cert-agent-chart-file string     Path to a local Helm chart for installing the Certificate Agent. If unset, this command will install the Certificate Agent from the publicly released Helm chart.
      --cert-agent-chart-values string   Path to a Helm values.yaml file for customizing the installation of the Certificate Agent. If unset, this command will install the Certificate Agent with default Helm values.
      --chart-file string                Path to a local Helm chart for installing Service Mesh Hub. If unset, this command will install Service Mesh Hub from the publicly released Helm chart.
      --chart-values-file string         File containing value overrides for the Service Mesh Hub Helm chart
      --cluster-domain string            The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. Defaults to 'cluster.local'. Read more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/
      --cluster-name string              Name with which to register the cluster running Service Mesh Hub, only applies if --register is also set (default "mgmt-cluster")
  -d, --dry-run                          Output installation manifest
  -h, --help                             help for install
      --kubeconfig string                path to the kubeconfig from which the management cluster will be accessed
      --kubecontext string               name of the kubeconfig context to use for the management cluster
      --namespace string                 namespace in which to install Service Mesh Hub (default "service-mesh-hub")
  -r, --register                         Register the cluster running Service Mesh Hub
      --release-name string              Helm release name (default "service-mesh-hub")
  -v, --verbose                          Enable verbose output
      --version string                   Version to install, defaults to latest if omitted
```

### SEE ALSO

* [meshctl](../meshctl)	 - The Command Line Interface for managing Service Mesh Hub.

