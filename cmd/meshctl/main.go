package main

import (
	"context"
	"log"

	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands"
)

func main() {
	if err := commands.RootCommand(context.Background()).Execute(); err != nil {
		log.Fatal(err)
	}
}
