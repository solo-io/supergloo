package main

import (
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/code-generator/cmd"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/options"
	"github.com/solo-io/solo-kit/pkg/code-generator/sk_anyvendor"
)

//go:generate go run generate.go
//go:generate bash ./api/external/smi/generate.sh

const GlooPkg = "github.com/solo-io/gloo"

func main() {

	imports := sk_anyvendor.CreateDefaultMatchOptions(
		[]string{
			"api/**/*.proto",
			sk_anyvendor.SoloKitMatchPattern,
		},
	)
	imports.External[GlooPkg] = []string{"projects/**/*.proto"}

	log.Printf("starting generate")
	docsOpts := &cmd.DocsOptions{
		Output: options.Hugo,
	}
	if err := cmd.Generate(cmd.GenerateOptions{
		RelativeRoot:  ".",
		CompileProtos: true,
		GenDocs:            docsOpts,
		SkipGeneratedTests: true,
		SkipGenMocks:       true,
		ExternalImports:    imports,
	}); err != nil {
		log.Fatalf("generate failed!: %v", err)
	}
	log.Printf("Finished generating code")
}
