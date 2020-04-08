<h2 align="center">
    <img src="https://www.solo.io/wp-content/uploads/2019/11/Solo_ServiceMesh_Logo_Dark_bg.svg">
    <br>
</h2>

Service Mesh Hub is a Kubernetes-native **management plane** that enables configuration 
and operational management of multiple heterogeneous service meshes across multiple 
clusters through a unified API. The Service Mesh Hub API integrates with the leading 
service meshes and  abstracts away differences between their disparate API's, allowing 
users to configure a set of different service meshes through a single API. Service 
Mesh Hub is engineered with a focus on its utility as an operational management 
tool, providing both graphical and command line UIs, observability features, and 
debugging tools.

![Architecture](docs/content/img/smh-diagram.png)

## Features

### Multi-mesh and multi-cluster

A core feature of Service Mesh Hub is its ability to configure and manage multiple 
service mesh deployments across multiple clusters. The Service Mesh Hub API provides 
an abstraction that allows users to configure groups of meshes as a single entity 
without having to deal with the underlying network configuration complexities.

### Heterogeneous meshes

Service Mesh Hub supports industry leading service mesh solutions. Its simple and 
powerful unified API allows users to easily utilize a variety of different service 
meshes without requiring expertise in any specific service mesh implementation.

### Simple and powerful API

Service Mesh Hub offers an API that emphasizes simplicity and ease of use without 
sacrificing functionality. Inherently complex service mesh configuration concepts 
(such as network routing, access control, security, etc.) are presented in a coherent 
representation supported by detailed, exhaustive documentation.

## Next Steps
- Join us on our Slack channel: [https://slack.solo.io/](https://slack.solo.io/)
- Follow us on Twitter: [https://twitter.com/soloio_inc](https://twitter.com/soloio_inc)
- Check out the [docs](https://docs.solo.io/service-mesh-hub/latest)
- Check out the code and contribute: [Contribution Guide](CONTRIBUTING.md)
- Contribute to the [Docs](docs/)
