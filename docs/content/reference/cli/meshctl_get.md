---
title: "meshctl get"
weight: 5
---
## meshctl get

Examine Service Mesh Hub resources

### Synopsis

Examine Service Mesh Hub resources

```
meshctl get [flags]
```

### Options

```
  -h, --help            help for get
  -o, --output string   Output format. Valid values: [pretty, json, yaml] (default "pretty")
```

### Options inherited from parent commands

```
      --context string          Specify which context from the kubeconfig should be used; uses current context if none is specified
      --kube-timeout duration   Specify the default timeout for requests to kubernetes API servers (default 5s)
      --kubeconfig string       Specify the kubeconfig for the current command
  -n, --namespace string        Specify the namespace where Service Mesh Hub resources should be written (default "service-mesh-hub")
  -v, --verbose                 Enable verbose mode, which outputs additional execution details that may be helpful for debugging
```

### SEE ALSO

* [meshctl](../meshctl)	 - CLI for Service Mesh Hub
* [meshctl get clusters](../meshctl_get_clusters)	 - Examine registered kubernetes clusters
* [meshctl get meshes](../meshctl_get_meshes)	 - Examine discovered meshes
* [meshctl get services](../meshctl_get_services)	 - Examine discovered mesh services
* [meshctl get virtualmeshcertificatesigningrequests](../meshctl_get_virtualmeshcertificatesigningrequests)	 - Examine configured virtual mesh ceriticate signing request
* [meshctl get virtualmeshes](../meshctl_get_virtualmeshes)	 - Examine configured virtual meshes
* [meshctl get workloads](../meshctl_get_workloads)	 - Examine discovered mesh workloads

