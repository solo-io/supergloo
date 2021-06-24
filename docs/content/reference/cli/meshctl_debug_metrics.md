---
title: "meshctl debug metrics"
weight: 5
---
## meshctl debug metrics

metrics for the discovery and networking pods.

```
meshctl debug metrics [flags]
```

### Options

```
      --dir string           dir to write file outputs to
  -f, --file string          file to write output to
  -h, --help                 help for metrics
      --kubeconfig string    Path to the kubeconfig from which the management cluster will be accessed
      --kubecontext string   Name of the kubeconfig context to use for the management cluster
      --zip string           zip file output
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging
```

### SEE ALSO

* [meshctl debug](../meshctl_debug)	 - Debug Gloo Mesh resources
* [meshctl debug metrics discovery](../meshctl_debug_metrics_discovery)	 - Input and output snapshots for the discovery pod
* [meshctl debug metrics enterprise-agent](../meshctl_debug_metrics_enterprise-agent)	 - Input and output snapshots for the enterprise agent pod
* [meshctl debug metrics enterprise-networking](../meshctl_debug_metrics_enterprise-networking)	 - Input and output snapshots for the enterprise networking pod
* [meshctl debug metrics networking](../meshctl_debug_metrics_networking)	 - Input and output snapshots for the networking pod

