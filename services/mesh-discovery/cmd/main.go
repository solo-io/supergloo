package main

import (
	mesh_discovery "github.com/solo-io/service-mesh-hub/services/mesh-discovery"
)

func main() {
	mesh_discovery.Run(nil)
}
