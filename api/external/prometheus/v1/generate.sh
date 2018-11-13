#!/usr/bin/env bash

set -ex

PROJECTS="$( cd -P "$( dirname "$PROJECTS" )" >/dev/null && pwd )"/../../../../..

OUT=${PROJECTS}/supergloo/pkg/api/external/prometheus/v1/

mkdir -p ${OUT}

PROMETHEUS_IN=${PROJECTS}/supergloo/api/external/prometheus/v1/

IMPORTS="-I=${PROMETHEUS_IN} \
    -I=${GOPATH}/src/github.com/solo-io/supergloo/api/external/gloo/v1 \
    -I=${GOPATH}/src \
    -I=${GOPATH}/src/github.com/solo-io/solo-kit/api/external/proto"

# Run protoc once for gogo
GOGO_FLAG="--gogo_out=Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types:${GOPATH}/src/"
SOLO_KIT_FLAG="--plugin=protoc-gen-solo-kit=${GOPATH}/bin/protoc-gen-solo-kit --solo-kit_out=${PWD}/project.json:${OUT}"

INPUT_PROTOS="${PROMETHEUS_IN}/*.proto"

protoc ${IMPORTS} \
    ${GOGO_FLAG} \
    ${SOLO_KIT_FLAG} \
    ${INPUT_PROTOS}


# currently unused in automation:
# proteus gen command
# used to construct part of config.proto semi-automatically
# proteus proto -f ${GOPATH}/pkg/api/external/prometheus/v1 -p github.com/solo-io/supergloo/api/external/prometheus/v1

