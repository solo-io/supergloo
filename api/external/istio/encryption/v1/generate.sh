#!/usr/bin/env bash

set -ex

ROOT=${GOPATH}/src
SUPERGLOO=${ROOT}/github.com/solo-io/supergloo
IN=${SUPERGLOO}/api/external/istio/encryption/v1/
OUT=${SUPERGLOO}/pkg/api/external/istio/encryption/v1/

IMPORTS="\
    -I=${IN} \
    -I=${SUPERGLOO}/api/external \
    -I=${ROOT}/github.com/solo-io/solo-kit/api/external \
    -I=${ROOT} \
    "

GOGO_FLAG="--gogo_out=Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types:${GOPATH}/src/"
SOLO_KIT_FLAG="--plugin=protoc-gen-solo-kit=${GOPATH}/bin/protoc-gen-solo-kit --solo-kit_out=${PWD}/project.json:${OUT}"
INPUT_PROTOS="${IN}/*.proto"

mkdir -p ${OUT}
protoc ${IMPORTS} \
    ${GOGO_FLAG} \
    ${SOLO_KIT_FLAG} \
    ${INPUT_PROTOS}
