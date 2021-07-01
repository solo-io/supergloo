// +build tools

// these import statements ensure that `go mod tidy` does not remove repos needed in Makefile targets
package tools

import (
	_ "github.com/envoyproxy/protoc-gen-validate"
	//_ "github.com/gogo/googleapis/google/rpc"
	//_ "github.com/googleapis/googleapis"
	_ "github.com/solo-io/solo-apis"
	_ "istio.io/tools/cmd/protoc-gen-jsonshim"
)
