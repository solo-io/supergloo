package main

import (
	"context"
	"github.com/solo-io/smh/pkg/cli/commands"
	"log"
)

func main() {
	if err := commands.RootCommand(context.Background()).Execute(); err != nil {
		log.Fatal(err)
	}
}