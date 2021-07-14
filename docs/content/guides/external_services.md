---
title: External Services
menuTitle: External Services
weight: 30
---

{{% notice note %}} Gloo Mesh Enterprise is required for this feature. {{% /notice %}}

## Before you begin
To illustrate these concepts, we will assume that:

* There is at least 1 cluster managed by Gloo Mesh named (`cluster-1` for example.)
* Istio is [installed on the managed cluster]({{% versioned_link_path fromRoot="/guides/installing_istio" %}})


{{% notice note %}}
Be sure to review the assumptions and satisfy the pre-requisites from the [Guides]({{% versioned_link_path fromRoot="/guides" %}}) top-level document.
{{% /notice %}}

## Declaring External Services

#### Motivation

Sometimes within a mesh, an application needs the ability to communicate with services which are "external" to that mesh. For instance a public API, or other external API providers. Istio exposes this functionality via their [`ServiceEntry`](https://istio.io/latest/docs/reference/config/networking/service-entry/) API. Gloo Mesh exposes the same functionality, in a way that feels natural within the rest of the Gloo Mesh API ecosystem.

#### Destintaion CRD

The [Destination CRD]({{% versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.destination/" %}}) can be thought of as a host. In other words an addressable host which may be backed by 0-n endpoints. Typically these `Destinations` are Discovered by Gloo Mesh, and are backed by kubernetes workloads. However, that is only 1 use case. An external service at it's core is just a hostname backed my 0-n endpoints as well, so in comes the `External Destination`.

The `External Destination` is not a seperate CRD in itself, but rather a way to configure the `Destination` CRD to accomplish our goal of routing to services which exist outside of our mesh.

### Demo

Enough explaining, let's get to a quick example.

To start let's create our external service. Simply apply the following yaml to the cluster:
```yaml
apiVersion: discovery.mesh.gloo.solo.io/v1
kind: Destination
metadata:
  name: my-ext-service-tt
spec:
  externalService:
    endpoints:
    - address: solo.io
      ports:
        http: 80
    hosts:
    - foo.bar.global
    name: my-external-service
    ports:
    - name: http
      number: 80
      protocol: HTTP
```

This is a very simple example, but it should help us understand the value, as well as how to use External Destinations. The core of the External Destination exists in the spec block:
```yaml
  externalService:
    endpoints:
    - address: solo.io
      ports:
        http: 80
    hosts:
    - foo.bar.global
    name: my-external-service
    ports:
    - name: http
      number: 80
      protocol: HTTP
```
For those who have used Istio's `ServiceEntry` API this should look very familiar.

For reference, let's take a look at what `ServiceEntry` Gloo Mesh generates from this resource.
```shell
kubectl get se -n istio-system my-external-service -oyaml
apiVersion: networking.istio.io/v1beta1
kind: ServiceEntry
metadata:
  annotations:
    parents.networking.mesh.gloo.solo.io: '{"networking.mesh.gloo.solo.io/v1, Kind=VirtualMesh":[{"name":"bookinfo-federation","namespace":"default"}]}'
  labels:
    cluster.multicluster.solo.io: ""
    owner.networking.mesh.gloo.solo.io: gloo-mesh
    relay-agent: mgmt-cluster
  name: my-external-service
  namespace: istio-system
spec:
  addresses:
  - 248.6.56.91
  endpoints:
  - address: solo.io
    ports:
      http: 80
  hosts:
  - foo.bar.global
  ports:
  - name: http
    number: 80
    protocol: HTTP
  resolution: DNS
```

As you can see we have translated this directly to the underlying ServiceEntry primitive, but now we can also add policies to this destination using other Gloo Mesh primitives such as the `TrafficPolicy` and/or `AccessPolicy`.

Just to make sure it's working let's quickly curl this host.
```shell
kubectl label namespace default istio-injection=enabled
kubectl run curl --image=radial/busyboxplus:curl -i --tty --rm
```

Once the previous command finishes initializting, it will leave you with a shell running in an injected Istio pod. This pod will also have curl so we can test our service.

Run the following curl from inside of our injected shell:
```shell
 curl foo.bar.global
<!DOCTYPE html>
<html>
...
</body>
</html>
```

If everything worked, the output should be the solo.io homepage! I cut out most of it above for brevity.