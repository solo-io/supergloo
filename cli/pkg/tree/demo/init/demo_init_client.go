package demo_init

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/solo-io/service-mesh-hub/cli/pkg/common/exec"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
func DemoInit(ctx context.Context, runner exec.Runner) error {
	return runner.Run("bash", "-c", initDemoScript, "init-demo.sh", fmt.Sprintf("%d", rand.Uint32()))
}

const (
	initDemoScript = `
set -e

# generate 1 allows to make several envs in parallel
managementPlane=management-plane-$1
remoteCluster=target-cluster-$1

# set up each cluster
# Create NodePort for remote cluster so it can be reachable from the management plane.
# This config is roughly based on: https://kind.sigs.k8s.io/docs/user/ingress/
cat <<EOF | kind create cluster --name $managementPlane --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
        authorization-mode: "AlwaysAllow"
  extraPortMappings:
  - containerPort: 32001
    hostPort: 32001
    protocol: TCP
EOF
# Create NodePort for remote cluster so it can be reachable from the management plane.
# This config is roughly based on: https://kind.sigs.k8s.io/docs/user/ingress/
cat <<EOF | kind create cluster --name $remoteCluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
        authorization-mode: "AlwaysAllow"
  extraPortMappings:
  - containerPort: 32000
    hostPort: 32000
    protocol: TCP
EOF

printf "\n\n---\n"
echo "Finished setting up cluster $managementPlane"
echo "Finished setting up cluster $remoteCluster"

# set up kubectl to be pointing to the proper cluster
kubectl config use-context kind-$managementPlane

# ensure service-mesh-hub ns exists
kubectl create ns --context kind-$managementPlane  service-mesh-hub
kubectl create ns --context kind-$remoteCluster  service-mesh-hub

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
helmVersion=$(git describe --tags --dirty | sed -E 's|^v(.*$)|\1|')
./_output/meshctl install --file ./_output/helm/charts/management-plane/service-mesh-hub-$helmVersion.tgz

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
./_output/meshctl cluster register \
  --remote-context kind-$managementPlane \
  --remote-cluster-name management-plane-cluster \
  --local-cluster-domain-override $CLUSTER_DOMAIN_MGMT \
  --dev-csr-agent-chart

#register the remote cluster, and install Istio onto the remote cluster
./_output/meshctl cluster register \
  --remote-context kind-$remoteCluster \
  --remote-cluster-name target-cluster \
  --local-cluster-domain-override $CLUSTER_DOMAIN_REMOTE \
  --dev-csr-agent-chart

./_output/meshctl mesh install istio --context kind-$remoteCluster --operator-spec=- <<EOF
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: example-istiooperator
  namespace: istio-operator
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
    security:
      selfSigned: false
EOF

./_output/meshctl mesh install istio --context kind-$managementPlane --operator-spec=- <<EOF
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: example-istiooperator
  namespace: istio-operator
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
    security:
      selfSigned: false
EOF

printf "\n\n---\n"
printf "Your demo clusters are now ready to go!"
printf "\nThe management plane cluster can be accessed via: kubectl config use-context kind-$managementPlane"
printf "\nThe remote cluster can be accessed via: kubectl config use-context kind-$remoteCluster"
printf "\nIf you plan are using this demo along with an official Service Mesh Hub guide the following may come in very handy:"
printf "\n\nexport MGMT_PLANE_CTX=kind-$managementPlane\nexport REMOTE_CTX=kind-$remoteCluster\n\n"
printf "\nWhen you are finished simply run:\nmeshctl demo cleanup\nor\nkind delete cluster --name $managementPlane && kind delete cluster --name $remoteCluster"
printf "\n\nHave fun playing around with Service Mesh Hub. Let us know what you think 😊"
`
)
