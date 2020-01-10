// +build tools

/*
	Explanation for tools pattern:
	https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
*/

package tools

import (
	_ "github.com/envoyproxy/protoc-gen-validate"
	_ "github.com/gogo/protobuf/gogoproto"
	_ "github.com/solo-io/protoc-gen-ext"
	_ "k8s.io/code-generator"
)
