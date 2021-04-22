---
title: Advanced Enterprise Installation
menuTitle: Troubleshooting
weight: 1
description: Advanced options for Gloo Mesh Enterprise installation
---

Gloo Mesh Enterprise's default install options are sufficient for managing your service meshes in production. However,
select organizations have specific requirements. This guide addresses so 

#### Manual Certificate Creation

By default, Gloo Mesh Enterprise's default behavior is to create self-signed certificates at install time to handle
bootstrapping mTLS connectivity between the [server and agent components]({{% versioned_link_path fromRoot="/concepts/relay" %}})
of Gloo Mesh Enterprise. To learn how to provision your own certificates for this purpose, read on.

The following script illustrates how you can create these certificates manually. If you do not create certificates yourself, both the Helm chart and `meshctl` will create self-signed certificates for you.

At a high level, the script achieves the following:

1. Creates a root certificate (`RELAY_ROOT_CERT_NAME`). This is distributed to managed clusters
   so that relay agents can use it verify the identity of the relay server.

2. Create a certificate for the relay server (`RELAY_SERVER_CERT_NAME`), derived from the root certificate,
   which is presented by the server to relay agents.

3. Create a signing certificate (`RELAY_SIGNING_CERT_NAME`), derived from the root certificate,
   which is used by the relay server to issue certificates for relay agents once initial trust has been established.

{{% notice note %}}
The following was tested with OpenSSL 1.1.1h  22 Sep 2020.

Mac users may have LibreSSL installed by default. If so, we recommend that you `brew install openssl` and verify that
`openssl version` returns an OpenSSL version before you proceed.
{{% /notice %}}

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
