---
title: "Open Service Mesh"
menuTitle: OSM
description: How to get started using OSM with Service Mesh Hub
weight: 10
---

[Open Service Mesh (OSM)](https://openservicemesh.io/) is the brand new Open Source envoy proxy based 
service mesh from the folks at microsoft.

OSM comes with a new API, and a new set of usage patterns, but luckily Service Mesh Hub allows us to 
interact with OSM using our API.

## Installing OSM

OSM can be installed multiple different ways. The easiest method is using `meshctl`, but it can also be
installed directly using the `osm` binary, along with an existing kubernetes cluster.

To get up and Running with OSM simply run:
```shell script
meshctl demo osm init
``` 

This command will create a local KinD cluster, then install OSM as well as a demo application onto it.


