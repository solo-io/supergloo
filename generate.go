package main

import (
	"github.com/solo-io/solo-kit/pkg/code-generator/cmd"
	"github.com/solo-io/solo-kit/pkg/utils/log"
)

//go:generate go run generate.go

func main() {
	log.Printf("starting generate")
	if err := cmd.Run(".", true, true, []string{"../gloo"}, []string{"pkg2", "api2"}); err != nil {
		log.Fatalf("generate failed!: %v", err)
	}
}
