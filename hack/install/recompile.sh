#!/usr/bin/env bash

# TODO(ilackarms): refactor this out into setup-new-minishift, apply.sh, and rebuild.sh

set -ex

BASEDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# won't work for ui...
# need to modify ui make target
make -C ${BASEDIR}/../.. supergloo-docker
docker save soloio/supergloo:dev | ( eval $(minikube docker-env) && docker load)