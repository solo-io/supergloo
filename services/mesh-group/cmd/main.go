package main

import (
	"github.com/solo-io/mesh-projects/services/internal/config"
	mesh_group "github.com/solo-io/mesh-projects/services/mesh-group"
)

func main() {
	ctx := config.CreateRootContext(nil, "mesh-group")
	mesh_group.Run(ctx)
}
