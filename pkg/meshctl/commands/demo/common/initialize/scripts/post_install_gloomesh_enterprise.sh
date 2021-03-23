#!/bin/bash -ex

cluster=$0

K="kubectl --context kind-${cluster}"

# wait for networking to roll out
${K} -n gloo-mesh rollout status deployment enterprise-networking

# sleep to allow CRDs to register
sleep 4

# install the istio ingress
${K} apply -n gloo-mesh -f- <<EOF
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
---
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
---
EOF
