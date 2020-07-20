package main

import (
	"context"
	"log"

	"github.com/solo-io/service-mesh-hub/pkg/common/bootstrap"
	mesh_networking "github.com/solo-io/service-mesh-hub/pkg/mesh-networking"
)

func main() {
	err := mesh_networking.Start(context.Background(), bootstrap.Options{DebugMode: true, MetricsBindAddress: "0"})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("exiting...")
}
