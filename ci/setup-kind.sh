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

if [ "$1" == "cleanup" ]; then
  kind get clusters | grep -E "${masterCluster}|${remoteCluster}" | while read -r r; do kind delete cluster --name "${r}"; done
  exit 0
fi

PROJECT_ROOT=$( cd "$( dirname "${0}" )" >/dev/null 2>&1 && pwd )/..
echo "Using project root ${PROJECT_ROOT}"

source ${PROJECT_ROOT}/ci/setup-funcs.sh

# NOTE(ilackarms): we run the setup_kind clusters sequentially due to this bug:
# related: https://github.com/kubernetes-sigs/kind/issues/1596
create_kind_cluster ${masterCluster} 32001
install_istio ${masterCluster} 32001 &

create_kind_cluster ${remoteCluster} 32000
install_istio ${remoteCluster} 32000 &

wait

echo successfully set up clusters.

# install service mesh hub
install_smh ${masterCluster}

# sleep to allow crds to register
sleep 4

# register remote cluster
register_cluster ${remoteCluster} &

wait

# set current context to master cluster
kubectl config use-context kind-${masterCluster}

