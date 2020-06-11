#!/bin/bash

######################################################################################
#
# Our project structure follows these constraints:
#   1. Code in `pkg/csr-agent`, `pkg/mesh-discovery`, and `pkg/mesh-networking`
#      only imports code from itself or `pkg/common`
#
#   Additional constraints can be enforced by extended the bash script below.
#
######################################################################################

function ensure_no_cross_imports() {
  root_pkg=$1
  badImportsInPkg=$(find $root_pkg -name '*.go' | \
    grep -v .pb.go | \
    xargs grep -E '"github.com/solo-io/service-mesh-hub/.*"' | \
    grep -v service-mesh-hub/$root_pkg | \
    grep -v service-mesh-hub/test | \
    grep -v service-mesh-hub/pkg/api | \
    grep -v service-mesh-hub/pkg/common)
  if [ $? = 0 ]; then
    printf "Code in $root_pkg should only import from $root_pkg or pkg/common\n"
    echo "Problem files:"
    echo "$badImportsInPkg"
    exit 1
  fi
}

ensure_no_cross_imports pkg/mesh-discovery
ensure_no_cross_imports pkg/mesh-networking
ensure_no_cross_imports pkg/csr-agent
