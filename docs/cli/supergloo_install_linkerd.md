---
title: "supergloo install linkerd"
weight: 5
---
## supergloo install linkerd

install the Linkerd control plane

### Synopsis

install the Linkerd control plane

```
supergloo install linkerd [flags]
```

### Options

```
      --auto-inject                     enable auto-injection? (default true)
  -h, --help                            help for linkerd
      --installation-namespace string   which namespace to install Linkerd into? (default "linkerd")
      --mtls                            enable mtls? (default true)
      --version string                  version of linkerd to install? available: [stable-2.3.0] (default "stable-2.3.0")
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

