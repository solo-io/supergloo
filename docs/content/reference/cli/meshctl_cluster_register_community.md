---
title: "meshctl cluster register community"
weight: 5
---
## meshctl cluster register community

Register using the community certificate agent

```
meshctl cluster register community [flags]
```

### Options

```
      --agent-crds-chart-file string     Path to a local Helm chart for installing CRDs needed by remote agents. If unset, this command will install the agent CRDs from the publicly released Helm chart.
      --api-server-address string        Swap out the address of the remote cluster's k8s API server for the value of this flag. Set this flag when the address of the cluster domain used by the Gloo Mesh is different than that specified in the local kubeconfig.
      --cert-agent-chart-file string     Path to a local Helm chart for installing the Certificate Agent. If unset, this command will install the Certificate Agent from the publicly released Helm chart.
      --cert-agent-chart-values string   Path to a Helm values.yaml file for customizing the installation of the Certificate Agent. If unset, this command will install the Certificate Agent with default Helm values.
      --cluster-domain string            The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. Defaults to 'cluster.local'. Read more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/
      --cluster-name string              name of the cluster to register
  -h, --help                             help for community
      --kubeconfig string                path to the kubeconfig from which the registered cluster will be accessed
      --mgmt-context string              name of the kubeconfig context to use for the management cluster
      --mgmt-namespace string            namespace of the Gloo Mesh control plane in which the secret for the registered cluster will be created (default "gloo-mesh")
      --remote-context string            name of the kubeconfig context to use for the remote cluster
      --remote-namespace string          namespace in the target cluster where a service account enabling remote access will be created. If the namespace does not exist it will be created. (default "gloo-mesh")
```

### Options inherited from parent commands

```
  -v, --verbose   enable verbose logging
```

### SEE ALSO

* [meshctl cluster register](../meshctl_cluster_register)	 - Register a Kubernetes cluster with Gloo Mesh

