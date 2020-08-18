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

PROJECT_ROOT=$( cd "$( dirname "${0}" )" >/dev/null 2>&1 && pwd )/..
echo "Using project root ${PROJECT_ROOT}"
source ${PROJECT_ROOT}/ci/setup-funcs.sh

if [ "$1" == "cleanup" ]; then
  kind get clusters | grep -E "${masterCluster}|${remoteCluster}" | while read -r r; do kind delete cluster --name "${r}"; done
  exit 0
fi

# NOTE(ilackarms): we run the setup_kind clusters sequentially due to this bug:
# related: https://github.com/kubernetes-sigs/kind/issues/1596
if [ "$1" == "osm" ]; then
  # optionally install open service mesh
  create_kind_cluster ${masterCluster} 32001
  install_osm ${masterCluster} 32001 &

  wait

  echo successfully set up clusters.

  # install service mesh hub
  install_smh ${masterCluster}

  # sleep to allow crds to register
  sleep 4

  # register clusters
  register_cluster ${masterCluster} &

  wait

else
  # default to istio install
  create_kind_cluster ${masterCluster} 32001
  install_istio ${masterCluster} 32001 &

  create_kind_cluster ${remoteCluster} 32000
  install_istio ${remoteCluster} 32000 &

  wait

  # create istio-injectable namespace
  kubectl --context kind-${masterCluster} create namespace bookinfo
  kubectl --context kind-${masterCluster} label ns bookinfo istio-injection=enabled --overwrite
  kubectl --context kind-${remoteCluster} create namespace bookinfo
  kubectl --context kind-${remoteCluster} label ns bookinfo istio-injection=enabled --overwrite

  # install bookinfo without reviews-v3 to master cluster
  kubectl --context kind-${masterCluster} -n bookinfo apply -f ./ci/bookinfo.yaml -l 'app notin (details),version notin (v3)'
  kubectl --context kind-${masterCluster} -n bookinfo apply -f ./ci/bookinfo.yaml -l 'account'

  # install only reviews-v3 to remote cluster
  kubectl --context kind-${remoteCluster} -n bookinfo apply -f ./ci/bookinfo.yaml -l 'app notin (details),version in (v3)'
  kubectl --context kind-${remoteCluster} -n bookinfo apply -f ./ci/bookinfo.yaml -l 'service=reviews'
  kubectl --context kind-${remoteCluster} -n bookinfo apply -f ./ci/bookinfo.yaml -l 'account=reviews'
  kubectl --context kind-${remoteCluster} -n bookinfo apply -f ./ci/bookinfo.yaml -l 'app=ratings'
  kubectl --context kind-${remoteCluster} -n bookinfo apply -f ./ci/bookinfo.yaml -l 'account=ratings'

  # wait for deployments to finish
  kubectl --context kind-${masterCluster} -n bookinfo rollout status deployment/productpage-v1 --timeout=300s
  kubectl --context kind-${masterCluster} -n bookinfo rollout status deployment/reviews-v1 --timeout=300s
  kubectl --context kind-${masterCluster} -n bookinfo rollout status deployment/reviews-v2 --timeout=300s

  kubectl --context kind-${remoteCluster} -n bookinfo rollout status deployment/reviews-v3 --timeout=300s

  echo successfully set up clusters.

  # install service mesh hub
  install_smh ${masterCluster}

  # sleep to allow crds to register
  sleep 4

  # register remote cluster
  register_cluster ${remoteCluster} &

  wait

fi

# set current context to master cluster
kubectl config use-context kind-${masterCluster}

