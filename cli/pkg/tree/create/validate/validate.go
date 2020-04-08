package validate

import (
	"strconv"
	"strings"

	"github.com/rotisserie/eris"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/validation"
)

// Ensure that user supplied name adheres to DNS subdomain name (RFC1123),
// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-subdomain-names
func K8sName(userInput string) error {
	errStrings := validation.IsDNS1123Subdomain(userInput)
	if len(errStrings) > 0 {
		return eris.New(strings.Join(errStrings, ". "))
	}
	return nil
}

// Validate that labels are supplied as a comma delimited list, and that label keys and values are valid
func Labels(userInput string) error {
	_, err := labels.ConvertSelectorToLabelsMap(userInput)
	return err
}

// Validate that namespaces are supplied as a comma delimited list, and that namespaces are valid k8s names
func Namespaces(userInput string) error {
	// Check if empty
	if strings.TrimSpace(userInput) == "" {
		return nil
	}
	namespaces := strings.Split(userInput, ",")
	var err error
	for _, namespace := range namespaces {
		if err = K8sName(strings.TrimSpace(namespace)); err != nil {
			return err
		}
	}
	return nil
}

func PositiveInteger(userInput string) error {
	intValue, err := strconv.Atoi(userInput)
	if err != nil || intValue < 1 {
		return eris.Errorf("Invalid value: %s. Value must be a positive integer.", userInput)
	}
	return nil
}

func AllowEmpty(validate func(userInput string) error) func(userInput string) error {
	return func(userInput string) error {
		if userInput == "" {
			return nil
		}
		return validate(userInput)
	}
}
