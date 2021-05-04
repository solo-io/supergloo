---
title: "meshctl cluster register enterprise"
weight: 5
---
## meshctl cluster register enterprise

Register a cluster for Gloo Mesh enterprise edition

### Synopsis

In the process of registering a cluster, an agent (called the relay agent)
will be installed on the remote cluster. To establish trust between the relay agent and
Gloo-Mesh, mTLS is used.

The relay agent can either be provided with a client certificate, or a bootstrap token. If provided
with a bootstrap token, the relay agent will then exchange it for a client certificate and save it
as a secret in the cluster. Once the client certificate secret exists, the bootstrap token is no
longer needed and can be discarded.

For the relay agent to trust Gloo Mesh a root CA is needed.

To make the registration process easy, this command will try to copy the root CA and 
bootstrap token from the gloo-mesh cluster, if these are not explicitly provided in the command line flags.


```
meshctl cluster register enterprise [cluster name] [flags]
```

### Examples

```
 meshctl cluster register --remote-context=my-context enterprise remote-cluster
```

### Options

```
      --client-cert-secret-name string         Secret name for the client cert for communication with relay server. Note that if a bootstrap token is provided, then a client certificate will be created automatically.
      --client-cert-secret-namespace string    Secret namespace for the client cert for communication with relay server.
      --enterprise-agent-chart-file string     Path to a local Helm chart for installing the Enterprise Agent.
                                               If unset, this command will install the Enterprise Agent from the publicly released Helm chart.
      --enterprise-agent-chart-values string   Path to a Helm values.yaml file for customizing the installation of the Enterprise Agent.
                                               If unset, this command will install the Enterprise Agent with default Helm values.
  -h, --help                                   help for enterprise
      --relay-server-address string            The address via which the enterprise agent will communicate with the relay server.
      --relay-server-insecure                  Communicate with the relay server over an insecure connection.
      --root-ca-secret-name string             Secret name for the root CA for communication with relay server.
      --root-ca-secret-namespace string        Secret namespace for the root CA for communication with relay server.
      --token-secret-key string                Secret data entry key for the bootstrap token. (default "token")
      --token-secret-name string               Secret name for the bootstrap token. This token will be used to bootstrap a client certificate from relay server. Not needed if you already have a client certificate.
      --token-secret-namespace string          Secret namespace for the bootstrap token.
```

### Options inherited from parent commands

```
      --cluster-domain string     The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. 
                                  Read more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/ (default "cluster.local")
      --kubeconfig string         path to the kubeconfig from which the registered cluster will be accessed
      --mgmt-context string       name of the kubeconfig context to use for the management cluster
      --mgmt-kubeconfig string    path to the kubeconfig for the managemtn plane if differente from which the registered cluster
      --mgmt-namespace string     namespace of the Gloo Mesh control plane in which the secret for the registered cluster will be created (default "gloo-mesh")
      --remote-context string     name of the kubeconfig context to use for the remote cluster
      --remote-namespace string   namespace in the target cluster where a service account enabling remote access will be created.
                                  If the namespace does not exist it will be created. (default "gloo-mesh")
  -v, --verbose                   Enable verbose logging
```

### SEE ALSO

* [meshctl cluster register](../meshctl_cluster_register)	 - Register a Kubernetes cluster with Gloo Mesh

