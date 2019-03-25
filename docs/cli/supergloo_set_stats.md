---
title: "supergloo set stats"
weight: 5
---
## supergloo set stats

configure one or more prometheus instances to scrape a mesh for metrics.

### Synopsis

Updates the target mesh to propagate metrics to (have them scraped by) one or more instances of Prometheus.

```
supergloo set stats [flags]
```

### Options

```
  -h, --help                                     help for stats
      --prometheus-configmap ResourceRefsValue   resource reference to a prometheus configmap (used to configure prometheus in kubernetes) to which supergloo will ensure metrics are propagated. if empty, the any existing metric propagation will be disconnected. format must be <NAMESPACE>.<NAME> (default [])
      --target-mesh ResourceRefValue             resource reference the mesh for which you wish to configure metrics. format must be <NAMESPACE>.<NAME> (default { })
```

### SEE ALSO

* [supergloo set](../supergloo_set)	 - subcommands set a configuration parameter for one or more meshes

