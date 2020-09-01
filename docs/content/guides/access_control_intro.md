---
title: Access Control
menuTitle: Access Control
weight: 30
---

In the [previous guide]({{% versioned_link_path fromRoot="/guides/federate_identity/" %}}), we federated multiple meshes and established a [shared root CA for a shared identity]({{% versioned_link_path fromRoot="/guides/federate_identity/#understanding-the-shared-root-process" %}}) domain. Now that we have a logical [VirtualMesh]({{% versioned_link_path fromRoot="/reference/api/virtual_mesh/" %}}), we need a way to establish **access** policies across the multiple meshes, without treating each of them individually. Service Mesh Hub helps by establishing a single, unified API that understands the logical VirtualMesh construct.


## Before you begin
To illustrate these concepts, we will assume that:

* Service Mesh Hub is [installed and running on the `mgmt-cluster`]({{% versioned_link_path fromRoot="/setup/#install-service-mesh-hub" %}})
* Istio is [installed on both `mgmt-cluster` and `remote-cluster`]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}) clusters
* Both `mgmt-cluster` and `remote-cluster` clusters are [registered with Service Mesh Hub]({{% versioned_link_path fromRoot="/guides/#two-registered-clusters" %}})
* The `bookinfo` app is [installed into the two clusters]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployed-on-two-clusters" %}})


{{% notice note %}}
Be sure to review the assumptions and satisfy the pre-requisites from the [Guides]({{% versioned_link_path fromRoot="/guides" %}}) top-level document.
{{% /notice %}}


## Enforcing Access Control


Ensure that your kubeconfig has the correct context set as its `currentContext`:

```shell
MGMT_CONTEXT=your_management_plane_context
REMOTE_CONTEXT=your_remote_context

kubectl config use-context $MGMT_CONTEXT
```

In another shell, start a port-forward to the bookinfo demo:

```shell
kubectl --context $MGMT_CONTEXT -n bookinfo port-forward deployment/productpage-v1 9080:9080
```

In a browser, visit [http://localhost:9080](http://localhost:9080) (potentially selecting "normal user" if this is your first time using the app) and verify that both the book details and the reviews are loading correctly. Depending on which review service is accessed you will see reviews with no stars or black stars. You can refresh the page to see the review source change.

Let's use the Service Mesh Hub [AccessPolicy]({{% versioned_link_path fromRoot="/reference/api/access_policy/" %}}) resource to enforce access control across the logical VirtualMesh. The default behavior is to `deny-all`.

In the previous guide, [we created a VirtualMesh resource]({{% versioned_link_path fromRoot="/guides/federate_identity/#creating-a-virtual-mesh" %}}), but we had access control disabled. Let's take a look at the same VirtualService, with access control enabled:

{{< tabs >}}
{{< tab name="YAML file" codelang="shell">}}
apiVersion: networking.smh.solo.io/v1alpha2
kind: VirtualMesh
metadata:
  name: virtual-mesh
  namespace: service-mesh-hub
spec:
  mtlsConfig:
    autoRestartPods: true
    shared:
      rootCertificateAuthority:
        generated: null
  federation: {}
  globalAccessPolicy: ENABLED
  meshes:
  - name: istiod-istio-system-mgmt-cluster
    namespace: service-mesh-hub
  - name: istiod-istio-system-remote-cluster
    namespace: service-mesh-hub
{{< /tab >}}
{{< tab name="CLI inline" codelang="shell" >}}
kubectl apply --context $MGMT_CONTEXT -f - << EOF
apiVersion: networking.smh.solo.io/v1alpha2
kind: VirtualMesh
metadata:
  name: virtual-mesh
  namespace: service-mesh-hub
spec:
  mtlsConfig:
    autoRestartPods: true
    shared:
      rootCertificateAuthority:
        generated: null
  federation: {}
  globalAccessPolicy: ENABLED
  meshes:
  - name: istiod-istio-system-mgmt-cluster
    namespace: service-mesh-hub
  - name: istiod-istio-system-remote-cluster
    namespace: service-mesh-hub
EOF
{{< /tab >}}
{{< /tabs >}}

If you saved this VirtualMesh CR to a file named `demo-virtual-mesh.yaml`, you can apply it like this:

```shell
kubectl --context $MGMT_CONTEXT apply -f demo-virtual-mesh.yaml

virtualmesh.networking.smh.solo.io/virtual-mesh configured
```

{{% notice note %}}
It may take a few moments for the access policy / authorizations propagate to the workloads. 
{{% /notice %}}

With the `globalAccessPolicy` setting `ENABLED` and with no other `AccessControlPolicies`, we should see a `deny-all` access posture. 

Try going back to [http://localhost:9080](http://localhost:9080) and refresh the bookinfo sample and you should see the `details` and `reviews` services blocked.

For Istio, global access control is enforced using an AuthorizationPolicy with an empty spec, placed in Istio's root namespace (usually `istio-system`). More details can be found in the [Istio AuthorizationPolicy documentation](https://istio.io/latest/docs/reference/config/security/authorization-policy/#AuthorizationPolicy).

Note that Service Mesh Hub will also create additional AuthorizationPolicies in order to allow all traffic through ingress gateways so that federated traffic can continue working as expected.

## Using `AccessPolicy`

To allow traffic to continue as before, we apply the following two updates our configuration. Respectively the additions do the following:

1. Allow the `productpage` service to talk to anything in the bookinfo namespace
2. Allow the `reviews` service to talk ony to the ratings service

In this configuration, we select sources (in this case the `productpage` service account) and allow traffic the any service in the `bookinfo` namespace. Be sure to update the `clusterName` value as needed.

{{< tabs >}}
{{< tab name="YAML file" codelang="shell">}}
apiVersion: networking.smh.solo.io/v1alpha2
kind: AccessPolicy
metadata:
  namespace: service-mesh-hub
  name: productpage
spec:
  sourceSelector:
  - kubeServiceAccountRefs:
      serviceAccounts:
        - name: bookinfo-productpage
          namespace: bookinfo
          clusterName: mgmt-cluster
  destinationSelector:
  - kubeServiceMatcher:
      namespaces:
      - bookinfo
EOF
{{< /tab >}}
{{< tab name="CLI inline" codelang="shell" >}}
kubectl apply --context $MGMT_CONTEXT -f - << EOF
apiVersion: networking.smh.solo.io/v1alpha2
kind: AccessPolicy
metadata:
  namespace: service-mesh-hub
  name: productpage
spec:
  sourceSelector:
  - kubeServiceAccountRefs:
      serviceAccounts:
        - name: bookinfo-productpage
          namespace: bookinfo
          clusterName: mgmt-cluster
  destinationSelector:
  - kubeServiceMatcher:
      namespaces:
      - bookinfo
EOF
{{< /tab >}}
{{< /tabs >}}

If you saved this VirtualMesh CR to a file named `demo-product-policy.yaml`, you can apply it like this:

```shell
kubectl --context $MGMT_CONTEXT apply -f demo-product-policy.yaml

accesspolicy.networking.smh.solo.io/productpage created
```

In this next configuration, we enable traffic from `reviews` to `ratings`:

{{< tabs >}}
{{< tab name="YAML file" codelang="shell">}}
apiVersion: networking.smh.solo.io/v1alpha2
kind: AccessPolicy
metadata:
  namespace: service-mesh-hub
  name: reviews
spec:
  sourceSelector:
  - kubeServiceAccountRefs:
      serviceAccounts:
        - name: bookinfo-reviews
          namespace: bookinfo
          clusterName: mgmt-cluster
  destinationSelector:
  - kubeServiceMatcher:
      namespaces:
      - bookinfo
      labels:
        service: ratings
{{< /tab >}}
{{< tab name="CLI inline" codelang="shell" >}}
kubectl apply --context $MGMT_CONTEXT -f - <<EOF
apiVersion: networking.smh.solo.io/v1alpha2
kind: AccessPolicy
metadata:
  namespace: service-mesh-hub
  name: reviews
spec:
  sourceSelector:
  - kubeServiceAccountRefs:
      serviceAccounts:
        - name: bookinfo-reviews
          namespace: bookinfo
          clusterName: mgmt-cluster
  destinationSelector:
  - kubeServiceMatcher:
      namespaces:
      - bookinfo
      labels:
        service: ratings    
EOF
{{< /tab >}}
{{< /tabs >}}

If you have this YAML saved to a file called `reviews-access.yaml`, you can apply it to take effect:

```yaml
kubectl --context $MGMT_CONTEXT apply -f reviews-access.yaml

accesspolicy.networking.smh.solo.io/reviews created
```

Traffic should be allowed again.

## See it in action

Check out "Part Three" of the ["Dive into Service Mesh Hub" video series](https://www.youtube.com/watch?v=4sWikVELr5M&list=PLBOtlFtGznBjr4E9xYHH9eVyiOwnk1ciK)
(note that the video content reflects Service Mesh Hub <b>v0.6.1</b>):

<iframe width="560" height="315" src="https://www.youtube.com/embed/cG1VCx9G408" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## Next steps

In this section, we enabled access policies for services running in the mesh. In the next section, let's see [how to control traffic routing]({{% versioned_link_path fromRoot="/guides/multicluster_communication/" %}}). 
