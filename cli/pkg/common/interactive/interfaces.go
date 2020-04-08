package interactive

//go:generate mockgen -source ./interfaces.go -destination mocks/mocks.go

type Validator func(userInput interface{}) error

type Required func() Validator

type InteractivePrompt interface {
	// Prompt user for a single value that they input. Input is required.
	PromptValue(message, defaultValue string) (string, error)
	// Prompt user for a single value that they input with validation
	PromptValueWithValidator(message, defaultValue string, validator Validator) (string, error)
	// Prompt user to select a single value from a list of options
	SelectValue(message string, options []string) (string, error)
	// Prompt user to select a multiple values from a list of options
	SelectMultipleValues(message string, options []string) ([]string, error)
}
