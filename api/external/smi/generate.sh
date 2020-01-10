
#!/usr/bin/env bash

set -e

set -o errexit
set -o nounset
set -o pipefail


# The following script is used to generate the smi protos.
# This script will work both in and out of the GOPATH, however, it does assume that the imported protos will be
# available in the root level vendor folder. This script will be run as part of `make generated-code` so there
# should be no need to run it otherwise. `make generated-code` will also vendor the necessary protos.
# This script is called from generate.go in the repo root
ROOT=$(dirname "${BASH_SOURCE[0]}")/../../../..
MESH_PROJECTS=${ROOT}/mesh-projects
IN=${MESH_PROJECTS}/api/external/smi
VENDOR_ROOT=${MESH_PROJECTS}/vendor_any/github.com

TEMP_DIR=$(mktemp -d)
cleanup() {
    echo ">> Removing ${TEMP_DIR}"
    rm -rf ${TEMP_DIR}
}
trap "cleanup" EXIT SIGINT

echo ">> Temporary output directory ${TEMP_DIR}"

IMPORTS="\
    -I=${IN} \
    -I=${ROOT} \
    -I=${VENDOR_ROOT}/gogo/protobuf \
    -I=${VENDOR_ROOT}/solo-io/protoc-gen-ext \
    -I=${VENDOR_ROOT}/solo-io"

GOGO_FLAG="--gogo_out=Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:${TEMP_DIR}"
HASH_FLAG="--ext_out=Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:${TEMP_DIR}"

TRAFFIC_TARGET_PROTOS="${IN}/traffictarget/v1alpha1/*.proto"
TRAFFIC_SPLIT_PROTOS="${IN}/trafficsplit/v1alpha2/*.proto"
ROUTE_GROUP_PROTOS="${IN}/httproutegroup/v1alpha1/*.proto"

protoc ${IMPORTS} \
    ${GOGO_FLAG} \
    ${HASH_FLAG} \
    ${TRAFFIC_TARGET_PROTOS}
protoc ${IMPORTS} \
    ${GOGO_FLAG} \
    ${HASH_FLAG} \
    ${ROUTE_GROUP_PROTOS}
protoc ${IMPORTS} \
    ${GOGO_FLAG} \
    ${HASH_FLAG} \
    ${TRAFFIC_SPLIT_PROTOS}

cp -r  ${TEMP_DIR}/github.com/solo-io/mesh-projects ${ROOT}

goimports -w pkg