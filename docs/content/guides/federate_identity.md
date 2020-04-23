---
title: Identity / Trust Domain
menuTitle: Identity / Trust Domain
weight: 25
---

Service Mesh Hub can help unify the root identity between multiple service mesh installations so any intermediates are signed by the same Root CA and end-to-end mTLS between clusters and services can be established correctly.

Service Mesh Hub will establish trust based on the [trust model](https://spiffe.io/spiffe/concepts/#trust-domain) defined by the user -- is there complete *shared trust* and a common root and identity? Or is there *limited trust* between clusters and traffic is gated by egress and ingress gateways? 

In this guide, we'll explore the *shared trust* model between two Istio clusters and how Service Mesh Hub simplifies and orchestrates the processes needed for this to happen.

## Before you begin
To illustrate these concepts, we will assume that:

* Service Mesh Hub is [installed and running on the `management-plane-context`]({{% versioned_link_path fromRoot="/setup/#install-service-mesh-hub" %}})
* Istio is [installed on both `management-plane-context` and `remote-cluster-context`]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}) clusters
* Both `management-plane-context` and `remote-cluster-context` clusters are [registered with Service Mesh Hub]({{% versioned_link_path fromRoot="/guides/#two-registered-clusters" %}})
* The `bookinfo` app is [installed into two Istio clusters]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployed-on-two-clusters" %}})


{{% notice note %}}
Be sure to review the assumptions and satisfy the pre-requisites from the [Guides]({{% versioned_link_path fromRoot="/guides" %}}) top-level document.
{{% /notice %}}

## Verify identity in two clusters is different

We can see the certificate chain used to establish mTLS between Istio services in `management-plane-context` cluster and `remote-cluster-context` cluster and can compare them to be different. One way to see the certificates, is to use the `openssl s_client` tool with the `-showcerts` param when calling between two services. Let's try it on the `management-plane-cluster`:

```shell
kubectl --context management-plane-context exec -it deploy/reviews-v1 -c istio-proxy \
-- openssl s_client -showcerts -connect ratings.default:9080
```
You should see an output of the certificate chain among other handshake-related information. You can review the last certificate in the chain and that's the root cert:

{{< highlight shell "hl_lines=24-43" >}}
---
Certificate chain
 0 s:
   i:O = cluster.local
-----BEGIN CERTIFICATE-----
MIIDKTCCAhGgAwIBAgIRAKZzWK3r7aZqVd0pXUalzKIwDQYJKoZIhvcNAQELBQAw
GDEWMBQGA1UEChMNY2x1c3Rlci5sb2NhbDAeFw0yMDA0MjMxMjM2MDFaFw0yMDA0
MjQxMjM2MDFaMAAwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCwCZft
3uavGRCv+ooKVWUod7Z3PWukPGIR0icI12+ghFygT9ZKlyu+LQ8iN93A6/soIo8r
ccp5YCyW9O4JCJSPg+iFqGeg9yNDLCATb+6vwTsHx0rvdLdX4803bjF9evWkr5yZ
AlPath6S/Wxihue2xrnw9mSF1nKRQxxw8ypysKiqLfVNBhCnBsN28gppYnl1pIiv
YamBeSiNA887BDnuXIc6t6yNJudlvefuixUhzBeR9zYNlstWBLsdqSubbPdPVxfv
7H6hRjeAmu0VB2oDpsWJ0OYGu42ZavCSHRIL2nD3fqk9DyWtXKIklU4B7rE4mySe
InDLbkj1lLv1FB1VAgMBAAGjgYUwgYIwDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQW
MBQGCCsGAQUFBwMBBggrBgEFBQcDAjAMBgNVHRMBAf8EAjAAMEMGA1UdEQEB/wQ5
MDeGNXNwaWZmZTovL2NsdXN0ZXIubG9jYWwvbnMvZGVmYXVsdC9zYS9ib29raW5m
by1yYXRpbmdzMA0GCSqGSIb3DQEBCwUAA4IBAQBEUi1ge/M3NlQ6xuizY7o/mkFe
+PXKjKT/vf21d6N5FTAJT4fjL6nEsa4NqJC7Txiz9kEjlqLy/SywtB3qYGuG0/+d
QGgWmN1NVOMtl2Kq++LOQOIaEV24mjHb5r38DVk4YyVH2E/1QAWByONDB54Ovlyf
l3qiE3gEeegKsgtsLuhzQCReU5evdmPhnCAMiZvUhQKxHIoCJEx5A+eB4q2zBDN2
H2CNJyWLPulBNCsZvCYXGLDRIy+Sp9AsXhqMTAxvqNS2NaNQ9fh7gSqOORO3kIKz
axoFg6neo+LAaYwoyBtO7/V9OvShd9TbkyPP4amR7k3zkdulFo2o+jKAqzCq
-----END CERTIFICATE-----
 1 s:O = cluster.local
   i:O = cluster.local
-----BEGIN CERTIFICATE-----
MIIC3TCCAcWgAwIBAgIQDZ3lILg70fkKSuuBpj3O7TANBgkqhkiG9w0BAQsFADAY
MRYwFAYDVQQKEw1jbHVzdGVyLmxvY2FsMB4XDTIwMDQyMjIzMjM0NVoXDTMwMDQy
MDIzMjM0NVowGDEWMBQGA1UEChMNY2x1c3Rlci5sb2NhbDCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBALX8anGTKtlpdbIlwGsxTW/ZJeqSM29eei5Lmsee
wll7xaNE4sNaj6HFyqAZomPDJm/4PYZ0fWmJ1FIXFqCXQ6PNf/J592D1x8oIHh50
88BbOkH7wYzEMymoP+2BqXQsY5kxjCg9N6xj4XygSunjXo3ctyVP11GhUew0j+Aw
U7dtZqWlpgMsZsPEn2V4JFid20q+0qz6iCzRh/a3iO98QSfvlpeQuVQkhLiPZOzA
q796C1HLWU7sefkXzVAsQGHA5FqSQLQbOqXBPWaf82Fw9cO4/skBH/qOIVtIh8Ks
rHMgrYkSXprev4bMafUAdfJ9GLity/4D2Mn0rK3k4GiLoL8CAwEAAaMjMCEwDgYD
VR0PAQH/BAQDAgIEMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEB
AGZVlJzyM9E14eucxf1UmJvt5NeGGmgf8YdUd9R693iffG8eAg44VJP6xZ8LTrj7
WoGoWC9SmclvAm2Lo1zh7JQWL2jca5X/aJSW4CZROGDMwWm9e+SaujsOKG3hhis6
iwTl1VsjV4o0UBRm5Z26T/gn1CoIPjQDJRb86RPr/6DHY8jFGvGjceEl+o6mf+gk
Q0xfk7VNxpxScJ/+lU5+IJrqQTBmrhk40eDe24D4zbtnk4YVRRbiMh4p9PIBySyp
gyMylEJ3SgwpVoWwV0e2UvNCG1AlZADiYPpgy2qANzJqtF/GYjfgcpR01r8LceIj
s2rL2u8nTerM5bjlurn1Z58=
-----END CERTIFICATE-----
{{< /highlight >}}

Run the same thing in the `remote-cluster-context` and explore the output and compare. For the `reviews` service running in the `remote-cluster-context` cluster, we have to use `deploy/reviews-v3` as `reviews-v1` which we used in the previous command doesn't exist on that cluster:


```shell
kubectl --context remote-cluster-context exec -it deploy/reviews-v3 -c istio-proxy \
-- openssl s_client -showcerts -connect ratings.default:9080
```

You should notice that the root certificates that signed the workload certificates are **different**. Let's unify those into a *shared trust* model of identity. 

## Creating a Virtual Mesh

Service Mesh Hub uses the [Virtual Mesh]({{% versioned_link_path fromRoot="/reference/api/virtual_mesh/" %}}) Custom Resource to configure a Virtual Mesh, which is a logical grouping of one or multiple service meshes for the purposes of federation according to some parameters. Let's take a look at a VirtualMesh configuration that can help unify our two service meshes and establish a *shared trust* model for identity:

{{< highlight yaml "hl_lines=8-15 17-21" >}}
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
  enforceAccessControl: false
  meshes:
  - name: istio-istio-system-management-plane 
    namespace: service-mesh-hub
  - name: istio-istio-system-new-remote-cluster
    namespace: service-mesh-hub
{{< /highlight >}}


##### Understanding VirtualMesh

In the first highlighted section, we can see the parameters to establishing shared identity and federation. In this case, we tell Service Mesh Hub to create a Root CA using the parameters specified above (ttl, key size, org name, etc). We could have also configured an existing Root CA by providing an existing secret:

```yaml
  certificateAuthority:
    provided:
      certificate:
        name: root-ca-name
        namespace: root-ca-namespace
```

We also specify the federation mode to be `PERMISSIVE`. This means we'll make services available between meshes. You can control this later by specifying different global service properties. 

Lastly, we are creating the VirtualMesh with two different service meshes: `istio-istio-system-management-plane` and `istio-istio-system-new-remote-cluster`. We can have any meshes defined here that should be part of this virtual grouping and federation.

##### Applying VirtualMesh
If you saved this VirtualMesh CR to a file named `demo-virtual-mesh.yaml`, you can apply it like this:

```shell
kubectl --context management-plane-context apply -f demo-virtual-mesh.yaml
```

At this point **we need to bounce the `istiod` control plane**. This is because the Istio control plane picks up the CA for Citadel and does not rotate it often enough. This is being [improved in future versions of Istio](https://github.com/istio/istio/issues/22993). 

```shell
kubectl --context management-plane-context \
delete pod -n istio-system -l app=istiod 
```

```shell
kubectl --context remote-cluster-context \
delete pod -n istio-system -l app=istiod 
```

{{% notice note %}}
Note, after you bounce the control plane, it may still take time for the workload certs to get re-issued with the new CA. You can force the workloads to re-load by bouncing them. For example, for the bookinfo sample running in the `default` namespace:

```shell
kubectl --context management-plane-context delete po --all
kubectl --context remote-cluster-context delete po --all
```
{{% /notice %}}

Creating this resource will instruct Service Mesh to establish a shared root identity across the clusters in the Virtual Mesh as well as federate the services. The next sections of this document help you understand some of the pieces of how this works.

## Understanding the shared root process

When we create the VirtualMesh CR and set the trust model to `shared` and configured the Root CA parameters, Service Mesh Hub will kick off the process to unify the identity to a shared root. First, Service Mesh Hub will either create the Root CA specified (if `builtin` is used). 

Then Service Mesh Hub will use a Certificate Signing Request (CSR) agent on each of the different clusters to create a new key/cert pair that will form an intermediate CA that will be used by the mesh on that cluster. It will then create a Certificate Signing Request, represented by the [VirtualMeshCertificateSigningRequest]() CR. Service Mesh Hub will sign the certificate with the Root CA specified in the VirtualMesh. At that point, we will want the mesh (Istio in this case) to pick up the new intermediate CA and start using that for its workloads.

![Service Mesh Hub Architecture]({{% versioned_link_path fromRoot="/img/smh-csr.png" %}})

To verify, let's check the `VirtualServiceCertificateSigningRequest` CR in `remote-cluster-context`:

```shell
kubectl --context remote-cluster-context \
get virtualmeshcertificatesigningrequest -n service-mesh-hub
```

We should see this on the remote cluster:

```shell
NAME                              AGE
istio-virtual-mesh-cert-request   3m15s
```

If we do the same on the `management-plane-cluster`, we should also see a `VirtualMeshCertificateSigningRequest` there as well.

Lastly, let's verify the correct `cacerts` was created in the `istio-system` namespace that can be used for Istio's Citadel:

```shell
kubectl --context management-plane-context get secret -n istio-system cacerts 
NAME      TYPE                      DATA   AGE
cacerts   solo.io/ca-intermediate   5      8m10s
```

```shell
kubectl --context remote-cluster-context get secret -n istio-system cacerts 
NAME      TYPE                      DATA   AGE
cacerts   solo.io/ca-intermediate   5      8m34s
```

In the previous section, we bounced the Istio control plane to pick up these intermediate certs. Again, this is being [improved in future versions of Istio](https://github.com/istio/istio/issues/22993). 


##### Multi-cluster mesh federation

Once trust has been established, Service Mesh Hub will start federating services so that they are accessible across clusters. Behind the scenes, Service Mesh Hub will handle the networking -- possibly through egress and ingress gateways, and possibly affected by user-defined traffic and access policies -- and ensure requests to the service will resolve and be routed to the right destination. Users can fine-tune which services are federated where by editing the virtual mesh. 

For example, you can see what Istio `ServiceEntry` objects were created. On the `management-plane-context` cluster you can see:

```shell
kubectl --context management-plane-context \
get serviceentry -n istio-system
```

```shell
NAME                                                   HOSTS                                                    LOCATION        RESOLUTION   AGE
istio-ingressgateway.istio-system.new-remote-cluster   [istio-ingressgateway.istio-system.new-remote-cluster]   MESH_INTERNAL   DNS          62m
ratings.default.new-remote-cluster                     [ratings.default.new-remote-cluster]                     MESH_INTERNAL   DNS          62m
reviews.default.new-remote-cluster                     [reviews.default.new-remote-cluster]                     MESH_INTERNAL   DNS          62m
```

On the `remote-cluster-context` cluster, you can see:

```shell
kubectl --context remote-cluster-context \
get serviceentry -n istio-system
```

```shell
NAME                                                 HOSTS                                                  LOCATION        RESOLUTION   AGE
details.default.management-plane                     [details.default.management-plane]                     MESH_INTERNAL   DNS          63m
istio-ingressgateway.istio-system.management-plane   [istio-ingressgateway.istio-system.management-plane]   MESH_INTERNAL   DNS          63m
productpage.default.management-plane                 [productpage.default.management-plane]                 MESH_INTERNAL   DNS          63m
ratings.default.management-plane                     [ratings.default.management-plane]                     MESH_INTERNAL   DNS          63m
reviews.default.management-plane                     [reviews.default.management-plane]                     MESH_INTERNAL   DNS          63m
```

## See it in action

Check out "Part Two" of the ["Dive into Service Mesh Hub" video series](https://www.youtube.com/watch?v=4sWikVELr5M&list=PLBOtlFtGznBjr4E9xYHH9eVyiOwnk1ciK):

<iframe width="560" height="315" src="https://www.youtube.com/embed/djcDaIsqIl8" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>


## Next steps

At this point, you should be able to route traffic across your clusters with end-to-end mTLS. You can verify the certs following the same [approach we did earlier in this section]({{% versioned_link_path fromRoot="/guides/federate_identity/#verify-identity-in-two-clusters-is-different" %}}).

Now that you have a single logical "virtual mesh" you can begin configuring it with an API that is aware of this VirtualMesh concept. In the next sections, you can apply [access control]({{% versioned_link_path fromRoot="/guides/access_control_intro/" %}}) and [traffic policies]({{% versioned_link_path fromRoot="/guides/multicluster_communication/" %}}). 

