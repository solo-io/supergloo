---
title: Configuring a Traffic Policy
menuTitle: Traffic Policy
weight: 25
---

Gloo Mesh can manage and configure multiple service meshes across multiple Kubernetes clusters. Gloo Mesh can control the configuration of traffic policies of associated services from a given mesh, including properties like timeouts, retries, CORS, and header manipulation.

In this guide we will examine how Gloo Mesh can configure Istio to apply retry and timeout settings to an existing service. We will be dealing with the same resource types that were introduced in the [Mesh Discovery]({{% versioned_link_path fromRoot="/guides/discovery_intro" %}}) guide.

1. **Kubernetes Clusters**
    - Representation of a cluster that Gloo Mesh is aware of and is authorized to talk to its Kubernetes API server
    - *note*: this resource is created by `meshctl` at cluster registration time
2. **Meshes**
    - Representation of a service mesh control plane that has been discovered 
3. **Workloads**
    - Representation of a pod that is a member of a service mesh; this is often determined by the presence of an injected proxy sidecar
4. **Destinations**
    - Representation of a Kubernetes service that is backed by Workload pods, e.g. pods that are a member of the service mesh


## Before you begin
To illustrate these concepts, we will assume that:

* Gloo Mesh is [installed and running on `cluster-1`]({{% versioned_link_path fromRoot="/setup/#install-gloo-mesh" %}})
* Istio is [installed on the `cluster-1`]({{% versioned_link_path fromRoot="/guides/installing_istio" %}})
* `cluster-1` is also [registered with Gloo Mesh]({{% versioned_link_path fromRoot="/guides/#two-registered-clusters" %}}), making it both the management cluster and a managed cluster.


{{% notice note %}}
Be sure to review the assumptions and satisfy the pre-requisites from the [Guides]({{% versioned_link_path fromRoot="/guides" %}}) top-level document.
{{% /notice %}}

### Create a service

First we are going to deploy the Pet Store application to the management cluster in the `default` namespace. Before deploying the application, we need to label the `default` namespace for Istio sidecar injection.

Ensure that your `kubeconfig` has the management cluster set as its current context:

```shell
MGMT_CONTEXT=your_management_plane_context (in this case cluster-1's context)
kubectl config use-context $MGMT_CONTEXT
```

Label the `default` namespace:

```shell
kubectl label namespace default istio-injection=enabled
```

Now we will deploy the Pet Store application:

```shell
kubectl apply -f https://raw.githubusercontent.com/solo-io/gloo/v1.2.9/example/petstore/petstore.yaml
```

We can verify the deployment and service by checking for resources in the default namespace:

```shell
kubectl get all
```

```shell
NAME                           READY   STATUS    RESTARTS   AGE
pod/petstore-fc84b46dd-c9794   2/2     Running   0          3m

NAME                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/kubernetes   ClusterIP   10.96.0.1       <none>        443/TCP    31m
service/petstore     ClusterIP   10.99.255.226   <none>        8080/TCP   3m

NAME                       READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/petstore   1/1     1            1           3m

NAME                                 DESIRED   CURRENT   READY   AGE
replicaset.apps/petstore-fc84b46dd   1         1         1       3m
```

Because we labeled the `default` namespace for proxy sidecar injection, Istio has added a container to the `petstore` pod. We can verify this by looking at the details of the pod.

```shell
kubectl get pod -l app=petstore -oyaml | grep sidecar
```

```shell
sidecar.istio.io/status: '{"version":"e2cb9d4837cda9584fd272bfa1f348525bcaacfadb7e9b9efbd21a3bb44ad7a1","initContainers":["istio-init"],"containers":["istio-proxy"],"volumes":["istio-envoy","istio-data","istio-podinfo"],"imagePullSecrets":null}'
```

Now we can verify that Gloo Mesh has discovered the Pet Store application and configure a *TrafficPolicy* for it.

### Configure Traffic Policy

With our Pet Store application deployed and wired up to Istio, let's make sure that Gloo Mesh has discovered it by checking for *Workload* and *Destination* resources.

```shell
kubectl get workloads -n gloo-mesh
```

```shell
NAME                                                              AGE
istio-ingressgateway-istio-system-cluster-1-deployment            3h4m
petstore-default-cluster-1-deployment                             3h4m
```

If you've also deployed the Bookstore application, you may see entries for that as well. We can see the naming for the Pet Store application is the deployment name, followed by the namespace, and then the cluster name. We can also check for the *Destination*, which represents the service associated with the pods in the Workload resource.

```shell
kubectl get destination -n gloo-mesh
```

```shell
NAME                                                   AGE
istio-ingressgateway-istio-system-cluster-1            3h7m
petstore-default-cluster-1                             3h7m
```

We are going to create a TrafficPolicy that uses the `petstore-default-cluster-1` as a Destination. Within the TrafficPolicy, we are going to set a retry limit and timeout for the service. You can find more information about the options available for [TrafficPolicy in the API reference section]({{% versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy/" %}}).

Here is the configuration we will apply:

{{< tabs >}}
{{< tab name="YAML file" codelang="yaml">}}
apiVersion: networking.mesh.gloo.solo.io/v1
kind: TrafficPolicy
metadata:
  namespace: gloo-mesh
  name: petstore
spec:
  destinationSelector:
  - kubeServiceRefs:
      services:
      - clusterName: cluster-1
        name: petstore
        namespace: default
  policy:
    requestTimeout: 100ms 
    retries:
      attempts: 5
      perTryTimeout: 5ms
{{< /tab >}}
{{< tab name="CLI inline" codelang="shell" >}}
kubectl apply --context $MGMT_CONTEXT -f - << EOF
apiVersion: networking.mesh.gloo.solo.io/v1
kind: TrafficPolicy
metadata:
  namespace: gloo-mesh
  name: petstore
spec:
  destinationSelector:
  - kubeServiceRefs:
      services:
      - clusterName: cluster-1
        name: petstore
        namespace: default
  policy:
    requestTimeout: 100ms 
    retries:
      attempts: 5
      perTryTimeout: 5ms
EOF
{{< /tab >}}
{{< /tabs >}}

We are using the `destinationSelector` property to specify a single Kubernetes service on the management cluster. We could specify multiple services across several clusters, along with setting up a Virtual Mesh and Multicluster communication. The request timeout is being set to 100ms and there is a maximum of five attempts before an error will be returned.

We can validate that the settings have been applied by checking the status of the Istio VirtualService for the Pet Store application:

```shell
kubectl get virtualservice petstore -oyaml
```

```shell
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"Service","metadata":{"annotations":{},"labels":{"service":"petstore"},"name":"petstore","namespace":"default"},"spec":{"ports":[{"port":8080,"protocol":"TCP"}],"selector":{"app":"petstore"}}}
    parents.networking.mesh.gloo.solo.io: '{"discovery.mesh.gloo.solo.io/v1,
      Kind=Destination":[{"name":"petstore-default-cluster-1","namespace":"gloo-mesh"}]}'
  creationTimestamp: "2020-12-16T20:39:22Z"
  generation: 1
  labels:
    cluster.multicluster.solo.io: cluster-1
    owner.networking.mesh.gloo.solo.io: gloo-mesh
  name: petstore
  namespace: default
  resourceVersion: "39986"
  selfLink: /apis/networking.istio.io/v1beta1/namespaces/default/virtualservices/petstore
  uid: 45fe34e2-0645-4cc7-97a6-0ab88dc9d164
spec:
  hosts:
  - petstore.default.svc.cluster.local
  http:
  - retries:
      attempts: 5
      perTryTimeout: 0.005s
    route:
    - destination:
        host: petstore.default.svc.cluster.local
    timeout: 0.100s
```

As we can see above, the proper retry and timeout settings have been applied to the VirtualService from the Gloo Mesh TrafficPolicy. This feature can be extended to configure many services across multiple service meshes and clusters. Many other features can be configured through the traffic policy as well, including fault injection and traffic mirroring. The [`TrafficPolicySpec`]({{% versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy/" %}}) in our API provides more information on using traffic policies.

## Configuring a Default Route

Consider the following TrafficPolicy that specifies a request matcher (note that in our environment `petstore-test`
does not actually exist, but we reference it here for illustration purposes):

```yaml
apiVersion: networking.mesh.gloo.solo.io/v1
kind: TrafficPolicy
metadata:
  name: petstore-canary
  namespace: gloo-mesh
spec:
  destinationSelector:
  - kubeServiceRefs:
      services:
      - clusterName: cluster-1
        name: petstore
        namespace: default
  httpRequestMatchers:
  - headers:
    - name: version
      value: test
  policy:
    trafficShift:
      destinations:
        - kubeService:
            clusterName: cluster-1
            name: petstore-test
            namespace: default
```

This will generate an Istio VirtualService that routes any requests with the header `version: test` to the `petstore-test` service.
**Any requests that do not have that header will, by default, will return a 404.**

This behavior may be suitable in certain situations, but in other scenarios you may also want
to route all non-matching requests to the original intended destination. This is a common pattern for canary
deployments, where a portion of the overall traffic gets routed to a canary deployment based on certain request
properties (i.e. headers), and all other traffic proceeds as normal to the original intended destination.

To declare a default fallback route for all non-matching requests, you must create an additional
TrafficPolicy with no request matcher and a traffic shift to the selected destination:

```yaml
apiVersion: networking.mesh.gloo.solo.io/v1
kind: TrafficPolicy
metadata:
  name: petstore-default
  namespace: gloo-mesh
spec:
  destinationSelector:
  - kubeServiceRefs:
      services:
      - clusterName: cluster-1
        name: petstore
        namespace: default
  policy:
    trafficShift:
      destinations:
        - kubeService:
            clusterName: cluster-1
            name: petstore
            namespace: default
```

## Next Steps

Now that we have seen a simple example of how Gloo Mesh can be used to configure traffic policies, we can expand that vision across multiple clusters in a [Virtual Mesh]({{% versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.virtual_mesh/" %}}). See the guide on [establishing shared trust domain for multiple meshes]({{% versioned_link_path fromRoot="/guides/federate_identity" %}}).
