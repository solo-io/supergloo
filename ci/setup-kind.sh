#!/bin/bash -ex

#####################################
#
# Set up two kind clusters:
#   1. a master cluster in which Service Mesh Hub is installed.
#   2. a remote cluster in which only Istio and the bookinfo app are installed.
#
# The master cluster will have the appropriate secret for communicating with the remote cluster
# Your kube context will be left pointing to the master cluster
# Each cluster will have Istio set up in the istio-system namespace in its minimal profile
#
# To clean up your kind clusters, run this script as: `bash ci/setup-kind.sh cleanup`. Use this if you notice the docker VM running out of disk space (for images).
#
#####################################

# clean up background processes on exit
trap "kill 0" EXIT

PROJECT_ROOT=$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )/..
echo "Using project root ${PROJECT_ROOT}"

# print debug info on error
trap 'catch' ERR
catch() {
  echo cluster setup failed! printing debug info and exiting...
  ${PROJECT_ROOT}/ci/print-kind-info.sh
  exit 1
}

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
}

function install_istio() {
  cluster=$1
  port=$2

  echo "installing istio to ${cluster}..."

  cat << EOF | istioctl manifest apply --context "kind-${cluster}" -f -
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: example-istiooperator
  namespace: istio-system
spec:
  profile: minimal
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
      - '{{ valueOrDefault .DeploymentMeta.Namespace "default" }}.global'
EOF

  # install bookinfo
  ${K} create namespace bookinfo
  ${K} label ns bookinfo istio-injection=enabled --overwrite
  ${K} apply -n bookinfo -f https://raw.githubusercontent.com/istio/istio/master/samples/bookinfo/platform/kube/bookinfo.yaml

  ${K} -n bookinfo rollout status deployment --timeout 150s details-v1
  ${K} -n bookinfo rollout status deployment --timeout 150s productpage-v1
  ${K} -n bookinfo rollout status deployment --timeout 150s ratings-v1
  ${K} -n bookinfo rollout status deployment --timeout 150s reviews-v1
  ${K} -n bookinfo rollout status deployment --timeout 150s reviews-v2
  ${K} -n bookinfo rollout status deployment --timeout 150s reviews-v3

  printf "\n\n---\n"
  echo "Finished setting up cluster ${cluster}"

}

function setup_kind_cluster() {
  # The default version of k8s under Linux is 1.18
  # https://github.com/solo-io/service-mesh-hub/issues/700
  kindImage=kindest/node:v1.17.5

  cluster=$1
  port=$2

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

  echo "installing istio to ${cluster}..."

  cat << EOF | istioctl manifest apply --context "kind-${cluster}" -f -
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: example-istiooperator
  namespace: istio-system
spec:
  profile: minimal
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
      - '{{ valueOrDefault .DeploymentMeta.Namespace "default" }}.global'
EOF

  # install bookinfo
  ${K} create namespace bookinfo
  ${K} label ns bookinfo istio-injection=enabled --overwrite
  ${K} apply -n bookinfo -f https://raw.githubusercontent.com/istio/istio/master/samples/bookinfo/platform/kube/bookinfo.yaml

  ${K} -n bookinfo rollout status deployment --timeout 150s details-v1
  ${K} -n bookinfo rollout status deployment --timeout 150s productpage-v1
  ${K} -n bookinfo rollout status deployment --timeout 150s ratings-v1
  ${K} -n bookinfo rollout status deployment --timeout 150s reviews-v1
  ${K} -n bookinfo rollout status deployment --timeout 150s reviews-v2
  ${K} -n bookinfo rollout status deployment --timeout 150s reviews-v3

  printf "\n\n---\n"
  echo "Finished setting up cluster ${cluster}"

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

  ${K} create ns service-mesh-hub || echo exists

  go run "${PROJECT_ROOT}/cmd/meshctl/main.go" cluster register \
    --cluster-name "${cluster}" \
    --master-context "kind-${masterCluster}" \
    --remote-context "kind-${cluster}" \
    --api-server-address "${apiServerAddress}"
}

# Note(ilackarms): these names are hard-coded in test/e2e/env.go
masterCluster=master-cluster
remoteCluster=remote-cluster

if [ "$1" == "cleanup" ]; then
  kind get clusters | grep -E "${masterCluster}|${remoteCluster}" | while read -r r; do kind delete cluster --name "${r}"; done
  exit 0
fi

# NOTE(ilackarms): we run the setup_kind clusters sequentially due to this bug:
# related: https://github.com/kubernetes-sigs/kind/issues/1596
setup_kind_cluster ${masterCluster} 32001
install_istio ${masterCluster} 32001 &
pids[0]=$!

setup_kind_cluster ${remoteCluster} 32000
install_istio ${remoteCluster} 32000 &
pids[1]=$!


# wait for all bg processes
for pid in ${pids[*]}; do
    wait $pid || echo failed; exit 1
done

echo setup successfully set up clusters.

# install service mesh hub
${PROJECT_ROOT}/ci/setup-smh.sh ${masterCluster}

# sleep to allow crds to register
sleep 4

# register clusters
register_cluster ${masterCluster} &
pids[0]=$!
register_cluster ${remoteCluster} &
pids[1]=$!

# wait for all bg processes
for pid in ${pids[*]}; do
    wait $pid || echo failed; exit 1
done

# set current context to master cluster
kubectl config use-context ${masterCluster}