package main

import (
	"context"
	"os"

	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands"
)

func main() {
	if err := commands.RootCommand(context.Background()).Execute(); err != nil {
		os.Exit(1)
	}
}
