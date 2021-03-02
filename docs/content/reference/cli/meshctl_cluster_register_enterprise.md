---
title: "meshctl cluster register enterprise"
weight: 5
---
## meshctl cluster register enterprise

Register a cluster for Gloo Mesh enterprise editio

```
meshctl cluster register enterprise [cluster name] [flags]
```

### Examples

```
  # Register the current context
  meshctl cluster register enterprise mgmt-cluster

  # Register a different context when the current one is the management cluster
  meshctl cluster register --remote-context=my-context enterprise remote-cluster
```

### Options

```
      --enterprise-agent-chart-file string     Path to a local Helm chart for installing the Enterprise Agent.
                                               If unset, this command will install the Enterprise Agent from the publicly released Helm chart.
      --enterprise-agent-chart-values string   Path to a Helm values.yaml file for customizing the installation of the Enterprise Agent.
                                               If unset, this command will install the Enterprise Agent with default Helm values.
  -h, --help                                   help for enterprise
      --relay-server-address string            The address via which the enterprise agent will communicate with the relay server.
```

### Options inherited from parent commands

```
      --cluster-domain string     The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. Defaults to 'cluster.local'.
                                  Read more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/
      --kubeconfig string         path to the kubeconfig from which the registered cluster will be accessed
      --mgmt-context string       name of the kubeconfig context to use for the management cluster
      --mgmt-namespace string     namespace of the Gloo Mesh control plane in which the secret for the registered cluster will be created (default "gloo-mesh")
      --remote-context string     name of the kubeconfig context to use for the remote cluster
      --remote-namespace string   namespace in the target cluster where a service account enabling remote access will be created.
                                  If the namespace does not exist it will be created. (default "gloo-mesh")
  -v, --verbose                   Enable verbose logging
```

### SEE ALSO

* [meshctl cluster register](../meshctl_cluster_register)	 - Register a Kubernetes cluster with Gloo Mesh

