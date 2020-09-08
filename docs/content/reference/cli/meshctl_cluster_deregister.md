---
title: "meshctl cluster deregister"
weight: 5
---
## meshctl cluster deregister

Deregister a Kubernetes cluster from Service Mesh Hub, cleaning up any associated resources

### Synopsis

Deregister a Kubernetes cluster from Service Mesh Hub, cleaning up any associated resources

```
meshctl cluster deregister [flags]
```

### Options

```
      --api-server-address string     Swap out the address of the remote cluster's k8s API server for the value of this flag. Set this flag when the address of the cluster domain used by the Service Mesh Hub is different than that specified in the local kubeconfig.
      --cluster-name string           name of the cluster to deregister
      --federation-namespace string   namespace of the Service-Mesh-Hub control plane in which the secret for the deregistered cluster will be created (default "service-mesh-hub")
  -h, --help                          help for deregister
      --kubeconfig string             path to the kubeconfig from which the registered cluster will be accessed
      --mgmt-context string           name of the kubeconfig context to use for the management cluster
      --remote-context string         name of the kubeconfig context to use for the remote cluster
      --remote-namespace string       namespace in the target cluster where a service account enabling remote access will be created. If the namespace does not exist it will be created. (default "service-mesh-hub")
      --verbose                       enable/disable verbose logging during installation of cert-agent (default true)
```

### SEE ALSO

* [meshctl cluster](../meshctl_cluster)	 - Interacting with remote Kubernetes clusters registered to Service Mesh Hub

