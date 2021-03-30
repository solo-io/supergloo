---
title: Routing Based On Locality
menuTitle: Routing Based On Locality
weight: 50
---

Providing high-availability of applications across clusters, zones, and regions can be a significant challenge. Source traffic should be routed to the closest available destination, or be routed to a failover destination if issues occur. In this guide, you will use a *VirtualDestination* to accomplish locality-based failover.

Gloo Mesh provides the ability to configure a *VirtualDestination*, which is a virtual traffic destination composed of a list of 1-n services selected.  The composing services are configured with outlier detection, the ability of the system to detect unresponsive services, [read more here](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/outlier). Traffic will automatically be sorted into priority levels by proximity to the orginiating service, and failover when priorities become unhealthy.

## Before you begin

To illustrate these concepts, we will assume that:

* Gloo Mesh is [installed and running on the `mgmt-cluster`]({{% versioned_link_path fromRoot="/setup/#install-gloo-mesh" %}})
* Istio is [installed on both the `mgmt-cluster` and `remote-cluster`]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}) clusters
* Both the `mgmt-cluster` and `remote-cluster` clusters are [registered with Gloo Mesh]({{% versioned_link_path fromRoot="/guides/#two-registered-clusters" %}}) under the names `mgmt-cluster` and `remote-cluster` respectively
* The `bookinfo` app is [installed into both clusters]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployed-on-two-clusters" %}})
* You have run through the guides for [Federated Trust and Identity]({{% versioned_link_path fromRoot="/guides/federate_identity/" %}}) and [Access Control]({{% versioned_link_path fromRoot="/guides/access_control_intro/" %}}).

{{% notice note %}}
Be sure to review the assumptions and satisfy the pre-requisites from the [Guides]({{% versioned_link_path fromRoot="/guides" %}}) top-level document.
{{% /notice %}}

Ensure you have the correct context names set in your environment:

```shell
MGMT_CONTEXT=your_management_plane_context
REMOTE_CONTEXT=your_remote_context
```

## Configure the Region and Zone for the Nodes

Gloo Mesh uses the configured [region and zone labels](https://v1-18.docs.kubernetes.io/docs/reference/kubernetes-api/labels-annotations-taints/#topologykubernetesioregion) on nodes to indicate locality for services. If you do not already have the region and zone labels set, you will need to do so now. In our example, we will set the `mgmt-cluster` node to use `us-east-1` for the region and `us-east-1a` for the zone. The `remote-cluster` node will be set to `us-east-2` and `us-east-2b` respectively. In a cloud-based deployment, these labels will typically be set by the cloud provider.

```bash
kubectl label node $MGMT_CLUSTER-control-plane --context kind-$MGMT_CLUSTER \
  topology.kubernetes.io/region=us-east-1 topology.kubernetes.io/zone=us-east-1a

kubectl label node $REMOTE_CLUSTER-control-plane --context kind-$REMOTE_CLUSTER \
  topology.kubernetes.io/region=us-east-2 topology.kubernetes.io/zone=us-east-2b
```

## Create the VirtualDestination

Now we will create the VirtualDestination for the `reviews` service, composed of the reviews service on the `mgmt-cluster` cluster and on the `remote-cluster`. If the `reviews` service on the local (`mgmt-cluster`) cluster is unhealthy, requests will automatically be routed to the reviews service on the remote cluster.

Apply the following config to the `mgmt-cluster` cluster:
 
{{< tabs >}}
{{< tab name="YAML file" codelang="yaml">}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualDestination
metadata:
  name: bookinfo-global
  namespace: gloo-mesh
spec:
  hostname: reviews.global
  port:
    number: 9080
    protocol: http
  localized:
    outlierDetection:
      consecutiveErrors: 1
      maxEjectionPercent: 100
      interval: 5s
      baseEjectionTime: 120s
    destinationSelectors:
    - kubeServiceMatcher:
        labels:
          app: reviews
  virtualMesh:
    name: virtual-mesh
    namespace: gloo-mesh
{{< /tab >}}
{{< tab name="CLI inline" codelang="shell" >}}
kubectl apply -f - << EOF
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualDestination
metadata:
  name: bookinfo-global
  namespace: gloo-mesh
spec:
  hostname: reviews.global
  port:
    number: 9080
    protocol: http
  localized:
    outlierDetection:
      consecutiveErrors: 1
      maxEjectionPercent: 100
      interval: 5s
      baseEjectionTime: 120s
    destinationSelectors:
    - kubeServiceMatcher:
        labels:
          app: reviews
  virtualMesh:
    name: virtual-mesh
    namespace: gloo-mesh
EOF
{{< /tab >}}
{{< /tabs >}}

{{% notice note %}}
For demonstration purposes, we're setting `consecutiveErrors` to 1 and `maxEjectionPercent` to 100 to more easily trigger the failover. However, these should most likely not be used in production scenarios.
{{% /notice %}}

The `virtualMesh` field indicates the control planes that the VirtualDestination is visible to. It will be visible to all meshes in the `VirtualMesh`. Alternatively, a list of meshes can be supplied here instead. 

Once applied, run the following:

```shell
kubectl -n gloo-mesh get virtualdestination/bookinfo-global -oyaml
```

and you should see the following status:

```yaml
status:
  observedGeneration: "1"
  state: ACCEPTED
```

## Demonstrating Locality Routing Functionality

To demonstrate locality routing functionality, we configure a traffic shift such that all requests targeting the `reviews` service will instead be routed to the reviews VirtualDestination we created above.

Apply the following TrafficPolicy:

{{< tabs >}}
{{< tab name="YAML file" codelang="yaml">}}
apiVersion: networking.mesh.gloo.solo.io/v1
kind: TrafficPolicy
metadata:
  name: reviews-shift-failover
  namespace: bookinfo
spec:
  destinationSelector:
  - kubeServiceRefs:
      services:
      - clusterName: mgmt-cluster
        name: reviews
        namespace: bookinfo
  policy:
    trafficShift:
      destinations:
      - virtualDestination:
          name: bookinfo-global
          namespace: gloo-mesh
{{< /tab >}}
{{< tab name="CLI inline" codelang="shell" >}}
kubectl apply -f - << EOF
apiVersion: networking.mesh.gloo.solo.io/v1
kind: TrafficPolicy
metadata:
  name: reviews-shift-failover
  namespace: bookinfo
spec:
  destinationSelector:
  - kubeServiceRefs:
      services:
      - clusterName: mgmt-cluster
        name: reviews
        namespace: bookinfo
  policy:
    trafficShift:
      destinations:
      - virtualDestination:
          name: bookinfo-global
          namespace: gloo-mesh
EOF
{{< /tab >}}
{{< /tabs >}}

Now we can test the TrafficShift by accessing the reviews service from the bookinfo's product page. Port forward the `productpage` pod with the following command and open your web browser to [localhost:9080](http://localhost:9080/productpage?u=normal).

```shell
kubectl -n bookinfo port-forward deployments/productpage-v1 9080
```

Reloading the page a few times should show the "Book Reviews" section with either no stars (for requests routed to the `reviews-v1` pod) or black stars (for requests routed to the `reviews-v2` pod). This shows that the `productpage` is routing to the local service. This is the desired behavior. The product page requests are coming from the local cluster and being routed to a local destination.

Recall from the [multicluster setup guide]({{% versioned_link_path fromRoot="/guides/#two-registered-clusters" %}}) that `reviews-v1` and `reviews-v2` only exist on the `mgmt-cluster` and `reviews-v3` only exists on the `remote-cluster`, which we'll use to distinguish requests routing to either cluster.

Now, to trigger the failover, we'll modify the `reviews-v1` and `reviews-v2` deployments to disable the web servers. 

Run the following commands:

```shell
kubectl -n bookinfo patch deploy reviews-v1 --patch '{"spec": {"template": {"spec": {"containers": [{"name": "reviews","command": ["sleep", "20h"]}]}}}}'
kubectl -n bookinfo patch deploy reviews-v2 --patch '{"spec": {"template": {"spec": {"containers": [{"name": "reviews","command": ["sleep", "20h"]}]}}}}'
```

Once the modified deployment has rolled out, refresh the `productpage` and you should see reviews with red stars, corresponding to `reviews-v3`, which only exists on `remote-cluster`, demonstrating that the requests are indeed failing locally, and so instead they are being routed to the remote instance.

To restore the disabled `reviews-v1` and `reviews-v2`, run the following:

```shell
kubectl -n bookinfo patch deployment reviews-v1  --type json   -p '[{"op": "remove", "path": "/spec/template/spec/containers/0/command"}]'
kubectl -n bookinfo patch deployment reviews-v2  --type json   -p '[{"op": "remove", "path": "/spec/template/spec/containers/0/command"}]'
```

Once the deployment has rolled out, reloading the `productpage` should show reviews with no stars or black stars, indicating that our localized virtual destination is routing requests to the local service in the `mgmt-cluster`.

## Next Steps

In this guide, you successfully configured cross-cluster failover using a VirtualMesh and Traffic Policies. To explore more about Gloo Mesh, we recommend checking out the [concepts section]({{% versioned_link_path fromRoot="/concepts" %}}) of the docs.
