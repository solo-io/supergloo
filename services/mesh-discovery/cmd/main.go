package main

import (
	"github.com/solo-io/mesh-projects/services/internal/config"
	mesh_discovery "github.com/solo-io/mesh-projects/services/mesh-discovery"
)

func main() {
	ctx := config.CreateRootContext(nil, "mesh-discovery")
	mesh_discovery.Run(ctx)
}
