---
title: "meshctl debug snapshot"
weight: 5
---
## meshctl debug snapshot

Input and Output snapshots for the discovery and networking pods. Requires jq to be installed if the --json flag is not being used.

```
meshctl debug snapshot [flags]
```

### Options

```
      --dir string           dir to write file outputs to
  -f, --file string          file to write output to
  -h, --help                 help for snapshot
      --json                 display the entire json snapshot. The output can be piped into a command like jq (https://stedolan.github.io/jq/tutorial/). For example:
                              meshctl debug snapshot discovery input | jq '.'
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
* [meshctl debug snapshot discovery](../meshctl_debug_snapshot_discovery)	 - Input and output snapshots for the discovery pod
* [meshctl debug snapshot enterprise-agent](../meshctl_debug_snapshot_enterprise-agent)	 - Input and output snapshots for the enterprise agent pod
* [meshctl debug snapshot enterprise-networking](../meshctl_debug_snapshot_enterprise-networking)	 - Input and output snapshots for the enterprise networking pod
* [meshctl debug snapshot networking](../meshctl_debug_snapshot_networking)	 - Input and output snapshots for the networking pod

