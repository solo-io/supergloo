package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands"
)

func main() {
	if err := commands.RootCommand(context.Background()).Execute(); err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}
}
