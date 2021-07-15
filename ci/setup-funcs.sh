#!/bin/bash -ex

#####################################
# Functions for setting up kind clusters
#####################################

PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )/.."

#### FUNCTIONS

function create_kind_cluster() {
  # The default version of k8s under Linux is 1.18
  # https://github.com/solo-io/gloo-mesh/issues/700
  kindImage=kindest/node:v1.17.5

  # gloo mesh cluster name
  cluster=$1

  # ingress ports
  port1=$2
  port2=$3

  # Set the network suffix based on the ingress port.
  # The ingress port params are either 32000, or 32001, so this will either be 1 or 2.
  # This number will the be used to construct the subnet.
  # For example: 10.96.${net}.0/24
  # This value will be used to cordon off, and later join the different pod subnets of the multiple clusters.
  ((net=$port1%32000+1))

  echo "creating cluster ${cluster} with ingress port ${port1}"

  K="kubectl --context=kind-${cluster}"

  # When running multi cluster kind with flat-networking, kind must be configured with a custom CNI.
  # https://kind.sigs.k8s.io/docs/user/configuration/#disable-default-cni
  # This allows us to use our own CNI, namely calico.
  disableDefaultCNI=false
  if [ ! -z ${FLAT_NETWORKING_ENABLED} ]; then
    disableDefaultCNI=true
  fi

  # This config is roughly based on: https://kind.sigs.k8s.io/docs/user/ingress/
  cat <<EOF | kind create cluster --name "${cluster}" --image $kindImage --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  serviceSubnet: "10.96.${net}.0/24"
  podSubnet: "192.168.2${net}.0/24"
  disableDefaultCNI: ${disableDefaultCNI}
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 6443
    hostPort: ${net}000
  - containerPort: ${port1}
    hostPort: ${port1}
    protocol: TCP
  - containerPort: ${port2}
    hostPort: ${port2}
    protocol: TCP
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        # Populate nodes with region/zone info, which are used by VirtualDestination locality-based failover (Enterprise-only)
        node-labels: "ingress-ready=true,topology.kubernetes.io/region=us-east-1,topology.kubernetes.io/zone=us-east-1c"
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

  # Only setup kind clusters with flat networking if ENV var is set
  if [ ! -z ${FLAT_NETWORKING_ENABLED} ]; then
    # Apply calico networking CNI
    ${K} apply -f https://docs.projectcalico.org/v3.15/manifests/calico.yaml
    ${K} -n kube-system set env daemonset/calico-node FELIX_IGNORELOOSERPF=true

    # Ensure calico node is ready before installing istio
    ${K} -n kube-system rollout status daemonset/calico-node --timeout=300s

    # Install metallb to each cluster. This is a bit of a "hack" to enable LoadBalancers in kind so that
    # the 2 clusters can directly communicate with each other.
    # For more info on metallb see: https://metallb.org/
    ${K} apply -f https://raw.githubusercontent.com/metallb/metallb/v0.9.3/manifests/namespace.yaml
    ${K} apply -f https://raw.githubusercontent.com/metallb/metallb/v0.9.3/manifests/metallb.yaml
    # Setup required memberlist config, see: https://github.com/hashicorp/memberlist
    ${K} -n metallb-system create secret generic memberlist --from-literal=secretkey="$(openssl rand -base64 128)"


    if hostname -i; then
      myip=$(hostname -i)
    else
      myip=$(ipconfig getifaddr en0)
    fi
    ipkind=$(docker inspect ${cluster}-control-plane | jq -r '.[0].NetworkSettings.Networks[].IPAddress')
    networkkind=$(echo ${ipkind} | sed 's/.$//')

    # Create the metallb address pool based on the cluster subnet
    cat << EOF | ${K} apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: metallb-system
  name: config
data:
  config: |
    address-pools:
    - name: default
      protocol: layer2
      addresses:
      - ${networkkind}2${net}0-${networkkind}2${net}9
EOF

  fi

}

function setup_flat_networking() {
  mgmt_cluster=$1
  mgmt_port=$2
  remote_cluster=$3
  remote_port=$4
  K_mgmt="kubectl --context=kind-${mgmt_cluster}"
  K_remote="kubectl --context=kind-${remote_cluster}"

  # Retrieve the subnets as defined by the setup_kind_cluster func above
  ((mgmt_net=$mgmt_port%32000+1))
  ((remote_net=$remote_port%32000+1))

  mgmt_cluster_ip=$(docker inspect ${mgmt_cluster}-control-plane | jq -r '.[0].NetworkSettings.Networks.kind.IPAddress')
  remote_cluster_ip=$(docker inspect ${remote_cluster}-control-plane | jq -r '.[0].NetworkSettings.Networks.kind.IPAddress')

  tmp=$(mktemp -d /tmp/gloo_mesh.XXXXXX)

  # Configuration for calico bird. More info can be found here: https://github.com/BIRD/bird
  # Example config can be found here: https://github.com/BIRD/bird/blob/master/doc/bird.conf.example
  # Bird functions as the network "bridge" between the 2 local kind clusters, allowing the pods to communicate directly
  # with each other
  cat << EOF > ${tmp}/bird.conf
log syslog { debug, trace, info, remote, warning, error, auth, fatal, bug };
log stderr all;
router id 172.18.0.100;
filter import_kernel {
  if ( net != 0.0.0.0/0 ) then {
  accept;
  }
reject;
}
debug protocols all;
protocol device {
  scan time 2;
}
protocol bgp mgmt_cluster_ip {
  description "${mgmt_cluster_ip}";
  local as 64513;
  neighbor ${mgmt_cluster_ip} port 31179 as 64513;
  multihop;
  rr client;
  graceful restart;
  import all;
  export all;
}
protocol bgp remote_cluster_ip {
  description "${remote_cluster_ip}";
  local as 64514;
  neighbor ${remote_cluster_ip} port 32179 as 64514;
  multihop;
  rr client;
  graceful restart;
  import all;
  export all;
}
EOF
  docker run --rm -d --name bird -p 179:179 -v ${tmp}/bird.conf:/etc/bird/bird.conf --entrypoint='' --network=kind pierky/bird bird -c /etc/bird/bird.conf -d
  bird=$(docker inspect bird | jq -r '.[0].NetworkSettings.Networks.kind.IPAddress')

  # The following 4 steps are done on both the management, and remote clusters to enable direct communication.
  # In order to accomplish this they setup the routing protocol between them known as BGP, so that they are "discoverable"
  # 1. A port is added to the calico-node
  # 2. A BGP peer is added, which is the "discoverable" network node: https://docs.projectcalico.org/reference/resources/bgppeer
  # 3. A BGP configuration is created with the available IPs for the cidr: https://docs.projectcalico.org/reference/resources/bgpconfig
  # 4. An IPPool is added for the other clusters pod cidr: https://docs.projectcalico.org/reference/resources/ippool
  ${K_mgmt}  -n kube-system patch ds calico-node --type=json -p='[{"op": "add", "path": "/spec/template/spec/containers/0/ports", "value": [{"containerPort": 179}]}]'
  ${K_mgmt} apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: bgp
  namespace: kube-system
spec:
  selector:
    k8s-app: calico-node
  ports:
    - port: 179
      targetPort: 179
      nodePort: 31179
  type: NodePort
EOF
  # calicoctl doesn't support kube context as a param, so need to manually switch the context before using it.
  kubectl config use-context kind-${mgmt_cluster}
  # calicoctl needs to exist before using
  calicoctl apply -f - <<EOF
kind: BGPPeer
apiVersion: projectcalico.org/v3
metadata:
  name: peer-to-rrs
spec:
  peerIP: ${bird}
  asNumber: 64513
EOF

  calicoctl apply -f - <<EOF
 apiVersion: projectcalico.org/v3
 kind: BGPConfiguration
 metadata:
   name: default
 spec:
   asNumber: 64513
   nodeToNodeMeshEnabled: true
   serviceClusterIPs:
   - cidr: 10.96.${mgmt_net}.0/24
EOF

  cat << EOF | calicoctl apply -f -
apiVersion: projectcalico.org/v3
kind: IPPool
metadata:
  name: remote-cluster-ip-pool
spec:
  cidr: 10.96.${remote_net}.0/24
  disabled: true
EOF

  ${K_remote} -n kube-system patch ds calico-node --type=json -p='[{"op": "add", "path": "/spec/template/spec/containers/0/ports", "value": [{"containerPort": 179}]}]'
  ${K_remote} apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: bgp
  namespace: kube-system
spec:
  selector:
    k8s-app: calico-node
  ports:
    - port: 179
      targetPort: 179
      nodePort: 32179
  type: NodePort
EOF
  kubectl config use-context kind-${remote_cluster}
  calicoctl apply -f - <<EOF
kind: BGPPeer
apiVersion: projectcalico.org/v3
metadata:
  name: peer-to-rrs
spec:
  peerIP: ${bird}
  asNumber: 64514
EOF
#  calicoctl delete BGPConfiguration default
  calicoctl apply -f - <<EOF
 apiVersion: projectcalico.org/v3
 kind: BGPConfiguration
 metadata:
   name: default
 spec:
   asNumber: 64514
   nodeToNodeMeshEnabled: true
   serviceClusterIPs:
   - cidr: 10.96.${remote_net}.0/24
EOF
  cat << EOF | calicoctl create -f -
apiVersion: projectcalico.org/v3
kind: IPPool
metadata:
  name: mgmt-cluster-ip-pool
spec:
  cidr: 10.96.${mgmt_net}.0/24
  disabled: true
EOF


}

# Operator spec for istio 1.7.x
function install_istio_1_7() {
  cluster=$1
  eastWestIngressPort=$2

  K="kubectl --context=kind-${cluster}"

  echo "installing istio to ${cluster}..."

  cat << EOF | istioctl install --context "kind-${cluster}" -y -f -
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: example-istiooperator
  namespace: istio-system
spec:
  hub: gcr.io/istio-release
  profile: minimal
  addonComponents:
    istiocoredns:
      enabled: true
  components:
    # Istio Gateway feature
    ingressGateways:
    - name: istio-ingressgateway
      enabled: true
      label:
        traffic: east-west
      k8s:
        env:
          # needed for Gateway TLS AUTO_PASSTHROUGH mode, reference: https://istio.io/latest/docs/reference/config/networking/gateway/#ServerTLSSettings-TLSmode
          - name: ISTIO_META_ROUTER_MODE
            value: "sni-dnat"
        service:
          type: NodePort
          selector:
            app: istio-ingressgateway
            istio: ingressgateway
            traffic: east-west
          ports:
            # in the future we may want to use this port for east west traffic in limited trust
            - port: 443
              targetPort: 8443
              name: https
            - port: 15443
              targetPort: 15443
              name: tls
              nodePort: ${eastWestIngressPort}
  meshConfig:
    enableAutoMtls: true
    defaultConfig:
      envoyAccessLogService:
        address: enterprise-agent.gloo-mesh:9977
      envoyMetricsService:
        address: enterprise-agent.gloo-mesh:9977
      proxyMetadata:
        # annotate Gloo Mesh cluster name for envoy requests (i.e. access logs, metrics)
        GLOO_MESH_CLUSTER_NAME: ${cluster}
  values:
    prometheus:
      enabled: false
    global:
      pilotCertProvider: istiod
      controlPlaneSecurityEnabled: true
      podDNSSearchNamespaces:
      - global
      # needed for annotating istio metrics with cluster
      multiCluster:
        clusterName: ${cluster}
EOF
}

# Operator spec for istio 1.8.x, 1.9.x, and 1.10x
function install_istio_1_8() {
  cluster=$1
  eastWestIngressPort=$2

  K="kubectl --context=kind-${cluster}"

  echo "installing istio to ${cluster}..."

  cat << EOF | istioctl manifest install -y --context "kind-${cluster}" -f -
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: example-istiooperator
  namespace: istio-system
spec:
  hub: gcr.io/istio-release
  profile: preview
  meshConfig:
    enableAutoMtls: true
    defaultConfig:
      envoyAccessLogService:
        address: enterprise-agent.gloo-mesh:9977
      envoyMetricsService:
        address: enterprise-agent.gloo-mesh:9977
      proxyMetadata:
        # Enable Istio agent to handle DNS requests for known hosts
        # Unknown hosts will automatically be resolved using upstream dns servers in resolv.conf
        ISTIO_META_DNS_CAPTURE: "true"
        # annotate Gloo Mesh cluster name for envoy requests (i.e. access logs, metrics)
        GLOO_MESH_CLUSTER_NAME: ${cluster}
      proxyStatsMatcher:
        inclusionPrefixes:
        - "http"
  components:
    # Istio Gateway feature
    ingressGateways:
    - name: istio-ingressgateway
      enabled: true
      label:
        traffic: east-west
      k8s:
        env:
          # needed for Gateway TLS AUTO_PASSTHROUGH mode, reference: https://istio.io/latest/docs/reference/config/networking/gateway/#ServerTLSSettings-TLSmode
          - name: ISTIO_META_ROUTER_MODE
            value: "sni-dnat"
        service:
          type: NodePort
          selector:
            app: istio-ingressgateway
            istio: ingressgateway
            traffic: east-west
          ports:
            # in the future we may want to use this port for east west traffic in limited trust
            - port: 443
              targetPort: 8443
              name: https
            - port: 15443
              targetPort: 15443
              name: tls
              nodePort: ${eastWestIngressPort}
  values:
    global:
      pilotCertProvider: istiod
      # needed for annotating istio metrics with cluster
      multiCluster:
        clusterName: ${cluster}
EOF
}

# updates the kube-system/coredns configmap in order to resolve hostnames with a ".global" suffix, needed for istio < 1.8
function install_istio_coredns() {

  cluster=$1
  port=$2
  K="kubectl --context=kind-${cluster}"
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
}

function install_istio() {
  cluster=$1
  eastWestIngressPort=$2
  K="kubectl --context=kind-${cluster}"

  if istioctl version | grep -E -- '1.7'
  then
    install_istio_1_7 $cluster $eastWestIngressPort
    install_istio_coredns $cluster $eastWestIngressPort
  elif istioctl version | grep -E -- '1.8'
  then
    install_istio_1_8 $cluster $eastWestIngressPort
  elif istioctl version | grep -E -- '1.9'
  then
    install_istio_1_8 $cluster $eastWestIngressPort
  elif istioctl version | grep -E -- '1.10'
  then
    install_istio_1_8 $cluster $eastWestIngressPort
  else
    echo "Encountered unsupported version of Istio: $(istioctl version)"
    exit 1
  fi


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

# must be called after make manifest-gen
function setChartVariables() {
  export INSTALL_DIR="${PROJECT_ROOT}/install/"
  export DEFAULT_MANIFEST="${INSTALL_DIR}/gloo-mesh-default.yaml"
  export AGENT_VALUES=${INSTALL_DIR}/helm/cert-agent/values.yaml
  export AGENT_IMAGE_REGISTRY=$(cat ${AGENT_VALUES} | grep "registry: " | awk '{print $2}')
  export AGENT_IMAGE_REPOSITORY=$(cat ${AGENT_VALUES} | grep "repository: " | awk '{print $2}')
  export AGENT_IMAGE_TAG=$(cat ${AGENT_VALUES} | grep "tag: " | awk '{print $2}' | sed 's/"//g')
  export AGENT_CHART=${INSTALL_DIR}/helm/_output/charts/cert-agent/cert-agent-${AGENT_IMAGE_TAG}.tgz
  export AGENT_IMAGE=${AGENT_IMAGE_REGISTRY}/${AGENT_IMAGE_REPOSITORY}:${AGENT_IMAGE_TAG}
  export GLOOMESH_VALUES=${INSTALL_DIR}/helm/gloo-mesh/values.yaml
  export GLOOMESH_IMAGE_REGISTRY=$(cat ${GLOOMESH_VALUES} | grep "registry: " | awk '{print $2}' | head -1)
  export GLOOMESH_IMAGE_REPOSITORY=$(cat ${GLOOMESH_VALUES} | grep "repository: " | awk '{print $2}' | head -1)
  export GLOOMESH_IMAGE_TAG=$(cat ${GLOOMESH_VALUES} | grep -m 1 "tag: " | awk '{print $2}' | sed 's/"//g')
  export GLOOMESH_IMAGE=${GLOOMESH_IMAGE_REGISTRY}/${GLOOMESH_IMAGE_REPOSITORY}:${GLOOMESH_IMAGE_TAG}
  export GLOOMESH_CHART=${INSTALL_DIR}/helm/_output/charts/gloo-mesh/gloo-mesh-${GLOOMESH_IMAGE_TAG}.tgz

  export AGENT_CRDS_CHART_YAML=${INSTALL_DIR}/helm/agent-crds/Chart.yaml
  export AGENT_CRDS_VERSION=$(cat ${AGENT_CRDS_CHART_YAML} | grep "version: " | awk '{print $2}' | sed 's/"//g')
  export AGENT_CRDS_CHART=${INSTALL_DIR}/helm/_output/charts/agent-crds/agent-crds-${AGENT_CRDS_VERSION}.tgz
}

function register_cluster() {
  cluster=$1
  apiServerAddress=$(get_api_address ${cluster})

  K="kubectl --context kind-${cluster}"

  echo "registering ${cluster} with local cert-agent image..."

  # needed for the agent chart
  setChartVariables

  # load cert-agent image
  kind load docker-image --name "${cluster}" "${AGENT_IMAGE}"

  go run "${PROJECT_ROOT}/cmd/meshctl/main.go" cluster register community "${cluster}" \
    --mgmt-context "kind-${mgmtCluster}" \
    --remote-context "kind-${cluster}" \
    --api-server-address "${apiServerAddress}" \
    --cert-agent-chart-file "${AGENT_CHART}" \
    --agent-crds-chart-file "${AGENT_CRDS_CHART}" ``
}

function install_gloomesh() {

  cluster=$1
  apiServerAddress=$(get_api_address ${cluster})

  bash -x ${PROJECT_ROOT}/ci/setup-gloomesh.sh ${cluster} ${apiServerAddress}

  if [ ! -z ${POST_INSTALL_SCRIPT} ]; then
    bash -x ${POST_INSTALL_SCRIPT}
  fi
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
