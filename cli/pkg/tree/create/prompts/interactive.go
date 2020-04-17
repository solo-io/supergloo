package prompts

import (
	"strings"

	"github.com/solo-io/service-mesh-hub/cli/pkg/common/interactive"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/create/validate"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"k8s.io/apimachinery/pkg/labels"
)

func SelectServiceSelector(
	message string,
	meshServiceNames []string,
	meshServiceNamesToRef map[string]*zephyr_core_types.ResourceRef,
	interactivePrompt interactive.InteractivePrompt,
) (*zephyr_core_types.ServiceSelector, error) {
	selections, err := interactivePrompt.SelectMultipleValues(message, meshServiceNames)
	if err != nil {
		return nil, err
	}
	var selectedTargets []*zephyr_core_types.ResourceRef
	for _, selection := range selections {
		selectedTargets = append(selectedTargets, meshServiceNamesToRef[selection])
	}
	return &zephyr_core_types.ServiceSelector{
		ServiceSelectorType: &zephyr_core_types.ServiceSelector_ServiceRefs_{
			ServiceRefs: &zephyr_core_types.ServiceSelector_ServiceRefs{
				Services: selectedTargets,
			},
		},
	}, nil
}

func PromptLabels(
	message string,
	interactivePrompt interactive.InteractivePrompt,
) (labels.Set, error) {
	selections, err := interactivePrompt.PromptValueWithValidator(message, "", validate.Labels)
	labelSet, err := labels.ConvertSelectorToLabelsMap(selections)
	if err != nil {
		return nil, err
	}
	// Avoid returning empty initialized map, instead return nil
	if len(labelSet) == 0 {
		labelSet = nil
	}
	return labelSet, nil
}

func PromptCommaDelimitedValues(
	message string,
	interactivePrompt interactive.InteractivePrompt,
) ([]string, error) {
	namespaces, err := interactivePrompt.PromptValueWithValidator(message, "", validate.Namespaces)
	if err != nil {
		return nil, err
	}
	values := strings.Split(namespaces, ",")
	var cleanedValues []string
	for _, value := range values {
		cleanedValues = append(cleanedValues, strings.TrimSpace(value))
	}
	return cleanedValues, nil
}
