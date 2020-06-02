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

badImportsInPkg=$(find pkg -name '*.go' | \
  grep -v .pb.go | \
  xargs grep -E '"github.com/solo-io/service-mesh-hub/.*"' | \
  grep -v service-mesh-hub/pkg | \
  grep -v service-mesh-hub/test)
if [ $? = 0 ]; then
  printf "Code in top-level pkg/ must only import Service Mesh Hub code from itself, not cli/ or services/\n\n"
  echo "Problem files:"
  echo "$badImportsInPkg"
  exit 1
fi

badImportsInCli=$(find cli -name '*.go' | \
  grep -v .pb.go | \
  xargs grep -E '"github.com/solo-io/service-mesh-hub/.*"' | \
  grep -v service-mesh-hub/cli | \
  grep -v service-mesh-hub/pkg | \
  grep -v service-mesh-hub/test \
)
if [ $? = 0 ]; then
  printf "Code in cli/ must only import Service Mesh Hub code from itself, pkg/ or test/\n\n"
  echo "Problem files:"
  echo "$badImportsInCli"
  exit 1
fi

badImportsInServices=$(find services -name '*.go' | \
  grep -v .pb.go | \
  xargs grep -E '"github.com/solo-io/service-mesh-hub/.*"' | \
  grep -v service-mesh-hub/services | \
  grep -v service-mesh-hub/pkg | \
  grep -v service-mesh-hub/test \
)
if [ $? = 0 ]; then
  printf "Code in services/ must only import Service Mesh Hub code from itself, pkg/ or test/\n\n"
  echo "Problem files:"
  echo "$badImportsInServices"
  exit 1
fi
