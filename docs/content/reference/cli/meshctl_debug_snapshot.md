---
title: "meshctl debug snapshot"
weight: 5
---
## meshctl debug snapshot

Input and Output snapshots for the discovery and networking pod

### Synopsis

The output can be piped into a command like jq. For example:
meshctl debug snapshot discovery input | jq 'to_entries | .[] | {kind: (.key), value: .value[]?} | {kind, name: .value.metadata?.name?, namespace: .value.metadata?.namespace?, cluster: .value.metadata?.clusterName?}'

```
meshctl debug snapshot [flags]
```

### Options

```
  -f, --file string   file to be read or written to
  -h, --help          help for snapshot
      --json          display the entire json snapshot (best used when piping the output into another command like jq)
      --zip           zip file output
```

### SEE ALSO

* [meshctl debug](../meshctl_debug)	 - Debug Service Mesh Hub resources
* [meshctl debug snapshot discovery](../meshctl_debug_snapshot_discovery)	 - for the discovery pod only
* [meshctl debug snapshot networking](../meshctl_debug_snapshot_networking)	 - for the networking pod only

