#!/usr/bin/env bash

# Use this script instead of running supergloo init in order to deploy the local supergloo server code for testing

set -ex

# Expects supergloo to be installed in the standard place
SUPERGLOO_DIR="${GOPATH}/src/github.com/solo-io/supergloo"
HACK_DIR="${SUPERGLOO_DIR}/hack/install"

cd $SUPERGLOO_DIR

sh $HACK_DIR/recompile.sh

sed 's/imagePullPolicy: Always/imagePullPolicy: IfNotPresent/g' $HACK_DIR/supergloo.yaml | kubectl apply -f -
make install-cli
