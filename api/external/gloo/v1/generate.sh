#!/usr/bin/env bash 

set -ex

ROOT=${GOPATH}/src
SUPERGLOO=${ROOT}/github.com/solo-io/supergloo
GLOO_IN=${SUPERGLOO}/api/external/gloo/v1/
GLOO_OUT=${SUPERGLOO}/pkg/api/external/gloo/v1/


GOGO_OUT_FLAG="--gogo_out=plugins=grpc,"
GOGO_OUT_FLAG+="Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor,"
GOGO_OUT_FLAG+="Menvoy/api/v2/discovery.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2"
GOGO_OUT_FLAG+=":${GOPATH}/src/"

SOLO_KIT_FLAG="--plugin=protoc-gen-solo-kit=${GOPATH}/bin/protoc-gen-solo-kit --solo-kit_out=${PWD}/project.json:${GLOO_OUT}"

mkdir -p ${GLOO_OUT}/plugins

PROTOC_FLAGS="-I=${GOPATH}/src \
    -I=${GOPATH}/src/github.com/solo-io/solo-kit/api/external/proto \
    ${GOGO_OUT_FLAG} \
    ${SOLO_KIT_FLAG}"

protoc -I=${GLOO_IN} ${PROTOC_FLAGS} ${GLOO_IN}/*.proto

IN=${GLOO_IN}/plugins

# protoc made me do it
protoc -I=${GLOO_IN} ${PROTOC_FLAGS} ${GOPATH}/src/github.com/solo-io/supergloo/api/external/gloo/v1/plugins/service_spec.proto

for plugin_dir in $(echo plugins/*/); do
# remove folder
plugin=${plugin_dir#"plugins/"}
# remove trailing slash
plugin=${plugin%"/"}

mkdir -p ${GLOO_OUT}/plugins/${plugin}

# we need ${GOPATH}/src/github.com/gogo/protobuf/protobuf
# as the filter's protobufs use validate/validate.proto
protoc -I=${GLOO_IN} ${PROTOC_FLAGS} ${GLOO_IN}/plugins/${plugin}/*.proto
done
