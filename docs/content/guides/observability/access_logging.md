---
title: Access Logging
menuTitle: Access Logging
weight: 30
---

{{% notice note %}} Gloo Mesh Enterprise is required for this feature. {{% /notice %}}

## Before you begin

This guide assumes the following:

* Gloo Mesh Enterprise is [installed in relay mode and running on the `mgmt-cluster`]({{% versioned_link_path fromRoot="/setup/install-gloo-mesh" %}})
  * `gloo-mesh` is the installation namespace for Gloo Mesh
  * `enterprise-networking` is deployed on the `mgmt-cluster` in the `gloo-mesh` namespace and exposes its gRPC server on port 9900
  * `enterprise-agent` is deployed on both clusters and exposes its gRPC server on port 9977
  * Both `mgmt-cluster` and `remote-cluster` clusters are [registered with Gloo Mesh]({{% versioned_link_path fromRoot="/guides/#two-registered-clusters" %}})
* Istio is [installed on both `mgmt-cluster` and `remote-cluster`]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}) clusters
  * `istio-system` is the root namespace for both Istio deployments
* The `bookinfo` app is [installed into the two clusters]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployed-on-two-clusters" %}}) under the `bookinfo` namespace
* the following environment variables are set:
    ```shell
    MGMT_CONTEXT=your_management_plane_context
    REMOTE_CONTEXT=your_remote_context
    ```

## Istio Configuration

Before we begin, we need to ensure that our Istio deployments in
both `mgmt-cluster` and `remote-cluster` have the necessary configuration for
Gloo Mesh access logging. View the Istio ConfigMap with the following command:

```shell
kubectl --context $MGMT_CONTEXT -n istio-system get configmap istio -oyaml
```

Ensure that the following highlighted entries exist in the ConfigMap:

{{< highlight yaml "hl_lines=5-7" >}}
data:
  mesh:
    defaultConfig:
      envoyAccessLogService:
        address: enterprise-agent.gloo-mesh:9977
      proxyMetadata:
        GLOO_MESH_CLUSTER_NAME: your-gloo-mesh-registered-cluster-name
{{< /highlight >}}

If you updated the ConfigMap, you must restart existing Istio injected workloads in order
for their sidecars to pick up the new config.

The `GLOO_MESH_CLUSTER_NAME` value is used to annotate the Gloo Mesh cluster name when emitting
access logs, which is then used by Gloo Mesh to correlate the envoy proxy to a discovered workload.

## AccessLogRecord CRD

Gloo Mesh uses the [AccessLogRecord]({{% versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.observability.v1alpha1.access_logging/" %}})
custom resource to configure access log collection. The API allows specifying the workloads
for which to enable collection as well as request/response level filter criteria (for only emitting a filtered subset of all access logs).

For demonstration purposes let's create the following object:

{{< tabs >}}
{{< tab name="YAML file" codelang="yaml">}}
apiVersion: observability.enterprise.mesh.gloo.solo.io/v1
kind: AccessLogRecord
metadata:
  name: access-log-all
  namespace: gloo-mesh
spec:
  filters:
    - headerMatcher:
        name: foo
        value: bar
{{< /tab >}}
{{< tab name="CLI inline" codelang="shell" >}}
kubectl apply --context $MGMT_CONTEXT -f - << EOF
apiVersion: observability.enterprise.mesh.gloo.solo.io/v1
kind: AccessLogRecord
metadata:
  name: access-log-all
  namespace: gloo-mesh
spec:
  filters:
  - headerMatcher:
      name: foo
      value: bar
EOF
{{< /tab >}}
{{< /tabs >}}

This will enable access log collection for all workloads in both clusters, but only
for requests containing the header `"foo": "bar"`.

## Retrieving Access Logs

Let's first generate some access logs by making requests in both clusters:

```shell
kubectl --context $MGMT_CONTEXT -n bookinfo exec -it deploy/ratings-v1 -c ratings --  curl -H "foo: bar" -v reviews:9080/reviews/1
```

```shell
kubectl --context $REMOTE_CONTEXT -n bookinfo exec -it deploy/ratings-v1 -c ratings --  curl -H "foo: bar" -v reviews:9080/reviews/1
```

Assuming the access logs were collected successfully, we can now retrieve them, either
through `meshctl` or, since enterprise-networking exposes a REST API, using `curl`.

### curl

Before proceeding, open a port forward from `enterprise-networking`'s HTTP
port to your local machine by running the following:

```shell
# forward port 8080 from enterprise-networking to localhost:8080
kubectl --context $MGMT_CONTEXT -n gloo-mesh port-forward deploy/enterprise-networking 8080
```

The following command will fetch up to 10 of the latest access logs.

```shell
curl -XPOST 'localhost:8080/v0/observability/logs?pretty'
```

The response will look similar to:

```json
{
  "result": {
    "workloadRef": {
      "name": "ratings-v1",
      "namespace": "bookinfo",
      "clusterName": "mgmt-cluster"
    },
    "httpAccessLog": {
      "commonProperties": {
        "downstreamRemoteAddress": {
          "socketAddress": {
            "address": "192.168.2.19",
            "portValue": 58196
          }
        },
        "downstreamLocalAddress": {
          "socketAddress": {
            "address": "10.96.2.228",
            "portValue": 9080
          }
        },
        "startTime": "2021-02-02T17:47:30.301634Z",
        "timeToLastRxByte": "0.000032300s",
        "timeToFirstUpstreamTxByte": "0.004618400s",
        "timeToLastUpstreamTxByte": "0.004637800s",
        "timeToFirstUpstreamRxByte": "0.028778300s",
        "timeToLastUpstreamRxByte": "0.029594900s",
        "timeToFirstDownstreamTxByte": "0.029173900s",
        "timeToLastDownstreamTxByte": "0.029632700s",
        "upstreamRemoteAddress": {
          "socketAddress": {
            "address": "192.168.2.18",
            "portValue": 9080
          }
        },
        "upstreamLocalAddress": {
          "socketAddress": {
            "address": "192.168.2.19",
            "portValue": 52538
          }
        },
        "upstreamCluster": "outbound|9080||reviews.bookinfo.svc.cluster.local",
        "routeName": "default",
        "downstreamDirectRemoteAddress": {
          "socketAddress": {
            "address": "192.168.2.19",
            "portValue": 58196
          }
        }
      },
      "protocolVersion": "HTTP11",
      "request": {
        "requestMethod": "GET",
        "scheme": "https",
        "authority": "reviews:9080",
        "path": "/reviews/1",
        "userAgent": "curl/7.52.1",
        "requestId": "676b631f-b6dd-4a57-b99b-de66b03c2813",
        "requestHeadersBytes": "1207"
      },
      "response": {
        "responseCode": 200,
        "responseHeadersBytes": "174",
        "responseBodyBytes": "295",
        "responseCodeDetails": "via_upstream"
      }
    }
  }
}
{
  "result": {
    "workloadRef": {
      "name": "reviews-v1",
      "namespace": "bookinfo",
      "clusterName": "mgmt-cluster"
    },
    ...
    }
  }
}
{
  "result": {
    "workloadRef": {
      "name": "ratings-v1",
      "namespace": "bookinfo",
      "clusterName": "remote-cluster"
    },
    ...
  }
}
{
  "result": {
    "workloadRef": {
      "name": "reviews-v3",
      "namespace": "bookinfo",
      "clusterName": "remote-cluster"
    },
    ...
  }
}
```

You can also filter the retrieved access logs by workload. The following
request retrieves access logs for any Kubernetes workload with label `app: reviews`, or
`app: productpage`.

```shell
curl -XPOST --data '{
   "workloadSelectors":[
      {
         "kubeWorkloadMatcher":{
            "labels":{
               "app":"reviews"
            }, "clusters": ["mgmt-cluster"]
         }
      },
      {
         "kubeWorkloadMatcher":{
            "labels":{
               "app":"productpage"
            }
         }
      }
   ]
}' "localhost:8080/v0/observability/logs?&pretty"
```

**Streaming Retrieval**

While debugging, it can be helpful to observe the access logs in real time as you manually
make requests. This can be achieved using the same REST endpoint and setting the
query parameter `?watch=1`, which will initiate a streaming connection.

```shell
curl -XPOST 'localhost:8080/v0/observability/logs?watch=1&pretty'
```

In a separate terminal context, perform curl requests and you will see access logs
being streamed back as they are received and processed by Gloo Mesh.

### meshctl accesslogs plugin

The `meshctl accesslogs` plugin can also be used to facilitate access log retrieval. 
Install the plugin and read its usage documentation for more details.

## Debugging

Because access logs provide detailed contextual information at the granularity of 
individual networking requests and responses, they are a valuable tool for debugging.
To showcase this, we will contrive a network error and see how access logs can help
in diagnosing the problem.

First ensure that the Gloo Mesh settings object disables Istio mTLS. This will allow
us to modify mTLS settings for specific Destinations.

{{< highlight yaml "hl_lines=10" >}}
apiVersion: settings.mesh.gloo.solo.io/v1
kind: Settings
metadata:
  name: settings
  namespace: gloo-mesh
spec:
  ...
  mtls:
    istio:
      tlsMode: DISABLE
{{< /highlight >}}

Next, create the following Istio DestinationRule which is intentionally erroroneous,
the referenced TLS secret data does not exist.

{{< tabs >}}
{{< tab name="YAML file" codelang="yaml">}}
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: ratings
  namespace: bookinfo
spec:
  host: ratings.bookinfo.svc.cluster.local
  trafficPolicy:
    tls:
      mode: MUTUAL
      # these files do not exist
      clientCertificate: /etc/certs/myclientcert.pem
      privateKey: /etc/certs/client_private_key.pem
      caCertificates: /etc/certs/rootcacerts.pem
{{< /tab >}}
{{< tab name="CLI inline" codelang="shell" >}}
kubectl apply --context $REMOTE_CONTEXT -f - << EOF
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: ratings
  namespace: bookinfo
spec:
  host: ratings.bookinfo.svc.cluster.local
    trafficPolicy:
      tls:
        mode: MUTUAL
        # these files do not exist
        clientCertificate: /etc/certs/myclientcert.pem
        privateKey: /etc/certs/client_private_key.pem
        caCertificates: /etc/certs/rootcacerts.pem
EOF
{{< /tab >}}
{{< /tabs >}}

Sending a request from the `productpage` pod to the ratings Destination should yield 
the following access log:

{{< highlight json "hl_lines=10" >}}
{
  "result": {
    "workloadRef": {
      "name": "productpage-v1",
      "namespace": "bookinfo",
      "clusterName": "mgmt-cluster"
    },
    "httpAccessLog": {
      ...
        "upstreamTransportFailureReason": "TLS error: Secret is not supplied by SDS",
        "routeName": "default",
        "downstreamDirectRemoteAddress": {
          "socketAddress": {
            "address": "192.168.2.14",
            "portValue": 52836
          }
        }
      },
      ...
    }
  }
}
{{< /highlight >}}

Envoy access logs contain a highly detailed information, the details of which can be found
in the [envoy access log documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage).
