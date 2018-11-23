package nsutil

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"gopkg.in/AlecAivazis/survey.v1"
)

//NOTE these functions are good candidates for code generation

// ChooseMesh allows users to interactively select a mesh
// Options are displayed in the format "<installation_namespace>, <name>" for each mesh
// Selections are returned as a resource ref (and the resource ref namespace may differ from the installation namespace)
func ChooseMesh(nsr options.NsResourceMap) (options.ResourceRef, error) {

	meshOptions, meshMap := generateMeshSelectOptions(nsr)
	if len(meshOptions) == 0 {
		return options.ResourceRef{}, fmt.Errorf("No meshs found. Please create a mesh")
	}

	question := &survey.Select{
		Message: "Select a mesh",
		Options: meshOptions,
	}

	var choice string
	if err := survey.AskOne(question, &choice, survey.Required); err != nil {
		// this should not error
		fmt.Println("error with input")
		return options.ResourceRef{}, err
	}

	return meshMap[choice].resourceRef, nil
}

func ChooseSecret(nsr options.NsResourceMap) (options.ResourceRef, error) {

	secretOptions, secretMap := generateSecretSelectOptions(nsr)
	if len(secretOptions) == 0 {
		return options.ResourceRef{}, fmt.Errorf("No secrets found. Please create a secret")
	}
	question := &survey.Select{
		Message: "Select a secret",
		Options: secretOptions,
	}

	var choice string
	if err := survey.AskOne(question, &choice, survey.Required); err != nil {
		// this should not error
		fmt.Println("error with input")
		return options.ResourceRef{}, err
	}

	return secretMap[choice].resourceRef, nil
}
