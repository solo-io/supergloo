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
  kind get clusters | grep -E '(management-plane|target-cluster)-[a-z0-9]*' | while read -r r; do kind delete cluster --name $r; done
  exit 0
fi

make clean

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

# Create NodePort for remote cluster so it can be reachable from the management plane.
# This config is roughly based on: https://kind.sigs.k8s.io/docs/user/ingress/
cat <<EOF | kind create cluster --name $remoteCluster --image $kindImage --config=-
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

wait

printf "\n\n---\n"
echo "Finished setting up cluster $managementPlane"
echo "Finished setting up cluster $remoteCluster"

# set up kubectl to be pointing to the proper cluster
kubectl config use-context kind-$managementPlane

# ensure service-mesh-hub ns exists
kubectl --context kind-$managementPlane create ns service-mesh-hub
kubectl --context kind-$remoteCluster create ns service-mesh-hub

# leaving this in for the time being as there is a race with helm installing CRDs
# register all our CRDs in the management plane
kubectl --context kind-$managementPlane apply -f install/helm/charts/custom-resource-definitions/crds
# register all the CRDs in the target cluster too
kubectl --context kind-$remoteCluster apply -f install/helm/charts/custom-resource-definitions/crds

# Build the docker images
make -B docker

# Load images into the management plane cluster
export CLUSTER_NAME=$managementPlane
make kind-load-images

# Load images into the target cluster
export CLUSTER_NAME=$remoteCluster
make kind-load-images

# create Helm packages
make -s package-index-mgmt-plane-helm -B
make -s package-index-csr-agent-helm -B

# generate the meshctl binary
make meshctl -B
# install the app
# the helm version needs to strip the leading v out of the git describe output
if [ -z "$VERSION" ]; then
  helmVersion=$(git describe --tags --dirty | sed -E 's|^v(.*$)|\1|')
else
  helmVersion=$VERSION
fi

./_output/meshctl --context kind-$managementPlane install --file ./_output/helm/charts/management-plane/service-mesh-hub-$helmVersion.tgz

if [ -n "$DEBUG_MODE" ]; then
  kubectl --context kind-$managementPlane set env -n service-mesh-hub deploy/mesh-networking DEBUG_MODE=1
fi

case $(uname) in
  "Darwin")
  {
      CLUSTER_DOMAIN_MGMT=host.docker.internal
      CLUSTER_DOMAIN_REMOTE=host.docker.internal
  } ;;
  "Linux")
  {
      CLUSTER_DOMAIN_MGMT=$(docker exec $managementPlane-control-plane ip addr show dev eth0 | sed -nE 's|\s*inet\s+([0-9.]+).*|\1|p'):6443
      CLUSTER_DOMAIN_REMOTE=$(docker exec $remoteCluster-control-plane ip addr show dev eth0 | sed -nE 's|\s*inet\s+([0-9.]+).*|\1|p'):6443
  } ;;
  *)
  {
      echo "Unsupported OS"
      exit 1
  } ;;
esac

#register the remote cluster, and install Istio onto the management plane cluster
./_output/meshctl --context kind-$managementPlane cluster register \
  --remote-context kind-$managementPlane \
  --remote-cluster-name management-plane-cluster \
  --local-cluster-domain-override $CLUSTER_DOMAIN_MGMT \
  --dev-csr-agent-chart

#register the remote cluster, and install Istio onto the remote cluster
./_output/meshctl --context kind-$managementPlane cluster register \
  --remote-context kind-$remoteCluster \
  --remote-cluster-name target-cluster \
  --local-cluster-domain-override $CLUSTER_DOMAIN_REMOTE \
  --dev-csr-agent-chart

./_output/meshctl --context kind-$remoteCluster mesh install istio1.5 --operator-spec=- <<EOF
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
      pilotCertProvider: kubernetes
      controlPlaneSecurityEnabled: true
      mtls:
        enabled: true
      podDNSSearchNamespaces:
      - global
      - '{{ valueOrDefault .DeploymentMeta.Namespace "default" }}.global'
EOF

./_output/meshctl --context kind-$managementPlane mesh install istio1.5 --operator-spec=- <<EOF
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
              nodePort: 32001
  values:
    prometheus:
      enabled: false
    gateways:
      istio-ingressgateway:
        type: NodePort
        ports:
          - targetPort: 15443
            name: tls
            nodePort: 32001
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


echo '>>> Waiting for Meshes to be created'
retries=50
count=0
ok=false
until ${ok}; do
    numResources=$(kubectl --context kind-$managementPlane -n service-mesh-hub get meshes | grep istio -c || true)
    if [[ ${numResources} -eq 2 ]]; then
        ok=true
        continue
    fi
    sleep 5
    count=$(($count + 1))
    if [[ ${count} -eq ${retries} ]]; then
        echo "No more retries left"
        exit 1
    fi
done

echo '✔ Meshes have been created'

kubectl --context kind-$managementPlane -n istio-system rollout status deployment istiod
kubectl --context kind-$remoteCluster -n istio-system rollout status deployment istiod

kubectl --context kind-$managementPlane apply -f - <<EOF
apiVersion: networking.smh.solo.io/v1alpha1
kind: VirtualMesh
metadata:
  name: virtual-mesh
  namespace: service-mesh-hub
spec:
  meshes:
  - name: istio-istio-system-target-cluster
    namespace: service-mesh-hub
  - name: istio-istio-system-management-plane-cluster
    namespace: service-mesh-hub
  certificate_authority:
    builtin: {}
EOF

echo ">>> Waiting for cacerts in ${managementPlane}"
retries=50
count=0
ok=false
until ${ok}; do
    numResources=$(kubectl --context kind-$managementPlane -n istio-system get secrets | grep cacerts -c || true)
    if [[ ${numResources} -eq 1 ]]; then
        ok=true
        continue
    fi
    sleep 5
    count=$(($count + 1))
    if [[ ${count} -eq ${retries} ]]; then
        echo "No more retries left"
        exit 1
    fi
done

echo "✔ cacerts in $managementPlane have been created"

echo ">>> Waiting for cacerts in $remoteCluster"
retries=50
count=0
ok=false
until ${ok}; do
    numResources=$(kubectl --context kind-$remoteCluster -n istio-system get secrets | grep cacerts -c || true)
    if [[ ${numResources} -eq 1 ]]; then
        ok=true
        continue
    fi
    sleep 5
    count=$(($count + 1))
    if [[ ${count} -eq ${retries} ]]; then
        echo "No more retries left"
        exit 1
    fi
done

echo "✔ cacerts in $remoteCluster have been created"


kubectl --context kind-$managementPlane -n istio-system delete pod -l app=istiod
kubectl --context kind-$remoteCluster -n istio-system delete pod -l app=istiod

kubectl --context kind-$managementPlane -n istio-system rollout status deployment istiod
kubectl --context kind-$remoteCluster -n istio-system rollout status deployment istiod

# label bookinfo namespaces for injection
kubectl --context kind-$managementPlane label namespace default istio-injection=enabled
kubectl --context kind-$remoteCluster label namespace default istio-injection=enabled

# Apply bookinfo deployments and services
kubectl --context kind-$managementPlane apply -f ./ci/bookinfo.yaml -l 'app,version notin (v3)'
kubectl --context kind-$managementPlane apply -f ./ci/bookinfo.yaml -l 'account'

kubectl --context kind-$remoteCluster apply -f ./ci/bookinfo.yaml -l 'app,version in (v3)'
kubectl --context kind-$remoteCluster apply -f ./ci/bookinfo.yaml -l 'service=reviews'
kubectl --context kind-$remoteCluster apply -f ./ci/bookinfo.yaml -l 'account=reviews'
kubectl --context kind-$remoteCluster apply -f ./ci/bookinfo.yaml -l 'app=ratings'
kubectl --context kind-$remoteCluster apply -f ./ci/bookinfo.yaml -l 'account=ratings'

# wait for deployments to finish
kubectl --context kind-$managementPlane rollout status deployment/productpage-v1
kubectl --context kind-$managementPlane rollout status deployment/reviews-v1
kubectl --context kind-$managementPlane rollout status deployment/reviews-v2

kubectl --context kind-$remoteCluster rollout status deployment/reviews-v3

echo '>>> Waiting for MeshWorkloads to be created'
retries=50
count=0
ok=false
until ${ok}; do
    numResources=$(kubectl --context kind-$managementPlane -n service-mesh-hub get meshworkloads | grep istio -c || true)
    if [[ ${numResources} -eq 9 ]]; then
        ok=true
        continue
    fi
    sleep 5
    count=$(($count + 1))
    if [[ ${count} -eq ${retries} ]]; then
        echo "Have ${numResources} resources, and no more retries left"
        exit 1
    fi
done

echo '✔ MeshWorkloads have been created'

echo '>>> Waiting for MeshServices to be created'
retries=50
count=0
ok=false
until ${ok}; do
    numResources=$(kubectl --context kind-$managementPlane -n service-mesh-hub get meshservices | grep default -c || true)
    if [[ ${numResources} -eq 6 ]]; then
        ok=true
        continue
    fi
    sleep 5
    count=$(($count + 1))
    if [[ ${count} -eq ${retries} ]]; then
        echo "Have ${numResources} resources, and no more retries left"
        exit 1
    fi
done

echo '✔ MeshServices have been created'


# echo context to tests if they watch us
# dont change this line without changing StartEnv in test/e2e/env.go
if [ -e /proc/self/fd/3 ]; then
echo kind-$managementPlane kind-$remoteCluster >&3
fi
