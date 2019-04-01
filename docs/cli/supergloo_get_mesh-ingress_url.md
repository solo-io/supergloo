---
title: "supergloo get mesh-ingress url"
weight: 5
---
## supergloo get mesh-ingress url

get proxy url for mesh ingress

### Synopsis

get proxy url for mesh ingress

```
supergloo get mesh-ingress url [flags]
```

### Options

```
  -h, --help                           help for url
  -l, --local-cluster                  use when the target kubernetes cluster is running locally, e.g. in minikube or minishift. this will default to true if LoadBalanced services are not assigned external IPs by your cluster
      --port string                    the name of the service port to connect to (default "http")
  -t, --target-mesh ResourceRefValue   target mesh (default { })
```

### Options inherited from parent commands

```
  -i, --interactive        run in interactive mode
      --name string        name for the resource
      --namespace string   namespace for the resource (default "supergloo-system")
```

### SEE ALSO

* [supergloo get mesh-ingress](../supergloo_get_mesh-ingress)	 - retrieve information regarding an installed mesh-ingress

