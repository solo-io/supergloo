package main

import (
	"context"

	mesh_apiserver "github.com/solo-io/service-mesh-hub/services/mesh-apiserver/pkg"
)

func main() {
	ctx := context.Background()
	mesh_apiserver.Run(ctx)
}
