package main

import (
	"github.com/solo-io/go-utils/clidoc"
	"github.com/solo-io/supergloo/cli/pkg/cmd"
	"github.com/solo-io/supergloo/pkg/version"
)

func main() {
	app := cmd.SuperglooCli(version.Version)
	clidoc.MustGenerateCliDocs(app)
}
