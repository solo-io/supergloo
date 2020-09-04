---
title: "Open Service Mesh"
menuTitle: Open Service Mesh
description: How to get started using OSM with Service Mesh Hub
weight: 80
---

[Open Service Mesh (OSM)](https://openservicemesh.io/) is the brand new Open Source, Envoy proxy based, service mesh from Microsoft.

OSM introduces a new API set of usage patterns for managing a service mesh. OSM is supported by Service Mesh Hub, which can translate, configure, and manage instances of OSM in your environment. 

In this guide, we will walk through the process of installing OSM and a sample application. Then we will use *Access Policies* and *Traffic Policies* from Service Mesh Hub to configure the settings in OSM to allow communication between the services in the sample application. The sample application being installed is a variant of the Bookstore application. You can view the topology of the application [here](https://github.com/openservicemesh/osm/blob/main/img/book-thief-app-topology.jpg).

## Before you begin
To illustrate these concepts, we will assume that you have the following:

* Kubernetes cluster for installation or have Docker and Kind to run a local cluster
* The `meshctl` CLI tool for managing Service Mesh Hub
* The [OSM installer](https://github.com/openservicemesh/osm/releases)

## Installing OSM

OSM can be installed multiple different ways. The easiest method is using `meshctl`, but it can also be installed directly using the `osm` binary, along with an existing Kubernetes cluster.

### Install using `meshctl`

To get up and running with OSM simply run:

```shell script
meshctl demo osm init
``` 

This command will create a local Kind cluster, install OSM, install Service Mesh Hub, and deploy a sample application.

### Install manually

If you prefer to install the components yourself, first you will need to install OSM using the `osm` CLI tool.

```shell
osm install
kubectl rollout status deployment --timeout 300s -n osm-system osm-controller
```

After a few moments you should see a successful deployment of OSM.

```shell
Waiting for deployment "osm-controller" rollout to finish: 0 of 1 updated replicas are available...
deployment "osm-controller" successfully rolled out
```

Next we will install Service Mesh Hub, as outlined in the Setup guide for Service Mesh Hub. Be sure to replace the `cluster-name` and `remote-context` values with the correct values for your environment.

```shell
meshctl install
meshctl cluster register --cluster-name mgmt-cluster --remote-context mgmt-cluster-context
```

Finally, we will deploy the sample application.

```shell
kubectl create ns bookstore
kubectl create ns bookthief 
kubectl create ns bookwarehouse 
kubectl create ns bookbuyer

osm namespace add bookstore
osm namespace add bookthief 
osm namespace add bookwarehouse 
osm namespace add bookbuyer

kubectl apply -f https://raw.githubusercontent.com/solo-io/service-mesh-hub/v0.7.2/ci/osm-demo.yaml

kubectl rollout status deployment --timeout 300s -n bookstore bookstore-v1
kubectl rollout status deployment --timeout 300s -n bookstore bookstore-v2
kubectl rollout status deployment --timeout 300s -n bookthief bookthief
kubectl rollout status deployment --timeout 300s -n bookwarehouse bookwarehouse
kubectl rollout status deployment --timeout 300s -n bookbuyer bookbuyer
```

## OSM Basics

The OSM default operating mode is "restrictive" mode. This means that services cannot to talk to each other unless specifically allowed via the API.

The sample application installed is a variant of the Bookstore application with the following components:

* Book buyer - Sends requests to the Bookstore services
* Book thief - Sends requests to the Bookstore services
* Bookstore v1 - Sends requests to the Book warehouse service, receives requests from the Book buyer and thief
* Bookstore v2 - Sends requests to the Book warehouse service, receives requests from the Book buyer and thief
* Book warehouse - Receives requests from the Bookstore service

The restrictive mode of OSM means that **none** of the services can talk to each other yet.

To check the status of our services, we can port-forward to our the bookthief component of the application. 

```shell
kubectl port-forward -n bookthief deploy/bookthief 8080:80
```

Once the port-forward is open navigate to `localhost:8080` in any browser. The numbers should read `0` and `0`
as we have not enabled the bookthief or bookbuyer to interact with the bookstores.

![Bookthief no books]({{% versioned_link_path fromRoot="/img/bookthief-no-books.png" %}})

## Configuring OSM

In order to configure OSM to allow traffic between the various services, and properly split traffic between the two bookstores, we need to apply the following two resources. Be sure to update the `cluster_name` value as needed.

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

The first resource above is a Service Mesh Hub `AccessPolicy`. This is the resource which governs access between different services of the application. In this case we are taking advantage of the ability to specify multiple sources and destinations on a rule to allow communication from the bookthief to both bookstore instances. The configuration settings will manifest as `HTTPRouteGroup` and `TrafficTarget` custom resources for OSM.

The second resource is our `TrafficPolicy`. This resource governs how traffic is controlled, and manipulated within our application. This is how we split the traffic between the two separate bookstore instances. The configuration settings will manifest as `TrafficSplits` custom resources for OSM.

Once you have applied the two policies, the bookthief service will start "stealing" books from the bookstore. You can validate this by checking the bookthief page we set up a port-forward to earlier.

![Bookthief success]({{% versioned_link_path fromRoot="/img/bookthief-lotsa-books.png" %}})

## Next Steps

In this guide we installed the Open Service Mesh (OMS) service mesh on a Kubernetes cluster and managed it with Service Mesh Hub. From this simple example, we can expand to creating a [Virtual Mesh]({{% versioned_link_path fromRoot="/guides/federate_identity/" %}}) across clusters and [enabling multicluster communications]({{% versioned_link_path fromRoot="/guides/multicluster_communication/" %}}).
