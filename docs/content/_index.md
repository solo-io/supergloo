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


<figure>
    <img src="{{% versioned_link_path fromRoot="/img/smh-diagram.png" %}}"/>
    <figcaption style="text-align: center"> Note: diagram reflects Service Mesh Hub <b>v0.6.1</b></figcaption>
</figure>


### Getting to know Service Mesh Hub

![Service Mesh Hub Architecture]({{% versioned_link_path fromRoot="/img/smh-3clusters.png" %}})

Service Mesh Hub can be run in its own cluster (or co-located with an existing mesh) and remotely operates and drives the configuration for specific service-mesh control planes. This allows Service Mesh Hub to discover meshes/workloads, establish federated identity, enable global traffic routing and load balancing, access control policy, centralized observability and more. We walk through each of these components in the following videos:

### Videos: Take a dive into Service Mesh Hub

We've put together [a handful of videos](https://www.youtube.com/watch?v=4sWikVELr5M&list=PLBOtlFtGznBjr4E9xYHH9eVyiOwnk1ciK) detailing the features of Service Mesh Hub.

## Contribution
There are many ways to get involved in an open source community and contribute to the project. Watch [this talk](https://www.youtube.com/watch?v=VE-igex6Lz4) to learn more about the architecture and how it works. 
- **Code:** If you're looking to hack on service mesh, check out the code and the contribution guide [here](https://docs.solo.io/service-mesh-hub/latest/contributing/) and look for the *good first issue* and *help wanted* labels in the GitHub issues. 
 - **Docs:** Contribute to the [Docs](docs/) or file issues for any docs bugs or requests [here](https://github.com/solo-io/service-mesh-hub/issues). 
 - **Talks and Blogs:** If you are interested in writing or speaking about Service Mesh Hub and would like access to content, images or help, [DM us here](https://solo-io.slack.com/archives/DHQ9J939V). Share your demos, tutorials and content back to the community. 

### Community Meetings 
Calls will be held every other Wednesday at **10am Pacific Time** and are open to the general public. These meetings will be recorded and posted publicly to YouTube. 
 - [Zoom meeting link](https://solo.zoom.us/j/98337720715) - open to the public and recorded
 - [Meeting Calendar](https://calendar.google.com/calendar/embed?src=solo.io_c144salt3ffnlfto3p1qnkbmdo%40group.calendar.google.com&ctz=America%2FLos_Angeles)
 - [Meeting notes](https://bit.ly/ServiceMeshHub-CommunityMeeting) - this document is open to the public
 - [Recorded meetings](https://www.youtube.com/playlist?list=PLBOtlFtGznBiF3Dti9WbPBjPj5KPmoalq) - YouTube playlist for all previous meetings

## Questions and Resources
If you have questions, please join the [#Service-Mesh-Hub channel](https://solo-io.slack.com/archives/CJQGK5TQ8) in the community slack. More information is available on the [website](https://www.solo.io/products/service-mesh-hub/) and in the [docs](https://docs.solo.io/service-mesh-hub/latest) 
- Learn more about [Open Source at Solo.io](https://www.solo.io/open-source/)
- Follow us on Twitter [@soloio_inc](https://twitter.com/soloio_inc)

## Thanks
Service Mesh Hub  would not be possible without the valuable open-source work of projects in the community.
