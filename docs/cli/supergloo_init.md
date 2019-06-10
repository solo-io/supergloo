---
title: "supergloo init"
weight: 5
---
## supergloo init

Installs SuperGloo to a Kubernetes cluster

### Synopsis

Installs SuperGloo using the official helm chart with default values.

The basic SuperGloo installation is composed of single-instance deployments for the supergloo-controller and discovery pods. 


```
supergloo init [flags]
```

### Options

```
  -d, --dry-run            Dump the raw installation yaml instead of applying it to kubernetes
  -f, --file string        Install SuperGloo from this Helm chart location (file path or URL). Target file must be a tarball
  -h, --help               help for init
  -n, --namespace string   Namespace to install supergloo into (default "supergloo-system")
      --release string     install from this release version. Should correspond with the name of the release on GitHub
  -v, --values string      Provide a custom values.yaml to override default values in the helm chart. Leave empty to use default values.
```

### SEE ALSO

* [supergloo](../supergloo)	 - CLI for Supergloo

