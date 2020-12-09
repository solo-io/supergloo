#!/bin/bash

set -ex

function check_diffs() {
    git status --porcelain --untracked-files=no
}

bash ./ci/check-test-suites.bash

protoc --version

if [ ! -f .gitignore ]; then
  echo "_output" > .gitignore
fi

make install-go-tools

go mod tidy

if [[ $(check_diffs | wc -l) -ne 0 ]]; then
  echo "Need to run go mod tidy before committing"
  git diff
  exit 1;
fi


set +e

make generated-code -B > /dev/null
if [[ $? -ne 0 ]]; then
  echo "Go code generation failed"
  exit 1;
fi

if [[ $(check_diffs | wc -l) -ne 0 ]]; then
  echo "Generating code produced a non-empty diff"
  echo "Try running 'make install-go-tools generated-code -B' then re-pushing."
  check_diffs
  git diff | cat
  exit 1;
fi
