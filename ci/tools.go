// +build tools

/*
	Explanation for tools pattern:
	https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
*/

package tools

import (
	_ "github.com/envoyproxy/protoc-gen-validate"
	_ "github.com/gogo/protobuf/gogoproto"
	_ "github.com/gogo/protobuf/protoc-gen-gogo"
	_ "github.com/golang/mock/gomock"
	_ "github.com/golang/mock/mockgen"
	_ "github.com/google/wire/cmd/wire"
	_ "github.com/solo-io/protoc-gen-ext"
	_ "golang.org/x/tools/cmd/goimports"
	_ "k8s.io/code-generator"
)
