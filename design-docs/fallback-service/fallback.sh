#!/bin/bash -e

#####################################
#
# Set up two kind clusters:
#   1. a management plane which will have a service-mesh-hub namespace, and
#   2. a target cluster
#
# The management plane will have the appropriate secret for communicating with the target cluster
# Your kube context will be left pointing to the management plane cluster
# The target cluster will have Istio set up in the istio-system namespace in its demo profile
#
# To clean up **ALL** of your kind clusters, run this script as: `bash ci/setup-kind.sh cleanup`
# I had some trouble with the docker VM running out of disk space- run this cleanup step often if you can
#
#####################################


if [ "$1" == "cleanup" ]; then
  kind get clusters | grep meshhub-$2- | while read -r r; do kind delete cluster --name "$r"; done
  exit 0
fi

# generate names: $1 allows to make several envs in parallel 
managementPlane=meshhub-$1-management-plane 
remoteCluster=meshhub-$1-target-cluster

# set up each cluster
(cat <<EOF | kind create cluster --name $managementPlane --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
# - role: worker

kubeadmConfigPatches:
- |
  kind: InitConfiguration
  nodeRegistration:
    kubeletExtraArgs:
      node-labels: "topology.kubernetes.io/region=us-east-1,topology.kubernetes.io/zone=us-east-1c"
      authorization-mode: "AlwaysAllow"
      feature-gates: "EphemeralContainers=true"
- |
  kind: KubeletConfiguration
  featureGates:
    EphemeralContainers: true
- |
  kind: KubeProxyConfiguration
  featureGates:
    EphemeralContainers: true
- |
  kind: ClusterConfiguration
  metadata:
    name: config
  apiServer:
    extraArgs:
      "feature-gates": "EphemeralContainers=true"
  scheduler:
    extraArgs:
      "feature-gates": "EphemeralContainers=true"
  controllerManager:
    extraArgs:
      "feature-gates": "EphemeralContainers=true"
EOF
)&


# Create NodePort for remote cluster so it can be reachable from the management plane.
# This config is roughly based on: https://kind.sigs.k8s.io/docs/user/ingress/
cat <<EOF | kind create cluster --name $remoteCluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 32000
    hostPort: 32000
    protocol: TCP
kubeadmConfigPatches:
- |
  kind: InitConfiguration
  nodeRegistration:
    kubeletExtraArgs:
      node-labels: "topology.kubernetes.io/region=eu-west-2,topology.kubernetes.io/zone=eu-west-2b"
      authorization-mode: "AlwaysAllow"
      feature-gates: "EphemeralContainers=true"
- |
  kind: KubeletConfiguration
  featureGates:
    EphemeralContainers: true
- |
  kind: KubeProxyConfiguration
  featureGates:
    EphemeralContainers: true
- |
  kind: ClusterConfiguration
  metadata:
    name: config
  apiServer:
    extraArgs:
      "feature-gates": "EphemeralContainers=true"
  scheduler:
    extraArgs:
      "feature-gates": "EphemeralContainers=true"
  controllerManager:
    extraArgs:
      "feature-gates": "EphemeralContainers=true"
EOF

wait

printf "\n\n---\n"
echo "Finished setting up cluster $managementPlane"
echo "Finished setting up cluster $remoteCluster"

kubectl create ns --context kind-$managementPlane istio-system
kubectl create ns --context kind-$remoteCluster istio-system

ISTIO_HOME=$HOME/bin/istio-1.5.1/

kubectl --context kind-$managementPlane create secret generic cacerts -n istio-system \
    --from-file=$ISTIO_HOME/samples/certs/ca-cert.pem \
    --from-file=$ISTIO_HOME/samples/certs/ca-key.pem \
    --from-file=$ISTIO_HOME/samples/certs/root-cert.pem \
    --from-file=$ISTIO_HOME/samples/certs/cert-chain.pem
kubectl --context kind-$remoteCluster create secret generic cacerts -n istio-system \
    --from-file=$ISTIO_HOME/samples/certs/ca-cert.pem \
    --from-file=$ISTIO_HOME/samples/certs/ca-key.pem \
    --from-file=$ISTIO_HOME/samples/certs/root-cert.pem \
    --from-file=$ISTIO_HOME/samples/certs/cert-chain.pem

values=$(mktemp devportal-XXXXXXX.yaml --tmpdir)
cat > $values <<EOF
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: example-istiooperator
  namespace: istio-operator
spec:
  profile: minimal
  addonComponents:
    istiocoredns:
      enabled: true
  components:
    pilot:
      k8s:
        env:
          - name: PILOT_CERT_PROVIDER
            value: "kubernetes"
    proxy:
      k8s:
        env:
          - name: PILOT_CERT_PROVIDER
            value: "kubernetes"
    # Istio Gateway feature
    ingressGateways:
    - name: istio-ingressgateway
      enabled: true
      k8s:
        env:
          - name: ISTIO_META_ROUTER_MODE
            value: "sni-dnat"
          - name: PILOT_CERT_PROVIDER
            value: "kubernetes"
        service:
          ports:
            - port: 80
              targetPort: 8080
              name: http2
            - port: 443
              targetPort: 8443
              name: https
            - port: 15443
              targetPort: 15443
              name: tls
              nodePort: 32000
  values:
    prometheus:
      enabled: false
    gateways:
      istio-ingressgateway:
        type: NodePort
        ports:
          - targetPort: 15443
            name: tls
            nodePort: 32000
            port: 15443
    global:
      multiCluster:
        enabled: true
      pilotCertProvider: kubernetes
      controlPlaneSecurityEnabled: true
      mtls:
        enabled: true
      podDNSSearchNamespaces:
      - global
      - '{{ valueOrDefault .DeploymentMeta.Namespace "default" }}.global'
    security:
      selfSigned: false
EOF

$ISTIO_HOME/bin/istioctl manifest apply --context kind-$remoteCluster --set profile=minimal -f $values &

values2=$(mktemp devportal-XXXXXXX.yaml --tmpdir)
cat > $values2 <<EOF
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: example-istiooperator
  namespace: istio-operator
spec:
  profile: minimal
  addonComponents:
    istiocoredns:
      enabled: true
  components:
    pilot:
      k8s:
        env:
          - name: PILOT_CERT_PROVIDER
            value: "kubernetes"
    proxy:
      k8s:
        env:
          - name: PILOT_CERT_PROVIDER
            value: "kubernetes"
  values:
    global:
      multiCluster:
        enabled: true
      pilotCertProvider: kubernetes
      controlPlaneSecurityEnabled: true
      mtls:
        enabled: true
      podDNSSearchNamespaces:
      - global
      - '{{ valueOrDefault .DeploymentMeta.Namespace "default" }}.global'
    security:
      selfSigned: false
EOF

$ISTIO_HOME/bin/istioctl manifest apply --context kind-$managementPlane --set profile=minimal -f $values2

wait

rm $values $values2
kubectl --context kind-$managementPlane apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
data:
  Corefile: |
    .:53 {
        errors
        health
        ready
        kubernetes cluster.local in-addr.arpa ip6.arpa {
           pods insecure
           upstream
           fallthrough in-addr.arpa ip6.arpa
        }
        prometheus :9153
        forward . /etc/resolv.conf
        cache 30
        loop
        reload
        loadbalance
    }
    global:53 {
        errors
        cache 30
        forward . $(kubectl --context kind-$managementPlane get svc -n istio-system istiocoredns -o jsonpath={.spec.clusterIP}):53
    }
EOF
kubectl --context kind-$remoteCluster apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
data:
  Corefile: |
    .:53 {
        errors
        health
        ready
        kubernetes cluster.local in-addr.arpa ip6.arpa {
           pods insecure
           upstream
           fallthrough in-addr.arpa ip6.arpa
        }
        prometheus :9153
        forward . /etc/resolv.conf
        cache 30
        loop
        reload
        loadbalance
    }
    global:53 {
        errors
        cache 30
        forward . $(kubectl --context kind-$remoteCluster get svc -n istio-system istiocoredns -o jsonpath={.spec.clusterIP}):53
    }
EOF

kubectl --context kind-$managementPlane -n istio-system rollout status deployment istiod
kubectl --context kind-$remoteCluster -n istio-system rollout status deployment istiod

function print_debug_info {
  kubectl --context kind-$managementPlane get pod -A
  kubectl --context kind-$remoteCluster get pod -A
}
trap print_debug_info ERR

# label bookinfo namespaces for injection
kubectl --context kind-$managementPlane label namespace default istio-injection=enabled
kubectl --context kind-$remoteCluster label namespace default istio-injection=enabled

# Apply bookinfo deployments and services
kubectl --context kind-$managementPlane apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'app,version notin (v3)'
kubectl --context kind-$managementPlane apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'account'

kubectl --context kind-$remoteCluster apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'app,version in (v3)'
kubectl --context kind-$remoteCluster apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'service=reviews'
kubectl --context kind-$remoteCluster apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'account=reviews'
kubectl --context kind-$remoteCluster apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'app=ratings'
kubectl --context kind-$remoteCluster apply -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml -l 'account=ratings'

# wait for deployments to finish
kubectl --context kind-$managementPlane rollout status deployment/productpage-v1
kubectl --context kind-$managementPlane rollout status deployment/reviews-v1
kubectl --context kind-$managementPlane rollout status deployment/reviews-v2

kubectl --context kind-$remoteCluster rollout status deployment/reviews-v3


CLUSTER2_GW_ADDR=$(docker exec $remoteCluster-control-plane ip addr show dev eth0 | sed -nE 's|\s*inet\s+([0-9.]+).*|\1|p')
kubectl apply --context kind-$managementPlane -n default -f - <<EOF
apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: reviews-default
spec:
  hosts:
  # must be of form name.namespace.global
  - reviews.default.global
  # Treat remote cluster services as part of the service mesh
  # as all clusters in the service mesh share the same root of trust.
  location: MESH_INTERNAL
  ports:
  - name: http1
    number: 9080
    protocol: http
  resolution: DNS
  addresses:
  # the IP address to which httpbin.bar.global will resolve to
  # must be unique for each remote service, within a given cluster.
  # This address need not be routable. Traffic for this IP will be captured
  # by the sidecar and routed appropriately.
  - 240.0.0.2
  endpoints:
  # This is the routable address of the ingress gateway in cluster2 that
  # sits in front of sleep.foo service. Traffic from the sidecar will be
  # routed to this address.
  - address: ${CLUSTER2_GW_ADDR}
    ports:
      http1: 32000 # Do not change this port value
EOF

kubectl apply --context kind-$managementPlane -n default -f - <<EOF
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: reviews-global
spec:
  host: reviews.default.global
  trafficPolicy:
    tls:
      mode: ISTIO_MUTUAL
    outlierDetection:
      consecutiveErrors: 7
      interval: 5m
      baseEjectionTime: 15m
EOF

kubectl apply --context kind-$managementPlane -n default -f - <<EOF
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: reviews-local
spec:
  host: reviews.default.svc.cluster.local
  trafficPolicy:
    outlierDetection:
      consecutiveErrors: 7
      interval: 5m
      baseEjectionTime: 15m
    tls:
      mode: ISTIO_MUTUAL
EOF

kubectl apply --context kind-$managementPlane -n default -f - <<EOF
apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: reviews2-global
spec:
  hosts:
  # must be of form name.namespace.global
  - reviews2.default.global
  # Treat remote cluster services as part of the service mesh
  # as all clusters in the service mesh share the same root of trust.
  location: MESH_INTERNAL
  ports:
  - name: http1
    number: 9080
    protocol: http
  resolution: DNS
  addresses:
  # the IP address to which httpbin.bar.global will resolve to
  # must be unique for each remote service, within a given cluster.
  # This address need not be routable. Traffic for this IP will be captured
  # by the sidecar and routed appropriately.
  - 240.0.0.20
EOF

kubectl apply --context kind-$managementPlane -n default -f - <<EOF
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: reviews2.default.global
spec:
  configPatches:
  - applyTo: CLUSTER
    match:
      context: ANY
      cluster:
        name: "outbound|9080||reviews2.default.global"
    patch:
      operation: REMOVE
  - applyTo: CLUSTER
    match:
      context: ANY
      cluster:
        name: "outbound|9080||reviews2.default.global"
    patch:
      operation: ADD
      value: # cluster specification
        name: "outbound|9080||reviews2.default.global"
        connect_timeout: 1s
        lb_policy: CLUSTER_PROVIDED
        cluster_type:
          name: envoy.clusters.aggregate
          typed_config:
            "@type": type.googleapis.com/udpa.type.v1.TypedStruct
            type_url: type.googleapis.com/envoy.config.cluster.aggregate.v2alpha.ClusterConfig
            value:
              clusters:
              - outbound|9080||reviews.default.svc.cluster.local
              - outbound|9080||reviews.default.global
EOF

kubectl apply --context kind-$managementPlane -n default -f - <<EOF
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: reviews3.default.global
spec:
  configPatches:
  - applyTo: CLUSTER
    match:
      context: ANY
      cluster: {}
    patch:
      operation: ADD
      value: # cluster specification
        name: outbound|9080||reviews3.default.global
        connect_timeout: 1s
        lb_policy: CLUSTER_PROVIDED
        cluster_type:
          name: envoy.clusters.aggregate
          typed_config:
            "@type": type.googleapis.com/udpa.type.v1.TypedStruct
            type_url: type.googleapis.com/envoy.config.cluster.aggregate.v2alpha.ClusterConfig
            value:
              clusters:
              - outbound|9080||reviews.default.svc.cluster.local
              - outbound|9080||reviews.default.global
EOF

# kubectl apply --context kind-$managementPlane -n default -f - <<EOF
# apiVersion: networking.istio.io/v1alpha3
# kind: VirtualService
# metadata:
#   name: reviews
# spec:
#   hosts:
#   - reviews
#   http:
#   - route:
#     - destination:
#         host: reviews2.default.global
#         port:
#           number: 9080
# EOF

# kubectl apply --context kind-$remoteCluster -f - <<EOF
# apiVersion: networking.istio.io/v1beta1
# kind: VirtualService
# metadata:
#   name: reviews2.default.global
# spec:
#   gateways:
#   - istio-multicluster-ingressgateway2
#   hosts:
#   - reviews2.default.global2
#   http:
#   - route:
#     - destination:
#         host: reviews.default.svc.cluster.local
#         port:
#           number: 9080
# EOF

# kubectl apply --context kind-$remoteCluster -f - <<EOF
# apiVersion: networking.istio.io/v1beta1
# kind: Gateway
# metadata:
#   labels:
#     app: istio-ingressgateway
#     istio: ingressgateway
#   name: istio-multicluster-ingressgateway2
# spec:
#   selector:
#     istio: ingressgateway
#   servers:
#   - hosts:
#     - '*.global2'
#     port:
#       name: tls
#       number: 15443
#       protocol: TLS
# EOF

echo
echo managementPlane=$managementPlane
echo remoteCluster=$remoteCluster
echo CLUSTER2_GW_ADDR=$CLUSTER2_GW_ADDR

kubectl port-forward --context kind-$remoteCluster -n istio-system deploy/istio-ingressgateway 15000 &
sleep 10
curl -XPOST "http://localhost:15000/logging?level=debug"
killall kubectl
echo kubectl --context kind-$remoteCluster logs -n istio-system deploy/istio-ingressgateway -f

kubectl port-forward --context kind-$managementPlane -n default deploy/details-v1 15000 &
sleep 10
curl -XPOST "http://localhost:15000/logging?level=debug"
killall kubectl
echo kubectl --context kind-$managementPlane logs -n default deploy/details-v1 -c istio-proxy -f


kubectl --context kind-$managementPlane exec -ti deploy/details-v1 -c details -- bash -c "apt update && apt install curl --yes"
kubectl --context kind-$managementPlane exec -ti deploy/details-v1 -c details -- curl http://reviews.default.global:9080/reviews/1 -v
# kubectl --context kind-$managementPlane exec -ti deploy/details-v1 -c details -- curl http://reviews2.default.global:9080/reviews/1 -v

set +x
echo
echo managementPlane=$managementPlane
echo remoteCluster=$remoteCluster
echo CLUSTER2_GW_ADDR=$CLUSTER2_GW_ADDR
