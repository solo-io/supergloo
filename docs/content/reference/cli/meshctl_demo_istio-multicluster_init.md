---
title: "meshctl demo istio-multicluster init"
weight: 5
---
## meshctl demo istio-multicluster init

Bootstrap a multicluster Istio demo with Service Mesh Hub

### Synopsis


Running the Service Mesh Hub demo setup locally requires 4 tools to be installed and 
accessible via your PATH: kubectl, kind, docker, and istioctl < v1.7.0.

This command will bootstrap 2 clusters, one of which will run the Service Mesh Hub
management-plane as well as Istio, and the other will just run Istio.


```
meshctl demo istio-multicluster init [flags]
```

### Options

```
  -h, --help   help for init
```

### SEE ALSO

* [meshctl demo istio-multicluster](../meshctl_demo_istio-multicluster)	 - Demo Service Mesh Hub functionality with two Istio control planes deployed on separate k8s clusters. Requires kubectl, kind, docker, and istioctl < v1.7.0.

