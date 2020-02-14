# Service Mesh Hub API

The Service Mesh Hub Api is broken into 4 main groups

1. **Discovery**: Resources representing discovered entities. These resources are (for the most part) not meant to be
configured directly by users, but rather discovered by service-mesh-hub. This currently includes the following:
    * `Meshes`
    * `MeshServices`
    * `MeshWorkloads`
    * `KubernetesCluster`

2. **Networking**: Resources representing user configuration of the management plane. This currently includes:
    * `AccessControlPolicy`
    * `TrafficPolicy`
    * `MeshGroup`

3. **Security**: Resources representing security entities and related workflows. Including:
    * `MeshGroupCertificateSigningRequest`

4. **Core**: This group has no custom resources of it's own, but rather serves as a shared library of types. This allows
the versions of the other groups to not be interdependent but rather jointly depend on a single core lib.
