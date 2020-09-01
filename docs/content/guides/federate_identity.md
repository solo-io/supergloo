---
title: Federated Trust and Identity
menuTitle: Federated Trust and Identity
weight: 25
---

Service Mesh Hub can help unify the root identity between multiple service mesh installations so any intermediates are signed by the same Root CA and end-to-end mTLS between clusters and services can be established correctly.

Service Mesh Hub will establish trust based on the [trust model](https://spiffe.io/spiffe/concepts/#trust-domain) defined by the user -- is there complete *shared trust* and a common root and identity? Or is there *limited trust* between clusters and traffic is gated by egress and ingress gateways? 

In this guide, we'll explore the *shared trust* model between two Istio clusters and how Service Mesh Hub simplifies and orchestrates the processes needed for this to happen.

## Before you begin
To illustrate these concepts, we will assume that:

* Service Mesh Hub is [installed and running on the `mgmt-cluster`]({{% versioned_link_path fromRoot="/setup/#install-service-mesh-hub" %}})
* Istio is [installed on both `mgmt-cluster` and `remote-cluster`]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}) clusters
* Both `mgmt-cluster` and `remote-cluster` clusters are [registered with Service Mesh Hub]({{% versioned_link_path fromRoot="/guides/#two-registered-clusters" %}})
* The `bookinfo` app is [installed into both clusters]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployed-on-two-clusters" %}})


{{% notice note %}}
Be sure to review the assumptions and satisfy the pre-requisites from the [Guides]({{% versioned_link_path fromRoot="/guides" %}}) top-level document.
{{% /notice %}}

## Verify identity in two clusters is different

We can see the certificate chain used to establish mTLS between Istio services in `mgmt-cluster` cluster and `remote-cluster` cluster and can compare them to be different. One way to see the certificates, is to use the `openssl s_client` tool with the `-showcerts` param when calling between two services. Let's try it on the `mgmt-cluster-cluster`:

```shell
MGMT_CONTEXT=your_management_plane_context
REMOTE_CONTEXT=your_remote_context

kubectl --context $MGMT_CONTEXT -n bookinfo exec -it deploy/reviews-v1 -c istio-proxy \
-- openssl s_client -showcerts -connect ratings.bookinfo:9080
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

Run the same thing in the `remote-cluster` and explore the output and compare. For the `reviews` service running in the `remote-cluster` cluster, we have to use `deploy/reviews-v3` as `reviews-v1` which we used in the previous command doesn't exist on that cluster:


```shell
kubectl --context $REMOTE_CONTEXT -n bookinfo exec -it deploy/reviews-v3 -c istio-proxy \
-- openssl s_client -showcerts -connect ratings.bookinfo:9080
```

You should notice that the root certificates that signed the workload certificates are **different**. Let's unify those into a *shared trust* model of identity. 

## Creating a Virtual Mesh

Service Mesh Hub uses the [Virtual Mesh]({{% versioned_link_path fromRoot="/reference/api/virtual_mesh/" %}}) Custom Resource to configure a Virtual Mesh, which is a logical grouping of one or multiple service meshes for the purposes of federation according to some parameters. Let's take a look at a VirtualMesh configuration that can help unify our two service meshes and establish a *shared trust* model for identity:

{{< highlight yaml "hl_lines=8-15 17-21" >}}
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
  meshes:
  - name: istiod-istio-system-mgmt-cluster 
    namespace: service-mesh-hub
  - name: istiod-istio-system-remote-cluster
    namespace: service-mesh-hub
{{< /highlight >}}


##### Understanding VirtualMesh

In the first highlighted section, we can see the parameters establishing shared identity and federation. In this case, we tell Service Mesh Hub to create a Root CA using the parameters specified above (ttl, key size, org name, etc).

We could have also configured an existing Root CA by providing an existing secret:

```yaml
  mtlsConfig:
    autoRestartPods: true
    shared:
      rootCertificateAuthority:
        secret:
          name: root-ca-name
          namespace: root-ca-namespace
```

See the section on [User Provided Certificates]({{% versioned_link_path fromRoot="/guides/federate_identity/#user-provided-certificates" %}}) below for details on how to format the certificate as a Kubernetes Secret.

We also specify the federation mode to be `PERMISSIVE`. This means we'll make services available between meshes. You can control this later by specifying different global service properties. 

Lastly, we are creating the VirtualMesh with two different service meshes: `istiod-istio-system-mgmt-cluster` and `istiod-istio-system-remote-cluster`. We can have any meshes defined here that should be part of this virtual grouping and federation.

##### User Provided Certificates

A root certificate for a VirtualMesh must be supplied to Service Mesh Hub 
as a Secret formatted as follows:

```yaml
kind: Secret
metadata:
  name: providedrootcert
  namespace: default
type: Opaque
data:
  key.pem: {private key file}
  root-cert.pem: {root CA certificate file}
```

Given a root certificate file `root-cert.pem` and its associated private key file `key.pem`,
this secret can be created by running:

`kubectl -n default create secret generic providedrootcert --from-file=root-cert.pem --from-file=key.pem`.

An example root certificate and private key file can be generated by following 
[this guide](https://github.com/istio/istio/tree/1.5.0/samples/certs) and running `make root-ca`.

Note that the name/namespace of the provided root cert cannot be `cacerts/istio-system` as that is used by Service Mesh Hub for carrying out the CSR ([certificate signing request](https://en.wikipedia.org/wiki/Certificate_signing_request)) procedure
that unifies the trust root between Meshes in the VirtualMesh.

##### Applying VirtualMesh

If you saved this VirtualMesh CR to a file named `demo-virtual-mesh.yaml`, you can apply it like this:

```shell
kubectl --context $MGMT_CONTEXT apply -f demo-virtual-mesh.yaml
```

Notice the `autoRestartPods: true` in the mtlsConfig stanza. This instructs Service Mesh Hub to restart the Istio pods in the relevant clusters. 

This is due to a limitation of Istio. The Istio control plane picks up the CA for Citadel and does not rotate it often enough. This is being [improved in future versions of Istio](https://github.com/istio/istio/issues/22993). 

If you wish to perform this step manually, set `autoRestartPods: false` and run the following:

```shell
meshctl mesh restart --mesh-name istiod-istio-system-mgmt-cluster
```

{{% notice note %}}
Note, after you bounce the control plane, it may still take time for the workload certs to get re-issued with the new CA. You can force the workloads to re-load by bouncing them. For example, for the bookinfo sample running in the `bookinfo` namespace:

```shell
kubectl --context $MGMT_CONTEXT -n bookinfo delete po --all
kubectl --context $REMOTE_CONTEXT -n bookinfo delete po --all
```
{{% /notice %}}

Creating this resource will instruct Service Mesh to establish a shared root identity across the clusters in the Virtual Mesh as well as federate the services. The next sections of this document help you understand some of the pieces of how this works.

## Understanding the Shared Root Process

When we create the VirtualMesh CR, set the trust model to `shared`, and configure the Root CA parameters, Service Mesh Hub will kick off the process to unify the identity to a shared root. First, Service Mesh Hub will either create the Root CA specified (if `generated` is used) or use the supplied CA information. 

Then Service Mesh Hub will use a Certificate Request (CR) agent on each of the affected clusters to create a new key/cert pair that will form an intermediate CA used by the mesh on that cluster. It will then create a Certificate Request, represented by the [CertificateRequest]({{% versioned_link_path fromRoot="/reference/api/certificate_request/" %}}) CR.

 Service Mesh Hub will sign the certificate with the Root CA specified in the VirtualMesh. At that point, we will want the mesh (Istio in this case) to pick up the new intermediate CA and start using that for its workloads.

![Service Mesh Hub Architecture]({{% versioned_link_path fromRoot="/img/smh-csr.png" %}})

To verify, let's check the `IssuedCertificates` CR in `remote-cluster-context`:

```shell
kubectl --context $REMOTE_CONTEXT \
get issuedcertificates -n service-mesh-hub
```

We should see this on the remote cluster:

```shell
NAME                                 AGE
istiod-istio-system-remote-cluster   3m15s
```

If we do the same on the `mgmt-cluster`, we should also see an `IssuedCertificates` entry there as well.

Lastly, let's verify the correct `cacerts` was created in the `istio-system` namespace that can be used for Istio's Citadel:

```shell
kubectl --context $MGMT_CONTEXT get secret -n istio-system cacerts 

NAME      TYPE                                          DATA   AGE
cacerts   certificates.smh.solo.io/issued_certificate   5      20s
```

```shell
kubectl --context $REMOTE_CONTEXT get secret -n istio-system cacerts 

NAME      TYPE                                          DATA   AGE
cacerts   certificates.smh.solo.io/issued_certificate   5      5m3s
```

In the previous section, we bounced the Istio control plane to pick up these intermediate certs. Again, this is being [improved in future versions of Istio](https://github.com/istio/istio/issues/22993). 


##### Multi-cluster mesh federation

Once trust has been established, Service Mesh Hub will start federating services so that they are accessible across clusters. Behind the scenes, Service Mesh Hub will handle the networking -- possibly through egress and ingress gateways, and possibly affected by user-defined traffic and access policies -- and ensure requests to the service will resolve and be routed to the right destination. Users can fine-tune which services are federated where by editing the virtual mesh. 

For example, you can see what Istio `ServiceEntry` objects were created. On the `mgmt-cluster` cluster you can see:

```shell
kubectl --context $MGMT_CONTEXT \
  get serviceentry -n istio-system
```

```shell
NAME                                                          HOSTS                                                           LOCATION        RESOLUTION   AGE
istio-ingressgateway.istio-system.svc.remote-cluster.global   [istio-ingressgateway.istio-system.svc.remote-cluster.global]   MESH_INTERNAL   DNS          6m2s
ratings.bookinfo.svc.remote-cluster.global                    [ratings.bookinfo.svc.remote-cluster.global]                    MESH_INTERNAL   DNS          6m2s
reviews.bookinfo.svc.remote-cluster.global                    [reviews.bookinfo.svc.remote-cluster.global]                    MESH_INTERNAL   DNS          6m2s
```

On the `remote-cluster-context` cluster, you can see:

```shell
kubectl --context $REMOTE_CONTEXT \
get serviceentry -n istio-system
```

```shell
NAME                                                            HOSTS                                                             LOCATION        RESOLUTION   AGE
details.bookinfo.svc.mgmt-cluster.global                    [details.bookinfo.svc.mgmt-cluster.global]                    MESH_INTERNAL   DNS          2m5s     
istio-ingressgateway.istio-system.svc.mgmt-cluster.global   [istio-ingressgateway.istio-system.svc.mgmt-cluster.global]   MESH_INTERNAL   DNS          5m18s    
productpage.bookinfo.svc.mgmt-cluster.global                [productpage.bookinfo.svc.mgmt-cluster.global]                MESH_INTERNAL   DNS          55s      
ratings.bookinfo.svc.mgmt-cluster.global                    [ratings.bookinfo.svc.mgmt-cluster.global]                    MESH_INTERNAL   DNS          7m2s     
reviews.bookinfo.svc.mgmt-cluster.global                    [reviews.bookinfo.svc.mgmt-cluster.global]                    MESH_INTERNAL   DNS          90s 
```

## See it in action

Check out "Part Two" of the ["Dive into Service Mesh Hub" video series](https://www.youtube.com/watch?v=4sWikVELr5M&list=PLBOtlFtGznBjr4E9xYHH9eVyiOwnk1ciK)
(note that the video content reflects Service Mesh Hub <b>v0.6.1</b>):

<iframe width="560" height="315" src="https://www.youtube.com/embed/djcDaIsqIl8" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>


## Next steps

At this point, you should be able to route traffic across your clusters with end-to-end mTLS. You can verify the certs following the same [approach we did earlier in this section]({{% versioned_link_path fromRoot="/guides/federate_identity/#verify-identity-in-two-clusters-is-different" %}}).

Now that you have a single logical "virtual mesh" you can begin configuring it with an API that is aware of this VirtualMesh concept. In the next sections, you can apply [access control]({{% versioned_link_path fromRoot="/guides/access_control_intro/" %}}) and [traffic policies]({{% versioned_link_path fromRoot="/guides/multicluster_communication/" %}}). 

