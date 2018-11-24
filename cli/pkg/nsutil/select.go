package nsutil

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/common"
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

// EnsureSecret validates a meshRef relative to static vs. interactive mode
// If in interactive mode (non-static mode) and a secret is not given, it will prompt the user to choose one
func EnsureMesh(meshRef *options.ResourceRef, opts *options.Options) error {
	if err := validateResourceRefForStaticMode("mesh", meshRef, opts); err != nil {
		return err
	}

	if meshRef.Name == "" || meshRef.Namespace == "" {
		chosenMeshRef, err := ChooseMesh(opts.Cache.NsResources)
		if err != nil {
			return err
		}
		*meshRef = chosenMeshRef
	}
	return nil
}

func ChooseResource(typeName string, nsr options.NsResourceMap) (options.ResourceRef, error) {

	resOptions, resMap := generateCommonResourceSelectOptions(typeName, nsr)
	if len(resOptions) == 0 {
		return options.ResourceRef{}, fmt.Errorf("No %v found. Please create a %v", typeName, typeName)
	}
	question := &survey.Select{
		Message: fmt.Sprintf("Select a %v", typeName),
		Options: resOptions,
	}

	var choice string
	if err := survey.AskOne(question, &choice, survey.Required); err != nil {
		// this should not error
		fmt.Println("error with input")
		return options.ResourceRef{}, err
	}

	return resMap[choice].resourceRef, nil
}

// EnsureCommonResource validates a resRef relative to static vs. interactive mode
// If in interactive mode (non-static mode) and a resourceRef is not given, it will prompt the user to choose one
func EnsureCommonResource(typeName string, resRef *options.ResourceRef, opts *options.Options) error {
	if err := validateResourceRefForStaticMode(typeName, resRef, opts); err != nil {
		return err
	}

	// interactive mode
	if resRef.Name == "" || resRef.Namespace == "" {
		chosenResRef, err := ChooseResource(typeName, opts.Cache.NsResources)
		if err != nil {
			return err
		}
		*resRef = chosenResRef
	}
	return nil
}

func validateResourceRefForStaticMode(typeName string, resRef *options.ResourceRef, opts *options.Options) error {
	if opts.Top.Static {
		// make sure we have a full resource ref
		if resRef.Name == "" {
			return fmt.Errorf("Please provide a %v name", typeName)
		}
		if resRef.Namespace == "" {
			return fmt.Errorf("Please provide a %v namespace", typeName)
		}

		// make sure they chose a valid namespace
		if !common.Contains(opts.Cache.Namespaces, resRef.Namespace) {
			return fmt.Errorf("Please specify a valid namespace. Namespace %v not found.", resRef.Namespace)
		}

		// make sure that the particular resource exists in the specified namespace
		switch typeName {
		case "mesh":
			if !common.Contains(opts.Cache.NsResources[resRef.Namespace].Meshes, resRef.Name) {
				return fmt.Errorf("Please specify a valid %v name. %v not found in namespace %v.", resRef.Name, typeName, resRef.Namespace)
			}
		case "secret":
			if !common.Contains(opts.Cache.NsResources[resRef.Namespace].Secrets, resRef.Name) {
				return fmt.Errorf("Please specify a valid %v name. %v not found in namespace %v.", resRef.Name, typeName, resRef.Namespace)
			}
		case "upstream":
			if !common.Contains(opts.Cache.NsResources[resRef.Namespace].Upstreams, resRef.Name) {
				return fmt.Errorf("Please specify a valid %v name. %v not found in namespace %v.", resRef.Name, typeName, resRef.Namespace)
			}
		default:
			panic(fmt.Errorf("typename %v not recognized", typeName))
		}
	}
	return nil
}
