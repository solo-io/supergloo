package main

import (
	"fmt"
	"os"
	"time"

	check "github.com/solo-io/go-checkpoint"
	"github.com/solo-io/supergloo/cli/pkg/cmd"
)

var Version = "0.0.1"

func main() {
	start := time.Now()
	defer check.Format1("supergloo", Version, start)

	app := cmd.App(Version)
	if err := app.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
