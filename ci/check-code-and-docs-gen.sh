#!/bin/bash

set -ex

protoc --version

if [ ! -f .gitignore ]; then
  echo "_output" > .gitignore
fi

git init
git add .
git commit -m "set up dummy repo for diffing" -q

git clone https://github.com/solo-io/solo-kit /workspace/gopath/src/github.com/solo-io/solo-kit
git clone https://github.com/solo-io/gloo /workspace/gopath/src/github.com/solo-io/gloo

make update-deps
make pin-repos

PATH=/workspace/gopath/bin:$PATH

set +e

make generated-code -B  > /dev/null
if [[ $? -ne 0 ]]; then
  echo "Code generation failed"
  exit 1;
fi
if [[ $(git status --porcelain | wc -l) -ne 0 ]]; then
  echo "Generating code produced a non-empty diff."
  echo "Try running 'dep ensure && make update-deps generated-code -B' then re-pushing."
  git status --porcelain
  git diff | cat
  exit 1;
fi

if [[ $(git --git-dir=/workspace/gopath/src/github.com/solo-io/gloo/.git --work-tree=/workspace/gopath/src/github.com/solo-io/gloo status --porcelain | wc -l) -ne 0 ]]; then
echo "Generating code produced a non-empty diff in the gloo repo"
  echo "Make sure the go_import directory in protos is set to directory in supergloo, not in gloo."
  git --git-dir=/workspace/gopath/src/github.com/solo-io/gloo/.git --work-tree=/workspace/gopath/src/github.com/solo-io/gloo status --porcelain
  git --git-dir=/workspace/gopath/src/github.com/solo-io/gloo/.git --work-tree=/workspace/gopath/src/github.com/solo-io/gloo diff | cat
  exit 1;
fi

