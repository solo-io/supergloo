package nsutil

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
)

type ResSelect struct {
	name      string
	namespace string
}

type ResMap map[string]ResSelect

func generateMeshSelectOptions(nsrMap options.NsResourceMap) ([]string, ResMap) {

	var meshOptions []string
	// map the key to the mesh select object
	// key is namespace, name
	meshMap := make(ResMap)

	for namespace, nsr := range nsrMap {
		for _, mesh := range nsr.Meshes {
			displayName := fmt.Sprintf("%v, %v", namespace, mesh)
			meshOptions = append(meshOptions, displayName)
			meshMap[displayName] = ResSelect{
				name:      mesh,
				namespace: namespace,
			}
		}
	}
	return meshOptions, meshMap
}

func generateSecretSelectOptions(nsrMap options.NsResourceMap) ([]string, ResMap) {

	var secretOptions []string
	// map the key to the secret select object
	// key is namespace, name
	secretMap := make(ResMap)

	for namespace, nsr := range nsrMap {
		for _, secret := range nsr.Secrets {
			displayName := fmt.Sprintf("%v, %v", namespace, secret)
			secretOptions = append(secretOptions, displayName)
			secretMap[displayName] = ResSelect{
				name:      secret,
				namespace: namespace,
			}
		}
	}
	return secretOptions, secretMap
}
