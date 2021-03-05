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

**Istio Ingress Setup**

For purposes of demonstration, the following describes how to configure a Kubernetes
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

To get the address of this ingress for use during [cluster registration]({{% versioned_link_path fromRoot="/setup/cluster_registration/enterprise_cluster_registration" %}}), run:

```shell
mgmtIngressAddress=$(kubectl get node -ojson | jq -r ".items[0].status.addresses[0].address")
mgmtIngressPort=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="https")].nodePort}')
ingressAddress=${mgmtIngressAddress}:${mgmtIngressPort}
```

## Establishing Trust Between Agents and Server

#### Manual Certificate Creation

As described in the [concept document]({{% versioned_link_path fromRoot="/concepts/relay" %}}),
gRPC communication between agents and the server is secured with TLS.

The following script illustrates how you can create these certificates manually.
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
   --context kind-${MGMT_CLUSTER} \
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

Last, in order to establish initial trust, we need to create a token shared by the relay server and all potential relay agents.
This is used by the server to initially verify the identity of a relay agent before issuing a certificate for that agent.
All registered clusters as well as the management cluster must have the same token value.

```shell
for context in ${MGMT_CONTEXT} ${REMOTE_CONTEXT}; do
  echo "creating shared identity token in cluster context ${context}..."
  kubectl apply --context ${context} -f- <<EOF
kind: Secret
apiVersion: v1
metadata:
  name: relay-identity-token-secret
  namespace: gloo-mesh
stringData:
  token: "your-token-value-of-choice"
EOF
done
```
