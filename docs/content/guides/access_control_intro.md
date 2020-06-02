---
title: Access Control Intro
menuTitle: Access Control Intro
weight: 30
---

In the [previous guide]({{% versioned_link_path fromRoot="/guides/federate_identity/" %}}), we federated multiple meshes and established a [shared root CA for a shared identity]({{% versioned_link_path fromRoot="/guides/federate_identity/#understanding-the-shared-root-process" %}}) domain. Now that we have a logical [VirtualMesh]({{% versioned_link_path fromRoot="/reference/api/virtual_mesh/" %}}), we need a way to establish **access** policies across the multiple meshes, without treating each of them individually. Service Mesh Hub helps by establishing a single, unified API that understands the logical VirtualMesh construct.


## Before you begin
To illustrate these concepts, we will assume that:

* Service Mesh Hub is [installed and running on the `management-plane-context`]({{% versioned_link_path fromRoot="/setup/#install-service-mesh-hub" %}})
* Istio is [installed on both `management-plane-context` and `remote-cluster-context`]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}) clusters
* Both `management-plane-context` and `remote-cluster-context` clusters are [registered with Service Mesh Hub]({{% versioned_link_path fromRoot="/guides/#two-registered-clusters" %}})
* The `bookinfo` app is [installed into two Istio clusters]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployed-on-two-clusters" %}})


{{% notice note %}}
Be sure to review the assumptions and satisfy the pre-requisites from the [Guides]({{% versioned_link_path fromRoot="/guides" %}}) top-level document.
{{% /notice %}}


## Enforcing Access Control


Ensure that your kubeconfig has the correct context set as its `currentContext`:

```shell
kubectl config use-context management-plane-context
```

In another shell, start a port-forward to the bookinfo demo:

```shell
kubectl -n default port-forward deployment/productpage-v1 9080:9080
```

In a browser, visit [http://localhost:9080](http://localhost:9080) (potentially selecting "normal user" if this is your first time using the app) and verify that both the book details and the reviews are loading correctly. Half of the responses should have no stars, while half of the responses should have black stars.

Let's use the Service Mesh Hub [AccessControlPolicy]({{% versioned_link_path fromRoot="/reference/api/access_control_policy/" %}}) resource to enforce access control across the logical VirtualMesh. The default behavior is to `deny-all`.

In the previous guide, [we created a VirtualMesh resource]({{% versioned_link_path fromRoot="/guides/federate_identity/#creating-a-virtual-mesh" %}}), but we had access control disabled. Let's take a look at the same VirtualService, with access control enabled:

{{< highlight yaml "hl_lines=16" >}}
apiVersion: networking.zephyr.solo.io/v1alpha1
kind: VirtualMesh
metadata:
  name: virtual-mesh
  namespace: service-mesh-hub
spec:
  displayName: "Demo Mesh Federation"
  certificateAuthority:
    builtin:
      ttlDays: 356
      rsaKeySizeBytes: 4096
      orgName: "service-mesh-hub"
  federation: 
    mode: PERMISSIVE
  shared: {}
  enforceAccessControl: ENABLED
  meshes:
  - name: istio-istio-system-management-plane 
    namespace: service-mesh-hub
  - name: istio-istio-system-new-remote-cluster
    namespace: service-mesh-hub
{{< /highlight >}}


If you saved this VirtualMesh CR to a file named `demo-virtual-mesh.yaml`, you can apply it like this:

```shell
kubectl --context management-plane-context apply -f demo-virtual-mesh.yaml

virtualmesh.networking.zephyr.solo.io/virtual-mesh configured
```

With the `enforceAccessControl` setting `enabled` and with no other `AccessControlPolicies`, we should see a `deny-all` access posture. Try going back to [http://localhost:9080](http://localhost:9080) and refresh the bookinfo sample and you should see the `details` and `reviews` services blocked.

{{% notice note %}}
It may take a few moments for the access policy / authorizations propagate to the workloads. 
{{% /notice %}}


## Using `AccessControlPolicy`

To allow traffic to continue as before, we apply the following two pieces of config. Respectively these
configs:

1. Allow the `productpage` service to talk to anything in the default namespace
2. Allow the `reviews` service to talk ony to the ratings service


{{< tabs >}}
{{< tab name="YAML file" codelang="shell">}}
apiVersion: networking.zephyr.solo.io/v1alpha1
kind: AccessControlPolicy
metadata:
  namespace: service-mesh-hub
  name: productpage
spec:
  sourceSelector:
    serviceAccountRefs:
      serviceAccounts:
        - name: bookinfo-productpage
          namespace: default
          cluster: management-plane
  destinationSelector:
    matcher:
      namespaces:
        - default       
EOF
{{< /tab >}}
{{< /tabs >}}

If you have this YAML saved to a file called `product-details-access.yaml`, you can apply it to take effect:

```yaml
kubectl --context management-plane-context \
 apply -f product-details-access.yaml
```


In this configuration, we select sources (in this case the `productpage` service account) and allow traffic the any service in the `default` namespace.

In this next configuration, we enable traffic from `reviews` to `ratings`:

{{< tabs >}}
{{< tab name="YAML file" codelang="shell">}}
apiVersion: networking.zephyr.solo.io/v1alpha1
kind: AccessControlPolicy
metadata:
  namespace: service-mesh-hub
  name: reviews
spec:
  sourceSelector:
    serviceAccountRefs:
      serviceAccounts:
        - name: bookinfo-reviews
          namespace: default
          cluster: remote-cluster
  destinationSelector:
    matcher:
      namespaces:
        - default
{{< /tab >}}
{{< tab name="CLI inline" codelang="shell" >}}
kubectl apply --context management-plane-context -f - << EOF
apiVersion: networking.zephyr.solo.io/v1alpha1
kind: AccessControlPolicy
metadata:
  namespace: service-mesh-hub
  name: productpage
spec:
  sourceSelector:
    serviceAccountRefs:
      serviceAccounts:
        - name: bookinfo-productpage
          namespace: default
          cluster: management-plane
  destinationSelector:
    matcher:
      namespaces:
        - default       
EOF
{{< /tab >}}
{{< /tabs >}}

If you have this YAML saved to a file called `product-details-access.yaml`, you can apply it to take effect:

```yaml
kubectl --context management-plane-context \
 apply -f product-details-access.yaml
```


In this configuration, we select sources (in this case the `productpage` service account) and allow traffic the any service in the `default` namespace.

In this next configuration, we enable traffic from `reviews` to `ratings`:

{{< tabs >}}
{{< tab name="YAML file" codelang="shell">}}
apiVersion: networking.zephyr.solo.io/v1alpha1
kind: AccessControlPolicy
metadata:
  namespace: service-mesh-hub
  name: reviews
spec:
  sourceSelector:
    serviceAccountRefs:
      serviceAccounts:
        - name: bookinfo-reviews
          namespace: default
          cluster: management-plane
  destinationSelector:
    matcher:
      namespaces:
        - default
      labels:
        service: ratings
{{< /tab >}}
{{< tab name="CLI inline" codelang="shell" >}}
kubectl apply --context management-plane-context -f - <<EOF
apiVersion: networking.zephyr.solo.io/v1alpha1
kind: AccessControlPolicy
metadata:
  namespace: service-mesh-hub
  name: reviews
spec:
  sourceSelector:
    serviceAccountRefs:
      serviceAccounts:
        - name: bookinfo-reviews
          namespace: default
          cluster: management-plane
  destinationSelector:
    matcher:
      namespaces:
        - default
      labels:
        service: ratings    
EOF
{{< /tab >}}
{{< /tabs >}}

If you have this YAML saved to a file called `reviews-access.yaml`, you can apply it to take effect:

```yaml
kubectl --context management-plane-context \
 apply -f reviews-access.yaml
```

Traffic should be allowed again.

## See it in action

Check out "Part Three" of the ["Dive into Service Mesh Hub" video series](https://www.youtube.com/watch?v=4sWikVELr5M&list=PLBOtlFtGznBjr4E9xYHH9eVyiOwnk1ciK):

<iframe width="560" height="315" src="https://www.youtube.com/embed/cG1VCx9G408" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## Next steps

In this section, we enabled access policies for services running in the mesh. In the next section, let's see [how to control traffic routing]({{% versioned_link_path fromRoot="/guides/multicluster_communication/" %}}). 
