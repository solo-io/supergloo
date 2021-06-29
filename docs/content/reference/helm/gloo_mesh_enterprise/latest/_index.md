
---
title: "v1.1.0-beta19"
description: Reference for Helm values. 
weight: 2
---

The following pages provide reference documentation for Helm values for the various Gloo Mesh
components. These components include:

1. **Open source Gloo Mesh**: the OSS version of Gloo Mesh
2. **Enterprise Networking (enterprise only)**: the management plane of Gloo Mesh Enterprise, deployed on the management cluster
3. **Enterprise Agent (enterprise only)**: the agent of Gloo Mesh Enterprise, deployed on each managed cluster
4. **RBAC Webhook (enterprise only)**: the Kubernetes webhook that enforces Gloo Mesh Enterprise's role-based API
5. **Gloo Mesh UI (enterprise only)**: the UI for Gloo Mesh Enterprise

Note that when providing Helm values for the bundled Gloo Mesh Enterprise chart 
(located at https://storage.googleapis.com/gloo-mesh-enterprise/gloo-mesh-enterprise),
values for each subchart must be prefixed accordingly:

1. Values for the RBAC Webhook must be prefixed with "rbac-webhook".
2. Values for Enterprise Networking must be prefixed with "enterprise-networking".
3. Values for the Gloo Mesh UI must be prefixed with "gloo-mesh-ui".


{{% children description="true" %}}
