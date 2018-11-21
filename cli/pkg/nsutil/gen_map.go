package nsutil

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
)

type MeshSelect struct {
	meshName      string
	meshNamespace string
}

type MeshMap map[string]MeshSelect

func generateSelectOptions(nsrMap options.NsResourceMap) ([]string, MeshMap) {

	var meshOptions []string
	// map the key to the mesh select object
	// key is namespace, name
	meshMap := make(MeshMap)

	for namespace, nsr := range nsrMap {
		for _, mesh := range nsr.Meshes {
			displayName := fmt.Sprintf("%v, %v", namespace, mesh)
			meshOptions = append(meshOptions, displayName)
			meshMap[displayName] = MeshSelect{
				meshName:      mesh,
				meshNamespace: namespace,
			}
		}
	}
	return meshOptions, meshMap
}
