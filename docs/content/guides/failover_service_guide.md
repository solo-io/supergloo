---
title: Failover Service
menuTitle: Failover Service
weight: 78
---

Service Mesh Hub provides the ability to configure a *FailoverService*. A FailoverService is a virtual traffic destination that is composed of a list of services ordered in decreasing priority. The composing services are configured with outlier detection, the ability of the system to detect unresponsive services, [read more here](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/outlier).Traffic will automatically be shifted over to services next in the priority order. Currently this feature is only supported for Istio meshes.

In this guide we will first enable outlier detection so the service mesh knows when a failure has ocurred. Then we will create the failover configuration to indicate which services are part of the failover process. Finally, we will test the failover configuration by generating errors on one of the instances of the service.

## Before you begin
To illustrate these concepts, we will assume that:

* Service Mesh Hub is [installed and running on the `mgmt-cluster`]({{% versioned_link_path fromRoot="/setup/#install-service-mesh-hub" %}})
* Istio is [installed on both the `mgmt-cluster` and `remote-cluster`]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}) clusters
* Both the `mgmt-cluster` and `remote-cluster` clusters are [registered with Service Mesh Hub]({{% versioned_link_path fromRoot="/guides/#two-registered-clusters" %}}) under the names `mgmt-cluster` and `remote-cluster` respectively
* The `bookinfo` app is [installed into both clusters]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployed-on-two-clusters" %}})
* You have run through the guides for [Federated Trust and Identity]({{% versioned_link_path fromRoot="/guides/federate_identity/" %}}) and [Access Control]({{% versioned_link_path fromRoot="/guides/access_control_intro/" %}}).


{{% notice note %}}
Be sure to review the assumptions and satisfy the pre-requisites from the [Guides]({{% versioned_link_path fromRoot="/guides" %}}) top-level document.
{{% /notice %}}

## Configure Outlier Detection

The services composing a FailoverService must be configured with outlier detection, which is done through a [TrafficPolicy]({{< versioned_link_path fromRoot="/reference/api/traffic_policy" >}}) custom resource. We are going to apply the following config on the `mgmt-cluster` cluster:

{{< highlight yaml "hl_lines=16-19" >}}
kubectl apply -f - << EOF
apiVersion: networking.smh.solo.io/v1alpha2
kind: TrafficPolicy
metadata:
  namespace: service-mesh-hub
  name: mgmt-reviews-outlier
spec:
  destinationSelector:
  - kubeServiceRefs:
      services:
      - name: reviews
        namespace: bookinfo
        clusterName: mgmt-cluster
      - name: reviews
        namespace: bookinfo
        clusterName: remote-cluster
  outlierDetection:
    consecutiveErrors: 1
EOF
{{< /highlight >}}

For demonstration purposes, we're setting `consecutiveErrors` to 1 to more easily trigger the failover. Once applied, run the following:

```shell
kubectl -n service-mesh-hub get trafficpolicy/mgmt-reviews-outlier -oyaml
```

You should see the following status indicating that the TrafficPolicy is valid and has been translated into mesh-specific config:

```yaml
status:
  observedGeneration: "1"
  state: ACCEPTED
  trafficTargets:
    reviews-bookinfo-mgmt-cluster.service-mesh-hub.:
      acceptanceOrder: 1
      state: ACCEPTED
    reviews-bookinfo-remote-cluster.service-mesh-hub.:
      state: ACCEPTED
```

## Create the FailoverService

Now we will create the FailoverService for the `reviews` service, composed of the reviews service on the `mgmt-cluster` cluster in first priority and on `remote-cluster` in second priority. If the `reviews` service on the `mgmt-cluster` cluster is unhealthy, requests will automatically be shifted over to the service on `remote-cluster`.

 Apply the following config to the `mgmt-cluster` cluster:
 
```yaml
kubectl apply -f - << EOF
apiVersion: networking.smh.solo.io/v1alpha2
kind: FailoverService
metadata:
  name: reviews-failover
  namespace: service-mesh-hub
spec:
  hostname: reviews-failover.bookinfo.global
  port:
    number: 9080
    protocol: http
  meshes:
    - name: istiod-istio-system-mgmt-cluster
      namespace: service-mesh-hub
  backingServices:
  - kubeService:
      name: reviews
      namespace: bookinfo
      clusterName: mgmt-cluster
  - kubeService:
      name: reviews
      namespace: bookinfo
      clusterName: remote-cluster
EOF
```

{{% notice note %}}
The `.global` suffix is needed if you want the failoverService's hostname to be resolvable in the Kubernetes context (e.g. with commands like `curl`), assuming your Istio installation has istiocoredns enabled. This leverages Istio's default DNS setup, which creates DNS configuration for hostnames with `.global` suffix.
{{% /notice %}}

Notice that the services referenced under `failoverServices` matches the services we just configured for outlier detection. This is a requirement for all services composing aFailoverService.

The `meshes` field indicates the control planes that the FailoverService is visible to. If multiple meshes are listed, they must be grouped under a common *VirtualMesh*.

Once applied, run the following:

```shell
kubectl -n service-mesh-hub get failoverservice/reviews-failover -oyaml
```

and you should see the following status:

```yaml
status:
    meshes:
      istiod-istio-system-mgmt-cluster.service-mesh-hub.:
        state: ACCEPTED
    observedGeneration: "1"
    state: ACCEPTED
```

## Demonstrating Failover Functionality

To demonstrate failover functionality, we configure a traffic shift such that all requests targeting the `reviews` service will instead be routed to the reviews FailoverService we created above.

Apply the following TrafficPolicy:

```yaml
kubectl apply -f - << EOF
apiVersion: networking.smh.solo.io/v1alpha2
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
  trafficShift:
    destinations:
    - failoverService:
        name: reviews-failover
        namespace: service-mesh-hub
EOF
```

Port forward the `productpage` pod with the following command and open your web browser to [localhost:9080](http://localhost:9080/productpage?u=normal).

```shell
kubectl -n bookinfo port-forward deployments/productpage-v1 9080
```

Reloading the page a few times should show the "Book Reviews" section showing either no stars (for requests routed to the `reviews-v1` pod) and black stars (for requests routed to the `reviews-v2` pod). This shows that the `productpage` is routing to the first service listed in the reviews-failover FailoverService. Recall from the [multicluster setup guide]({{% versioned_link_path fromRoot="/guides/#two-registered-clusters" %}}) that `reviews-v1` and `reviews-v2` only exist on the `mgmt-cluster` and `reviews-v3` only exists on the `remote-cluster`, which we'll use to distinguish requests routing to either cluster.

Now, to trigger the failover, we'll modify the `reviews-v1` and `reviews-v2` deployment to disable the web servers. 

Run the following commands:

```shell
kubectl -n bookinfo patch deploy reviews-v1 --patch '{"spec": {"template": {"spec": {"containers": [{"name": "reviews","command": ["sleep", "20h"]}]}}}}'
kubectl -n bookinfo patch deploy reviews-v2 --patch '{"spec": {"template": {"spec": {"containers": [{"name": "reviews","command": ["sleep", "20h"]}]}}}}'
```

Once the modified deployment has rolled out, refresh the `productpage` and you should see reviews with red stars, corresponding to `reviews-v3`, which only exists on `remote-cluster`, demonstrating that the requests are indeed failing over to the second service in the list.

To restore the disabled `reviews-v1` and `reviews-v2`, run the following:

```shell
kubectl -n bookinfo patch deployment reviews-v1  --type json   -p '[{"op": "remove", "path": "/spec/template/spec/containers/0/command"}]'
kubectl -n bookinfo patch deployment reviews-v2  --type json   -p '[{"op": "remove", "path": "/spec/template/spec/containers/0/command"}]'
```

Once the deployment has rolled out, reloading the `productpage` should show reviews with no stars or black stars, indicating that the failover service is routing requests to the first listed service in the `mgmt-cluster`.

## Next Steps
In this guide, you successfully configured cross-cluster failover using a VirtualMesh and Traffic Policies. To explore more about Service Mesh Hub, we recommend checking out the [concepts section]({{% versioned_link_path fromRoot="/concepts" %}}) of the docs.
