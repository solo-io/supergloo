---
title: "supergloo register appmesh"
weight: 5
---
## supergloo register appmesh

Register an AWS App Mesh with SuperGloo

### Synopsis

Creates a SuperGloo Mesh object representing an AWS App Mesh. The object will contain the information required to 
connect to the App Mesh control plane in AWS and, optionally, to automatically inject new pods with the AWS App Mesh sidecar proxy

```
supergloo register appmesh [flags]
```

### Options

```
      --auto-inject string                   determines whether auto-injection will be enabled for this mesh (default "true")
      --configmap ResourceRefValue           config map that contains the patch to be applied to the pods matching the selector. Format must be NAMESPACE.NAME (default { })
  -h, --help                                 help for appmesh
      --region string                        AWS region the AWS App Mesh control plane resources (Virtual Nodes, Virtual Routers, etc.) will be created in
      --secret ResourceRefValue              secret holding AWS access credentials. Format must be NAMESPACE.NAME (default { })
      --select-labels MapStringStringValue   auto-inject pods with these labels. Format must be KEY=VALUE (default [])
      --select-namespaces strings            auto-inject pods matching these labels
      --virtual-node-label string            If auto-injection is enabled, the value of the pod label with this key will be used to calculate the value of APPMESH_VIRTUAL_NODE_NAME environment variable that is set on the injected sidecar proxy container.
```

### Options inherited from parent commands

```
  -i, --interactive        run in interactive mode
      --name string        name for the resource
      --namespace string   namespace for the resource (default "supergloo-system")
  -o, --output string      output format: (yaml, json, table)
```

### SEE ALSO

* [supergloo register](../supergloo_register)	 - commands for registering meshes with SuperGloo

