---
title: "meshctl install enterprise"
weight: 5
---
## meshctl install enterprise

Install the enterprise agent

```
meshctl install enterprise [flags]
```

### Options

```
      --agent-crds-chart-file string           Path to a local Helm chart for installing CRDs needed by remote agents. If unset, this command will install the agent CRDs from the publicly released Helm chart.
      --api-server-address string              Swap out the address of the remote cluster's k8s API server for the value of this flag. Set this flag when the address of the cluster domain used by the Gloo Mesh is different than that specified in the local kubeconfig.
      --chart-file string                      Path to a local Helm chart for installing Gloo Mesh. If unset, this command will install Gloo Mesh from the publicly released Helm chart.
      --chart-values-file string               File containing value overrides for the Gloo Mesh Helm chart
      --cluster-domain string                  The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. Defaults to 'cluster.local'. Read more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/
      --cluster-name string                    Name with which to register the cluster running Gloo Mesh, only applies if --register is also set (default "mgmt-cluster")
  -d, --dry-run                                Output installation manifest
      --enterprise-agent-chart-file string     Path to a local Helm chart for installing the Enterprise Agent. If unset, this command will install the Enterprise Agent from the publicly released Helm chart.
      --enterprise-agent-chart-values string   Path to a Helm values.yaml file for customizing the installation of the Enterprise Agent. If unset, this command will install the Enterprise Agent with default Helm values.
  -h, --help                                   help for enterprise
      --kubeconfig string                      path to the kubeconfig from which the management cluster will be accessed
      --kubecontext string                     name of the kubeconfig context to use for the management cluster
      --license string                         Gloo Mesh Enterprise license key
      --namespace string                       namespace in which to install Gloo Mesh (default "gloo-mesh")
  -r, --register                               Register the cluster running Gloo Mesh
      --release-name string                    Helm release name (default "gloo-mesh")
      --skip-rbac                              Skip installation of the RBAC Webhook
      --skip-ui                                Skip installation of the Gloo Mesh UI
      --version string                         Version to install, defaults to latest if omitted
```

### Options inherited from parent commands

```
  -v, --verbose   enable verbose logging
```

### SEE ALSO

* [meshctl install](../meshctl_install)	 - Install Gloo Mesh

