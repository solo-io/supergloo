package main

import (
	"context"
	"log"
	"os"

	"github.com/solo-io/service-mesh-hub/cli/pkg/wire"
	docgen "github.com/solo-io/service-mesh-hub/docs"
)

//go:generate go run docsgen.go

func main() {
	log.Printf("Started docs generation\n")
	// generate docs
	rootCmd := wire.InitializeCLI(context.TODO(), os.Stdin, os.Stdout)
	docsGen := docgen.Options{
		Proto: docgen.ProtoOptions{
			OutputDir: "content/reference/api",
		},
		Cli: docgen.CliOptions{
			RootCmd:   rootCmd,
			OutputDir: "content/reference/cli",
		},
		DocsRoot: "docs",
	}

	if err := docgen.Execute(docsGen); err != nil {
		log.Fatal(err)
	}

	log.Printf("Finished generating docs\n")
}
