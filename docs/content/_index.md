---
weight: 99
title: Service Mesh Hub
---

# A Multi Mesh Management Tool

## What is Service Mesh Hub

Service Mesh Hub is a Kubernetes-native **management plane** that enables configuration 
and operational management of multiple heterogeneous service meshes across multiple 
clusters through a unified API. The Service Mesh Hub API integrates with the leading 
service meshes and  abstracts away differences between their disparate API's, allowing 
users to configure a set of different service meshes through a single API. Service 
Mesh Hub is engineered with a focus on its utility as an operational management 
tool, providing both graphical and command line UIs, observability features, and 
debugging tools.

![Service Mesh Hub Architecture]({{% versioned_link_path fromRoot="/img/smh-diagram.png" %}})

## Community

* Join us on our Slack channel: [https://slack.solo.io/](https://slack.solo.io/)
* Follow us on Twitter: [https://twitter.com/soloio_inc](https://twitter.com/soloio_inc)
* Contribute on Github: [https://github.com/solo-io/service-mesh-hub](https://github.com/solo-io/service-mesh-hub)

### Getting to know Service Mesh Hub

![Service Mesh Hub Architecture]({{% versioned_link_path fromRoot="/img/smh-3clusters.png" %}})

Service Mesh Hub can be run in its own cluster (or co-located with an existing mesh) and remotely operates and drives the configuration for specific service-mesh control planes. This allows Service Mesh Hub to discover meshes/workloads, establish federated identity, enable global traffic routing and load balancing, access control policy, centralized observability and more. We walk through each of these components in the following videos:

#### Part One

In this video, we take a look at installing Service Mesh Hub, registering some clusters, and discovering Istio meshes, workloads and services. 

<br><br>
<iframe width="560" height="315" src="https://www.youtube.com/embed/4sWikVELr5M" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>


#### Part Two

In this video, we look at federating shared identity across the two clusters we registered in the previous video. We look at all the steps that Service Mesh Hub to securely establish a root trust domain including creating intermediate certificates, signing them, and never distributing keys across the network:

<br><br>
<iframe width="560" height="315" src="https://www.youtube.com/embed/djcDaIsqIl8" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>


#### Part Three

In this video we 

<br><br>
<iframe width="560" height="315" src="https://www.youtube.com/embed/cG1VCx9G408" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>


#### Part Four

<br><br>
<iframe width="560" height="315" src="https://www.youtube.com/embed/HAr1Mw1bxB4" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>




---

## Thanks

Service Mesh Hub  would not be possible without the valuable open-source work of projects in the community.