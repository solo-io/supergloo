#!/bin/bash

set -ex

ROOTS=$@

for i in $(find ${ROOTS} | grep "\.test$"); do (cd $(dirname $i); ./$(basename $i) -ginkgo.failFast) || exit 1; done
