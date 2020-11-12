package main

import (
	"context"
	"log"

	"github.com/solo-io/anyvendor/pkg/manager"
	"github.com/solo-io/gloo-mesh/codegen/anyvendor"
	"github.com/solo-io/gloo-mesh/docs/docsgen"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands"
)

//go:generate go run docsgen.go

func main() {
	log.Printf("Started docs generation\n")

	ctx := context.TODO()
	mgr, err := manager.NewManager(ctx, "")
	if err != nil {
		log.Fatal("failed to initialize vendor_any manager")
	}

	anyvendorImports := anyvendor.AnyVendorImports()

	if err = mgr.Ensure(ctx, anyvendorImports.ToAnyvendorConfig()); err != nil {
		log.Fatal("failed to import protos")
	}
	// generate docs
	docsGen := docsgen.Options{
		Proto: docsgen.ProtoOptions{
			OutputDir: "content/reference/api",
		},
		Cli: docsgen.CliOptions{
			RootCmd:   commands.RootCommand(ctx),
			OutputDir: "content/reference/cli",
		},
		DocsRoot: "docs",
	}

	if err := docsgen.Execute(docsGen); err != nil {
		log.Fatal(err)
	}

	log.Printf("Finished generating docs\n")
}
