package uninstall

import (
	"gopkg.in/AlecAivazis/survey.v1"
)

func selectMeshByName(names []string) ([]string, error) {
	meshNames := []string{}
	prompt := &survey.MultiSelect{
		Message: "Which meshes would you like to delete?",
		Options: names,
	}
	err := survey.AskOne(prompt, &meshNames, nil)

	return meshNames, err
}

func deleteAllMeshes() (bool, error) {
	var deleteAll bool
	prompt := &survey.Confirm{
		Message: "Delete all meshes?",
	}
	err := survey.AskOne(prompt, &deleteAll, nil)
	return deleteAll, err
}
