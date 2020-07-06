package validationutils

import (
	"github.com/gogo/protobuf/types"
	"github.com/rotisserie/eris"
)

var (
	InvalidRetryPolicyNumAttempts = func(n int32) error {
		return eris.Errorf("Number of retry attempts must be >= 0, got %d", n)
	}
	InvalidHttpStatus = func(httpStatus int32) error {
		return eris.Errorf("Invalid HTTP status: %d", httpStatus)
	}
	InvalidPercentageError = func(pct float64) error {
		return eris.Errorf("Percentage must be between 0.0 and 100.0 inclusive, got %f", pct)
	}
	DestinationsNotFound = func(selector *smh_core_types.ServiceSelector) error {
		return eris.Errorf("No destinations found with Selector %+v", selector)
	}
	SubsetSelectorNotFound = func(meshService *smh_discovery.MeshService, subsetKey string, subsetValue string) error {
		return eris.Errorf("Subset selector with key: %s, value: %s not found on k8s service of name: %s, namespace: %s",
			subsetKey, subsetValue, meshService.GetName(), meshService.GetNamespace())
	}
	NilDestinationRef = eris.New("Destination reference must be non-nil")
	MinDurationError  = eris.New("Duration must be >= 1 millisecond")
)

// the Validator implements some common
// Validation methods that are shared by translator plugins.
type Validator interface {
	// validate that a duration is of a valid length
	ValidateDuration(duration *types.Duration) error
}

type trafficPolicyValidator struct {

}

func (t *trafficPolicyValidator) validateDuration(duration *types.Duration) error {
	if duration.GetSeconds() < 0 || (duration.GetSeconds() == 0 && duration.GetNanos() < 1000000) {
		return MinDurationError
	}
	return nil
}