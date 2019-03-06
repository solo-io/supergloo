---
title: "supergloo install istio"
weight: 5
---
## supergloo install istio

install the Istio control plane

### Synopsis

install the Istio control plane

```
supergloo install istio [flags]
```

### Options

```
      --auto-inject                     enable auto-injection? (default true)
      --grafana                         add grafana to the install? (default true)
  -h, --help                            help for istio
      --installation-namespace string   which namespace to install Istio into? (default "istio-system")
      --jaeger                          add jaeger to the install? (default true)
      --mtls                            enable mtls? (default true)
      --name string                     name for the resource
      --namespace string                namespace for the resource (default "supergloo-system")
  -o, --output string                   output format: (yaml, json, table)
      --prometheus                      add prometheus to the install? (default true)
      --version string                  version of istio to install? available: [1.0.3 1.0.5 1.0.6] (default "1.0.6")
```

### Options inherited from parent commands

```
  -i, --interactive   use interactive mode
```

### SEE ALSO

* [supergloo install](../supergloo_install)	 - install a service mesh using Supergloo

