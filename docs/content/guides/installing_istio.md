---
title: Installing Istio Multicluster
menuTitle: Introductory Guides
weight: 10
---


We can use `meshctl mesh` command to easily install any supported service mesh. For Istio, this is accomplished by installing the [Istio Operator](https://istio.io/blog/2019/introducing-istio-operator/) to the cluster and letting it handle the complex Istio installation process.


{{% notice note %}}
Be sure to review the assumptions and satisfy the pre-requisites from the [Guides]({{% versioned_link_path fromRoot="/guides" %}}) top-level document.
{{% /notice %}}


{{% notice warning %}}
We do not yet support automatically upgrading Istio in-place from versions 1.4 and earlier to 1.5, due to a number of
breaking changes across that version change. If you are currently running Istio prior to 1.5, you may have to
fully uninstall the mesh before attempting an installation of 1.5. 

<br/>

Users have reported seeing the following when attempting to upgrade in-place:

<br/>

https://discuss.istio.io/t/istio-upgrade-from-1-4-6-1-5-0-throws-istiod-errors-remote-error-tls-error-decrypting-message/5727

{{% /notice %}}

## Istio quick install (single cluster)

An easy way to get up and running quickly with Istio (**but insufficient for a multi-cluster demo**) is by installing Istio in its "demo" profile ([profile documentation](https://istio.io/docs/setup/additional-setup/config-profiles/)) is:


```shell
meshctl mesh install istio --profile=demo --context remote-cluster-context
```
{{% notice note %}}
This will NOT install Istio suitable for a multi-cluster installation. For a correct multi-cluster installation, see the next section.
{{% /notice %}}

All configuration profiles supported by Istio should be supported by `meshctl`.

To uninstall Istio, you can leverage the `--dry-run` command from `meshctl` and pass to `kubectl delete`

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

## Istio quick install (multi cluster)

For following some of the other Istio guides, we assume two clusters with Istio installed for mult-cluster communication in both of them. We will install Istio with a suitable configuration for a multi-cluster demonstration by overriding some of the Istio Operator values. 

Let's install Istio into both the `management-plane-context` **AND** the `remote-cluster-context`

```shell
meshctl mesh install istio --context management-plane-context --operator-spec=- <<EOF
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

```shell
meshctl mesh install istio --context remote-cluster-context --operator-spec=- <<EOF
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

With Service Mesh Hub and Istio installed into the `management-plane-context`, and Istio installed into `remote-cluster-context`, we have an architecture that looks like this:

![Service Mesh Hub Architecture]({{% versioned_link_path fromRoot="/img/smh-2clusters.png" %}})

When the Istio Operator has finished the installation (can take up to 90 seconds),
you should see the Istio control plane pods running successfully:

```shell
NAME                                    READY   STATUS    RESTARTS   AGE
istio-ingressgateway-749468d8cd-x5w9p   1/1     Running   0          4h16m
istiod-58696868d5-gtvg8                 1/1     Running   0          4h16m
```

## Next steps

Now that we have Istio and Service Mesh Hub installed ([and appropriate clusters registered]({{% versioned_link_path fromRoot="/setup/#register-a-cluster" %}})), we can continue to explore the [discovery capabilities]({{% versioned_link_path fromRoot="/guides/discovery_intro" %}}) of Service Mesh Hub. 