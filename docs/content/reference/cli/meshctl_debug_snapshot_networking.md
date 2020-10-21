---
title: "meshctl debug snapshot networking"
weight: 5
---
## meshctl debug snapshot networking

Input and output snapshots for the networking pod

### Synopsis

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
  -f, --file string   file to write output to
      --json          display the entire json snapshot. The output can be piped into a command like jq (https://stedolan.github.io/jq/tutorial/). For example:
                       meshctl debug snapshot discovery input | jq '.'
      --verbose       enables verbose/debug logging
      --zip string    zip file output
```

### SEE ALSO

* [meshctl debug snapshot](../meshctl_debug_snapshot)	 - Input and Output snapshots for the discovery and networking pod. Requires jq to be installed if the --json flag is not being used.
* [meshctl debug snapshot networking input](../meshctl_debug_snapshot_networking_input)	 - Input snapshot for the networking pod
* [meshctl debug snapshot networking output](../meshctl_debug_snapshot_networking_output)	 - Output snapshot for the networking pod

