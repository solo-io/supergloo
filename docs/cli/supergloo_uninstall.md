---
title: "supergloo uninstall"
weight: 5
---
## supergloo uninstall

uninstall a service mesh using Supergloo

### Synopsis

Disables an Install resource which the Supergloo controller 
will use to uninstall an installed service mesh.

This only works for meshes that Supergloo installed.

Installs represent a desired installation of a supported mesh.
Supergloo watches for installs and synchronizes the managed installations
with the desired configuration in the install object. When an install is 
disabled, Supergloo will remove corresponding installed components from the cluster.


```
supergloo uninstall [flags]
```

### Options

```
  -h, --help               help for uninstall
      --name string        name for the resource
      --namespace string   namespace for the resource (default "supergloo-system")
  -o, --output string      output format: (yaml, json, table)
```

### Options inherited from parent commands

```
  -i, --interactive   use interactive mode
```

### SEE ALSO

* [supergloo](../supergloo)	 - CLI for Supergloo

