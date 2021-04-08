---
title: "Prerequisites for Gloo Mesh Enterprise"
menuTitle: Prerequisites for Gloo Mesh Enterprise
description: Prerequisites for Gloo Mesh Enterprise
weight: 100
---

{{% notice note %}} Gloo Mesh Enterprise is required for this feature. {{% /notice %}}

This document describes the environmental prerequisites for Gloo Mesh Enterprise.
A conceptual overview of the Gloo Mesh Enterprise architecture can be found [here]({{% versioned_link_path fromRoot="/concepts/relay" %}}).

## Ingress Setup

In order for relay agents to communicate with the relay server, the management cluster
(i.e. the cluster on which the relay server is deployed) must be configured to accept
ingress traffic, the exact procedure of which depends on your particular environment.

**Option 1: LoadBalancer Service Type**

A simple way to expose the relay server to remote clusters is by setting the `enterprise-networking` service type to
[LoadBalancer](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types).
This can be done via Helm, by setting `enterprise-networking.enterpriseNetworking.serviceType=LoadBalancer` on the
Gloo Mesh Enterprise Helm chart.

LoadBalancer services are exposed to the externally via your Kubernetes cloud provider's load balancer. Note that this
approach **does not** work by default with Kind cluster deployments, but is a good option for getting started if you're
running your clusters via a managed service like Google Kubernetes Engine or Amazon's Elastic Kubernetes Service.

If you have your enterprise-networking service set as a LoadBalancer type,
get the relay address for [cluster registration]({{% versioned_link_path fromRoot="/setup/cluster_registration/enterprise_cluster_registration" %}}) by running:

```shell script
MGMT_INGRESS_ADDRESS=$(kubectl  -n gloo-mesh get service enterprise-networking -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
MGMT_INGRESS_PORT=$(kubectl -n gloo-mesh get service enterprise-networking -o jsonpath='{.spec.ports[?(@.name=="grpc")].port}')
RELAY_ADDRESS=${MGMT_INGRESS_ADDRESS}:${MGMT_INGRESS_PORT}
```

**Option 2: Istio Ingress Setup**

The enterprise-networking service can also be exposed via an ingress. The following describes how to configure a Kubernetes
cluster ingress assuming [Istio's ingress gateway model](https://istio.io/latest/docs/tasks/traffic-management/ingress/ingress-control/).

Create the following resources in the namespace of your choosing. Note that
we assume that the `enterprise-networking` deployment exposes its gRPC port on `9900`.

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: gloo-mesh-ingress
  namespace: gloo-mesh
spec:
  selector:
    istio: ingressgateway
  servers:
    - port:
        number: 443
        name: https
        protocol: HTTPS
      tls:
        mode: PASSTHROUGH
      hosts:
        - "enterprise-networking.gloo-mesh"
```

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: gloo-mesh-ingress
  namespace: gloo-mesh
spec:
  hosts:
    - "enterprise-networking.gloo-mesh"
  gateways:
    - gloo-mesh/gloo-mesh-ingress
  tls:
    - match:
        - port: 443
          sniHosts:
          - enterprise-networking.gloo-mesh
      route:
        - destination:
            host: enterprise-networking.gloo-mesh.svc.cluster.local
            port:
              number: 9900
```

You can set your ingress service as either a NodePort or LoadBalancer type.
From here, you need to find the address of this ingress to use
during [cluster registration]({{% versioned_link_path fromRoot="/setup/cluster_registration/enterprise_cluster_registration" %}}).

Some useful scripts for finding the ingress address can be found below. Note
that these scripts assume that 1) there is only one node, and 2) the ingress port is named "https":

{{< tabs >}}
{{< tab name="NodePort" codelang="yaml">}}
MGMT_INGRESS_ADDRESS=$(kubectl get node -ojson | jq -r ".items[0].status.addresses[0].address")
MGMT_INGRESS_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="https")].nodePort}')
RELAY_ADDRESS=${MGMT_INGRESS_ADDRESS}:${MGMT_INGRESS_PORT}
{{< /tab >}}
{{< tab name="LoadBalancer" codelang="yaml">}}
MGMT_INGRESS_ADDRESS=$(kubectl  -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
MGMT_INGRESS_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="https")].port}')
RELAY_ADDRESS=${MGMT_INGRESS_ADDRESS}:${MGMT_INGRESS_PORT}
{{< /tab >}}
{{< /tabs >}}

If you are using a NodePort and are failing to connect to relay,
that means you have multiple nodes, and the istio-ingressgateway isn't
on the first one. Track down the host for the correct worker node and
use that as the $MGMT_INGRESS_ADDRESS instead.

For more on determining the correct ingress address, see the [envoy docs](https://istio.io/latest/docs/tasks/traffic-management/ingress/ingress-control/).

## Establishing Trust Between Agents and Server

#### Manual Certificate Creation (Optional)

As described in the [concept document]({{% versioned_link_path fromRoot="/concepts/relay" %}}),
gRPC communication between agents and the server is authenticate and secured with mutual TLS.

The following script illustrates how you can create these certificates manually. If you do not create certificates yourself, both the Helm chart and `meshctl` will create self-signed certificates for you.

At a high level, the script achieves the following:

1. Creates a root certificate (`RELAY_ROOT_CERT_NAME`). This is distributed to managed clusters
   so that relay agents can use it verify the identity of the relay server.

2. Create a certificate for the relay server (`RELAY_SERVER_CERT_NAME`), derived from the root certificate,
   which is presented by the server to relay agents.

3. Create a signing certificate (`RELAY_SIGNING_CERT_NAME`), derived from the root certificate,
   which is used by the relay server to issue certificates for relay agents once initial trust has been established.

{{% notice note %}} The following was tested with OpenSSL 1.1.1h  22 Sep 2020. {{% /notice %}}

```bash
RELAY_ROOT_CERT_NAME=relay-root
RELAY_SERVER_CERT_NAME=relay-server-tls
RELAY_SIGNING_CERT_NAME=relay-tls-signing

echo "creating root cert ..."

openssl req -new -newkey rsa:4096 -x509 -sha256 \
        -days 3650 -nodes -out ${RELAY_ROOT_CERT_NAME}.crt -keyout ${RELAY_ROOT_CERT_NAME}.key \
        -subj "/CN=*.gloo-mesh/O=${RELAY_ROOT_CERT_NAME}" \
        -addext "subjectAltName = DNS:*.gloo-mesh"


echo "creating grpc server tls cert ..."

# server cert
cat > "${RELAY_SERVER_CERT_NAME}.conf" <<EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth, serverAuth
subjectAltName = @alt_names
[alt_names]
DNS = *.gloo-mesh
EOF

openssl genrsa -out "${RELAY_SERVER_CERT_NAME}.key" 2048
openssl req -new -key "${RELAY_SERVER_CERT_NAME}.key" -out ${RELAY_SERVER_CERT_NAME}.csr -subj "/CN=*.gloo-mesh/O=${RELAY_SERVER_CERT_NAME}" -config "${RELAY_SERVER_CERT_NAME}.conf"
openssl x509 -req \
  -days 3650 \
  -CA ${RELAY_ROOT_CERT_NAME}.crt -CAkey ${RELAY_ROOT_CERT_NAME}.key \
  -set_serial 0 \
  -in ${RELAY_SERVER_CERT_NAME}.csr -out ${RELAY_SERVER_CERT_NAME}.crt \
  -extensions v3_req -extfile "${RELAY_SERVER_CERT_NAME}.conf"

echo "creating identity server signing cert ..."

# signing cert
cat > "${RELAY_SIGNING_CERT_NAME}.conf" <<EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = critical,CA:TRUE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment, keyCertSign
extendedKeyUsage = clientAuth, serverAuth
subjectAltName = @alt_names
[alt_names]
DNS = *.gloo-mesh
EOF

openssl genrsa -out "${RELAY_SIGNING_CERT_NAME}.key" 2048
openssl req -new -key "${RELAY_SIGNING_CERT_NAME}.key" -out ${RELAY_SIGNING_CERT_NAME}.csr -subj "/CN=*.gloo-mesh/O=${RELAY_SIGNING_CERT_NAME}" -config "${RELAY_SIGNING_CERT_NAME}.conf"
openssl x509 -req \
  -days 3650 \
  -CA ${RELAY_ROOT_CERT_NAME}.crt -CAkey ${RELAY_ROOT_CERT_NAME}.key \
  -set_serial 0 \
  -in ${RELAY_SIGNING_CERT_NAME}.csr -out ${RELAY_SIGNING_CERT_NAME}.crt \
  -extensions v3_req -extfile "${RELAY_SIGNING_CERT_NAME}.conf"
```

Next, we'll create secrets on the management and remote clusters with all the requisite generated certificates. Let's
define the management cluster as the cluster where we'll install the Gloo Mesh control plane, and the remote cluster as
a cluster running a service mesh we would like Gloo Mesh to discover and configure. On the management cluster, we need
the root cert, server cert, and signing cert. On the remote clusters, we just need the root cert.

```bash
MGMT_CLUSTER=mgmt-cluster
MGMT_CONTEXT=kind-mgmt-cluster

REMOTE_CLUSTER=remote-cluster
REMOTE_CONTEXT=kind-remote-cluster

# ensure gloo-mesh namespace exists on both mgmt and remote clusters
for context in ${MGMT_CONTEXT} ${REMOTE_CONTEXT}; do
  kubectl --context ${context} create namespace gloo-mesh
done

# create secrets from certs

# Note: ${RELAY_SERVER_CERT_NAME}-secret must match the server Helm value `relayTlsSecret.Name`
kubectl create secret generic ${RELAY_SERVER_CERT_NAME}-secret \
  --from-file=tls.key=${RELAY_SERVER_CERT_NAME}.key \
  --from-file=tls.crt=${RELAY_SERVER_CERT_NAME}.crt \
  --from-file=ca.crt=${RELAY_ROOT_CERT_NAME}.crt \
  --dry-run=client -oyaml | kubectl apply -f- \
   --context ${MGMT_CONTEXT} \
  --namespace gloo-mesh

# Note: ${RELAY_ROOT_CERT_NAME}-tls-secret must match the agent Helm value `relay.rootTlsSecret.Name`
for context in ${MGMT_CONTEXT} ${REMOTE_CONTEXT}; do
echo "creating matching root cert for agent in cluster context ${context}..."
  kubectl create secret generic ${RELAY_ROOT_CERT_NAME}-tls-secret \
  --from-file=ca.crt=${RELAY_ROOT_CERT_NAME}.crt \
  --dry-run=client -oyaml | kubectl apply -f- \
   --context ${context} \
  --namespace gloo-mesh
done

# Note: ${RELAY_SIGNING_CERT_NAME}-secret must match the server Helm value `signingTlsSecret.Name`
kubectl create secret generic ${RELAY_SIGNING_CERT_NAME}-secret \
  --from-file=tls.key=${RELAY_SIGNING_CERT_NAME}.key \
  --from-file=tls.crt=${RELAY_SIGNING_CERT_NAME}.crt \
  --from-file=ca.crt=${RELAY_ROOT_CERT_NAME}.crt \
  --dry-run=client -oyaml | kubectl apply -f- \
  --context ${MGMT_CONTEXT} \
  --namespace gloo-mesh
```

You can also choose to use your own internal PKI to create and assign certificates to Gloo Mesh Enterprise. At a minimum, you would need the certificate chain to your PKI root CA, a server certificate for relay server communication, and a signing certificate the relay server can use to generate relay agent certificates. Please refer to your PKI documentation for more information.