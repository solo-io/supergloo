package main

import (
	"context"

	mesh_networking "github.com/solo-io/service-mesh-hub/services/mesh-networking"
)

func main() {
	ctx := context.Background()
	mesh_networking.Run(ctx)
}
