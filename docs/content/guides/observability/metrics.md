---
title: Metrics
menuTitle: Metrics
description: Guide on Gloo Mesh's metrics features.
weight: 30
---

This guide describes how to get started with Gloo Mesh Enterprise's out of the box metrics suite.

{{% notice note %}} Gloo Mesh Enterprise is required for this feature. {{% /notice %}}

{{% notice note %}} This feature currently only supports Istio meshes. {{% /notice %}}

## Before you begin

This guide assumes the following:

* Istio is [installed on both `cluster-1` and `cluster-2`]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}) clusters
    * Istio is configured according to what's described in "Environment Prerequisites" below
    * `istio-system` is the root namespace for both Istio deployments
* The `bookinfo` app is [installed into the two clusters]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployed-on-two-clusters" %}}) under the `bookinfo` namespace

## Environment Prerequisites

### Istio

Each managed Istio control plane must be installed with the following configuration in the [`IstioOperator` manifest](https://istio.io/latest/docs/reference/config/istio.operator.v1alpha1/).

```yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: example-istiooperator
  namespace: istio-system
spec:
  meshConfig:
    defaultConfig:
      envoyMetricsService:
        address: enterprise-agent.gloo-mesh:9977
      proxyMetadata:
        # needed for annotating Gloo Mesh cluster name on envoy requests (i.e. access logs, metrics)
        GLOO_MESH_CLUSTER_NAME: ${gloo-mesh-registered-cluster-name}
  values:
    global:
      # needed for annotating istio metrics with cluster
      multiCluster:
        clusterName: ${gloo-mesh-registered-cluster-name}
```

The `envoyMetricsService` config ensures that all Envoy proxies are configured to emit their metrics to 
the Enterprise Agent, which acts as an [Envoy metrics service sink](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/metrics/v3/metrics_service.proto#extension-envoy-stat-sinks-metrics-service).
The Enterprise Agents then forward all received metrics to Enterprise Networking, where metrics across all managed clusters are centralized.

The `multiCluster` config enables Istio collected metrics to be annotated with the Gloo Mesh registered cluster name.
This allows for proper attribution of metrics in multicluster environments, and is particularly important for attributing
requests that cross cluster boundaries.

### Gloo Mesh Enterprise

When installing Gloo Mesh Enterprise, the `metricsBackend.prometheus.enabled` Helm value must be set to true. This can be done by providing
the following argument to `Helm install`, `--set metricsBackend.prometheus.enabled=true`.

This configures Gloo Mesh to install a Prometheus server which comes preconfigured to scrape the centralized metrics from the Enterprise Networking
metrics endpoint.

After installation of the Gloo Mesh management plane into `cluster-1`, you should see the following deployments:

```shell
gloo-mesh      enterprise-networking-69d74c9744-8nlkd               1/1     Running   0          23m
gloo-mesh      prometheus-server-68b58c79f8-rlq54                   2/2     Running   0          23m
```

#### OpenShift Integration

If you are installing Gloo Mesh Enterprise on an [OpenShift](https://www.openshift.com/) cluster, you will need some additional helm values to make Prometheus run, as Openshift will require a user ID:

 - `gloo-mesh-ui.GlooMeshDashboard.apiserver.floatingUserId=true`
 - `enterprise-networking.prometheus.server.securityContext.runAsUser=$OPENSHIFT_ID`
 - `enterprise-networking.prometheus.server.securityContext.runAsGroup=$OPENSHIFT_ID`
 - `enterprise-networking.prometheus.server.securityContext.fsGroup=$OPENSHIFT_ID`
 
 Where `$OPENSHIFT_ID` is a single valid ID from the range that OpenShift has assigned to your intended Gloo Mesh Enterprise namespace. The valid ID range can be found by examining your namespace's metadata. Note that this
 requires that your intended installation namespace already exists. If it does not, you must create it first:
 ```shell script
% MESH_NAMESPACE='gloo-mesh' # Replace with your namespace if are installing Gloo Mesh Enterprise elsewhere.
% oc create ns $MESH_NAMESPACE 
```
 Once your namespace is established, check its metadata:
 ```shell 
% oc get ns $MESH_NAMESPACE -ojsonpath='{.metadata.annotations}' 
map[openshift.io/sa.scc.mcs: s0:c27,c9 openshift.io/sa.scc.supplemental-groups: 1000720000/10000 openshift.io/sa.scc.uid-range: 1000720000/10000]
```

OpenShift's range syntax is N through N + M - 1 inclusive, given the format N/M. So in this case, the valid ID range would be 1000720000 through 1000729999. Select a number from this range to be your ID. Assuming the number`1000720000` is a valid option, an example installation command would look like this: 

```shell 
% OPENSHIFT_ID=1000720000
% helm install gloo-mesh-enterprise gloo-mesh-enterprise/gloo-mesh-enterprise --namespace gloo-mesh \
--set licenseKey=${GLOO_MESH_LICENSE_KEY} \
--set enterprise-networking.metricsBackend.prometheus.enabled=true \
--set gloo-mesh-ui.GlooMeshDashboard.apiserver.floatingUserId=true \
--set enterprise-networking.prometheus.server.securityContext.runAsUser=$OPENSHIFT_ID \
--set enterprise-networking.prometheus.server.securityContext.runAsGroup=$OPENSHIFT_ID \
--set enterprise-networking.prometheus.server.securityContext.fsGroup=$OPENSHIFT_ID
```

## Functionality

### Generate Traffic

Before any meaningful metrics are collected, traffic has to be generated in the system.

Port forward the productpage deployment (the productpage workload is
convenient because it makes requests to the other workloads, but any workload of your choice will suffice).

```shell
kubectl -n bookinfo port-forward deploy/productpage-v1 9080
```

Then using a utility like [hey](https://github.com/rakyll/hey), send requests to that destination:

```shell
# send 1 request per second
hey -z 1h -c 1 -q 1 http://localhost:9080/productpage\?u\=normal
```

Note that you may need to wait a few minutes before the metrics are returned from the Gloo Mesh API discussed below.
The metrics need time to propagate from the Enovy proxies to the Gloo Mesh server, and for the Prometheus server to scrape the data from Gloo Mesh.

### Enterprise Networking Metrics Scrape Endpoint

The Enterprise Networking deployment exposes its aggregated metrics through an HTTP endpoint at `/metrics`, formatted as Prometheus metrics.
You can view them with the following commands:

```shell
# port-forward enterprise networking
kubectl -n gloo-mesh port-forward deploy/enterprise-networking 8080

# request metrics
curl localhost:8080/metrics
```

These metrics can be consumed by any custom Prometheus server.

### Prometheus UI

The Prometheus server comes with a builtin UI suitable for basic metrics querying. You can view it with the following commands:

```shell
# port forward prometheus server
kubectl -n gloo-mesh port-forward deploy/prometheus-server 9090
```

Then open `localhost:9090` in your browser of choice. 
Here is a simple promql query to get you started with navigating the collected metrics.
This query fetches the `istio_requests_total` metric (which counts the total number of requests) emitted by the
`productpage-v1.bookinfo.cluster-1` workload's Envoy proxy. You can read more about PromQL in the [official documentation](https://prometheus.io/docs/prometheus/latest/querying/basics/).

```promql
sum(
  increase(
    istio_requests_total{
      gm_workload_ref="productpage-v1.bookinfo.cluster-1",
    }[2m]
  )
) by (
  gm_workload_ref,
  gm_destination_workload_ref,
  response_code,
)
```

### Enterprise Networking Metrics Retrieval Endpoint

Enterprise Networking also exposes endpoints for retrieving metrics through an API
that models the service graph as a set of nodes and edges. A node is a workload in the network
that is capable of sending traffic, receiving traffic, or both. An edge consists of a source and target node,
representing metrics for requests originating at the source node and targeting the target node.

The `/v0/observability/metrics/node` endpoint returns metrics by node, and the `/v0/observability/metrics/edge` returns metrics by edge.

Here are some example queries, assuming that Enterprise Networking is port forwarded to 8080:

This request returns all incoming and outgoing requests for the productpage-v1 deployment, collecting metrics
samples spanning the last 5 minutes with a sample for each minute.

```shell
curl -XPOST --data '{
  "nodeSelectors": [
    {

        "workloadRef": {
          "name": "productpage-v1",
          "namespace": "bookinfo",
          "clusterName": "cluster-1"

      }
    }
  ],
  "window": "300s", "step": "60s"
}' "localhost:8080/v0/observability/metrics/node?pretty"
```

This request returns all requests originating from productpage-v1 and targeting details-v1:

```shell
curl -XPOST --data '{
   "edgeSelectors":[
      {
         "source": {
           "workloadRef":{
              "name":"productpage-v1",
              "namespace":"bookinfo",
              "clusterName":"cluster-1"
           }
         },
         "target": {
           "workloadRef":{
             "name":"details-v1",
             "namespace":"bookinfo",
             "clusterName":"cluster-1"
           }
         }
      }
   ],
   "window":"300s"
}' "localhost:8080/v0/observability/metrics/edge?pretty"
```

For full documentation on the access log retrieval endpoint, see the
[Swagger specification]({{% versioned_link_path fromRoot="/reference/swagger/metrics.swagger.json" %}}).
