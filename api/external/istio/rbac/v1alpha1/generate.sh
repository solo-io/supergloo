#!/usr/bin/env bash

set -ex

ROOT=${GOPATH}/src
SUPERGLOO=${ROOT}/github.com/solo-io/supergloo
GLOO_IN=${SUPERGLOO}/api/external/gloo/v1/

IN=${SUPERGLOO}/api/external/istio/rbac/v1alpha1/
OUT=${SUPERGLOO}/pkg/api/external/istio/rbac/v1alpha1/

IMPORTS="
    -I=${GLOO_IN} \
    -I=${IN} \
    -I=${ROOT}/github.com/solo-io/solo-kit/api/external \
    -I=${ROOT}/github.com/solo-io/supergloo/api/external/gloo/v1 \
    -I=${ROOT} \
    "

# Run protoc once for gogo
GOGO_FLAG="--gogo_out=Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types:${GOPATH}/src/"
SOLO_KIT_FLAG="--plugin=protoc-gen-solo-kit=${GOPATH}/bin/protoc-gen-solo-kit --solo-kit_out=${PWD}/project.json:${OUT}"

INPUT_PROTOS="${ISTIO_IN}/*.proto"

mkdir -p ${OUT}

protoc ${IMPORTS} \
    ${GOGO_FLAG} \
    ${SOLO_KIT_FLAG} \
    ${INPUT_PROTOS}

