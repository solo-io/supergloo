---
title: "meshctl debug snapshot networking"
weight: 5
---
## meshctl debug snapshot networking

Input and output snapshots for the networking pod

```
meshctl debug snapshot networking [flags]
```

### Options

```
  -h, --help   help for networking
```

### Options inherited from parent commands

```
  -f, --file string          file to write output to
      --json                 display the entire json snapshot. The output can be piped into a command like jq (https://stedolan.github.io/jq/tutorial/). For example:
                              meshctl debug snapshot discovery input | jq '.'
      --kubeconfig string    Path to the kubeconfig from which the management cluster will be accessed
      --kubecontext string   Name of the kubeconfig context to use for the management cluster
  -v, --verbose              Enable verbose logging
      --zip string           zip file output
```

### SEE ALSO

* [meshctl debug snapshot](../meshctl_debug_snapshot)	 - Input and Output snapshots for the discovery and networking pods. Requires jq to be installed if the --json flag is not being used.
* [meshctl debug snapshot networking input](../meshctl_debug_snapshot_networking_input)	 - Input snapshot for the networking pod
* [meshctl debug snapshot networking output](../meshctl_debug_snapshot_networking_output)	 - Output snapshot for the networking pod

