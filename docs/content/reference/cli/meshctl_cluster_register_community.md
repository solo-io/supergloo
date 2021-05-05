---
title: "meshctl cluster register community"
weight: 5
---
## meshctl cluster register community

Register a cluster for Gloo Mesh community edition

### Synopsis

 In the process of registering a cluster, an agent to issue certificates will be
installed on the remote cluster.

```
meshctl cluster register community [cluster name] [flags]
```

### Examples

```
  meshctl cluster register --remote-context=my-context community remote-cluster
```

### Options

```
      --agent-crds-chart-file string     Path to a local Helm chart for installing CRDs needed by remote agents.
                                         If unset, this command will install the agent CRDs from the publicly released Helm chart.
      --api-server-address string        Swap out the address of the remote cluster's k8s API server for the value of this flag.
                                         Set this flag when the address of the cluster domain used by the Gloo Mesh is different than that specified in the local kubeconfig.
      --cert-agent-chart-file string     Path to a local Helm chart for installing the Certificate Agent.
                                         If unset, this command will install the Certificate Agent from the publicly released Helm chart.
      --cert-agent-chart-values string   Path to a Helm values.yaml file for customizing the installation of the Certificate Agent.
                                         If unset, this command will install the Certificate Agent with default Helm values.
  -h, --help                             help for community
```

### Options inherited from parent commands

```
      --cluster-domain string     The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. 
                                  Read more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/ (default "cluster.local")
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

