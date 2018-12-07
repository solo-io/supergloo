#!/usr/bin/env bash

cd $GOPATH/src/github.com/solo-io/supergloo/cli/cmd
go build -o $GOPATH/src/github.com/solo-io/supergloo/test/cli/_output/supergloo
