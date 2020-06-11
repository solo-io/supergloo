#!/bin/bash

######################################################################################
#
# We want to enforce sane import paths; that means in our case that we want to
# enforce three separate properties on our imports:
#   1. Code in pkg/ must be agnostic of how it is run (eg CLI vs pods vs unknown
#      future use cases). That means pkg/ is disallowed from importing from SMH
#      packages not contained within itself. The only exception is test/.
#   2. Code in cli/ can only import from cli/, pkg/, and test/.
#   3. Code in services/ can only import from services/, pkg/, and test/.
#
# An example antipattern that would violate this is if code in services/ depended
# on importing a constant defined in our CLI code.
#
# I haven't been able to find a way in Go tooling to enforce this, so bash it is >:)
#
######################################################################################
badImportsInPkg=$(find pkg/mesh-discovery -name '*.go' | \
  grep -v .pb.go | \
  xargs grep -E '"github.com/solo-io/service-mesh-hub/.*"' | \
  grep service-mesh-hub/pkg/mesh-discovery | \
  grep service-mesh-hub/pkg/common)
if [ $? = 0 ]; then
  printf "Code in pkg/mesh-discovery should only import from pkg/mesh-discovery or pkg/common"
  echo "Problem files:"
  echo "$badImportsInPkg"
  exit 1
fi
badImportsInPkg=$(find pkg/mesh-networking -name '*.go' | \
  grep -v .pb.go | \
  xargs grep -E '"github.com/solo-io/service-mesh-hub/.*"' | \
  grep service-mesh-hub/pkg/mesh-networking | \
  grep service-mesh-hub/pkg/common)
if [ $? = 0 ]; then
  printf "Code in pkg/mesh-networking should only import from pkg/mesh-networking or pkg/common"
  echo "Problem files:"
  echo "$badImportsInPkg"
  exit 1
fi
badImportsInPkg=$(find pkg/csr-agent -name '*.go' | \
  grep -v .pb.go | \
  xargs grep -E '"github.com/solo-io/service-mesh-hub/.*"' | \
  grep service-mesh-hub/pkg/csr-agent | \
  grep service-mesh-hub/pkg/common)
if [ $? = 0 ]; then
  printf "Code in pkg/csr-agent should only import from pkg/csr-agent or pkg/common"
  echo "Problem files:"
  echo "$badImportsInPkg"
  exit 1
fi
