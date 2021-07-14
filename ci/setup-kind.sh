#!/bin/bash -ex

set -o xtrace

#####################################
#
# Set up two kind clusters:
#   1. a management cluster in which Gloo Mesh is installed.
#   2. a remote cluster in which only Istio and the bookinfo app are installed.
#
# The management cluster will have the appropriate secret for communicating with the remote cluster
# Your kube context will be left pointing to the management cluster
# Each cluster will have Istio set up in the istio-system namespace in its minimal profile
#
# To clean up your kind clusters, run this script as: `bash ci/setup-kind.sh cleanup`. Use this if you notice the docker VM running out of disk space (for images).
#
#####################################

PROJECT_ROOT=$( cd "$( dirname "${0}" )" >/dev/null 2>&1 && pwd )/..
echo "Using project root ${PROJECT_ROOT}"

# print build info
make -C ${PROJECT_ROOT} print-info

source ${PROJECT_ROOT}/ci/setup-funcs.sh

if [ "$1" == "cleanup" ]; then
  kind get clusters | grep -E "${mgmtCluster}|${remoteCluster}" | while read -r r; do kind delete cluster --name "${r}"; done

  # Only cleanup bird container if running with flat-networking
  if [ ! -z ${FLAT_NETWORKING_ENABLED} ]; then
    docker stop bird
  fi

  exit 0
fi

# default mgmt/remote cluster ingress ports
mgmtEastWestIngressPort=32001
((mgmtNorthSouthIngressPort=mgmtEastWestIngressPort+10)) # 32011
remoteEastWestIngressPort=32000
((remoteNorthSouthIngressPort=remoteEastWestIngressPort+10)) # 32010

# set mgmt/remote cluster ingress ports from environment variables if they exist
if [ ! -z ${MGMT_INGRESS_PORT_1} ]; then
    mgmtEastWestIngressPort="${MGMT_INGRESS_PORT_1}"
fi
if [ ! -z ${MGMT_INGRESS_PORT_2} ]; then
    mgmtNorthSouthIngressPort="${MGMT_INGRESS_PORT_2}"
fi
if [ ! -z ${REMOTE_INGRESS_PORT_1} ]; then
    remoteEastWestIngressPort="${REMOTE_INGRESS_PORT_1}"
fi
if [ ! -z ${REMOTE_INGRESS_PORT_2} ]; then
    remoteNorthSouthIngressPort="${REMOTE_INGRESS_PORT_2}"
fi

if [ "$1" == "osm" ]; then
  # optionally install open service mesh
  create_kind_cluster ${mgmtCluster} ${mgmtEastWestIngressPort} ${mgmtNorthSouthIngressPort}
  install_osm ${mgmtCluster} ${mgmtEastWestIngressPort}

  echo successfully set up clusters.

  # install gloo mesh
  install_gloomesh ${mgmtCluster}

  # sleep to allow crds to register
  sleep 4

  # register clusters
  register_cluster ${mgmtCluster}

else

  # NOTE(ilackarms): we run the setup_kind clusters sequentially due to this bug:
  # related: https://github.com/kubernetes-sigs/kind/issues/1596
  create_kind_cluster ${mgmtCluster} ${mgmtEastWestIngressPort} ${mgmtNorthSouthIngressPort}
  install_istio ${mgmtCluster} ${mgmtEastWestIngressPort} ${mgmtNorthSouthIngressPort}

  create_kind_cluster ${remoteCluster} ${remoteEastWestIngressPort} ${remoteNorthSouthIngressPort}
  install_istio ${remoteCluster} ${remoteEastWestIngressPort} ${remoteNorthSouthIngressPort}

  if [ ! -z ${FLAT_NETWORKING_ENABLED} ]; then
    setup_flat_networking ${mgmtCluster} ${mgmtEastWestIngressPort} ${remoteCluster} ${remoteEastWestIngressPort}
  fi

  # create istio-injectable namespace
  kubectl --context kind-${mgmtCluster} create namespace bookinfo
  kubectl --context kind-${mgmtCluster} label ns bookinfo istio-injection=enabled --overwrite
  kubectl --context kind-${remoteCluster} create namespace bookinfo
  kubectl --context kind-${remoteCluster} label ns bookinfo istio-injection=enabled --overwrite

  # install bookinfo with reviews-v1 and reviews-v2 to management cluster
  kubectl --context kind-${mgmtCluster} -n bookinfo apply -f ./ci/bookinfo.yaml -l 'app notin (details),version in (v1, v2)'
  kubectl --context kind-${mgmtCluster} -n bookinfo apply -f ./ci/bookinfo.yaml -l 'service=reviews'
  kubectl --context kind-${mgmtCluster} -n bookinfo apply -f ./ci/bookinfo.yaml -l 'account=reviews'
  kubectl --context kind-${mgmtCluster} -n bookinfo apply -f ./ci/bookinfo.yaml -l 'app=ratings'
  kubectl --context kind-${mgmtCluster} -n bookinfo apply -f ./ci/bookinfo.yaml -l 'account=ratings'
  kubectl --context kind-${mgmtCluster} -n bookinfo apply -f ./ci/bookinfo.yaml -l 'app=productpage'
  kubectl --context kind-${mgmtCluster} -n bookinfo apply -f ./ci/bookinfo.yaml -l 'account=productpage'

  # install bookinfo with reviews-v1 and reviews-v3 to remote cluster
  kubectl --context kind-${remoteCluster} -n bookinfo apply -f ./ci/bookinfo.yaml -l 'app notin (details),version in (v1, v3)'
  kubectl --context kind-${remoteCluster} -n bookinfo apply -f ./ci/bookinfo.yaml -l 'service=reviews'
  kubectl --context kind-${remoteCluster} -n bookinfo apply -f ./ci/bookinfo.yaml -l 'account=reviews'
  kubectl --context kind-${remoteCluster} -n bookinfo apply -f ./ci/bookinfo.yaml -l 'app=ratings'
  kubectl --context kind-${remoteCluster} -n bookinfo apply -f ./ci/bookinfo.yaml -l 'account=ratings'

  # wait for deployments to finish
  kubectl --context kind-${mgmtCluster} -n bookinfo rollout status deployment/productpage-v1 --timeout=300s
  kubectl --context kind-${mgmtCluster} -n bookinfo rollout status deployment/reviews-v1 --timeout=300s
  kubectl --context kind-${mgmtCluster} -n bookinfo rollout status deployment/reviews-v2 --timeout=300s

  kubectl --context kind-${remoteCluster} -n bookinfo rollout status deployment/reviews-v3 --timeout=300s

  echo successfully set up clusters.

  # skip installing and registering gloomesh components from source (just leaves clusters set up with istio+bookinfo)
  if [ "$SKIP_DEPLOY_FROM_SOURCE" == "1" ]; then
    echo "skipping deploy of gloomesh components."
  else
    # install gloo mesh
    install_gloomesh ${mgmtCluster}

    # sleep to allow crds to register
    sleep 4

    # register remote cluster
    register_cluster ${remoteCluster} &

    wait
  fi

fi
# set current context to management cluster
kubectl config use-context kind-${mgmtCluster}

