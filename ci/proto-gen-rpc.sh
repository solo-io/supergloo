
#!/usr/bin/env bash

set -e

VENDOR_FOLDER=vendor_any
PROTOC_IMPORT_PATH="${VENDOR_FOLDER}"

COMMON_PROTO_IMPORTS="-I${PROTOC_IMPORT_PATH}/github.com/solo-io -I${PROTOC_IMPORT_PATH}/github.com/envoyproxy/protoc-gen-validate -I${PROTOC_IMPORT_PATH}/github.com/gogo/protobuf -I${PROTOC_IMPORT_PATH}/github.com/solo-io/protoc-gen-ext -I${PROTOC_IMPORT_PATH}"

TEMP_DIR=$(mktemp -d)
cleanup() {
    echo ">> Removing ${TEMP_DIR}"
    rm -rf ${TEMP_DIR}
}
trap "cleanup" EXIT SIGINT

echo ">> Temporary output directory ${TEMP_DIR}"

COMMON_GO_PROTO_FLAGS="--gogo_out=plugins=grpc:${TEMP_DIR} ${COMMON_PROTO_IMPORTS}"

protoc ${COMMON_GO_PROTO_FLAGS} ${PROTOC_IMPORT_PATH}/github.com/solo-io/service-mesh-hub/services/apiserver/api/v1/*.proto

cp -a "${TEMP_DIR}/github.com/solo-io/service-mesh-hub/." "."