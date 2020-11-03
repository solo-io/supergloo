#!/bin/bash -ex

#####################################
# Functions for setting up kind clusters
#####################################

#!/bin/bash

INSTALL_DIR=${PROJECT_ROOT}/install
AGENT_VALUES=${INSTALL_DIR}/helm/cert-agent/values.yaml
AGENT_IMAGE_REGISTRY=$(cat ${AGENT_VALUES} | grep "registry: " | awk '{print $2}')
AGENT_IMAGE_REPOSITORY=$(cat ${AGENT_VALUES} | grep "repository: " | awk '{print $2}')
AGENT_IMAGE_TAG=$(cat ${AGENT_VALUES} | grep "tag: " | awk '{print $2}' | sed 's/"//g')

AGENT_IMAGE=${AGENT_IMAGE_REGISTRY}/${AGENT_IMAGE_REPOSITORY}:${AGENT_IMAGE_TAG}
AGENT_CHART=${INSTALL_DIR}/helm/_output/charts/cert-agent/cert-agent-${AGENT_IMAGE_TAG}.tgz

SMH_VALUES=${INSTALL_DIR}/helm/service-mesh-hub/values.yaml
SMH_IMAGE_TAG=$(cat ${SMH_VALUES} | grep -m 1 "tag: " | awk '{print $2}' | sed 's/"//g')
SMH_CHART=${INSTALL_DIR}/helm/_output/charts/service-mesh-hub/service-mesh-hub-${SMH_IMAGE_TAG}.tgz

#### FUNCTIONS

function create_kind_cluster() {
  # The default version of k8s under Linux is 1.18
  # https://github.com/solo-io/service-mesh-hub/issues/700
  kindImage=kindest/node:v1.17.5

  cluster=$1
  port=$2

  echo "creating cluster ${cluster} with ingress port ${port}"

  K="kubectl --context=kind-${cluster}"

  # This config is roughly based on: https://kind.sigs.k8s.io/docs/user/ingress/
  cat <<EOF | kind create cluster --name "${cluster}" --image $kindImage --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: ${port}
    hostPort: ${port}
    protocol: TCP
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
kubeadmConfigPatches:
- |
  kind: InitConfiguration
  nodeRegistration:
    kubeletExtraArgs:
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

  # NOTE: we delete the local-path-storage ns to free up CPU for ci
  ${K} delete ns local-path-storage
}

# Operator spec for istio 1.5.x and 1.6.x
function install_istio_1_5() {
  cluster=$1
  port=$2
  K="kubectl --context=kind-${cluster}"

  echo "installing istio to ${cluster}..."

  cat << EOF | istioctl manifest apply --context "kind-${cluster}" -f -
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: example-istiooperator
  namespace: istio-system
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
              nodePort: ${port}
    egressGateways:
    - name: istio-egressgateway
      enabled: true
  values:
    prometheus:
      enabled: false
    gateways:
      istio-ingressgateway:
        type: NodePort
        ports:
          - targetPort: 15443
            name: tls
            nodePort: ${port}
            port: 15443
    global:
      pilotCertProvider: kubernetes
      controlPlaneSecurityEnabled: true
      mtls:
        enabled: true
      podDNSSearchNamespaces:
      - global
EOF
}

# Operator spec for istio 1.7.x
function install_istio_1_7() {
  cluster=$1
  port=$2
  K="kubectl --context=kind-${cluster}"

  echo "installing istio to ${cluster}..."

  cat << EOF | istioctl manifest install --context "kind-${cluster}" -f -
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: example-istiooperator
  namespace: istio-system
spec:
  profile: minimal
  addonComponents:
    istiocoredns:
      enabled: true
  components:
    # Istio Gateway feature
    ingressGateways:
    - name: istio-ingressgateway
      enabled: true
      k8s:
        env:
          - name: ISTIO_META_ROUTER_MODE
            value: "sni-dnat"
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
              nodePort: ${port}
    egressGateways:
    - name: istio-egressgateway
      enabled: true
  meshConfig:
    enableAutoMtls: true
    accessLogFile: "/dev/stdout"
  values:
    prometheus:
      enabled: false
    gateways:
      istio-ingressgateway:
        type: NodePort
        ports:
          - targetPort: 15443
            name: tls
            nodePort: ${port}
            port: 15443
    global:
      pilotCertProvider: kubernetes
      controlPlaneSecurityEnabled: true
      podDNSSearchNamespaces:
      - global
EOF
}

function install_istio() {
  cluster=$1
  port=$2
  K="kubectl --context=kind-${cluster}"

  if istioctl version | grep 1.7
  then
    install_istio_1_7 $cluster $port
  else
    install_istio_1_5 $cluster $port
  fi

  # enable istio dns for .global stub domain:
  ISTIO_COREDNS=$(${K} get svc -n istio-system istiocoredns -o jsonpath={.spec.clusterIP})
  ${K} apply -f - <<EOF
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
        forward . ${ISTIO_COREDNS}:53
    }
EOF

  printf "\n\n---\n"
  echo "Finished setting up cluster ${cluster}"

}

function install_osm() {
  cluster=$1
  port=$2
  K="kubectl --context=kind-${cluster}"

  echo "installing osm to ${cluster}..."
  ROLLOUT="${K} rollout status deployment --timeout 300s"

  VERSION=$(osm version)
  V3='v0.3.0'
  # install in permissive mode for testing
  if [[ "$VERSION" == *"$V3"* ]]; then
    osm install --enable-metrics-stack=false --deploy-zipkin=false
  else
    osm install
  fi


  ${ROLLOUT} -n osm-system osm-controller

  for i in bookstore bookthief bookwarehouse bookbuyer; do ${K} create ns $i; done

  if [[ "$VERSION" == *"$V3"* ]]; then
    for i in bookstore bookthief bookwarehouse bookbuyer; do osm namespace add $i; done
  else
  # for OSM versions >= v0.4.0
    for i in bookstore bookthief bookwarehouse bookbuyer; do osm namespace add $i --enable-sidecar-injection; done
  fi


  ${K} apply -f ${PROJECT_ROOT}/ci/osm-demo.yaml

  ${ROLLOUT} -n bookstore bookstore-v1
  ${ROLLOUT} -n bookstore bookstore-v2
  ${ROLLOUT} -n bookthief bookthief
  ${ROLLOUT} -n bookwarehouse bookwarehouse
  ${ROLLOUT} -n bookbuyer bookbuyer
}

function get_api_address() {
  cluster=$1
  case $(uname) in
    "Darwin")
    {
        apiServerAddress=host.docker.internal
    } ;;
    "Linux")
    {
        apiServerAddress=$(docker exec "${cluster}-control-plane" ip addr show dev eth0 | sed -nE 's|\s*inet\s+([0-9.]+).*|\1|p'):6443
    } ;;
    *)
    {
        echo "Unsupported OS"
        exit 1
    } ;;
  esac
  echo ${apiServerAddress}
}

function register_cluster() {
  cluster=$1
  apiServerAddress=$(get_api_address ${cluster})

  K="kubectl --context kind-${cluster}"

  echo "registering ${cluster} with local cert-agent image..."

  # load cert-agent image
  kind load docker-image --name "${cluster}" "${AGENT_IMAGE}"

  go run "${PROJECT_ROOT}/cmd/meshctl/main.go" cluster register \
    --cluster-name "${cluster}" \
    --mgmt-context "kind-${mgmtCluster}" \
    --remote-context "kind-${cluster}" \
    --api-server-address "${apiServerAddress}" \
    --cert-agent-chart-file "${AGENT_CHART}"
}

function install_smh() {
  cluster=$1
  apiServerAddress=$(get_api_address ${cluster})

  ${PROJECT_ROOT}/ci/setup-smh.sh ${cluster} ${SMH_CHART} ${AGENT_CHART} ${AGENT_IMAGE} ${apiServerAddress}
}

#### START SCRIPT

# Note(ilackarms): these names are hard-coded in test/e2e/env.go
mgmtCluster=mgmt-cluster
remoteCluster=remote-cluster

### DEBUG FUNCS
function debug_proxy() {
  set -x
  deployment=$1
  namespace=$2
  cluster=$3
  port=$4
  kpf --context=kind-$cluster -n $namespace deployment/$deployment $port:15000 &
  sleep 2
  curl "localhost:${port}/logging?level=debug" -XPOST
  curl "localhost:${port}/logging?filter=trace" -XPOST
  k logs --context="kind-$cluster" -n "$namespace" "deployment/$deployment" -c istio-proxy -f
  set +x
}


function create_virtual_mesh() {
  cluster=$1
  K="kubectl --context=kind-${cluster}"
  ${K} apply -f - <<EOF
apiVersion: networking.smh.solo.io/v1alpha2
kind: VirtualMesh
metadata:
  name: virtual-mesh
  namespace: service-mesh-hub
spec:
  mtlsConfig:
    autoRestartPods: true
    limited:
      rootCertificateAuthority:
        generated: {}
  federation: {}
  meshes:
  - name: istiod-istio-system-mgmt-cluster
    namespace: service-mesh-hub
  - name: istiod-istio-system-remote-cluster
    namespace: service-mesh-hub
EOF
}
