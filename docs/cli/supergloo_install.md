---
title: "supergloo install"
weight: 5
---
## supergloo install

install a service mesh using Supergloo

### Synopsis

Creates an Install resource which the supergloo controller 
will use to install a service mesh.

Installs represent a desired installation of a supported mesh.
Supergloo watches for installs and synchronizes the managed installations
with the desired configuration in the install object.

Updating the configuration of an install object will cause supergloo to 
modify the corresponding mesh.




### Options

```
  -h, --help               help for install
  -i, --interactive        run in interactive mode
      --name string        name for the resource
      --namespace string   namespace for the resource (default "supergloo-system")
  -o, --output string      output format: (yaml, json, table)
      --timeout duration   maximum time to wait for a mesh installation to complete (default 5m0s)
      --update             update an existing install?
```

### SEE ALSO

* [supergloo](../supergloo)	 - CLI for Supergloo
* [supergloo install gloo](../supergloo_install_gloo)	 - gloo installation
* [supergloo install istio](../supergloo_install_istio)	 - install the Istio control plane
* [supergloo install linkerd](../supergloo_install_linkerd)	 - install the Linkerd control plane

