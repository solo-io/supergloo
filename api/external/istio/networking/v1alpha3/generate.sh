#!/usr/bin/env bash

set -ex

BASE_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && cd ../../../../../.. >/dev/null && pwd )

OUT=${BASE_DIR}/supergloo/pkg/api/external/istio/networking/v1alpha3/

mkdir -p ${OUT}

ISTIO_IN=${BASE_DIR}/supergloo/api/external/istio/networking/v1alpha3/

IMPORTS="-I=${ISTIO_IN} \
    -I=${GOPATH}/src/github.com/solo-io/supergloo/api/external/gloo/v1 \
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

#!/usr/bin/env bash

set -ex

ROOT=${GOPATH}/src
SUPERGLOO=${ROOT}/github.com/solo-io/supergloo
IN=${SUPERGLOO}/api/external/istio/networking/v1alpha3/
OUT=${SUPERGLOO}/pkg/api/external/istio/networking/v1alpha3/

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

