---
title: "Registered ClusterRole Definition"
menuTitle: Registered ClusterRole
description: ClusterRole created when registering a cluster with Gloo Mesh.
weight: 10
---

The following YAML shows the ClusterRole created on a target cluster when it is registered with Gloo Mesh.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: gloomesh-remote-access
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - services
  - configmaps
  - nodes
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - appmesh.k8s.aws
  resources:
  - meshes
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - apps
  resources:
  - deployments
  - replicasets
  - daemonsets
  - statefulsets
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - '*'
- apiGroups:
  - certificates.mesh.gloo.solo.io
  resources:
  - issuedcertificates
  - podbouncedirectives
  verbs:
  - '*'
- apiGroups:
  - networking.istio.io
  resources:
  - destinationrules
  - virtualservices
  - envoyfilters
  - serviceentries
  - gateways
  verbs:
  - '*'
- apiGroups:
  - security.istio.io
  resources:
  - authorizationpolicies
  verbs:
  - '*'
- apiGroups:
  - access.smi-spec.io
  resources:
  - traffictargets
  verbs:
  - '*'
- apiGroups:
  - specs.smi-spec.io
  resources:
  - httproutegroups
  verbs:
  - '*'
- apiGroups:
  - split.smi-spec.io
  resources:
  - trafficsplits
  verbs:
  - '*'
- apiGroups:
  - certificates.mesh.gloo.solo.io
  resources:
  - issuedcertificates
  - certificaterequests
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - certificates.mesh.gloo.solo.io
  resources:
  - issuedcertificates/status
  - certificaterequests/status
  verbs:
  - get
  - update
```