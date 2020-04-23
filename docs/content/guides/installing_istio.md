---
title: Installing Istio Multicluster
menuTitle: Introductory Guides
weight: 10
---


{{% notice note %}}
Be sure to satisfy the pre-requisites listed in the top-level page [here]({{% versioned_link_path fromRoot="/guides" %}})
{{% /notice %}}

We will assume in this guide that we have the following two contexts available in our kubeconfig file.
Your actual context names will likely be different.

* `management-plane-context`
    - kubeconfig context pointing to a cluster where we will install and operate Service Mesh Hub
* `remote-cluster-context`
    - kubeconfig context pointing to a cluster where we will install and manage a service mesh using Service Mesh Hub 
    
Note that these contexts need not be different. You may install and manage a service mesh in the same cluster as Service Mesh Hub. For the purposes of this guide, though, we will assume they are different, and that we have Service Mesh Hub deployed in the cluster pointed to by `management-plane-context`. See the [Setup Guide]({{% versioned_link_path fromRoot="/setup/" %}}) for installation options.

## Installing a mesh 

{{% notice warning %}}
We do not yet support automatically upgrading Istio in-place from versions 1.4 and earlier to 1.5, due to a number of
breaking changes across that version change. If you are currently running Istio prior to 1.5, you may have to
fully uninstall the mesh before attempting an installation of 1.5. 

<br/>

Users have reported seeing the following when attempting to upgrade in-place:

<br/>

https://discuss.istio.io/t/istio-upgrade-from-1-4-6-1-5-0-throws-istiod-errors-remote-error-tls-error-decrypting-message/5727

{{% /notice %}}

We can use `meshctl mesh` command to easily install a service mesh. This is accomplished by installing the
[Istio Operator](https://istio.io/blog/2019/introducing-istio-operator/) to the cluster and letting it handle the complex Istio installation process.

An easy way to get up and running quickly with Istio (**but insufficient for a multi-cluster demo**)
is by installing Istio in its "demo" profile 
([profile documentation](https://istio.io/docs/setup/additional-setup/config-profiles/)) is:

{{% notice note %}}
This will NOT install Istio suitable for a multi-cluster installation. For a correct multi-cluster installation, see the next section.
{{% /notice %}}

```shell
meshctl mesh install istio --profile=demo --context remote-cluster-context
```

All configuration profiles supported by Istio should be supported by `meshctl`.

To uninstall, you can leverage the `--dry-run` command from `meshctl` and pass to `kubectl delete`

```shell
meshctl mesh install istio --profile=demo --context remote-cluster-context --dry-run \
| k delete -f - --context remote-cluster-context
```

{{% notice note %}}
At times, the *finalizer* on the Istio CRD for IstioOperator hangs and can halt an uninstall. You can fix this by deleting the finalizer in this CR:

```
kubectl edit istiooperators.install.istio.io -n istio-operator
```
{{% /notice %}}

## Installing Istio for Multi Cluster Communication

We will install Istio with a suitable configuration for a multi-cluster demonstration by overriding some of the Istio Operator values. Let's install Istio in both the `management-plane-context` **AND** the `remote-cluster-context`

```shell
meshctl mesh install istio --context <context-name> --operator-spec=- <<EOF
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
