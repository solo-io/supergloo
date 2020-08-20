package main

import (
	"context"
	"os"

	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands"
)

func main() {
	if err := commands.RootCommand(context.Background()).Execute(); err != nil {
		os.Exit(1)
	}
}
