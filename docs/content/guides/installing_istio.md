---
title: Installing Istio Multicluster
menuTitle: Introductory Guides
weight: 10
---

We can use `istioctl` CLI to easily install Istio in our registered cluster. You can find `istioctl` on the [Getting Started page](https://istio.io/latest/docs/setup/getting-started/) of the Istio site. Currently, Service Mesh Hub supports Istio versions 1.5 and 1.6.

{{% notice note %}}
Be sure to review the assumptions and satisfy the pre-requisites from the [Guides]({{% versioned_link_path fromRoot="/guides" %}}) top-level document. If you used the `meshctl demo init` command, Istio has already been installed for you.
{{% /notice %}}


In this guide we will walk you through two options for installing Istio for use with Service Mesh Hub in a single cluster and multi-cluster setting. The instructions here are for reference only, and your installation process for Istio will likely be different depending on your organization's policies and procedures.

## Istio quick install (single cluster)

An easy way to get up and running quickly with Istio (**but insufficient for a multi-cluster demo**) is by installing Istio in its "demo" profile ([profile documentation](https://istio.io/docs/setup/additional-setup/config-profiles/)). Using the Istio command line tool, the command is as follows:


```shell
istioctl install --set profile=demo
```

{{% notice note %}}
This will NOT install Istio suitable for a multi-cluster installation. For a correct multi-cluster installation, see the next section.
{{% /notice %}}

## Istio quick install (multi cluster)

For following some of the other Service Mesh Hub guides, we assume two clusters with Istio installed for multi-cluster communication across both of them. 

We will install Istio with a suitable configuration for a multi-cluster demonstration by overriding some of the Istio Operator values.

Let's install Istio into both the management plane **AND** the remote cluster. The configuration below assumes that you are using the Kind clusters created in the Setup guide. The contexts for those clusters should be `kind-management-plane` and `kind-remote-cluster`. Both clusters are using a NodePort to expose the ingress gateway for Istio. If you are deploying Istio on different cluster setup, update your context and gateway settings accordingly.

```shell
cat << EOF | istioctl manifest apply --context kind-management-plane -f -
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
    pilot:
      k8s:
        env:
          - name: PILOT_CERT_PROVIDER
            value: "kubernetes"
    proxy:
      k8s:
        env:
          - name: PILOT_CERT_PROVIDER
            value: "kubernetes"
    # Istio Gateway feature
    ingressGateways:
    - name: istio-ingressgateway
      enabled: true
      k8s:
        env:
          - name: ISTIO_META_ROUTER_MODE
            value: "sni-dnat"
          - name: PILOT_CERT_PROVIDER
            value: "kubernetes"
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
      pilotCertProvider: kubernetes
      controlPlaneSecurityEnabled: true
      mtls:
        enabled: true
      podDNSSearchNamespaces:
      - global
EOF
```

```shell
cat << EOF | istioctl manifest apply --context kind-remote-cluster -f -
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
    pilot:
      k8s:
        env:
          - name: PILOT_CERT_PROVIDER
            value: "kubernetes"
    proxy:
      k8s:
        env:
          - name: PILOT_CERT_PROVIDER
            value: "kubernetes"
    # Istio Gateway feature
    ingressGateways:
    - name: istio-ingressgateway
      enabled: true
      k8s:
        env:
          - name: ISTIO_META_ROUTER_MODE
            value: "sni-dnat"
          - name: PILOT_CERT_PROVIDER
            value: "kubernetes"
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
      pilotCertProvider: kubernetes
      controlPlaneSecurityEnabled: true
      mtls:
        enabled: true
      podDNSSearchNamespaces:
      - global
EOF
```

With Service Mesh Hub and Istio installed into the `management-plane-context`, and Istio installed into `remote-cluster-context`, we have an architecture that looks like this:

![Service Mesh Hub Architecture]({{% versioned_link_path fromRoot="/img/smh-2clusters.png" %}})

When the Istio Operator has finished the installation (can take up to 90 seconds),
you should see the Istio control plane pods running successfully:

```shell
kubectl get pods -n istio-system

NAME                                    READY   STATUS    RESTARTS   AGE
istio-ingressgateway-746d597f7c-g6whv   1/1     Running   0          5d23h
istiocoredns-7ffc9b7fcf-crhr2           2/2     Running   0          5d23h
istiod-7795ccf9dc-vr4cq                 1/1     Running   0          5d22h
```

## Next steps

Now that we have Istio and Service Mesh Hub installed ([and appropriate clusters registered]({{% versioned_link_path fromRoot="/setup/#register-a-cluster" %}})), we can continue to explore the [discovery capabilities]({{% versioned_link_path fromRoot="/guides/discovery_intro" %}}) of Service Mesh Hub. 
