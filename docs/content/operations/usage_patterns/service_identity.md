---
title: "Multicluster Service Identity"
menuTitle: Multicluster Service Identity
description: Guidance on using Service Mesh Hub to apply consistent policies to services across clusters.
weight: 10
---

Service Mesh Hub was created to simplify the management of multiple service mesh deployments across multiple clusters. One of the key abstractions of Service Mesh Hub is the VirtualMesh, essentially the combination of multiple service meshes into a single logical entity. The VirtualMesh enables cross-cluster, secured communication with a common root certificate authority serving to provide trusted mTLS communications.

To further manage communication within and between clusters, Service Mesh Hub includes Access Policies and Traffic Policies. The Access Policies define what communication is allowed between sources (identities) and destinations. In addition to source and destination, the Access Policy can also specify paths, methods, and ports for a request. Traffic Policies define how communication between a source and destination is handled, including things like traffic shifting, fault injection, and header manipulation. 

Both the Access and Traffic Policies use `sourceSelector` for workloads/identities and `destinationSelector` for services. The selection syntax provides an opportunity to abstract a workload, identity, or service and apply consistent policies across multiple resources in different meshes and clusters.

In this document, we will walk through an example of using each selector type to implement consistent settings and policies across multiple resources.

## Identity Selector

The Identity Selector is used by Access Policies to identify the source for a communication. The source can be defined using a `KubeIdentityMatcher` or `KubeServiceAccountRef`. The `KubeIdentityMatcher` includes a list of allowed namespaces and a list of allowed clusters. The `KubeServiceAccountRef` refers to a singular service account in a namespace and cluster.

It is possible to use the `KubeIdentityMatcher` with just a list of namespaces and no cluster, which would have the effect of selecting traffic from those namespaces in any cluster. This could be useful when trying to apply an Access Policy to a source application that uses a common namespace across all clusters. The following code snippet illustrates such a configuration:

```yaml
kubeIdentityMatcher:
  namespaces:
    - app1
    - app2
```

In a situation where there are two clusters, `cluster-one` and `cluster-two`, the Access Policy would select traffic from the namespace `app1` and `app2` in either cluster. More importantly, if `cluster-three` were added with the same namespaces, it would also be selected as a source. If we would like to be more restrictive, the clusters in question can be added to the code snippet as well:

```yaml
kubeIdentityMatcher:
  namespaces:
    - app1
    - app2
  clusters:
    - cluster-one
    - cluster-two
```

Now if `cluster-three` is added to Service Mesh Hub, the `app1` and `app2` namespaces in that cluster would not be part of the Access Policy until it was updated.

## Workload Selector

The Workload Selector is used by Traffic Policies to identify the source of traffic to be processed by the policy. The `WorkloadSelector` spec has two fields, `labels` and `namespaces`. Source workloads must have all the labels specified and exist in one of the namespaces. The `labels` field provides a high degree of flexibility, as the workload can exist in any cluster or namespace, and as long as it has the proper labels, the Traffic Policy will be applied.

Take for instance the workload `productpage` in the Bookstore sample application. It accesses the `reviews` service to load book reviews. By applying consistent labels to all instances of the `productpage` workload, it would be simple to define a consistent Traffic Policy. In the sample spec below, we are going to assume that the `productpage` workload has the labels `app=bookstore` and `service=productpage` applied.

```yaml
sourceSelector:
  - labels:
       app: bookstore
       service: productpage
```

Since we have not specified a namespace, the Traffic Policy will apply to all namespaces.

## Service Selector

The Service Selector is used by both Access and Traffic Policies for destination selection. Both the source and destination must match for a policy to apply. Let's review how the Service Selector works and then how it could apply to both Access and Traffic Policies.

The `ServiceSelector` spec defines two selection mechanisms, `KubeServiceMatcher` and `KubeServiceRefs`. The `KubeServiceRefs` is a list of direct references to Kubernetes services; using the cluster, namespace, and service name as required fields for identification. A simple example of the `reviews` service is shown below:

```yaml
kubeServiceRefs:
  - name: reviews
    namespace: bookinfo
    clusterName: cluster-one
```

The `KubeServiceMatcher` selector provides a higher degree of freedom when it comes to picking destinations. There are three fields included in the matcher: `labels`, `namespaces`, and `clusters`. All three fields are optional, which means that it would be simple to select a destination based on the labels applied to the service regardless of what namespace or cluster the service was running in. 

By adding a namespace and/or cluster, the matching becomes more refined. For example, let's say we wanted to match on the `details` service for the Bookstore application, and we know it will have the labels `app=bookstore` and `service=details`.

```yaml
kubeServiceMatcher:
  labels:
    app: bookstore
    service: details
```

The above code snippet would match on the `details` service regardless of namespace or cluster. If we are only looking to apply the policy to destinations in `cluster-one`, we could update the matcher as shown below:

```yaml
kubeServiceMatcher:
  labels:
    app: bookstore
    service: details
  clusters:
    - cluster-one
```

### Access Policy

In an Access Policy, the flexibility of destination selection means that you can apply consistent access rules to services as they are added to your environment. By making effective use of labels, an Access Policy will automatically be applied to new instances of a service regardless of namespace and cluster. By default, all traffic is denied, so by creating a service identity using labels, you know that approved traffic will be allowed.

Sticking with our Bookstore application example. Let's say that we want to allow traffic from the `bookinfo` namespace to all instances of the `details` service, and the only allowed method should be `GET`. The resulting policy would look like this:

```yaml
apiVersion: networking.smh.solo.io/v1alpha2
kind: AccessPolicy
metadata:
  namespace: service-mesh-hub
  name: allow-details
spec:
  sourceSelector:
  - kubeIdentityMatcher:
      namespaces:
        - bookinfo
  destinationSelector:
  - kubeServiceMatcher:
      labels:
        app: bookstore
        service: details
  allowedMethods:
    - GET
```

As long as new instances of the `details` service have the correct labels, the Access Policy will apply to them as well.

### Traffic Policy

In a Traffic Policy, the flexibility of both the workload and destination selection means that you can apply consistent traffic rules to workloads and services as they are added to your environment. By making effective use of labels, a Traffic Policy will automatically be applied to new instances of a workload or service regardless of namespace and cluster.

Let's say that we wanted to mirror traffic between all instances of the `productpage` workload and `reviews` service for analysis. Assuming that labels have been applied consistently, the Traffic Policy would look like this:

```yaml
apiVersion: networking.smh.solo.io/v1alpha2
kind: TrafficPolicy
metadata:
  namespace: service-mesh-hub
  name: traffic-mirror
spec:
  sourceSelector:
    - labels:
        app: bookstore
        service: productpage
  destinationSelector:
  - kubeServiceMatcher:
      labels:
        app: bookstore
        service: reviews
  requestTimeout: 100ms
  mirror:
    kubeService:
      name: http-monitoring
      namespace: monitoring
      clusterName: cluster-one
    port: 80
```

Traffic Policies can also apply settings like fault injection, request timeouts, retries, CORS policy and more. By using consistently applied labels, the proper Traffic Policies can be automatically applied.

## Next Steps

You can dig deeper into these topics by checking out the Access Policies guide, or the Traffic Policy and Failover Service guides.
