package main

import (
	"context"
	"log"

	"github.com/solo-io/smh/pkg/common/bootstrap"
	mesh_discovery "github.com/solo-io/smh/pkg/mesh-discovery"
)

func main() {
	err := mesh_discovery.Start(context.Background(), bootstrap.Options{DebugMode: true})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("exiting...")
}
