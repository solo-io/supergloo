package interactive

import (
	"github.com/AlecAivazis/survey/v2"
)

type surveyInteractivePrompt struct{}

func NewSurveyInteractivePrompt() InteractivePrompt {
	return &surveyInteractivePrompt{}
}

func (s *surveyInteractivePrompt) PromptValue(message, defaultValue string) (string, error) {
	return s.PromptValueWithValidator(message, defaultValue, survey.Required)
}

func (s *surveyInteractivePrompt) PromptValueWithValidator(message, defaultValue string, validator Validator) (string, error) {
	value := ""
	prompt := &survey.Input{
		Message: message,
	}
	if defaultValue != "" {
		prompt.Default = defaultValue
	}
	err := survey.AskOne(prompt, &value, survey.WithValidator(ensureValidatorNonNil(validator)))
	if err != nil {
		return "", err
	}
	return value, nil
}

func (s *surveyInteractivePrompt) SelectValue(message string, options []string) (string, error) {
	selection := ""
	prompt := &survey.Select{
		Message: message,
		Options: options,
	}
	err := survey.AskOne(prompt, &selection, nil)
	if err != nil {
		return "", err
	}
	return selection, nil
}

func (s *surveyInteractivePrompt) SelectMultipleValues(message string, options []string) ([]string, error) {
	var selections []string
	prompt := &survey.MultiSelect{
		Message: message,
		Options: options,
	}
	err := survey.AskOne(prompt, &selections, nil)
	if err != nil {
		return nil, err
	}
	return selections, nil
}

func ensureValidatorNonNil(validator Validator) func(ans interface{}) error {
	if validator != nil {
		return validator
	}
	return func(ans interface{}) error {
		return nil
	}
}
