---
title: "Gloo Mesh Enterprise on Openshift"
menuTitle: Enterprise
description: Get started running Gloo Mesh Enterprise on an OpenShift Cluster.
weight: 30
---

The following guide describes how to get started with Gloo Mesh Enterprise on an [Openshift](https://www.openshift.com/) Cluster,
covering installation and workload management.

<figure>
    <img src="{{% versioned_link_path fromRoot="/img/enterprise-getting-started-diagram.png" %}}"/>
</figure>

## Prerequisites

Before we get started, ensure that you have the following tools installed:

- [oc](https://docs.openshift.com/container-platform/4.7/cli_reference/openshift_cli/getting-started-cli.html) - Command line utility for OpenShift. This contains most of `kubectl`'s funtionality, plus additional OpenShift capabilities.
- [meshctl]({{% versioned_link_path fromRoot="/getting_started" %}}) - Command line utility for Gloo Mesh
- [istioctl](https://istio.io/latest/docs/setup/getting-started/#download) - Command line utility for Istio. This document assumes you are using istioctl v1.8.

Provision three OpenShift clusters with contexts stored in the following environment variables:
- `MGMT_CONTEXT` - Context for the cluster where you'll be running the Gloo Mesh Enterprise management plane.
- `REMOTE_CONTEXT1` - Context for a cluster where you'll be running a service mesh and injected workloads.
- `REMOTE_CONTEXT2` - Context for a second cluster where you'll be running a service mesh and injected workloads.


Lastly, ensure that you have a Gloo Mesh Enterprise license key stored in the environment variable `GLOO_MESH_LICENSE_KEY`.

## Installing Istio

[Istio has its own documentation for Openshift installation](https://istio.io/latest/docs/setup/platform-setup/openshift/#additional-requirements-for-the-application-namespace). 
This guide includes those steps in the context of installing Istio for Gloo Mesh Enterprise.

Gloo Mesh Enterprise will discover and configure Istio workloads running on all registered clusters. Let's begin by installing
Istio on two of your clusters.

These installation profiles are provided for their simplicity, but Gloo Mesh can discover and manage Istio deployments
regardless of their installation options. However, to facilitate multicluster traffic later on in this guide, you should
ensure that each Istio deployment has an externally accessible ingress gateway.

Note: An important takeaway is that Istio installations on OpenShift clusters
need the `profile: openshift` value in their spec to function properly.

To install Istio on cluster 1, run: 

```shell script
cat << EOF | istioctl install -y --context $REMOTE_CONTEXT1 -f -
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: example-istiooperator
  namespace: istio-system
spec:
  profile: openshift # <-- OpenShift-specific
  # Install Gloo Mesh Istio
  hub: gcr.io/istio-enterprise
  tag: 1.8.5
  meshConfig:
    defaultConfig:
      envoyMetricsService:
        address: enterprise-agent.gloo-mesh:9977
      envoyAccessLogService:
        address: enterprise-agent.gloo-mesh:9977
      proxyMetadata:
        ISTIO_META_DNS_CAPTURE: "true"
        GLOO_MESH_CLUSTER_NAME: cluster1
  components:
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
      # needed for annotating istio metrics with cluster name
      multiCluster:
        clusterName: cluster1
EOF
```

And then cluster 2:

```shell script
cat << EOF | istioctl install -y --context $REMOTE_CONTEXT2 -f -
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: example-istiooperator
  namespace: istio-system
spec:
  profile: openshift
  # Install Gloo Mesh Istio
  hub: gcr.io/istio-enterprise
  tag: 1.8.5
  meshConfig:
    defaultConfig:
      envoyMetricsService:
        address: enterprise-agent.gloo-mesh:9977
      envoyAccessLogService:
        address: enterprise-agent.gloo-mesh:9977
      proxyMetadata:
        ISTIO_META_DNS_CAPTURE: "true"
        GLOO_MESH_CLUSTER_NAME: cluster2
  components:
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
      # needed for annotating istio metrics with cluster name
      multiCluster:
        clusterName: cluster2
EOF
```

If installation was successful, you should see the following output after each command:
```
✔ Istio core installed
✔ Istiod installed
✔ Ingress gateways installed
✔ Installation complete
```

## Istio in OpenShift Setup - Service Account Privileges

Istio sidecars make use of a user ID that is normally restricted by OpenShift. To allow enable this ID, we must elevate the permissions
of the `istio-system` and `istio-operator` service accounts. Run the following commands for each set of clusters you're using, 
replacing `<CONTEXT_NAME>` with the current cluster's name. In this example, you will need to run these commands on both `REMOTE_CONTEXT1` and `REMOTE_CONTEXT2`.

```shell script
oc --context <CONTEXT_NAME> adm policy add-scc-to-group anyuid system:serviceaccounts:istio-system
oc --context <CONTEXT_NAME> adm policy add-scc-to-group anyuid system:serviceaccounts:istio-operator
oc --context <CONTEXT_NAME> adm policy add-scc-to-group privileged system:serviceaccounts:default
oc --context <CONTEXT_NAME> adm policy add-scc-to-group anyuid system:serviceaccounts:default
```



### Istio in OpenShift Maintenence - NetworkAttachmentDefinition Upkeep

Before Istio can safely create workloads in a given namespace, that namespace needs a NetworkAttachmentDefinition. 
For each cluster and namespace you intend to have Istio minitor, run the following command, replacing `<CONTEXT_NAME>` and `<NAMESPACE>` with the appropriate values. 

In this example, you will need to run this command on the `default` cluster on both `REMOTE_CONTEXT1` and `REMOTE_CONTEXT2`.

```shell script
cat <<EOF | oc --context <CONTEXT_NAME> -n <NAMESPACE> create -f -
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: istio-cni
EOF
```

Keep in mind that if you expand your system to encompass new namespaces, you'll need to add this CRD to those new namespaces.


## Installing the Gloo Mesh management components


The OpenShift-specific changes for installing Gloo Mesh Enterprise are related to metrics. For that reason, the instructions for installing
 Gloo Mesh Enterprise in Openshift are grouped with other metrics topics [here]({{% versioned_link_path fromRoot="/guides/observability/metrics/#openshift-integration" %}}). Follow those steps to 
install Gloo Mesh Enterprise in OpenShift.


### Verify install
Once you've installed Gloo Mesh, verify that the following components were installed:

```shell
oc get pods -n gloo-mesh
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


## Register clusters

In order to register your remote clusters with the Gloo Mesh management plane via [Relay]({{% versioned_link_path fromRoot="/concepts/relay/" %}}),
you'll need to know the external address of the `enterprise-networking` service. Because the service
is of type LoadBalancer by default, your cloud provider will expose the service outside the cluster. You can determine the public address of the service with the following:

{{< tabs >}}
{{< tab name="IP LoadBalancer address (GKE)" codelang="yaml">}}
ENTERPRISE_NETWORKING_DOMAIN=$(oc get svc -n gloo-mesh enterprise-networking -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
ENTERPRISE_NETWORKING_PORT=$(oc -n gloo-mesh get service enterprise-networking -o jsonpath='{.spec.ports[?(@.name=="grpc")].port}')
ENTERPRISE_NETWORKING_ADDRESS=${ENTERPRISE_NETWORKING_DOMAIN}:${ENTERPRISE_NETWORKING_PORT}
echo $ENTERPRISE_NETWORKING_ADDRESS
{{< /tab >}}
{{< tab name="Hostname LoadBalancer address (EKS)" codelang="shell" >}}
ENTERPRISE_NETWORKING_DOMAIN=$(oc get svc -n gloo-mesh enterprise-networking -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
ENTERPRISE_NETWORKING_PORT=$(oc -n gloo-mesh get service enterprise-networking -o jsonpath='{.spec.ports[?(@.name=="grpc")].port}')
ENTERPRISE_NETWORKING_ADDRESS=${ENTERPRISE_NETWORKING_DOMAIN}:${ENTERPRISE_NETWORKING_PORT}
echo $ENTERPRISE_NETWORKING_ADDRESS
{{< /tab >}}
{{< /tabs >}}

This address will be accessed via secure connection by the `enterprise-agent` component deployed to each registered cluster.
Note that it may take a minute for your cloud provider to add an external address for the `enterprise-networking`
service. Ensure that the `ENTERPRISE_NETWORKING_ADDRESS` output above is fully qualified before proceeding to register
your clusters.

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

{{% notice note %}}
Ensure that the `gloo-mesh` namespace in each remote cluster is not being injected by Istio. This can be done by
labelling the `gloo-mesh` namespace with `istio-injection=disabled`.
{{% /notice %}}

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
oc get mesh -n gloo-mesh
```

```
NAME                           AGE
istiod-istio-system-cluster1   68s
istiod-istio-system-cluster2   28s
```

Cluster registration installs the Gloo Mesh Enterprise agent on the registered cluster to handle Kubernetes I/O on behalf of the
management plane. It also creates a KubernetesCluster resource on the management cluster to represent the cluster and store
relevant data such as the cluster's local domain (e.g. "cluster.local"). To learn more about cluster registration and how
it can be performed via Helm rather than meshctl, review the [enterprise cluster registration guide]({{% versioned_link_path fromRoot="/setup/cluster_registration/enterprise_cluster_registration/" %}}).

## Create a Virtual Mesh

At this point, you have a fully-functioning Gloo Mesh Enterprise environment complete with a demo application. Feel free
to browse the complete set of Gloo Mesh [guides]({{% versioned_link_path fromRoot="/guides" %}}), or follow along here
as we configure Gloo Mesh for a common multicluster use case.

Next, let's bootstrap connectivity between the two distinct Istio service meshes by creating a Virtual Mesh.

```shell
oc apply --context $MGMT_CONTEXT -f - << EOF
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
  federation:
    # federate all Destinations to all external meshes
    selectors:
    - {}
  # Disable global access policy enforcement for demonstration purposes.
  globalAccessPolicy: DISABLED
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
oc get virtualmesh -n gloo-mesh virtual-mesh -oyaml
```

Note that if the Virtual Mesh creation fails with a permissions error, then you likely installed the mesh using the `--include-rbac` option.  In that case, you will need to [configure]({{% versioned_link_path fromRoot="/guides/configure_role_based_api/" %}}) the RBAC facilities properly.

If there are no errors, then after a few moments the Virtual Mesh status will be "Accepted", indicating your meshes are configured for multicluster traffic.

```
...
status:
  meshes:
    istiod-istio-system-cluster1.gloo-mesh.:
      state: ACCEPTED
    istiod-istio-system-cluster2.gloo-mesh.:
      state: ACCEPTED
  observedGeneration: 1
  state: ACCEPTED
```

## Multicluster traffic

With our distinct Istio service meshes unified under a single Virtual Mesh, let's demonstrate how Gloo Mesh can facilitate
multicluster traffic.

### Deploy bookinfo across clusters

To demonstrate how Gloo Mesh configures multicluster traffic, we will deploy the bookinfo application to both cluster 1
and cluster 2. Cluster 1 will run the application with versions 1 and 2 of the reviews service, and cluster 2 will run
version 3. In order to access version 3 from the product page hosted on cluster 1, we will have to route to the 
reviews-v3 workload on cluster 2.

First, ensure that you have properly created a [NetworkAttachmentDefinition]({{% versioned_link_path fromRoot="/getting_started/openshift/#istio-in-openshift-maintenence---networkattachmentdefinition-upkeep" %}}) in the relevant namespaces before creating workloads.

To install bookinfo with reviews-v1 and reviews-v2 on cluster 1, run:

```shell script
# prepare the default namespace for Istio sidecar injection
oc --context $REMOTE_CONTEXT1 label namespace default istio-injection=enabled
# deploy bookinfo application components for all versions less than v3
oc --context $REMOTE_CONTEXT1 apply -f https://raw.githubusercontent.com/istio/istio/1.8.2/samples/bookinfo/platform/kube/bookinfo.yaml -l 'app,version notin (v3)'
# deploy all bookinfo service accounts
oc --context $REMOTE_CONTEXT1 apply -f https://raw.githubusercontent.com/istio/istio/1.8.2/samples/bookinfo/platform/kube/bookinfo.yaml -l 'account'
# configure ingress gateway to access bookinfo
oc --context $REMOTE_CONTEXT1 apply -f https://raw.githubusercontent.com/istio/istio/1.8.2/samples/bookinfo/networking/bookinfo-gateway.yaml
```

Verify that the bookinfo application is online with:

```shell script
oc --context $REMOTE_CONTEXT1 get pods
```

```
NAME                              READY   STATUS    RESTARTS   AGE
details-v1-558b8b4b76-w9qp8       2/2     Running   0          2m33s
productpage-v1-6987489c74-54lvk   2/2     Running   0          2m34s
ratings-v1-7dc98c7588-pgsxv       2/2     Running   0          2m34s
reviews-v1-7f99cc4496-lwtsr       2/2     Running   0          2m34s
reviews-v2-7d79d5bd5d-mpsk2       2/2     Running   0          2m34s
```

To deploy reviews-v3 and the ratings service it depends on to cluster 2, run:

```shell script
# prepare the default namespace for Istio sidecar injection
oc --context $REMOTE_CONTEXT2 label namespace default istio-injection=enabled
# deploy reviews and ratings services
oc --context $REMOTE_CONTEXT2 apply -f https://raw.githubusercontent.com/istio/istio/1.8.2/samples/bookinfo/platform/kube/bookinfo.yaml -l 'service in (reviews)'
# deploy reviews-v3
oc --context $REMOTE_CONTEXT2 apply -f https://raw.githubusercontent.com/istio/istio/1.8.2/samples/bookinfo/platform/kube/bookinfo.yaml -l 'app in (reviews),version in (v3)'
# deploy ratings
oc --context $REMOTE_CONTEXT2 apply -f https://raw.githubusercontent.com/istio/istio/1.8.2/samples/bookinfo/platform/kube/bookinfo.yaml -l 'app in (ratings)'
# deploy reviews and ratings service accounts
oc --context $REMOTE_CONTEXT2 apply -f https://raw.githubusercontent.com/istio/istio/1.8.2/samples/bookinfo/platform/kube/bookinfo.yaml -l 'account in (reviews, ratings)'
```

And verify success with:

```shell script
oc --context $REMOTE_CONTEXT2 get pods
```

```
NAME                          READY   STATUS    RESTARTS   AGE
ratings-v1-7dc98c7588-qbmmh   2/2     Running   0          3m11s
reviews-v3-7dbcdcbc56-w4kbf   2/2     Running   0          3m11s
```

{{% notice note %}}
If your bookinfo deployments are stuck in a pending state with the following error:
```
admission webhook "sidecar-injector.istio.io" denied the request: template:
      inject:1: function "Template_Version_And_Istio_Version_Mismatched_Check_Installation"
      not defined
```
This means that there is a mismatch with istioctl version and IstioOperator during
install. Check that your `istioctl` version matches the version of istio running on the server.
{{% /notice %}}

To access the bookinfo application, first determine the address of the ingress on cluster 1:

{{< tabs >}}
{{< tab name="IP LoadBalancer address (GKE)" codelang="yaml">}}
CLUSTER_1_INGRESS_ADDRESS=$(oc --context $REMOTE_CONTEXT1 get svc -n istio-system istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
echo http://$CLUSTER_1_INGRESS_ADDRESS/productpage
{{< /tab >}}
{{< tab name="Hostname LoadBalancer address (EKS)" codelang="shell" >}}
CLUSTER_1_INGRESS_ADDRESS=$(oc --context $REMOTE_CONTEXT1 get svc -n istio-system istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
echo http://$CLUSTER_1_INGRESS_ADDRESS/productpage
{{< /tab >}}
{{< /tabs >}}

Navigate to `http://$CLUSTER_1_INGRESS_ADDRESS/productpage` with the web browser of your choice. Refresh the page a few times
and you will see the black stars on the "Book Reviews" column of the page appear and disappear. These represent v1 and
v2 of the reviews service.

<figure>
    <img src="{{% versioned_link_path fromRoot="/img/bookinfo/star-rotation.png" %}}"/>
</figure>

### Split traffic across clusters

Since we did not deploy the reviews-v3 service to cluster 1, we must route to the reviews-v3 instance on cluster 2. We
will enable this functionality with a Gloo Mesh TrafficPolicy that will divert 75% of `reviews` traffic to reviews-v3
running on cluster2. To apply the traffic policy, run the following:

```shell script
cat << EOF | oc --context $MGMT_CONTEXT apply -f -
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

<figure>
    <img src="{{% versioned_link_path fromRoot="/img/bookinfo/redstars.png" %}}"/>
</figure>

{{% notice note %}}
If you are seeing connection issues, and running on `EKS`, please read the following:

\
EKS LoadBalancer Health Checks will by default use the first port in the Kubernetes Service. This can potentially be a problem because
istio-ingressgateway will deploy with the `port: 80` as first in the list. See the following for an example:
```yaml
spec:
  clusterIP: 10.100.108.166
  externalTrafficPolicy: Cluster
  ports:
  - name: http2
    nodePort: 31143
    port: 80
    protocol: TCP
    targetPort: 8080
  - name: https
    nodePort: 30131
    port: 443
    protocol: TCP
    targetPort: 8443
  - name: tls
    nodePort: 32287
    port: 15443
    protocol: TCP
    targetPort: 15443
  selector:
    app: istio-ingressgateway
    istio: ingressgateway
```

To fix this simply move the `port: 15443` to the top of the list, like so:
```yaml
spec:
  clusterIP: 10.100.108.166
  externalTrafficPolicy: Cluster
  ports:
  - name: tls
    nodePort: 32287
    port: 15443
    protocol: TCP
    targetPort: 15443
  - name: http2
    nodePort: 31143
    port: 80
    protocol: TCP
    targetPort: 8080
  - name: https
    nodePort: 30131
    port: 443
    protocol: TCP
    targetPort: 8443
  selector:
    app: istio-ingressgateway
    istio: ingressgateway
```
{{% /notice %}}

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

## Cleanup

Now you're all set up with a fully-functioning, multicluster Gloo Mesh Enteprise environment. Skip this section if
you're interested in using this environment to explore the other features Gloo Mesh has to offer.

The Gloo Mesh management plane and registered clusters each require separate cleanup steps. Administrators can deregister
all clusters or individual clusters from the system and uninstall the management plane components.

### Deregister the remote clusters

To deregister a cluster, you must uninstall the `enterprise-agent` running on the remote cluster as well as the
corresponding KubernetesCluster resource on the management cluster. Like cluster registration, meshctl can handle this for you.

To deregister cluster 1, run:

```shell script
meshctl cluster deregister enterprise \
  --mgmt-context $MGMT_CONTEXT \
  --remote-context $REMOTE_CONTEXT1 \
  cluster1
```

```
Finished uninstalling release enterprise-agent
```

At this point, the management cluster has no knowledge of or connection to cluster 1. To delete the CustomResourceDefinitions
installed by meshctl on the registered cluster at registration time, as well as the gloo-mesh namespace run:

```shell script
for crd in $(oc get crd --context $REMOTE_CONTEXT1 | grep mesh.gloo | awk '{print $1}'); do oc --context $REMOTE_CONTEXT1 delete crd $crd; done
oc --context $REMOTE_CONTEXT1 delete namespace gloo-mesh
```

To deregister cluster 2, run:

```shell script
meshctl cluster deregister enterprise \
  --mgmt-context $MGMT_CONTEXT \
  --remote-context $REMOTE_CONTEXT2 \
  cluster2
```

```
Finished uninstalling release enterprise-agent
```

At this point, the management cluster has no knowledge of or connection to cluster 2. To delete the CustomResourceDefinitions
installed by meshctl on the registered cluster at registration time, as well as the gloo-mesh namespace, run:

```shell script
for crd in $(oc get crd --context $REMOTE_CONTEXT2 | grep mesh.gloo | awk '{print $1}'); do oc --context $REMOTE_CONTEXT2 delete crd $crd; done
oc --context $REMOTE_CONTEXT2 delete namespace gloo-mesh
```

### Uninstall the management components

To uninstall the Gloo Mesh management plane components, run:

```shell script
meshctl uninstall --kubecontext $MGMT_CONTEXT
```

```
Uninstalling Helm chart
Finished uninstalling release gloo-mesh
```

To delete the Gloo Mesh CustomResourceDefinitions from the management cluster, as well as the gloo-mesh namespace, run:

```shell script
for crd in $(oc get crd --context $MGMT_CONTEXT | grep mesh.gloo | awk '{print $1}'); do oc --context $MGMT_CONTEXT delete crd $crd; done
oc --context $MGMT_CONTEXT delete namespace gloo-mesh
```

### Uninstall bookinfo

The following commands will uninstall the bookinfo components from clusters 1 and 2.

```shell script
# remove the sidecar injection label from the default namespace
oc --context $REMOTE_CONTEXT1 label namespace default istio-injection-
# remove bookinfo application components for all versions less than v3
oc --context $REMOTE_CONTEXT1 delete -f https://raw.githubusercontent.com/istio/istio/1.8.2/samples/bookinfo/platform/kube/bookinfo.yaml -l 'app,version notin (v3)'
# remove all bookinfo service accounts
oc --context $REMOTE_CONTEXT1 delete -f https://raw.githubusercontent.com/istio/istio/1.8.2/samples/bookinfo/platform/kube/bookinfo.yaml -l 'account'
# remove ingress gateway configuration for accessing bookinfo
oc --context $REMOTE_CONTEXT1 delete -f https://raw.githubusercontent.com/istio/istio/1.8.2/samples/bookinfo/networking/bookinfo-gateway.yaml
```

```shell script
# remove the sidecar injection label from the default namespace
oc --context $REMOTE_CONTEXT2 label namespace default istio-injection-
# remove reviews and ratings services
oc --context $REMOTE_CONTEXT2 delete -f https://raw.githubusercontent.com/istio/istio/1.8.2/samples/bookinfo/platform/kube/bookinfo.yaml -l 'service in (reviews)'
# remove reviews-v3
oc --context $REMOTE_CONTEXT2 delete -f https://raw.githubusercontent.com/istio/istio/1.8.2/samples/bookinfo/platform/kube/bookinfo.yaml -l 'app in (reviews),version in (v3)'
# remove ratings
oc --context $REMOTE_CONTEXT2 delete -f https://raw.githubusercontent.com/istio/istio/1.8.2/samples/bookinfo/platform/kube/bookinfo.yaml -l 'app in (ratings)'
# remove reviews and ratings service accounts
oc --context $REMOTE_CONTEXT2 delete -f https://raw.githubusercontent.com/istio/istio/1.8.2/samples/bookinfo/platform/kube/bookinfo.yaml -l 'account in (reviews, ratings)'
```

### Uninstall Istio

To uninstall Istio completely from clusters 1 and 2, run:

```shell script
istioctl --context $REMOTE_CONTEXT1 x uninstall --purge
oc --context $REMOTE_CONTEXT1 delete namespace istio-system
```

```shell script
istioctl --context $REMOTE_CONTEXT2 x uninstall --purge
oc --context $REMOTE_CONTEXT2 delete namespace istio-system
```

#### Revoke ID Permissions

Once Istio has been removed, we can revoke the Istio namespace ID permissions we allowed earlier:

```shell script
oc --context <CONTEXT_NAME> adm policy remove-scc-from-group anyuid system:serviceaccounts:istio-system
oc --context <CONTEXT_NAME> adm policy remove-scc-from-group anyuid system:serviceaccounts:istio-operator
oc --context <CONTEXT_NAME> adm policy remove-scc-from-group privileged system:serviceaccounts:default
oc --context <CONTEXT_NAME> adm policy remove-scc-from-group anyuid system:serviceaccounts:default
```

#### Remove NetworkAttachmentDefinitions

Once Istio has been removed, we can remove the NetworkAttachmentDefinitions it depended on. Run the following for each
cluster/namespace that you had previously added a NetworkAttachmentDefinition to:

```
oc --context <CONTEXT_NAME> -n <NAMESPACE> delete network-attachment-definition istio-cni
```


## Next Steps

This is just the beginning of what's possible with Gloo Mesh Enterprise. Review the [guides]({{% versioned_link_path fromRoot="/guides" %}})
to explore additional features, or check out the [Setup Guide]({{% versioned_link_path fromRoot="/setup/" %}}) for
advanced installation and cluster registration options.
