#!/usr/bin/env bash

set -ex

ROOT=${GOPATH}/src
SUPERGLOO=${ROOT}/github.com/solo-io/supergloo
OUT=${SUPERGLOO}/pkg/api/v1/
GLOO_IN=${SUPERGLOO}/api/external/gloo/v1/
SUPERGLOO_IN=${SUPERGLOO}/api/v1/

IMPORTS="-I=${GLOO_IN} \
    -I=${SUPERGLOO_IN} \
    -I=${SUPERGLOO}/api/external \
    -I=${ROOT}/github.com/solo-io/solo-kit/api/external \
    -I=${ROOT}"

GOGO_FLAG="--gogo_out=Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types:${GOPATH}/src/"

mkdir -p ${OUT}

# Run protoc once for gogo
INPUT_PROTOS="${SUPERGLOO_IN}/*.proto"
protoc ${IMPORTS} \
    ${GOGO_FLAG} \
    ${INPUT_PROTOS}

# Run protoc once for solo kit
SOLO_KIT_FLAG="--plugin=protoc-gen-solo-kit=${GOPATH}/bin/protoc-gen-solo-kit --solo-kit_out=${PWD}/project.json:${OUT}"
INPUT_PROTOS="${SUPERGLOO_IN}/*.proto ${GLOO_IN}/upstream.proto"

protoc ${IMPORTS} \
    ${SOLO_KIT_FLAG} \
    ${INPUT_PROTOS}

