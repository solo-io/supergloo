---
title: "meshctl demo osm init"
weight: 5
---
## meshctl demo osm init

Bootstrap an OSM demo with Service Mesh Hub

### Synopsis


Bootstrap an  OSM demo with Service Mesh Hub.

Running the Service Mesh Hub demo setup locally requires 4 tools to be installed and 
accessible via your PATH: kubectl >= v1.18.8, kind >= v0.8.1, OSM >= v0.3.0, and docker.
We recommend allocating at least 4GB of RAM for Docker.

This command will initialize a local kubernetes cluster using KinD. It will then install
all default OSM resources, which include the control-plane, prometheus, grafana, and zipkin. 
It will also install Service Mesh Hub, which includes the discovery, networking, and cert-agent
deployments.


```
meshctl demo osm init [flags]
```

### Options

```
  -h, --help   help for init
```

### SEE ALSO

* [meshctl demo osm](../meshctl_demo_osm)	 - Demo Service Mesh Hub functionality one OSM control plane deployed

