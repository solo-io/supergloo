#!/bin/bash

set -ex

go mod tidy

if [[ $(git status --porcelain | wc -l) -ne 0 ]]; then
  echo "Need to run go mod tidy before committing"
  git diff
  exit 1;
fi

protoc --version

if [ ! -f .gitignore ]; then
  echo "_output" > .gitignore
fi

make update-deps

set +e

REGENERATE_TYPES=1 make generated-code -B > /dev/null
if [[ $? -ne 0 ]]; then
  echo "Go code generation failed"
  exit 1;
fi

if [[ $(git status --porcelain | wc -l) -ne 0 ]]; then
  echo "Generating code produced a non-empty diff"
  echo "Try running 'make update-deps generated-code -B' then re-pushing."
  git status --porcelain
  git diff | cat
  exit 1;
fi
