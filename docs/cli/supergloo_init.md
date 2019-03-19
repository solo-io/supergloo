---
title: "supergloo init"
weight: 5
---
## supergloo init

install SuperGloo to a Kubernetes cluster

### Synopsis

Installs SuperGloo using default values based on the official helm chart located in install/helm/supergloo

The basic SuperGloo installation is composed of single-instance deployments for the supergloo-controller and discovery pods. 


```
supergloo init [flags]
```

### Options

```
  -d, --dry-run            Dump the raw installation yaml instead of applying it to kubernetes
  -f, --file string        Install SuperGloo from this Helm chart rather than from a release. Target file must be a tarball
  -h, --help               help for init
  -n, --namespace string   namespace to install supergloo into (default "supergloo-system")
      --release string     install from this release version. Should correspond with the name of the release on GitHub
  -v, --values string      Provide a custom values.yaml overrides for the installed helm chart. Leave empty to use default values from the chart.
```

### Options inherited from parent commands

```
  -i, --interactive   use interactive mode
```

### SEE ALSO

* [supergloo](../supergloo)	 - CLI for Supergloo

