---
title: "Getting Started with Gloo Mesh Enterprise"
menuTitle: Enterprise
description: How to get started using Gloo Mesh Enterprise
weight: 10
---

The following guide describes how to get started with Gloo Mesh Enterprise on a managed Kubernetes environment such as GKE or EKS.

## Before we begin

- meshctl
- istioctl
- jq (optional, but useful for rendering and navigating json)
- 3 clusters with contexts MGMT_CONTEXT, REMOTE_CONTEXT1, REMOTE_CONTEXT2, where MGMT_CONTEXT points to the cluster that will
run the Gloo Mesh management plane and REMOTE_CONTEXT1/2 point to clusters that are running Istio and application workloads. If
desired, the management cluster can also run a service mesh and workloads to be discovered and managed by Gloo Mesh.
- license key at GLOO_MESH_LICENSE_KEY

TODO joekelley diagram

## TODO joekelley install istio

Let's install Istio on cluster 1:

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

## Installing the management components

Installing Gloo Mesh Enterprise with `meshctl` is a simple process. You will use the command `meshctl install enterprise` and supply the license key, as well as any chart values you want to update, and arguments pointing to the cluster where Gloo Mesh Enterprise will be installed. For our example, we are going to install Gloo Mesh Enterprise on the cluster `mgmt-cluster`. First, let's set a variable for the license key.

```shell
GLOO_MESH_LICENSE_KEY=<your_key_here> # You'll need to supply your own key
```

If you are running Gloo Mesh Enterprise's management plane on a cluster you intend to register (i.e. also run a service mesh), set the `enterprise-networking.cluster` value to the cluster name you intend to set for the management cluster at registration time.

TODO joekelley change default svc type to load balancer

```shell
meshctl install enterprise --license $GLOO_MESH_LICENSE_KEY
```

TODO joekelley override the default user or disable rbac

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

```shell
NAME                                     READY   STATUS    RESTARTS   AGE
dashboard-6d6b944cdb-jcpvl               3/3     Running   0          4m2s
enterprise-networking-84fc9fd6f5-rrbnq   1/1     Running   0          4m2s
rbac-webhook-84865cb7dd-sbwp7            1/1     Running   0          4m2s
```

Running the check command from meshctl will also verify everything was installed correctly:

```shell
meshctl check
```

```shell
Gloo Mesh
-------------------
✅ Gloo Mesh pods are running

Management Configuration
---------------------------
✅ Gloo Mesh networking configuration resources are in a valid state
```

TODO joekelley link to advanced options / install guide

## Register your remote clusters

In order to register your remote clusters with the Gloo Mesh management plane, you'll need to know the external address
of the `enterprise-networking` service. ** TODO joekelley link to relay architecture description. ** Because the service
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

```shell script
kubectl get mesh -n gloo-mesh
```

```
NAME                           AGE
istiod-istio-system-cluster1   68s
istiod-istio-system-cluster2   28s
```

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

This kicks off a process by which each mesh is configured with certificates that share a common root of trust. To learn
more about Gloo Mesh's Virtual Mesh concept, see TODO joekelley link here.

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
splitting traffic across clusters.

### Deploy a distributed application

To demonstrate how Gloo Mesh configures multi-cluster traffic, we will deploy the bookinfo application to both cluster 1
and cluster 2. However, on cluster 1, we will only deploy versions 1 and 2 of the reviews service. In order to access
version 3 from the product page hosted on cluster 1, we will have to route to the reviews-v3 workload on cluster 2.

To install bookinfo with reviews-v1 and reviews-v2 on cluster 1, run:

```shell script
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

TODO joekelley maybe remove these since the UI will show it

Note that at this stage, Gloo Mesh will have discovered all bookinfo workloads and services (destinations). To review
those discovery artifacts, run:

```shell script
kubectl get workloads -n gloo-mesh
```

```
details-v1-default-cluster1-deployment                  8m1s
details-v1-default-cluster2-deployment                  68s
istio-ingressgateway-istio-system-cluster1-deployment   106m
istio-ingressgateway-istio-system-cluster2-deployment   105m
productpage-v1-default-cluster1-deployment              8m1s
productpage-v1-default-cluster2-deployment              67s
ratings-v1-default-cluster1-deployment                  8m2s
ratings-v1-default-cluster2-deployment                  68s
reviews-v1-default-cluster1-deployment                  8m2s
reviews-v1-default-cluster2-deployment                  67s
reviews-v2-default-cluster1-deployment                  8m2s
reviews-v2-default-cluster2-deployment                  67s
reviews-v3-default-cluster2-deployment                  67s
```

```shell script
kubectl get destination -n gloo-mesh
```

```
NAME                                         AGE
details-default-cluster1                     8m35s
details-default-cluster2                     102s
istio-ingressgateway-istio-system-cluster1   107m
istio-ingressgateway-istio-system-cluster2   106m
productpage-default-cluster1                 8m35s
productpage-default-cluster2                 101s
ratings-default-cluster1                     8m36s
ratings-default-cluster2                     102s
reviews-default-cluster1                     8m36s
reviews-default-cluster2                     101s
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

## Launch dashboard

To access the Gloo Mesh Enterprise dashboard, run:

```shell script
meshctl dashboard
```

From here you can review the clusters registered to Gloo Mesh, the status of the meshes, workloads, and services
running there, as well as all the config we've applied in this guide.

## Clean up
