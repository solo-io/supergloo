package main

import (
	"github.com/solo-io/solo-kit/pkg/code-generator/cmd"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/options"
	"github.com/solo-io/solo-kit/pkg/utils/log"
)

//go:generate go run generate.go

func main() {
	log.Printf("starting generate")
	docsOpts := &cmd.DocsOptions{
		Output: options.Hugo,
	}
	if err := cmd.Run(".", true, docsOpts, []string{"../gloo"}, []string{"pkg2", "api2"}); err != nil {
		log.Fatalf("generate failed!: %v", err)
	}
}
