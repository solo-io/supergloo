---
title: Vault PKI
menuTitle: Vault PKI
description: Guide to using vault with gloo-mesh to secure istio traffic
weight: 30
---

{{% notice note %}} Gloo Mesh Enterprise is required for this feature. {{% /notice %}}

## Before you begin

This guide assumes the following:

  * Gloo Mesh Enterprise is [installed in relay mode and running on the `cluster-1`]({{% versioned_link_path fromRoot="/setup/install-gloo-mesh" %}})
  * `gloo-mesh` is the installation namespace for Gloo Mesh
  * `enterprise-networking` is deployed on `cluster-1` in the `gloo-mesh` namespace and exposes its gRPC server on port 9900
  * `enterprise-agent` is deployed on both clusters and exposes its gRPC server on port 9977
  * Both `cluster-1` and `cluster-2` are [registered with Gloo Mesh]({{% versioned_link_path fromRoot="/guides/#two-registered-clusters" %}})
  * Istio is [installed on both clusters]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}) clusters
  * `istio-system` is the root namespace for both Istio deployments
  * The `bookinfo` app is [installed into the two clusters]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployed-on-two-clusters" %}}) under the `bookinfo` namespace
  * the following environment variables are set:
    ```shell
    CONTEXT_1=cluster_1's_context
    CONTEXT_2=cluster_2's_context
    ```

## Installing Vault

{{% notice note %}} Installing Vault is an optional step. An existing Vault deployment may be used if your organization already has one. {{% /notice %}}

1. Install Vault using helm
```shell
helm repo add hashicorp https://helm.releases.hashicorp.com
```

2. Generate root-cert and key for Vault
```shell
openssl req -new -newkey rsa:4096 -x509 -sha256 \
    -days 3650 -nodes -out root-cert.pem -keyout root-key.pem \
    -subj "/O=my-org"
```

3. Let's install vault, and add our root-ca to each deployment
```shell
for cluster in ${CONTEXT_1} ${CONTEXT_2}; do

  # For more info on vault in kubernetes, please see: https://learn.hashicorp.com/tutorials/vault/kubernetes-cert-manager

  # install vault in dev mode
  helm install -n vault  vault hashicorp/vault --set "injector.enabled=false" --set "server.dev.enabled=true" --kube-context="${cluster}" --create-namespace

  # Wait for vault to come up, can't use kubectl rollout because it's a stateful set without rolling deployment
  kubectl --context="${cluster}" wait --for=condition=Ready -n vault pod/vault-0

  # Enable vault kubernetes Auth
  kubectl --context="${cluster}" exec -n vault vault-0 -- /bin/sh -c 'vault auth enable kubernetes'

  # Set the kubernetes Auth config for vault to the mounted token
  kubectl --context="${cluster}" exec -n vault vault-0 -- /bin/sh -c 'vault write auth/kubernetes/config \
    token_reviewer_jwt="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
    kubernetes_host="https://$KUBERNETES_PORT_443_TCP_ADDR:443" \
    kubernetes_ca_cert=@/var/run/secrets/kubernetes.io/serviceaccount/ca.crt'

  # Bind the istiod service account to the pki policy below
  kubectl --context="${cluster}" exec -n vault vault-0 -- /bin/sh -c 'vault write auth/kubernetes/role/gen-int-ca-istio \
    bound_service_account_names=istiod-service-account \
    bound_service_account_namespaces=istio-system \
    policies=gen-int-ca-istio \
    ttl=2400h'

  # Initialize vault PKI
  kubectl --context="${cluster}" exec -n vault vault-0 -- /bin/sh -c 'vault secrets enable pki'

  # set the vault CA to our pem_bundle
  kubectl --context="${cluster}" exec -n vault vault-0 -- /bin/sh -c "vault write -format=json pki/config/ca pem_bundle=\"$(cat root-key.pem root-cert.pem)\""

  # Initialize vault intermediate path
  kubectl --context="${cluster}" exec -n vault vault-0 -- /bin/sh -c 'vault secrets enable -path pki_int pki'

  # Set the policy for intermediate cert path
  kubectl --context="${cluster}" exec -n vault vault-0 -- /bin/sh -c 'vault policy write gen-int-ca-istio - <<EOF
path "pki_int/*" {
capabilities = ["create", "read", "update", "delete", "list"]
}
path "pki/cert/ca" {
capabilities = ["read"]
}
path "pki/root/sign-intermediate" {
capabilities = ["create", "read", "update", "list"]
}
EOF'

done
```

## Enabling Vault as an intermediate CA

Now we need to federate our 2 meshes together using Vault to federate identity. To do this we will need to create/edit a `VirtualMesh` with the new Vault shared mTLS config

{{< highlight yaml "hl_lines=10-20" >}}
apiVersion: networking.mesh.gloo.solo.io/v1
kind: VirtualMesh
metadata:
  name: virtual-mesh
  namespace: gloo-mesh
spec:
  mtlsConfig:
    autoRestartPods: true
    shared:
      intermediateCertificateAuthority:
        vault:
          # Vault path to the CA endpoint
          caPath: "pki/root/sign-intermediate"
          # Vault path to the CSR endpoint
          csrPath: "pki_int/intermediate/generate/exported"
          # Vault server endpoint
          server: "http://vault.vault:8200"
          # Auth mechanism to use
          kubernetesAuth:
            role: "gen-int-ca-istio"
  federation:
    # federate all Destinations to all external meshes
    selectors:
    - {}
  meshes:
  - name: istiod-istio-system-cluster-1 
    namespace: gloo-mesh
  - name: istiod-istio-system-cluster-2
    namespace: gloo-mesh
{{< /highlight >}}