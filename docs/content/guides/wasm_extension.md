---
title: Wasm Extension Guide for Gloo Mesh Enterprise
menuTitle: Wasm extension
weight: 110
---

{{% notice note %}}
Gloo Mesh Enterprise is required for this feature.
{{% /notice %}}

With the Gloo Mesh Enterprise CLI, you can initialize, build, and push proprietary Wasm filters. Choose your preferred programming language and we’ll generate all the source code you need to get started implementing custom mesh behavior. To publish your work use the `build` and `push` commands. These will compile your Wasm module and make it available via webassemblyhub.io or the OCI registry of your choice.

To add your new Wasm filter to the mesh, all you need is a `WasmDeployment` Kubernetes custom resource. Specify which Workloads should be configured and with which Wasm filters, then let Gloo Mesh handle the rest. A Gloo Mesh Enterprise extension server will watch for WasmDeployments and manage the lifecycle of all your Wasm filters accordingly.

In this guide we will enable a Wasm filter for use by an Envoy proxy. The filter will add a custom header to the response from the reviews service in the bookinfo application. To do this, we will walk through the following steps:

1. Prepare the Envoy sidecar to fetch Wasm filters
1. Ensure the Enterprise Extender feature is enabled
1. Ensure the Wasm agent is installed
1. Deploy the Wasm filter and validate

## Before you begin
To illustrate these concepts, we will assume that:

* Gloo Mesh Enterprise is [installed and running on the `mgmt-cluster`]({{% versioned_link_path fromRoot="/setup/#install-gloo-mesh" %}})
* Istio **1.8** is [installed on both the `mgmt-cluster` and `remote-cluster`]({{% versioned_link_path fromRoot="/guides/installing_istio" %}})
* Both `mgmt-cluster` and `remote-cluster` clusters are [registered with Gloo Mesh Enterprise]({{% versioned_link_path fromRoot="/guides/#two-registered-clusters" %}})
* The `bookinfo` app is [installed into the two clusters]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployed-on-two-clusters" %}})


{{% notice note %}}
Be sure to review the assumptions and satisfy the pre-requisites from the [Guides]({{% versioned_link_path fromRoot="/guides" %}}) top-level document.
{{% /notice %}}

Set your environment variables like so to reference the management and remote clusters:

```shell
export MGMT_CONTEXT=kind-mgmt-cluster
export REMOTE_CONTEXT=kind-remote-cluster
export ENTERPRISE_EXTENDER_VERSION=0.4.0
```

## Prepare the Envoy sidecar to fetch Wasm filters

Our Envoy instances will fetch their wasm filters from an [envoy cluster](https://www.envoyproxy.io/docs/envoy/latest/api-v2/clusters/clusters) that must be defined in the static bootstrap config. We must therefore perform a one-time operation to add the `wasm-agent` as a cluster in the Envoy bootstrap.

To do so, let's create a ConfigMap containing the custom additions to the Envoy bootstrap:

```bash
cat <<EOF | kubectl apply --context ${REMOTE_CONTEXT} -n bookinfo -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: gloo-mesh-custom-envoy-bootstrap
  namespace: bookinfo
data:
  custom_bootstrap.json: |
    {
      "static_resources": {
        "clusters": [{
          "name": "wasm_agent_cluster",
          "type" : "STRICT_DNS",
          "connect_timeout": "1s",
          "lb_policy": "ROUND_ROBIN",
          "load_assignment": {
            "cluster_name": "wasm_agent_cluster",
            "endpoints": [{
              "lb_endpoints": [{
                "endpoint": {
                  "address":{
                    "socket_address": {
                      "address": "wasm-agent.gloo-mesh.svc.cluster.local",
                      "port_value": 9977
                    }
                  }
                }
              }]
            }]
          },
          "circuit_breakers": {
            "thresholds": [
              {
                "priority": "DEFAULT",
                "max_connections": 100000,
                "max_pending_requests": 100000,
                "max_requests": 100000
              },
              {
                "priority": "HIGH",
                "max_connections": 100000,
                "max_pending_requests": 100000,
                "max_requests": 100000
              }
            ]
          },
          "upstream_connection_options": {
            "tcp_keepalive": {
              "keepalive_time": 300
            }
          },
          "max_requests_per_connection": 1,
          "http2_protocol_options": { }
        }]
      }
    }
EOF
```

Next we'll patch the `reviews-v3` deployment to include this custom boostrap in the sidecar:

```bash
kubectl patch deployment -n bookinfo reviews-v3 --context ${REMOTE_CONTEXT} \
  --patch='{"spec":{"template": {"metadata": {"annotations": {"sidecar.istio.io/bootstrapOverride": "gloo-mesh-custom-envoy-bootstrap"}}}}}' \
  --type=merge
```

Now our deployment is wasm-ready.

##  Ensure the Enterprise Extender feature is enabled

The default installation of Gloo Mesh Enterprise should already have the Enterprise Extender feature included. We can check by running the following:

```shell
kubectl get deployment/enterprise-extender -n gloo-mesh
```

You should see the following:

```shell
NAME                  READY   UP-TO-DATE   AVAILABLE   AGE
enterprise-extender   1/1     1            1           53m
```

If there is no output, you will need to update your installation to include the Enterprise Extender feature. You can add the feature in by creating the following YAML file. Be sure to update the license key value before running the Helm upgrade.

```
# create values.yaml file to configure gloo-mesh to use the enterprise-extender
cat > gloo-mesh-values.yaml << EOF
licenseKey: <your-license-key>

# Set to false to omit installing the Gloo Mesh UI
gloo-mesh-ui:
  enabled: true

# Set to false to omit installing the RBAC webhook
rbac-webhook:
  enabled: true

# Set to false to omit installing the Gloo Mesh Enterprise Extender
gloo-mesh-extender:
  enabled: true

gloo-mesh:
  settings:
    networkingExtensionServers:
      - address: enterprise-extender:9900
        insecure: true
        reconnectOnNetworkFailures: true
EOF

# install upgrade from helm chart
helm upgrade --install gloo-mesh-enterprise gloo-mesh-enterprise/gloo-mesh-enterprise \
  --namespace gloo-mesh --kube-context $MGMT_CONTEXT \
  --values gloo-mesh-values.yaml

```

You can run the following command to verify the deployment was successful.

```shell
kubectl get deployment/enterprise-extender -n gloo-mesh
```

The next step is to install the `wasm-agent` on the remote cluster.

## Ensure the Wasm agent is installed

If you registered the `remote-cluster` after having installed the Enterprise Extender, meshctl should have already installed the Wasm agent on your behalf. Run the following command to verify presence of the Wasm agent.

```shell
kubectl get deployment/wasm-agent -n gloo-mesh --context $REMOTE_CONTEXT
```

You should see the following:

```shell
NAME         READY   UP-TO-DATE   AVAILABLE   AGE
wasm-agent   1/1     1            1           55m
```

If not, we will register the `remote-cluster` to install the wasm-agent. Even if you already registered the cluster, we will re-run the registration command and include the `--install-wasm-agent` flag to add the Wasm agent.

If using `kind` or another docker-based Kubernetes distro, the cluster registration command requires an additional flag `--api-server-address` along with the API server address and port. Use the command on the Kind tab if that is the case.

{{< tabs >}}
{{< tab name="Kubernetes" codelang="shell" >}}
meshctl cluster register \
    --cluster-name remote-cluster \
    --mgmt-context "${MGMT_CONTEXT}" \
    --remote-context "${REMOTE_CONTEXT}" \
    --install-wasm-agent --wasm-agent-chart-file=https://storage.googleapis.com/gloo-mesh-enterprise/wasm-agent/wasm-agent-${ENTERPRISE_EXTENDER_VERSION}.tgz
{{< /tab >}}
{{< tab name="Kind" codelang="shell" >}}
# For macOS
ADDRESS=host.docker.internal

# For Linux
ADDRESS=$(docker exec "remote-cluster-control-plane" ip addr show dev eth0 | sed -nE 's|\s*inet\s+([0-9.]+).*|\1|p')

meshctl cluster register \
    --cluster-name remote-cluster \
    --mgmt-context "${MGMT_CONTEXT}" \
    --remote-context "${REMOTE_CONTEXT}" \
    --api-server-address ${ADDRESS}:6443 \
    --install-wasm-agent --wasm-agent-chart-file=https://storage.googleapis.com/gloo-mesh-enterprise/wasm-agent/wasm-agent-${ENTERPRISE_EXTENDER_VERSION}.tgz
{{< /tab >}}
{{< /tabs >}}

We can validate the agent has been deployed by running the following:

```shell
kubectl get pods -n gloo-mesh --context $REMOTE_CONTEXT

NAME                          READY   STATUS    RESTARTS   AGE
cert-agent-d449599d9-26mz7    1/1     Running   0          38m
wasm-agent-7f56898555-lc5pn   1/1     Running   0          18s
```

## Deploy the Wasm filter and validate

We've got everything in place to use the Wasm filter, but first let's see what things look like without the filter added.

### Test without Wasm Filter

As a sanity check, let's run a `curl` without any wasm filter deployed. First we'll create a temporary container to run curl from in the same namespace as the review service.

```bash
kubectl run -it -n bookinfo --context $REMOTE_CONTEXT curl \
  --image=curlimages/curl:7.73.0 --rm  -- sh

# From the new terminal run the following
curl http://reviews:9080/reviews/1 -v
```

You should see the following response:

```bash
   Trying 10.96.151.245:9080...
* Connected to reviews (10.96.151.245) port 9080 (#0)
> GET /reviews/1 HTTP/1.1
> Host: reviews:9080
> User-Agent: curl/7.73.0-DEV
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< x-powered-by: Servlet/3.1
< content-type: application/json
< date: Thu, 10 Dec 2020 20:54:27 GMT
< content-language: en-US
< content-length: 375
< x-envoy-upstream-service-time: 22
< server: envoy
<
* Connection #0 to host reviews left intact
{"id": "1","reviews": [{  "reviewer": "Reviewer1",  "text": "An extremely entertaining play by Shakespeare. The slapstick humour is refreshing!", "rating": {"stars": 5, "color": "red"}},{  "reviewer": "Reviewer2",  "text": "Absolutely fun and entertaining. The play lacks thematic depth when compared to other plays by Shakespeare.", "rating": {"stars": 4, "color": "red"}}]}
```

Go ahead and exit the pod and it will delete itself. Next we'll try the same after we deploy the Wasm filter.

### Deploy the Filter

Now let's deploy a Wasm filter with a WasmDeployment:

```bash
cat <<EOF | kubectl apply --context ${MGMT_CONTEXT} -f-
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: WasmDeployment
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: remote-reviews-wasm
  namespace: bookinfo
spec:
  filters:
  - filterContext: SIDECAR_INBOUND
    wasmImageSource:
      wasmImageTag: webassemblyhub.io/ilackarms/assemblyscript-test:istio-1.8
  workloadSelector:
  - clusters:
    - remote-cluster
    labels:
      app: reviews
      version: v3
    namespaces:
    - bookinfo
EOF
```

You can verify the filter has been deployed successfully by checking on the new WasmDeployment:

```shell
kubectl get wasmdeployment -n bookinfo remote-reviews-wasm -oyaml
```

At the bottom of the output, you should see the following status:

```yaml
status:
  observedGeneration: 1
  workloadStates:
    reviews-v3-bookinfo-remote-cluster-deployment.gloo-mesh.: FILTERS_DEPLOYED
```

Let's try our curl again:

```bash
kubectl run -it -n bookinfo --context $REMOTE_CONTEXT curl \
  --image=curlimages/curl:7.73.0 --rm  -- sh

# From the new terminal run the following
curl http://reviews:9080/reviews/1 -v
```

Expected response:
{{< highlight shell "hl_lines=16" >}}
   Trying 10.96.151.245:9080...
* Connected to reviews (10.96.151.245) port 9080 (#0)
> GET /reviews/1 HTTP/1.1
> Host: reviews:9080
> User-Agent: curl/7.73.0-DEV
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< x-powered-by: Servlet/3.1
< content-type: application/json
< date: Thu, 10 Dec 2020 20:54:27 GMT
< content-language: en-US
< content-length: 375
< x-envoy-upstream-service-time: 22
< hello: world!
< server: envoy
<
* Connection #0 to host reviews left intact
{"id": "1","reviews": [{  "reviewer": "Reviewer1",  "text": "An extremely entertaining play by Shakespeare. The slapstick humour is refreshing!", "rating": {"stars": 5, "color": "red"}},{  "reviewer": "Reviewer2",  "text": "Absolutely fun and entertaining. The play lacks thematic depth when compared to other plays by Shakespeare.", "rating": {"stars": 4, "color": "red"}}]}
{{< /highlight >}}

We should see the `< hello: world!` header in our response if the filter was deployed successfully.

## Summary and Next Steps

In this guide you used Gloo Mesh Enterprise and the Wasm extension to push a Wasm filter to a service managed by Gloo Mesh.

This is a simple example of a Wasm filter to illustrate the concept. The flexibility of Wasm filters coupled with Envoy provides a platform for incredible innovation. Check out our docs on [Web Assembly Hub](https://docs.solo.io/web-assembly-hub/latest) for more information.
