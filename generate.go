package main

import (
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/code-generator/cmd"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/options"
	"github.com/solo-io/supergloo/pkg/version"
)

//go:generate go run generate.go

func main() {
	err := version.CheckVersions()
	if err != nil {
		log.Fatalf("generate failed!: %v", err)
	}
	log.Printf("starting generate")
	docsOpts := &cmd.DocsOptions{
		Output: options.Hugo,
	}
	if err := cmd.Generate(cmd.GenerateOptions{
		RelativeRoot:       ".",
		CompileProtos:      true,
		GenDocs:            docsOpts,
		CustomImports:      []string{"../gloo"},
		SkipGenMocks:       false,
		SkipGeneratedTests: true,
	}); err != nil {
		log.Fatalf("generate failed!: %v", err)
	}
}
