#!/bin/bash -ex

#####################################
# Functions for setting up kind clusters
#####################################

#!/bin/bash

INSTALL_DIR=${PROJECT_ROOT}/install
AGENT_VALUES=${INSTALL_DIR}/helm/cert-agent/values.yaml
AGENT_IMAGE_REGISTRY=$(cat ${AGENT_VALUES} | grep "registry: " | awk '{print $2}')
AGENT_IMAGE_REPOSITORY=$(cat ${AGENT_VALUES} | grep "repository: " | awk '{print $2}')
AGENT_IMAGE_TAG=$(cat ${AGENT_VALUES} | grep "tag: " | awk '{print $2}')

AGENT_IMAGE=${AGENT_IMAGE_REGISTRY}/${AGENT_IMAGE_REPOSITORY}:${AGENT_IMAGE_TAG}
AGENT_CHART=${INSTALL_DIR}/helm/_output/charts/cert-agent/cert-agent-${AGENT_IMAGE_TAG}.tgz

SMH_VALUES=${INSTALL_DIR}/helm/service-mesh-hub/values.yaml
SMH_IMAGE_TAG=$(cat ${SMH_VALUES} | grep -m 1 "tag: " | awk '{print $2}')
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

function install_istio() {
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

  # install (modified) bookinfo
  ${K} create namespace bookinfo
  ${K} label ns bookinfo istio-injection=enabled --overwrite
  ${K} apply -n bookinfo -f ${PROJECT_ROOT}/ci/bookinfo.yaml

  # NOTE: we delete the details service to free up CPU for ci
  ${K} delete -n bookinfo deployment details-v1

  ROLLOUT="${K} -n bookinfo rollout status deployment --timeout 300s"

  ${ROLLOUT} ratings-v1
  ${ROLLOUT} productpage-v1
  ${ROLLOUT} reviews-v1
  ${ROLLOUT} reviews-v2
  ${ROLLOUT} reviews-v3

  printf "\n\n---\n"
  echo "Finished setting up cluster ${cluster}"

}

function install_osm() {
  cluster=$1
  port=$2
  K="kubectl --context=kind-${cluster}"

  echo "installing osm to ${cluster}..."

  # install in permissive mode for testing
  osm install --enable-permissive-traffic-policy

  for i in bookstore bookbuyer bookthief bookwarehouse; do kubectl create ns $i; done

  for i in bookstore bookbuyer bookthief bookwarehouse; do osm namespace add $i; done

  ${K} apply -f https://raw.githubusercontent.com/openservicemesh/osm/main/docs/example/manifests/apps/bookbuyer.yaml
  ${K} apply -f https://raw.githubusercontent.com/openservicemesh/osm/main/docs/example/manifests/apps/bookstore-v1.yaml
  ${K} apply -f https://raw.githubusercontent.com/openservicemesh/osm/main/docs/example/manifests/apps/bookthief.yaml
  ${K} apply -f https://raw.githubusercontent.com/openservicemesh/osm/main/docs/example/manifests/apps/bookwarehouse.yaml
  ${K} apply -f https://raw.githubusercontent.com/openservicemesh/osm/main/docs/example/manifests/access/traffic-access.yaml
  ${K} apply -f https://raw.githubusercontent.com/openservicemesh/osm/main/docs/example/manifests/bookstore-v2/bookstore-v2.yaml
  ${K} apply -f https://raw.githubusercontent.com/openservicemesh/osm/main/docs/example/manifests/bookstore-v2/traffic-access-v2.yaml
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
  K="kubectl --context=kind-${cluster}"

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

  INSTALL_DIR="${PROJECT_ROOT}/install/"
  AGENT_VALUES=${INSTALL_DIR}/helm/cert-agent/values.yaml
  AGENT_IMAGE_REGISTRY=$(cat ${AGENT_VALUES} | grep "registry: " | awk '{print $2}')
  AGENT_IMAGE_REPOSITORY=$(cat ${AGENT_VALUES} | grep "repository: " | awk '{print $2}')
  AGENT_IMAGE_TAG=$(cat ${AGENT_VALUES} | grep "tag: " | awk '{print $2}')

  AGENT_IMAGE="${AGENT_IMAGE_REGISTRY}/${AGENT_IMAGE_REPOSITORY}:${AGENT_IMAGE_TAG}"
  AGENT_CHART="${INSTALL_DIR}/helm/_output/charts/cert-agent-${AGENT_IMAGE_TAG}.tgz"

  # load cert-agent image
  kind load docker-image --name "${cluster}" "${AGENT_IMAGE}"

  go run "${PROJECT_ROOT}/cmd/meshctl/main.go" cluster register \
    --cluster-name "${cluster}" \
    --master-context "kind-${masterCluster}" \
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
masterCluster=master-cluster
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
