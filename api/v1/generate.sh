#!/usr/bin/env bash

set -ex

ROOT=${GOPATH}/src
SUPERGLOO=${ROOT}/github.com/solo-io/supergloo
GLOO_IN=${SUPERGLOO}/api/external/gloo/v1/
ISTIO_ENCRYPTION_IN=${SUPERGLOO}/api/external/istio/encryption/v1/

IN=${SUPERGLOO}/api/v1/
OUT=${SUPERGLOO}/pkg/api/v1/

IMPORTS="\
    -I=${ISTIO_ENCRYPTION_IN} \
    -I=${GLOO_IN} \
    -I=${IN} \
    -I=${SUPERGLOO}/api/external \
    -I=${ROOT}/github.com/solo-io/solo-kit/api/external \
    -I=${ROOT}"

GOGO_FLAG="--gogo_out=Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types:${GOPATH}/src/"

mkdir -p ${OUT}

# Run protoc once for gogo
INPUT_PROTOS="${IN}/*.proto"
protoc ${IMPORTS} \
    ${GOGO_FLAG} \
    ${INPUT_PROTOS}

# Run protoc once for solo kit
RELATIVE_ROOT=../../..
SOLO_KIT_FLAG="--plugin=protoc-gen-solo-kit=${GOPATH}/bin/protoc-gen-solo-kit --solo-kit_out=${OUT} --solo-kit_opt=${PWD}/project.json,${RELATIVE_ROOT}/doc/docs/v1"
INPUT_PROTOS="${IN}/*.proto ${GLOO_IN}/upstream.proto \
    ${ISTIO_ENCRYPTION_IN}/istio_secret.proto \
    ${GLOO_IN}/secret.proto"


protoc ${IMPORTS} \
    ${SOLO_KIT_FLAG} \
    ${INPUT_PROTOS}