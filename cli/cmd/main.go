package main

import (
	"fmt"
	"os"
	"time"

	check "github.com/solo-io/go-checkpoint"
	"github.com/solo-io/supergloo/cli/pkg/cmd"
	"github.com/solo-io/supergloo/pkg/version"
)

func main() {
	start := time.Now()
	defer check.CallReport("supergloo", version.Version, start)

	app := cmd.App(version.Version)
	if err := app.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
