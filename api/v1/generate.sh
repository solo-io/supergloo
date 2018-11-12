#!/usr/bin/env bash

set -ex

ROOT=${GOPATH}/src

PROJECTS=${ROOT}/github.com/solo-io/solo-projects/projects
SUPERGLOO=${ROOT}/github.com/solo-io/supergloo

OUT=${SUPERGLOO}/pkg/api/v1/

mkdir -p ${OUT}

GLOO_IN=${PROJECTS}/gloo/api/v1/

SUPERGLOO_IN=${SUPERGLOO}/api/v1/

IMPORTS="-I=${GLOO_IN} \
    -I=${SUPERGLOO_IN} \
    -I=${ROOT}/github.com/solo-io/solo-projects/projects/gloo/api/v1 \
    -I=${ROOT} \
    -I=${ROOT}/github.com/solo-io/solo-kit/api/external/proto"

# Run protoc once for gogo
GOGO_FLAG="--gogo_out=Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types:${GOPATH}/src/"

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

