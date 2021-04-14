---
title: "Getting Started with Gloo Mesh Enterprise"
menuTitle: Enterprise
description: How to get started using Gloo Mesh Enterprise
weight: 10
---

The following guide describes how to get started with Gloo Mesh Enterprise on a managed Kubernetes environment such as GKE or EKS,
covering installation, cluster registration, and multi-cluster traffic.

<figure>
    <img src="{{% versioned_link_path fromRoot="/img/enterprise-getting-started-diagram.png" %}}"/>
</figure>

## Prerequisites

Before we get started, ensure that you have the following tools installed:

- kubectl - Command line utility for Kubernetes
- meshctl - Command line utility for Gloo Mesh 
- istioctl - Command line utility for Istio. This document assumes you are using istioctl v1.8 or v1.9.
- jq (optional) - Utility for manipulating JSON

Three Kubernetes clusters, with contexts stored in the following environment variables:
- MGMT_CONTEXT - Context for the cluster where you'll be running the Gloo Mesh Enterprise management plane.
- REMOTE_CONTEXT1 - Context for the cluster where you'll be running a service mesh and injected workloads.
- REMOTE_CONTEXT2 - Context for a second cluster where you'll be running a service mesh and injected workloads.


Lastly, ensure that you have a Gloo Mesh Enterprise license key stored in the environment variable `GLOO_MESH_LICENSE_KEY`.

## Installing Istio

Gloo Mesh Enterprise will discover and configure Istio workloads running on all registered clusters. Let's begin by installing
Istio on two of your clusters.

These installation profiles are provided for their simplicity, but Gloo Mesh can discover and manage Istio deployments
regardless of their installation options. However, to facilitate multi-cluster traffic later on in this guide, you should
ensure that each Istio deployment has an externally accessible ingress gateway.

To install Istio on cluster 1, run: 

```shell script
cat << EOF | istioctl manifest install -y --context $REMOTE_CONTEXT1 -f -
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
          type: LoadBalancer
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
  values:
    global:
      pilotCertProvider: istiod
EOF
```

And then cluster 2:

```shell script
cat << EOF | istioctl manifest install -y --context $REMOTE_CONTEXT2 -f -
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
          type: LoadBalancer
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
  values:
    global:
      pilotCertProvider: istiod
EOF
```

If your installs were successful, you should see the following output after each command:
```
✔ Istio core installed
✔ Istiod installed
✔ Ingress gateways installed
✔ Installation complete
```

## Installing the management components

The Gloo Mesh management plane is where all service mesh configuration will be provided and all discovery artifacts,
including Meshes, Workloads, and Destinations will be aggregated. If you wish you may also run a service mesh and
application workloads on the management cluster, but we will not for the purposes of this guide. To learn more about
installation options for Gloo Mesh Enterprise, including how to deploy Gloo Mesh via helm, review the [Gloo Mesh Enterprise install guide]({{% versioned_link_path fromRoot="/setup/installation/enterprise_installation/" %}}).

To get you up and running as quickly as possible, we will not install the Gloo Mesh Enterprise component responsible
for enforcing the [role-based API]({{% versioned_link_path fromRoot="/concepts/role_based_api/" %}}) by including the
`--skip-rbac` flag. If you wish to enable it, simply invoke the install command without this flag.


```shell
meshctl install enterprise --license $GLOO_MESH_LICENSE_KEY --skip-rbac
```

You should see the following output from the command:

```shell
Installing Helm chart
Finished installing chart 'gloo-mesh-enterprise' as release gloo-mesh:gloo-mesh
```

The installer has created the namespace `gloo-mesh` and installed Gloo Mesh Enterprise into the namespace using a Helm chart with default values.

### Verify install
Once you've installed Gloo Mesh, verify that the following components were installed:

```shell
kubectl get pods -n gloo-mesh
```

```
NAME                                     READY   STATUS    RESTARTS   AGE
dashboard-6d6b944cdb-jcpvl               3/3     Running   0          4m2s
enterprise-networking-84fc9fd6f5-rrbnq   1/1     Running   0          4m2s
```

Running the check command from meshctl will also verify everything was installed correctly:

```shell
meshctl check
```

```
Gloo Mesh
-------------------
✅ Gloo Mesh pods are running

Management Configuration
---------------------------
✅ Gloo Mesh networking configuration resources are in a valid state
```

## Register your remote clusters

In order to register your remote clusters with the Gloo Mesh management plane via [Relay]({{% versioned_link_path fromRoot="/concepts/relay/" %}}),
you'll need to know the external address of the `enterprise-networking` service. Because the service
is of type LoadBalancer by default, your cloud provider will expose the service outside the cluster. You can determine
the public address of the service with the following:

```shell
ENTERPRISE_NETWORKING_IP=$(kubectl get svc -n gloo-mesh enterprise-networking -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
ENTERPRISE_NETWORKING_PORT=$(kubectl -n gloo-mesh get service enterprise-networking -o jsonpath='{.spec.ports[?(@.name=="grpc")].port}')
ENTERPRISE_NETWORKING_ADDRESS=${ENTERPRISE_NETWORKING_IP}:${ENTERPRISE_NETWORKING_PORT}
```

To register cluster 1, run:

```shell
meshctl cluster register enterprise \
  --mgmt-context=$MGMT_CONTEXT \
  --remote-context=$REMOTE_CONTEXT1 \
  --relay-server-address $ENTERPRISE_NETWORKING_ADDRESS \
  cluster1
```

And for cluster 2, run:

```shell
meshctl cluster register enterprise \
  --mgmt-context=$MGMT_CONTEXT \
  --remote-context=$REMOTE_CONTEXT2 \
  --relay-server-address $ENTERPRISE_NETWORKING_ADDRESS \
  cluster2
```

{{% notice note %}}Ensure that the `gloo-mesh` namespace in each remote cluster is not being injected by Istio.{{% /notice %}}

For each cluster, you should see the following:

```shell
Registering cluster
📃 Copying root CA relay-root-tls-secret.gloo-mesh to remote cluster from management cluster
📃 Copying bootstrap token relay-identity-token-secret.gloo-mesh to remote cluster from management cluster
💻 Installing relay agent in the remote cluster
Finished installing chart 'enterprise-agent' as release gloo-mesh:enterprise-agent
📃 Creating remote-cluster KubernetesCluster CRD in management cluster
⌚ Waiting for relay agent to have a client certificate
         Checking...
         Checking...
🗑 Removing bootstrap token
✅ Done registering cluster!
```

Because we already installed Istio on each registered cluster, you can run the following to verify that Gloo Mesh
has successfully discovered the remote meshes.

```shell script
kubectl get mesh -n gloo-mesh
```

```
NAME                           AGE
istiod-istio-system-cluster1   68s
istiod-istio-system-cluster2   28s
```

To learn more about cluster registration and how it can be performed via Helm rather than meshctl, review the
[enterprise cluster registration guide]({{% versioned_link_path fromRoot="/setup/cluster_registration/enterprise_cluster_registration/" %}}).

## Create a virtual mesh

Next, let's bootstrap connectivity between the two distinct Istio service meshes by creating a Virtual Mesh.

```shell
kubectl apply --context $MGMT_CONTEXT -f - << EOF
apiVersion: networking.mesh.gloo.solo.io/v1
kind: VirtualMesh
metadata:
  name: virtual-mesh
  namespace: gloo-mesh
spec:
  mtlsConfig:
    autoRestartPods: true
    shared:
      rootCertificateAuthority:
        generated: {}
  federation: {}
  meshes:
  - name: istiod-istio-system-cluster1
    namespace: gloo-mesh
  - name: istiod-istio-system-cluster2
    namespace: gloo-mesh
EOF
```

This kicks off a process by which each mesh is configured with certificates that share a common root of trust. Learn more
about Virtual Meshes in the [concepts documentation]({{% versioned_link_path fromRoot="/concepts/concepts/" %}})

To verify that the Virtual Mesh has taken effect, run the following:

```shell script
kubectl get virtualmesh -n gloo-mesh virtual-mesh -o jsonpath='{.status}' | jq
```

The Virtual Mesh status will be "Accepted" when your meshes are configured for multicluster traffic.

```
{
  "meshes": {
    "istiod-istio-system-cluster1.gloo-mesh.": {
      "state": "ACCEPTED"
    },
    "istiod-istio-system-cluster2.gloo-mesh.": {
      "state": "ACCEPTED"
    }
  },
  "observedGeneration": 1,
  "state": "ACCEPTED"
}
```

## Multi-cluster traffic

With our distinct Istio service meshes unified under a single Virtual Mesh, let's demonstrate how Gloo Mesh can facilitate
multi-cluster traffic.

### Deploy a distributed application

To demonstrate how Gloo Mesh configures multi-cluster traffic, we will deploy the bookinfo application to both cluster 1
and cluster 2. However, on cluster 1, we will only deploy versions 1 and 2 of the reviews service. In order to access
version 3 from the product page hosted on cluster 1, we will have to route to the reviews-v3 workload on cluster 2.

To install bookinfo with reviews-v1 and reviews-v2 on cluster 1, run:

```shell script
# prepare the default namespace for Istio sidecar injection
kubectl --context $REMOTE_CONTEXT1 label namespace default istio-injection=enabled
# deploy bookinfo application components for all versions less than v3
kubectl --context $REMOTE_CONTEXT1 apply -f https://raw.githubusercontent.com/istio/istio/1.8.2/samples/bookinfo/platform/kube/bookinfo.yaml -l 'app,version notin (v3)'
# deploy all bookinfo service accounts
kubectl --context $REMOTE_CONTEXT1 apply -f https://raw.githubusercontent.com/istio/istio/1.8.2/samples/bookinfo/platform/kube/bookinfo.yaml -l 'account'
# configure ingress gateway to access bookinfo
kubectl --context $REMOTE_CONTEXT1 apply -f https://raw.githubusercontent.com/istio/istio/1.8.2/samples/bookinfo/networking/bookinfo-gateway.yaml
```

Verify that the bookinfo application is online with:

```shell script
kubectl --context $REMOTE_CONTEXT1 get pods
```

```
NAME                              READY   STATUS    RESTARTS   AGE
details-v1-558b8b4b76-w9qp8       2/2     Running   0          2m33s
productpage-v1-6987489c74-54lvk   2/2     Running   0          2m34s
ratings-v1-7dc98c7588-pgsxv       2/2     Running   0          2m34s
reviews-v1-7f99cc4496-lwtsr       2/2     Running   0          2m34s
reviews-v2-7d79d5bd5d-mpsk2       2/2     Running   0          2m34s
```

To install bookinfo with reviews-v1, reviews-v2, and reviews-v3 on cluster 2, run:

```shell script
# prepare the default namespace for Istio sidecar injection
kubectl --context $REMOTE_CONTEXT2 label namespace default istio-injection=enabled
# deploy all bookinfo service accounts and application components for all versions
kubectl --context $REMOTE_CONTEXT2 apply -f https://raw.githubusercontent.com/istio/istio/1.8.2/samples/bookinfo/platform/kube/bookinfo.yaml
# configure ingress gateway to access bookinfo
kubectl --context $REMOTE_CONTEXT2 apply -f https://raw.githubusercontent.com/istio/istio/1.8.2/samples/bookinfo/networking/bookinfo-gateway.yaml
```

And verify success with:

```shell script
kubectl --context $REMOTE_CONTEXT2 get pods
```

```
NAME                              READY   STATUS    RESTARTS   AGE
details-v1-558b8b4b76-gs9z2       2/2     Running   0          2m22s
productpage-v1-6987489c74-x45vd   2/2     Running   0          2m21s
ratings-v1-7dc98c7588-2n6bg       2/2     Running   0          2m21s
reviews-v1-7f99cc4496-4r48m       2/2     Running   0          2m21s
reviews-v2-7d79d5bd5d-cx9lp       2/2     Running   0          2m22s
reviews-v3-7dbcdcbc56-trjdx       2/2     Running   0          2m22s
```

To access the bookinfo application, first determine the address of the ingress on cluster 1:

```shell script
CLUSTER_1_INGRESS_ADDRESS=$(kubectl --context $REMOTE_CONTEXT1 get svc -n istio-system istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
echo $CLUSTER_1_INGRESS_ADDRESS
```

Navigate to `$CLUSTER_1_INGRESS_ADDRESS/productpage` with the web browser of your choice. Refresh the page a few times
and you will see the black stars on the "Book Reviews" column of the page appear and disappear. These represent v1 and
v2 of the reviews service.

### Split traffic across clusters

Since we did not deploy the reviews-v3 service to cluster 1, we must route to the reviews-v3 instance on cluster 2. We
will enable this functionality with a Gloo Mesh TrafficPolicy that will divert 75% of `reviews` traffic to reviews-v3
running on cluster2. To apply the traffic policy, run the following:

```shell script
cat << EOF | kubectl --context $MGMT_CONTEXT apply -f -
apiVersion: networking.mesh.gloo.solo.io/v1
kind: TrafficPolicy
metadata:
  namespace: gloo-mesh
  name: simple
spec:
  sourceSelector:
  - kubeWorkloadMatcher:
      namespaces:
      - default
  destinationSelector:
  - kubeServiceRefs:
      services:
        - clusterName: cluster1
          name: reviews
          namespace: default
  policy:
    trafficShift:
      destinations:
        - kubeService:
            clusterName: cluster2
            name: reviews
            namespace: default
            subset:
              version: v3
          weight: 75
        - kubeService:
            clusterName: cluster1
            name: reviews
            namespace: default
            subset:
              version: v1
          weight: 15
        - kubeService:
            clusterName: cluster1
            name: reviews
            namespace: default
            subset:
              version: v2
          weight: 10
EOF
```

Return to the bookinfo app in your web browser, refresh the page a few times, and you will see reviews-v3's red stars
appear alongside the book reviews.

## Launch the Gloo Mesh Enterprise dashboard

Gloo Mesh Enterprise ships with a dashboard which provides a single pane of glass through which you can observe the status
of the meshes, workloads, and services running across all your clusters, as well as all the policies configuring the
behavior of your network. To learn more about the dashboard, view [the guide]({{% versioned_link_path fromRoot="/guides/accessing_enterprise_ui/" %}}).

<figure>
    <img src="{{% versioned_link_path fromRoot="/img/dashboard.png" %}}"/>
</figure>

Access the Gloo Mesh Enterprise dashboard with:

```shell script
meshctl dashboard
```

## Next Steps

This is just the beginning of what's possible with Gloo Mesh Enterprise. Review the [guides]({{% versioned_link_path fromRoot="/guides" %}}) to explore additional features.