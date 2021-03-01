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
  namespace: your-namespace
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
  namespace: your-namespace
spec:
  hosts:
    - "enterprise-networking.gloo-mesh"
  gateways:
    - your-namespace/gloo-mesh-ingress
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


```bash
RELAY_ROOT_CERT_NAME=relay-root

openssl req -new -newkey rsa:4096 -x509 -sha256 \
        -days 3650 -nodes -out ${RELAY_ROOT_CERT_NAME}.crt -keyout ${RELAY_ROOT_CERT_NAME}.key \
        -subj "/CN=*.gloo-mesh/O=${RELAY_ROOT_CERT_NAME}" \
        -addext "subjectAltName = DNS:*.gloo-mesh"


echo "creating grpc server tls cert ..."

RELAY_SERVER_CERT_NAME=relay-server-tls

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
RELAY_SIGNING_CERT_NAME=relay-tls-signing
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
for cluster in ${MGMT_CLUSTER} ${REMOTE_CLUSTER}; do
echo "creating matching root cert for agent in cluster ${cluster}..."
  kubectl create secret generic ${RELAY_ROOT_CERT_NAME}-tls-secret \
  --from-file=ca.crt=${RELAY_ROOT_CERT_NAME}.crt \
  --dry-run=client -oyaml | kubectl apply -f- \
   --context kind-${cluster} \
  --namespace gloo-mesh
done

# Note: ${RELAY_SIGNING_CERT_NAME}-secret must match the server Helm value `signingTlsSecret.Name`
kubectl create secret generic ${RELAY_SIGNING_CERT_NAME}-secret \
  --from-file=tls.key=${RELAY_SIGNING_CERT_NAME}.key \
  --from-file=tls.crt=${RELAY_SIGNING_CERT_NAME}.crt \
  --from-file=ca.crt=${RELAY_ROOT_CERT_NAME}.crt \
  --dry-run=client -oyaml | kubectl apply -f- \
  --context kind-${MGMT_CLUSTER} \
  --namespace gloo-mesh
```

Lastly, in order to establish initial trust, we need to create a token shared by the relay server and all potential relay agents.
This is used by the server to initially verify the identity of a relay agent before issuing a certificate for that agent.

```shell
for cluster in ${MGMT_CLUSTER} ${REMOTE_CLUSTER}; do
  echo "creating shared identity token in ${cluster}..."
  kubectl apply --context kind-${cluster} -f- <<EOF
kind: Secret
apiVersion: v1
metadata:
  name: relay-identity-token-secret
  namespace: gloo-mesh
data:
  token: "your-token-value-of-choice"
EOF
```

#### meschtl

[comment]: <> (TODO document meshctl command once available)
