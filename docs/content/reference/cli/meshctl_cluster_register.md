---
title: "meshctl cluster register"
weight: 5
---
## meshctl cluster register

Register a Kubernetes cluster with Gloo Mesh

### Synopsis

Registering a cluster installs the necessary components and does other setup
such as creating service accounts and cluster roles in order to start manging
the service mesh(es) on it.

The name of the context of the target cluster must be provided via the
--remote-context flag.

If the management cluster is different than the one that the current context
points then it an be provided via the --mgmt-context flag.

The edition registered must match the edition installed on the management cluster

### Options

```
      --cluster-domain string     The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. 
                                  Read more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/ (default "cluster.local")
  -h, --help                      help for register
      --kubeconfig string         path to the kubeconfig from which the registered cluster will be accessed
      --mgmt-context string       name of the kubeconfig context to use for the management cluster
      --mgmt-kubeconfig string    path to the kubeconfig file to use for the management cluster if different from control plane kubeconfig file location
      --mgmt-namespace string     namespace of the Gloo Mesh control plane in which the secret for the registered cluster will be created (default "gloo-mesh")
      --remote-context string     name of the kubeconfig context to use for the remote cluster
      --remote-namespace string   namespace in the target cluster where a service account enabling remote access will be created.
                                  If the namespace does not exist it will be created. (default "gloo-mesh")
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging
```

### SEE ALSO

* [meshctl cluster](../meshctl_cluster)	 - Manage Kubernetes clusters registered to Gloo Mesh
* [meshctl cluster register community](../meshctl_cluster_register_community)	 - Register a cluster for Gloo Mesh community edition
* [meshctl cluster register enterprise](../meshctl_cluster_register_enterprise)	 - Register a cluster for Gloo Mesh enterprise edition

