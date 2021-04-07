---
title: Installing Istio Multicluster
menuTitle: Introductory Guides
weight: 10
---

We can use `istioctl` CLI to easily install Istio in our registered cluster. You can find `istioctl` on the [Getting Started page](https://istio.io/latest/docs/setup/getting-started/) of the Istio site. Currently, Gloo Mesh supports Istio versions 1.7, 1.8, and 1.9.

{{% notice note %}}
Be sure to review the assumptions and satisfy the pre-requisites from the [Guides]({{% versioned_link_path fromRoot="/guides" %}}) top-level document. If you used the `meshctl demo init` command, Istio has already been installed for you.
{{% /notice %}}

{{% notice note %}}
Gloo Mesh Enterprise users may be interested in [installing Gloo Mesh Istio]({{% versioned_link_path fromRoot="/setup/gloo_mesh_istio" %}}).
{{% /notice %}}

{{% notice note %}}
Istio versions 1.8.0, 1.8.1, and 1.8.2 have a [known issue](https://github.com/istio/istio/issues/28620) where sidecar proxies may fail to start
under specific circumstances. This bug may surface in sidecars configured by Failover Services. This issue is resolved in Istio 1.8.3.
{{% /notice %}}

In this guide we will walk you through two options for installing Istio for use with Gloo Mesh in a single cluster and multi-cluster setting. The instructions here are for reference only, and your installation process for Istio will likely be different depending on your organization's policies and procedures.

## Istio quick install (single cluster)

An easy way to get up and running quickly with Istio (**but insufficient for a multi-cluster demo**) is by installing Istio in its "demo" profile ([profile documentation](https://istio.io/docs/setup/additional-setup/config-profiles/)). Using the Istio command line tool, the command is as follows:


```shell
istioctl install --set profile=demo
```

{{% notice note %}}
This will NOT install Istio suitable for a multi-cluster installation. For a correct multi-cluster installation, see the next section.
{{% /notice %}}

## Istio quick install (multi cluster)

For following some of the Gloo Mesh guides, we assume two clusters with Istio installed for multi-cluster communication across both of them. 

We will install Istio with a suitable configuration for a multi-cluster demonstration by overriding some of the Istio Operator values.

Let's install Istio into both the management plane **AND** the remote cluster. The configuration below assumes that you are using the Kind clusters created in the Setup guide. The contexts for those clusters should be `kind-mgmt-cluster` and `kind-remote-cluster`. Both clusters are using a NodePort to expose the ingress gateway for Istio. If you are deploying Istio on different cluster setup, update your context and gateway settings accordingly.

The manifest for Istio 1.7 is different from other versions, so be sure to select the correct tab for the version of Istio you plan to install. The manifests below were tested using Istio 1.7.3, 1.8.3, and 1.9.0.

Before running the following, make sure your context names are set as environment variables:
```shell
MGMT_CONTEXT=your_management_plane_context
REMOTE_CONTEXT=your_remote_context
```

#### Management plane install

{{< tabs >}}
{{< tab name="Istio 1.8 and Istio 1.9" codelang="shell" >}}
cat << EOF | istioctl manifest install -y --context $MGMT_CONTEXT -f -
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: example-istiooperator
  namespace: istio-system
spec:
  profile: minimal
  meshConfig:
    enableAutoMtls: true
    defaultConfig:
      proxyMetadata:
        # Enable Istio agent to handle DNS requests for known hosts
        # Unknown hosts will automatically be resolved using upstream dns servers in resolv.conf
        ISTIO_META_DNS_CAPTURE: "true"
  components:
    # Istio Gateway feature
    ingressGateways:
    - name: istio-ingressgateway
      enabled: true
      k8s:
        env:
          - name: ISTIO_META_ROUTER_MODE
            value: "sni-dnat"
        service:
          type: NodePort
          ports:
            - port: 80
              targetPort: 8080
              name: http2
            - port: 443
              targetPort: 8443
              name: https
            - port: 15443
              targetPort: 15443
              name: tls
              nodePort: 32001
  values:
    global:
      pilotCertProvider: istiod
EOF
{{< /tab >}}
{{< tab name="Istio 1.7" codelang="shell" >}}
cat << EOF | istioctl manifest install --context $MGMT_CONTEXT -f -
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: mgmt-plane-istiooperator
  namespace: istio-system
spec:
  profile: minimal
  addonComponents:
    istiocoredns:
      enabled: true
  components:
    # Istio Gateway feature
    ingressGateways:
    - name: istio-ingressgateway
      enabled: true
      k8s:
        env:
          - name: ISTIO_META_ROUTER_MODE
            value: "sni-dnat"
        service:
          ports:
            - port: 80
              targetPort: 8080
              name: http2
            - port: 443
              targetPort: 8443
              name: https
            - port: 15443
              targetPort: 15443
              name: tls
              nodePort: 32001
  meshConfig:
    enableAutoMtls: true
  values:
    prometheus:
      enabled: false
    gateways:
      istio-ingressgateway:
        type: NodePort
        ports:
          - targetPort: 15443
            name: tls
            nodePort: 32001
            port: 15443
    global:
      pilotCertProvider: istiod
      controlPlaneSecurityEnabled: true
      podDNSSearchNamespaces:
      - global
EOF
{{< /tab >}}
{{< /tabs >}}

#### Remote cluster install

{{< tabs >}}
{{< tab name="Istio 1.8 and Istio 1.9" codelang="shell" >}}
cat << EOF | istioctl manifest install -y --context $REMOTE_CONTEXT -f -
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: example-istiooperator
  namespace: istio-system
spec:
  profile: minimal
  meshConfig:
    enableAutoMtls: true
    defaultConfig:
      proxyMetadata:
        # Enable Istio agent to handle DNS requests for known hosts
        # Unknown hosts will automatically be resolved using upstream dns servers in resolv.conf
        ISTIO_META_DNS_CAPTURE: "true"
  components:
    # Istio Gateway feature
    ingressGateways:
    - name: istio-ingressgateway
      enabled: true
      k8s:
        env:
          - name: ISTIO_META_ROUTER_MODE
            value: "sni-dnat"
        service:
          type: NodePort
          ports:
            - port: 80
              targetPort: 8080
              name: http2
            - port: 443
              targetPort: 8443
              name: https
            - port: 15443
              targetPort: 15443
              name: tls
              nodePort: 32000
  values:
    global:
      pilotCertProvider: istiod
EOF
{{< /tab >}}
{{< tab name="Istio 1.7" codelang="shell" >}}
cat << EOF | istioctl manifest install --context $REMOTE_CONTEXT -f -
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: remote-cluster-istiooperator
  namespace: istio-system
spec:
  profile: minimal
  addonComponents:
    istiocoredns:
      enabled: true
  components:
    # Istio Gateway feature
    ingressGateways:
    - name: istio-ingressgateway
      enabled: true
      k8s:
        env:
          - name: ISTIO_META_ROUTER_MODE
            value: "sni-dnat"
        service:
          ports:
            - port: 80
              targetPort: 8080
              name: http2
            - port: 443
              targetPort: 8443
              name: https
            - port: 15443
              targetPort: 15443
              name: tls
              nodePort: 32000
  meshConfig:
    enableAutoMtls: true
  values:
    prometheus:
      enabled: false
    gateways:
      istio-ingressgateway:
        type: NodePort
        ports:
          - targetPort: 15443
            name: tls
            nodePort: 32000
            port: 15443
    global:
      pilotCertProvider: istiod
      controlPlaneSecurityEnabled: true
      podDNSSearchNamespaces:
      - global
EOF
{{< /tab >}}
{{< /tabs >}}

With Gloo Mesh and Istio installed into the `mgmt-cluster`, and Istio installed into `remote-cluster`, we have an architecture that looks like this:

![Gloo Mesh Architecture]({{% versioned_link_path fromRoot="/img/gloomesh-2clusters.png" %}})

When the Istio Operator has finished the installation (can take up to 90 seconds),
you should see the Istio control plane pods running successfully:

```shell
kubectl get pods -n istio-system --context $REMOTE_CONTEXT
```

```shell
NAME                                    READY   STATUS    RESTARTS   AGE
istio-ingressgateway-746d597f7c-g6whv   1/1     Running   0          5d23h
istiod-7795ccf9dc-vr4cq                 1/1     Running   0          5d22h
```

{{% notice note %}}
The following section of the doc describes how to modify `coredns` to enable Istio DNS for the `.global` stub domain. This is not necessary when running Istio > 1.8.x
with Istio's new Smart DNS enabled. More information on this new DNS can be found in the following [blog post](https://istio.io/latest/blog/2020/dns-proxy/).
If you have installed Istio using the operator specs outlined above, this new DNS will be enabled by default.

When running Istio < 1.8.x, or when not using Istio DNS, the following changes add the `.global` stub domain for multicluster communication.
The following two blocks will enable Istio DNS for both clusters.
{{% /notice %}}

```shell
ISTIO_COREDNS=$(kubectl --context $MGMT_CONTEXT -n istio-system get svc istiocoredns -o jsonpath={.spec.clusterIP})
kubectl --context $MGMT_CONTEXT apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
data:
  Corefile: |
    .:53 {
        errors
        health
        ready
        kubernetes cluster.local in-addr.arpa ip6.arpa {
           pods insecure
           upstream
           fallthrough in-addr.arpa ip6.arpa
        }
        prometheus :9153
        forward . /etc/resolv.conf
        cache 30
        loop
        reload
        loadbalance
    }
    global:53 {
        errors
        cache 30
        forward . ${ISTIO_COREDNS}:53
    }
EOF
```

```shell
ISTIO_COREDNS=$(kubectl --context $REMOTE_CONTEXT -n istio-system get svc istiocoredns -o jsonpath={.spec.clusterIP})
kubectl --context $REMOTE_CONTEXT apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
data:
  Corefile: |
    .:53 {
        errors
        health
        ready
        kubernetes cluster.local in-addr.arpa ip6.arpa {
           pods insecure
           upstream
           fallthrough in-addr.arpa ip6.arpa
        }
        prometheus :9153
        forward . /etc/resolv.conf
        cache 30
        loop
        reload
        loadbalance
    }
    global:53 {
        errors
        cache 30
        forward . ${ISTIO_COREDNS}:53
    }
EOF
```

## Next steps

Now that we have Istio and Gloo Mesh installed ([and appropriate clusters registered]({{% versioned_link_path fromRoot="/setup/#register-a-cluster" %}})), we can continue to explore the [discovery capabilities]({{% versioned_link_path fromRoot="/guides/discovery_intro" %}}) of Gloo Mesh. 
