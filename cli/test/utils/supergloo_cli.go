package utils

import (
	"strings"

	"github.com/solo-io/supergloo/cli/pkg/cmd"
)

func Supergloo(args string) error {
	app := cmd.SuperglooCli("test")
	app.SetArgs(strings.Split(args, " "))
	return app.Execute()
}
