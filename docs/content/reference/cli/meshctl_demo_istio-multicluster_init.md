---
title: "meshctl demo istio-multicluster init"
weight: 5
---
## meshctl demo istio-multicluster init

Bootstrap a multicluster Istio demo with Service Mesh Hub

### Synopsis


Bootstrap a multicluster Istio demo with Service Mesh Hub.

Running the Service Mesh Hub demo setup locally requires 4 tools to be installed and 
accessible via your PATH: kubectl >= v1.18.8, kind >= v0.8.1, istioctl, and docker.
We recommend allocating at least 8GB of RAM for Docker.

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

* [meshctl demo istio-multicluster](../meshctl_demo_istio-multicluster)	 - Demo Service Mesh Hub functionality with two Istio control planes deployed on separate k8s clusters

