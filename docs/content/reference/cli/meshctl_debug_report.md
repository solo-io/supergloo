---
title: "meshctl debug report"
weight: 5
---
## meshctl debug report

meshctl debug report selectively captures cluster information and logs into an archive to help diagnose problems.

### Synopsis


Running this command requires

1) istioctl to be installed and accessible via your PATH.
2) a meshctl-config-file to be passed in. You can configure this file by running 'meschtl cluster config'.


```
meshctl debug report [flags]
```

### Options

```
  -f, --file string                                                name of the output tgz file (default "meshctl-bug-report.tgz")
  -h, --help                                                       help for report
  -c, --meshctl-config-file $HOME/.gloo-mesh/meshctl-config.yaml   path to the meshctl config file. defaults to $HOME/.gloo-mesh/meshctl-config.yaml
  -n, --namespace string                                           gloo-mesh namespace (default "gloo-mesh")
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging
```

### SEE ALSO

* [meshctl debug](../meshctl_debug)	 - Debug Gloo Mesh resources

