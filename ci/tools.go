// +build tools

// these import statements ensure that `go mod tidy` does not remove repos needed in vendor_any for protobuf generation
package tools

import (
	_ "github.com/cncf/udpa/test/build"
	_ "github.com/envoyproxy/data-plane-api/test/build"
)
