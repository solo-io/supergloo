#!/bin/bash 

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
  kind get clusters | while read -r r; do kind delete cluster --name "$r"; done
  exit 0
fi

make clean

# generate 16-character random suffix on these names
managementPlane=management-plane-$(xxd -l16 -ps /dev/urandom)
remoteCluster=target-cluster-$(xxd -l16 -ps /dev/urandom)

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
ls install/helm/charts/custom-resource-definitions/crds | while read f; do kubectl --context kind-$managementPlane apply -f install/helm/charts/custom-resource-definitions/crds/$f; done
# register all the CRDs in the target cluster too
ls install/helm/charts/custom-resource-definitions/crds | while read f; do kubectl --context kind-$remoteCluster apply -f install/helm/charts/custom-resource-definitions/crds/$f; done


# make all the docker images
# write the output to a temp file so that we can grab the image names out of it
# also ensure we clean up the file once we're done
tempFile=/tmp/images
make docker -B | tee $tempFile

function cleanup {
  rm $tempFile
}

trap cleanup EXIT

# grab the image names out of the `make docker` output
# the kind cluster name is just $managementPlane, not kind-$managementPlane; the latter is how kubectl is aware of it
sed -nE 's|Successfully tagged (.*$)|\1|p' $tempFile | while read f; do kind load docker-image --name $managementPlane $f; kind load docker-image --name $remoteCluster $f; done

# create Helm packages
make -s package-index-mgmt-plane-helm -B
make -s package-index-csr-agent-helm -B

# generate the meshctl binary
make meshctl -B
# install the app
# the helm version needs to strip the leading v out of the git describe output
helmVersion=$(git describe --tags --dirty | sed -E 's|^v(.*$)|\1|')
./_output/meshctl install --file ./_output/helm/charts/management-plane/service-mesh-hub-$helmVersion.tgz

#register the remote cluster, and install Istio onto the management plane cluster
./_output/meshctl cluster register \
  --remote-context kind-$managementPlane \
  --remote-cluster-name management-plane-cluster \
  --local-cluster-domain-override host.docker.internal \
  --dev-csr-agent-chart

#register the remote cluster, and install Istio onto the remote cluster
./_output/meshctl cluster register \
  --remote-context kind-$remoteCluster \
  --remote-cluster-name target-cluster \
  --local-cluster-domain-override host.docker.internal \
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


echo '>>> Waiting for Meshes to be created'
retries=50
count=0
ok=false
until ${ok}; do
    numResources=`kubectl --context kind-$managementPlane -n service-mesh-hub get meshes | grep istio -c`
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
apiVersion: networking.zephyr.solo.io/v1alpha1
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
    numResources=`kubectl --context kind-$managementPlane -n istio-system get secrets | grep cacerts -c`
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
    numResources=`kubectl --context kind-$remoteCluster -n istio-system get secrets | grep cacerts -c`
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
kubectl apply --context kind-$managementPlane -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml
kubectl apply --context kind-$remoteCluster -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/bookinfo/platform/kube/bookinfo.yaml

echo '>>> Waiting for MeshWorkloads to be created'
retries=50
count=0
ok=false
until ${ok}; do
    numResources=`kubectl --context kind-$managementPlane -n service-mesh-hub get meshworkloads | grep istio -c`
    if [[ ${numResources} -eq 14 ]]; then
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

echo '✔ MeshWorkloads have been created'

echo '>>> Waiting for MeshServices to be created'
retries=50
count=0
ok=false
until ${ok}; do
    numResources=`kubectl --context kind-$managementPlane -n service-mesh-hub get meshservices | grep default -c`
    if [[ ${numResources} -eq 8 ]]; then
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

echo '✔ MeshServices have been created'
