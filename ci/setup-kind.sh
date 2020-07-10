#!/bin/bash -ex


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
  kind get clusters | grep -E '(management-plane|target-cluster)-[a-z0-9]*' | while read -r r; do kind delete cluster --name $r; done
  exit 0
fi

# allow to make several envs in parallel
managementPlane=management-plane-$1
remoteCluster=target-cluster-$1

# The default version of k8s under Linux is 1.18
# https://github.com/solo-io/service-mesh-hub/issues/700
kindImage=kindest/node:v1.17.5

# set up each cluster
# Create NodePort for remote cluster so it can be reachable from the management plane.
# This config is roughly based on: https://kind.sigs.k8s.io/docs/user/ingress/
(cat <<EOF | kind create cluster --name $managementPlane --image $kindImage --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 32001
    hostPort: 32001
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
)&

##TODO:uncomment
## Create NodePort for remote cluster so it can be reachable from the management plane.
## This config is roughly based on: https://kind.sigs.k8s.io/docs/user/ingress/
#cat <<EOF | kind create cluster --name $remoteCluster --image $kindImage --config=-
#kind: Cluster
#apiVersion: kind.x-k8s.io/v1alpha4
#nodes:
#- role: control-plane
#  extraPortMappings:
#  - containerPort: 32000
#    hostPort: 32000
#    protocol: TCP
#  kubeadmConfigPatches:
#  - |
#    kind: InitConfiguration
#    nodeRegistration:
#      kubeletExtraArgs:
#        node-labels: "ingress-ready=true"
#kubeadmConfigPatches:
#- |
#  kind: InitConfiguration
#  nodeRegistration:
#    kubeletExtraArgs:
#      authorization-mode: "AlwaysAllow"
#      feature-gates: "EphemeralContainers=true"
#- |
#  kind: KubeletConfiguration
#  featureGates:
#    EphemeralContainers: true
#- |
#  kind: KubeProxyConfiguration
#  featureGates:
#    EphemeralContainers: true
#- |
#  kind: ClusterConfiguration
#  metadata:
#    name: config
#  apiServer:
#    extraArgs:
#      "feature-gates": "EphemeralContainers=true"
#  scheduler:
#    extraArgs:
#      "feature-gates": "EphemeralContainers=true"
#  controllerManager:
#    extraArgs:
#      "feature-gates": "EphemeralContainers=true"
#EOF

wait

printf "\n\n---\n"
echo "Finished setting up cluster $managementPlane"
##TODO:uncomment echo "Finished setting up cluster $remoteCluster"

# set up kubectl to be pointing to the proper cluster
kubectl config use-context kind-$managementPlane

# install istio
# istiod + ingressgateway only

values=istio-helm-values.yaml
cat > "${values}" << EOF
# istio 1.6 values
spec:
  meshConfig:
    accessLogFile: "/dev/stdout"
  components:
    ingressGateways:
    - name: istio-ingressgateway
      enabled: true
      k8s:
        service:
          type: NodePort
          ports:
            - port: 80
              targetPort: 80
              name: http2
              nodePort: 32000
EOF

istioctl manifest apply --set profile=demo -f "${values}"
rm "${values}"

cat << EOF | kubectl apply -f-
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  labels:
    release: istio
  name: istio-ingressgateway
  namespace: istio-system
spec:
  selector:
    app: istio-ingressgateway
    istio: ingressgateway
  servers:
  - hosts:
    - '*'
    port:
      name: http
      number: 80
      protocol: HTTP
EOF

# bookinfo
kubectl create namespace bookinfo
kubectl label ns bookinfo istio-injection=enabled --overwrite
kubectl apply -n bookinfo -f https://raw.githubusercontent.com/istio/istio/master/samples/bookinfo/platform/kube/bookinfo.yaml

kubectl -n bookinfo rollout status deployment details-v1
kubectl -n bookinfo rollout status deployment productpage-v1
kubectl -n bookinfo rollout status deployment ratings-v1
kubectl -n bookinfo rollout status deployment reviews-v1
kubectl -n bookinfo rollout status deployment reviews-v2
kubectl -n bookinfo rollout status deployment reviews-v3

echo setup successfully set up smh prereqs

# echo context to tests if they watch us
# dont change this line without changing StartEnv in test/e2e/env.go
if [ -e /proc/self/fd/3 ]; then
  echo kind-$managementPlane >&3
fi


# dev-portal
"$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )/setup-smh.sh" "${ARG}"
