package main

import (
	"context"
	mesh_discovery "github.com/solo-io/smh/pkg/mesh-discovery"
	"log"
)

func main() {
	err := mesh_discovery.Start(context.Background(), mesh_discovery.Options{})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("exiting...")
}
