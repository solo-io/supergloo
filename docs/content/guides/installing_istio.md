---
title: Installing Istio
menuTitle: Introductory Guides
weight: 1
---

## Pre-Guide Notes

{{% notice note %}}
Be sure to satisfy the pre-requisites listed in the top-level page [here]({{% versioned_link_path fromRoot="/guides" %}})
{{% /notice %}}

We will assume in this guide that we have the following two contexts available in our kubeconfig file.
Your actual context names will likely be different.

* management-plane-context
    - kubeconfig context pointing to a cluster where we will install and operate Service Mesh Hub
* remote-cluster-context
    - kubeconfig context pointing to a cluster where we will install and manage a service mesh using Service Mesh Hub 
    
Note that these contexts need not be different; you may install and manage a service mesh in the same cluster as Service Mesh Hub.
For the purposes of this guide, though, we will assume they are different, and that we have Service Mesh
Hub deployed in the cluster pointed to by `management-plane-context`.

## Guide

We can use `meshctl` to easily install Istio. This is accomplished by installing the
[Istio Operator](https://istio.io/blog/2019/introducing-istio-operator/) to the cluster
and letting it handle the complex Istio installation process.

An easy way to get up and running quickly with Istio (but insufficient for a multicluster demo)
is by installing Istio in its "demo" profile 
([profile documentation](https://istio.io/docs/setup/additional-setup/config-profiles/)) is:

```shell
meshctl mesh install istio --profile=demo --context remote-cluster-context
```

All configuration profiles supported by Istio should be supported by `meshctl`.

However, we will install Istio in a configuration that will lend itself well to our
multicluster demonstration:

```shell
meshctl mesh install istio --operator-spec=- <<EOF
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: istiocontrolplane-default
  namespace: istio-operator
spec:
  profile: default
  values:
    global:
      controlPlaneSecurityEnabled: true
      mtls:
        enabled: true
      pilotCertProvider: kubernetes
      podDNSSearchNamespaces:
      - global
      - '{{ valueOrDefault .DeploymentMeta.Namespace "default" }}.global'
    prometheus:
      enabled: false
    security:
      selfSigned: false
EOF
```

When the Istio Operator has finished the installation (can take up to 90 seconds),
you should see the Istio control plane pods running successfully:

```shell
NAME                                    READY   STATUS    RESTARTS   AGE
istio-ingressgateway-749468d8cd-x5w9p   1/1     Running   0          4h16m
istiod-58696868d5-gtvg8                 1/1     Running   0          4h16m
```
