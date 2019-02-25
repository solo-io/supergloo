package iutil

import (
	"fmt"

	"gopkg.in/AlecAivazis/survey.v1"
)

func GetStringInput(msg string, value *string, v survey.Validator) error {
	prompt := &survey.Input{Message: msg}
	if err := survey.AskOne(prompt, value, v); err != nil {
		return err
	}
	return nil
}

func ChooseFromList(message string, choice *string, options []string) error {

	question := &survey.Select{
		Message: message,
		Options: options,
	}

	if err := survey.AskOne(question, choice, survey.Required); err != nil {
		// this should not error
		fmt.Println("error with input")
		return err
	}

	return nil
}
