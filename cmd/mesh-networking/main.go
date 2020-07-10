package main

import (
	"context"
	"github.com/solo-io/smh/pkg/common/bootstrap"
	mesh_networking "github.com/solo-io/smh/pkg/mesh-networking"
	"log"
)

func main() {
	err := mesh_networking.Start(context.Background(), bootstrap.Options{DebugMode: true, MetricsBindAddress:"0"})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("exiting...")
}
