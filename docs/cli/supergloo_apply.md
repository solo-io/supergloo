---
title: "supergloo apply"
weight: 5
---
## supergloo apply

apply a rule to a mesh

### Synopsis

Creates or updates Rule resources which the SuperGloo controller 
will use to configure an installed mesh.

This set of commands creates Kubernetes CRDs which the SuperGloo controller
reads asynchronously.

To view these crds:

kubectl get routingrule [-n supergloo-system] 



### Options

```
  -h, --help   help for apply
```

### Options inherited from parent commands

```
  -i, --interactive   use interactive mode
```

### SEE ALSO

* [supergloo](../supergloo)	 - CLI for Supergloo
* [supergloo apply routingrule](../supergloo_apply_routingrule)	 - Apply a routing rule to one or more meshes.

