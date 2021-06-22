// +build tools

// these import statements ensure that `go mod tidy` does not remove repos needed in Makefile targets
package tools

import (
	_ "istio.io/tools/cmd/protoc-gen-jsonshim"
	_ "github.com/solo-io/solo-apis"
)
