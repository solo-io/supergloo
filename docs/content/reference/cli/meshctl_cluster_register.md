---
title: "meshctl cluster register"
weight: 5
---
## meshctl cluster register

Register a Kubernetes cluster with Service Mesh Hub

### Synopsis

Register a Kubernetes cluster with Service Mesh Hub

```
meshctl cluster register [flags]
```

### Options

```
      --api-server-address string        Swap out the address of the remote cluster's k8s API server for the value of this flag. Set this flag when the address of the cluster domain used by the Service Mesh Hub is different than that specified in the local kubeconfig.
      --cert-agent-chart-file string     Path to a local Helm chart for installing the Certificate Agent. If unset, this command will install the Certificate Agent from the publicly released Helm chart.
      --cert-agent-chart-values string   Path to a Helm values.yaml file for customizing the installation of the Certificate Agent. If unset, this command will install the Certificate Agent with default Helm values.
      --cluster-domain string            The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. Defaults to 'cluster.local'. Read more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/
      --cluster-name string              name of the cluster to register
      --federation-namespace string      namespace of the Service-Mesh-Hub control plane in which the secret for the registered cluster will be created (default "service-mesh-hub")
  -h, --help                             help for register
      --install-wasm-agent               If true, install the wasm-agent on the cluster being registered.
      --kubeconfig string                path to the kubeconfig from which the registered cluster will be accessed
      --mgmt-context string              name of the kubeconfig context to use for the management cluster
      --remote-context string            name of the kubeconfig context to use for the remote cluster
      --remote-namespace string          namespace in the target cluster where a service account enabling remote access will be created. If the namespace does not exist it will be created. (default "service-mesh-hub")
      --verbose                          enable/disable verbose logging during installation of cert-agent (default true)
      --wasm-agent-chart-file string     Path to a local Helm chart for installing the Wasm Agent. If unset, this command will install the Wasm Agent from the publicly released Helm chart.
      --wasm-agent-chart-values string   Path to a Helm values.yaml file for customizing the installation of the Wasm Agent. If unset, this command will install the Wasm Agent with default Helm values.
```

### SEE ALSO

* [meshctl cluster](../meshctl_cluster)	 - Interacting with remote Kubernetes clusters registered to Service Mesh Hub

