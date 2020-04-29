---
title: Troubleshooting
menuTitle: Troubleshooting
weight: 10
description: Understanding how to troubleshoot Service Mesh Hub with tips on logging, FAQ, and understanding how things work
---

In this guide we explore how to troubleshoot when things don't behave the way you're expecting. We also try to explain how things work under the covers so you can use your own troubleshooting skills to find why things may not behave as expected. We would love to hear from you if you if you get stuck on the [Solo.io Slack](https://slack.solo.io) or if you figure out how to solve something not covered here, please consider a [Pull Request to the docs](https://github.com/solo-io/service-mesh-hub/tree/master/docs) to add it.

## Helpful debugging tips

So you've created some `TrafficPolicy` rules or just grouped your first `VirtualMesh` and it's not behaving as expected. What do you do?

One of the first places to always start is `meshctl check`. This command is your friend. Here's what it does:

* Test [connectivity to the Kubernetes cluster](https://github.com/solo-io/service-mesh-hub/blob/master/cli/pkg/tree/check/healthcheck/internal/kube_connectivity_check.go#L20)
* Checking the [minimum supported Kubernetes minor version](https://github.com/solo-io/service-mesh-hub/blob/master/pkg/version/version.go#L13)
* Checking that the `install` [namespace exists (default, `service-mesh-hub`)](https://github.com/solo-io/service-mesh-hub/blob/master/cli/pkg/tree/check/healthcheck/internal/install_namespace_existence.go#L23)
* Verifying the Service Mesh Hub components [are installed and running](https://github.com/solo-io/service-mesh-hub/blob/master/cli/pkg/tree/check/healthcheck/internal/smh_components_health.go#L36)
* Verify none of the `MeshServices` have [any federation errors](https://github.com/solo-io/service-mesh-hub/blob/master/cli/pkg/tree/check/healthcheck/internal/federation_decision_check.go#L43)

The last bullet in the list, checking federation status, is likely the most helpful especially after you've tried to apply a `VirtualMesh`. Often it's best to check the `VirtualMesh` CR `status` field to make sure it doesn't see any issues:

```shell
kubectl get virtualmesh -n service-mesh-hub -o jsonpath='{.items[].status}'
```

NOTE: all of the Service Mesh Hub CRs have a `status` field that give you an indication of what has happened with its processing. Service Mesh Hub, like most operator-style controllers, implements a state machine and helps resources transition between states. All of the current state transitions are available on this `status` field of any of the Service Mesh Hub CRs.

## Logging

Knowing how to get logs from the Service Mesh Hub components is crucial for getting feedback about what's happening. Service Mesh Hub has three core components:

* `csr-agent` - responsible for creating certificate/key pairs and issuing a Certificate Signing Request for a particular mesh -- this is used in Mesh identity federation; this component is not available in the default install, only when a cluster is registered
* `mesh-discovery` - responsible for discovering Meshes, periodically querying clusters and control planes for their `MeshService`s and `MeshWorkload`s. 
* `mesh-networking` - responsible for orchestrating federation events, traffic policy updates, and access-control policy rules

When troubleshooting various parts of the Service Mesh Hub functionality, you will likely want to see the logs for some of these components. For example, when creating a new `VirtualMesh` and you see something like `PROCESSING_ERROR` in the `federation` status, check the logs for the `mesh-networking` component like this:

```shell
kubectl logs -f deploy/mesh-networking -n service-mesh-hub
```

This will give you valuable insight into the control loop that manages the state of a `VirtualMesh`

##### Debug Logging

You can set the `env` variable `DEBUG_MODE` to "true" in any of the Service Mesh Hub pods to increase the logging level. You can set an explicit logging level using the `LOG_LEVEL` env variable. See [https://pkg.go.dev/go.uber.org/zap/zapcore?tab=doc#Level](https://pkg.go.dev/go.uber.org/zap/zapcore?tab=doc#Level) for the log levels.

## Common issues

When following the guides for multi-cluster and multi-mesh communication, the following things could happen of which you should be aware:

##### Cluster shared identity failed

You can check that the remote cluster is reachable, that it has the `service-mesh-hub` namespace, that it has the `csr-agent` running successfully, and that a `VirtualMeshCertificateSigningRequest` has been created:

```shell
kubectl get virtualmeshcertificatesigningrequest -n service-mesh-hub --context remote-cluster-context
```

If everything looks fine, ie the `status` field is in `ACCEPTED` state, make sure the Istio cacerts were created in the `istio-system-namespace`

```shell
kubectl get secret -n istio-system cacerts --context remote-cluster-context
```

{{% notice warning %}}
One thing to NOT overlook is the fact that Istio's control plane, `istiod` needs to be restarted once the `cacerts` has been created or changed. 

For example:

```shell
kubectl --context remote-cluster-context delete pod -n istio-system -l app=istiod 
```

This is being [improved in future versions of Istio](https://github.com/istio/istio/issues/22993)
{{% /notice %}}

##### Communication across cluster isn't working

For communication in a [replicated control plane](https://istio.io/docs/setup/install/multicluster/gateways/), shared root-trust cluster, to work, Istio needs to be installed correctly with its Ingress Gateway. At the moment, either a LoadBalancer or NodePort can be used for the Ingress Gateway service. Either way, the Host for the NodePort or the IP for the LoadBalancer needs to be reachable between clusters.

For example, on a cloud provider like GKE, we may see something like this:

```shell
kubectl get svc -n istio-system
NAME                   TYPE           CLUSTER-IP    EXTERNAL-IP      PORT(S)                                                                                                                                      AGE
istio-ingressgateway   LoadBalancer   10.8.49.163   35.233.234.111   15020:32347/TCP,80:32414/TCP,443:32713/TCP,15029:31362/TCP,15030:32242/TCP,15031:31899/TCP,15032:32471/TCP,31400:31570/TCP,15443:30416/TCP   4d2h
istio-pilot            ClusterIP      10.8.54.191   <none>           15010/TCP,15011/TCP,15012/TCP,8080/TCP,15014/TCP,443/TCP                                                                                     4d2h
istiod                 ClusterIP      10.8.51.94    <none>           15012/TCP,443/TCP      

```

Note, the external-ip. Any `ServiceEntry` created for cross-cluster service discovery will use this external ip. The logic for getting the `ingressgateway` IP can be found in the [federation code base](https://github.com/solo-io/service-mesh-hub/blob/master/services/mesh-networking/pkg/federation/dns/external_access_point_getter.go#L85) of the `mesh-networking` service.

##### Do cross-cluster service entries resolve DNS?

They can, but they are not automatically configured by Service Mesh Hub *yet*. Services in deployment targets (clusters) that are registered with Service Mesh Hub are created and are routable within the Istio (or any mesh) sidecar proxy, but not directly (ie, `nslookup` will fail). 

However, you can manually set up the DNS yourself. See the [customizing DNS for Istio routing]({{% versioned_link_path fromRoot="/operations/customize_dns" %}}) for more.


This automation, to set up the DNS stubbing, is coming very soon (and this doc might be outdated by then). We will keep the docs as up to date as possible. 

##### What Istio versions are supported?

Right now, Service Mesh Hub supports Istio 1.5.x. More versions can be supported based on user feedback. Additional meshes like Linkerd and AppMesh are also supported, each in various stages of completeness. 

{{% notice warning %}}
We do not yet support automatically upgrading Istio in-place from versions 1.4 and earlier to 1.5, due to a number of
breaking changes across that version change. If you are currently running Istio prior to 1.5, you may have to
fully uninstall the mesh before attempting an installation of 1.5. 

<br/>

Users have reported seeing the following when attempting to upgrade in-place:

<br/>

https://discuss.istio.io/t/istio-upgrade-from-1-4-6-1-5-0-throws-istiod-errors-remote-error-tls-error-decrypting-message/5727

{{% /notice %}}




##### Found something else?

If you've run into something else, please reach out the `@ceposta`,  `@Graham Goudeau` or `@Harvey Xia` on the [Solo.io Slack](https://slack.solo.io) in the #service-mesh-hub channel

If you see an error like `resource version conflict` please chime in on [https://github.com/solo-io/service-mesh-hub/issues/635](https://github.com/solo-io/service-mesh-hub/issues/635)