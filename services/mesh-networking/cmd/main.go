package main

import (
	"context"

	mesh_networking "github.com/solo-io/mesh-projects/services/mesh-networking"
)

func main() {
	ctx := context.Background()
	mesh_networking.Run(ctx)
}
