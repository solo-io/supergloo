#!/usr/bin/env bash

set -ex

BASE_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && cd ../../../../../.. >/dev/null && pwd )

OUT=${BASE_DIR}/supergloo/pkg/api/external/istio/rbac/v1alpha1/

mkdir -p ${OUT}

ISTIO_IN=${BASE_DIR}/supergloo/api/external/istio/rbac/v1alpha1/

IMPORTS="-I=${ISTIO_IN} \
    -I=${GOPATH}/src/github.com/solo-io/solo-kit/projects/gloo/api/v1 \
    -I=${GOPATH}/src \
    -I=${GOPATH}/src/github.com/solo-io/solo-kit/api/external"

# Run protoc once for gogo
GOGO_FLAG="--gogo_out=Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types:${GOPATH}/src/"
SOLO_KIT_FLAG="--plugin=protoc-gen-solo-kit=${GOPATH}/bin/protoc-gen-solo-kit --solo-kit_out=${PWD}/project.json:${OUT}"

INPUT_PROTOS="${ISTIO_IN}/*.proto"

protoc ${IMPORTS} \
    ${GOGO_FLAG} \
    ${SOLO_KIT_FLAG} \
    ${INPUT_PROTOS}

