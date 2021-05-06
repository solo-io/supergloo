---
title: "meshctl install enterprise"
weight: 5
---
## meshctl install enterprise

Install Gloo Mesh Enterprise (requires a license)

```
meshctl install enterprise [flags]
```

### Examples

```
  # Install to the currently selected Kubernetes context
  meshctl install enterprise --license=<my_license>

  # Install to and register the currently selected Kubernetes context
  meshctl install enterprise --license=<my_license> --register

  # Don't install the UI
  meshctl install enterprise --license=<my_license> --skip-ui
```

### Options

```
      --enterprise-agent-chart-file string     Path to a local Helm chart for installing the Enterprise Agent.
                                               If unset, this command will install the Enterprise Agent from the publicly released Helm chart.
      --enterprise-agent-chart-values string   Path to a Helm values.yaml file for customizing the installation of the Enterprise Agent.
                                               If unset, this command will install the Enterprise Agent with default Helm values.
  -h, --help                                   help for enterprise
      --license string                         Gloo Mesh Enterprise license key (required)
      --relay-server-address string            The address that the enterprise agentw will communicate with the relay server via.
      --release-name string                    Helm release name (default "gloo-mesh")
      --skip-rbac                              Skip installation of the RBAC Webhook
      --skip-ui                                Skip installation of the Gloo Mesh UI
```

### Options inherited from parent commands

```
      --chart-file string          Path to a local Helm chart for installing Gloo Mesh.
                                   If unset, this command will install Gloo Mesh from the publicly released Helm chart.
      --chart-values-file string   File containing value overrides for the Gloo Mesh Helm chart
      --cluster-domain string      The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. 
                                   Read more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/ (default "cluster.local")
      --cluster-name string        Name with which to register the cluster running Gloo Mesh, only applies if --register is also set (default "mgmt-cluster")
  -d, --dry-run                    Output installation manifest
      --kubeconfig string          Path to the kubeconfig from which the management cluster will be accessed
      --kubecontext string         Name of the kubeconfig context to use for the management cluster
      --namespace string           namespace in which to install Gloo Mesh (default "gloo-mesh")
  -r, --register                   Also register the cluster
  -v, --verbose                    Enable verbose logging
      --version string             Version to install.
                                   Community defaults to meshctl version, enterprise defaults to latest stable
```

### SEE ALSO

* [meshctl install](../meshctl_install)	 - Install Gloo Mesh

