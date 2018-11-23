package nsutil

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
)

// If we are selecting resources by their name and the namespace in which they
// are installed, displayName and displayNamespace are identical to the
// resourceRef. However, meshes are selected by the ns in which they were
// installed, so we need both representations
// NOTE: if we add select helper utils for other resources we should make a
// general "select by resource ref" util
type ResSelect struct {
	displayName      string
	displayNamespace string
	resourceRef      options.ResourceRef
}

type ResMap map[string]ResSelect

func generateMeshSelectOptions(nsrMap options.NsResourceMap) ([]string, ResMap) {

	var meshOptions []string
	// map the key to the mesh select object
	// key is namespace, name
	meshMap := make(ResMap)

	for installNs, nsr := range nsrMap {
		for _, meshRef := range nsr.MeshesByInstallNs {
			selectMenuString := fmt.Sprintf("%v, %v", installNs, meshRef.Name)
			meshOptions = append(meshOptions, selectMenuString)
			meshMap[selectMenuString] = ResSelect{
				displayName:      meshRef.Name,
				displayNamespace: installNs,
				resourceRef: options.ResourceRef{
					Name:      meshRef.Name,
					Namespace: meshRef.Namespace,
				},
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
			selectMenuString := fmt.Sprintf("%v, %v", namespace, secret)
			secretOptions = append(secretOptions, selectMenuString)
			secretMap[selectMenuString] = ResSelect{
				displayName:      secret,
				displayNamespace: namespace,
				resourceRef: options.ResourceRef{
					Name:      secret,
					Namespace: namespace,
				},
			}
		}
	}
	return secretOptions, secretMap
}
