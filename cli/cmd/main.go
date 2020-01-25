package main

import (
	"context"
	"os"

	"github.com/solo-io/mesh-projects/cli/pkg/wire"
)

func main() {
	cliApp := wire.InitializeCLI(context.Background(), os.Stdout)
	err := cliApp.Execute()
	if err != nil {
		os.Exit(1)
	}
}
