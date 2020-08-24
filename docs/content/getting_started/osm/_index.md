---
title: "Open Service Mesh"
menuTitle: OSM
description: How to get started using OSM with Service Mesh Hub
weight: 10
---

[Open Service Mesh (OSM)](https://openservicemesh.io/) is the brand new Open Source envoy proxy based 
service mesh from Microsoft.

OSM comes with a new API and a new set of usage patterns which Service Mesh Hub can translate and manage.

## Installing OSM

OSM can be installed multiple different ways. The easiest method is using `meshctl`, but it can also be
installed directly using the `osm` binary, along with an existing kubernetes cluster.

To get up and Running with OSM simply run:
```shell script
meshctl demo osm init
``` 

This command will create a local KinD cluster, then install OSM, as well as a demo application onto it.
A topology of the app which was installed can be found [here](https://github.com/openservicemesh/osm/blob/main/img/book-thief-app-topology.jpg).

## OSM Basics

The OSM default operating mode is "restrictive" mode. This means that services cannot to talk to 
each other unless specifically allowed via the API.

To check the status of our services, we can port-forward to our application. `kubectl port-forward -n bookthief deploy/bookthief 8080:80`.
Once the port-forward is open navigate to `localhost:8080` in any browser. The numbers should read `0` and `0`
as we have not enabled the bookthief, or bookbuyer to interact with the bookstores.

## Configuring OSM

In order to configure OSM to allow traffic between the various services, and properly split traffic between
the 2 bookstores, we need to apply the following 2 resources.
```shell script
kubectl apply -f - <<EOF
apiVersion: networking.smh.solo.io/v1alpha2
kind: AccessPolicy
metadata:
  name: my-policy
  namespace: default
spec:
  destination_selector:
  - kube_service_refs:
      services:
      - cluster_name: mgmt-cluster
        name: bookstore-v1
        namespace: bookstore
      - cluster_name: mgmt-cluster
        name: bookstore-v2
        namespace: bookstore
  source_selector:
  - kube_service_account_refs:
      service_accounts:
      - cluster_name: mgmt-cluster
        name: bookthief
        namespace: bookthief
      - cluster_name: mgmt-cluster
        name: bookstore
        namespace: bookstore

---

apiVersion: networking.smh.solo.io/v1alpha2
kind: TrafficPolicy
metadata:
  name: my-policy
  namespace: default
spec:
  traffic_shift:
    destinations:
    - kube_service:
        cluster_name: mgmt-cluster
        name: bookstore-v1
        namespace: bookstore
      weight: 50
    - kube_service:
        cluster_name: mgmt-cluster
        name: bookstore-v2
        namespace: bookstore
      weight: 50
  destination_selector:
  - kube_service_refs:
      services:
      - cluster_name: mgmt-cluster
        name: bookstore
        namespace: bookstore
EOF
```

The first resource above is a Service Mesh Hub `AccessPolicy`. This is the resource which governs access
between different services of the application. In this case we are taking advantage of the ability to specify
multiple sources and destinations on any rule to allow communication from the bookbuyer, and bookthief to both 
bookstores.

The second resource is our `TrafficPolicy`. This resource governs how traffic is controlled, and manipulated 
within our application. This is how we split the traffic between the 2 separate bookstore versions.
