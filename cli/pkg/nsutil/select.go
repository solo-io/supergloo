package nsutil

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"gopkg.in/AlecAivazis/survey.v1"
)

func ChooseMesh(nsr options.NsResourceMap) (string, string, error) {

	meshOptions, meshMap := generateSelectOptions(nsr)
	question := &survey.Select{
		Message: "Select a mesh",
		Options: meshOptions,
	}

	var choice string
	if err := survey.AskOne(question, &choice, survey.Required); err != nil {
		// this should not error
		fmt.Println("error with input")
		return "", "", err
	}

	return meshMap[choice].meshName, meshMap[choice].meshNamespace, nil
}
