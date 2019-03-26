---
title: "supergloo install gloo"
weight: 5
---
## supergloo install gloo

gloo installation

### Synopsis

gloo installation

```
supergloo install gloo [flags]
```

### Options

```
  -h, --help                            help for gloo
      --installation-namespace string   which namespace to install Gloo into? (default "gloo-system")
  -m, --meshes strings                  Which meshes to target (comma seperated list) <namespace>.<name> (default [supergloo-system.istio])
      --version string                  version of gloo to install? available: [latest v0.13.0] (default "latest")
```

### Options inherited from parent commands

```
  -i, --interactive        run in interactive mode
      --name string        name for the resource
      --namespace string   namespace for the resource (default "supergloo-system")
  -o, --output string      output format: (yaml, json, table)
      --update             update an existing install?
```

### SEE ALSO

* [supergloo install](../supergloo_install)	 - install a service mesh using Supergloo

