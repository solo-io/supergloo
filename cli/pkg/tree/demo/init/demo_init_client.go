package demo_init

import (
	"context"

	"github.com/solo-io/mesh-projects/cli/pkg/common/exec"
)

func DemoInit(ctx context.Context, runner exec.Runner) error {
	return runner.Run("bash", "-c", initDemoScript)
}

const (
	initDemoScript = `

# generate 16-character random suffix on these names
managementPlane=management-plane-$(xxd -l16 -ps /dev/urandom)
remoteCluster=remote-cluster-$(xxd -l16 -ps /dev/urandom)

# set up each cluster
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

meshctl install

#register the remote cluster, and install Istio onto the management plane cluster
meshctl cluster register \
  --remote-context kind-$managementPlane \
  --remote-cluster-name management-plane \
  --local-cluster-domain-override host.docker.internal

#register the remote cluster, and install Istio onto the remote cluster
meshctl cluster register \
  --remote-context kind-$remoteCluster \
  --remote-cluster-name remote-cluster \
  --local-cluster-domain-override host.docker.internal

meshctl istio install --context kind-$remoteCluster --operator-spec=- <<EOF
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: example-istiooperator
  namespace: istio-operator
spec:
  profile: minimal
  components:
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

meshctl istio install --context kind-$managementPlane --operator-spec=- <<EOF
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: example-istiooperator
  namespace: istio-operator
spec:
  profile: minimal
  components:
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
printf "\nWhen you are finished simply run:\nmeshctl demo cleanup\nor\nkind delete cluster $managementPlane && kind delete cluster $remoteCluster"
printf "\n\nHvae fun playing around with Service Mesh Hub. Let us know what you think 😊"
`
)
