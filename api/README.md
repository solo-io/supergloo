# Service Mesh Hub API

The Service Mesh Hub Api is broken into 4 main groups

1. **Discovery**: Resources representing discovered entities. These resources are (for the most part) not meant to be
configured directly by users, but rather discovered by service-mesh-hub. This currently includes the following:
    * `Meshes`
    * `Workloads`
    * `TrafficTargets`

2. **Networking**: Resources representing user configuration of the management plane. This currently includes:
    * `AccessPolicy`
    * `TrafficPolicy`
    * `VirtualMesh`
    * `FailoverService`

3. **Certificates**: Resources representing security entities and related workflows. Including:
    * `CertificateRequest`
    * `IssuedCertificate`

