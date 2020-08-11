#!/bin/bash

set -ex

bash ./ci/check-test-suites.bash

protoc --version

if [ ! -f .gitignore ]; then
  echo "_output" > .gitignore
fi

make install-go-tools

set +e

make generated-code -B > /dev/null
if [[ $? -ne 0 ]]; then
  echo "Go code generation failed"
  exit 1;
fi

if [[ $(git status --porcelain | wc -l) -ne 0 ]]; then
  echo "Generating code produced a non-empty diff"
  echo "Try running 'make install-go-tools generated-code -B' then re-pushing."
  git status --porcelain
  git diff | cat
  exit 1;
fi

go mod tidy

if [[ $(git status --porcelain | wc -l) -ne 0 ]]; then
  echo "Need to run go mod tidy before committing"
  git diff
  exit 1;
fi
