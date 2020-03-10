package preprocess

import (
	"context"
	"net/http"

	"github.com/gogo/protobuf/types"
	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/stringutils"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
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
	DestinationsNotFound = func(selector *core_types.Selector) error {
		return eris.Errorf("No destinations found with Selector %+v", selector)
	}
	SubsetSelectorNotFound = func(meshService *discovery_v1alpha1.MeshService, subsetKey string, subsetValue string) error {
		return eris.Errorf("Subset selector with key: %s, value: %s not found on k8s service of name: %s, namespace: %s",
			subsetKey, subsetValue, meshService.GetName(), meshService.GetNamespace())
	}
	NilDestinationRef          = eris.New("Destination reference must be non-nil")
	InvalidDestinationSelector = eris.New("Destination must be selected by either (labels, namespaces, cluster) or references, encountered both")
	MinDurationError           = eris.New("Duration must be >= 1 millisecond")
)

type trafficPolicyValidator struct {
	meshServiceClient   zephyr_discovery.MeshServiceClient
	meshServiceSelector MeshServiceSelector
}

func NewTrafficPolicyValidator(
	meshServiceClient zephyr_discovery.MeshServiceClient,
	meshServiceSelector MeshServiceSelector,
) TrafficPolicyValidator {
	return &trafficPolicyValidator{
		meshServiceClient:   meshServiceClient,
		meshServiceSelector: meshServiceSelector,
	}
}

func (t *trafficPolicyValidator) Validate(ctx context.Context, trafficPolicy *networking_v1alpha1.TrafficPolicy) error {
	var multiErr *multierror.Error
	if err := t.validateDestination(ctx, trafficPolicy.Spec.GetDestinationSelector()); err != nil {
		multiErr = multierror.Append(multiErr, eris.Wrap(err, "Error found in Destination"))
	}
	if err := t.validateTrafficShift(ctx, trafficPolicy.Spec.GetTrafficShift()); err != nil {
		multiErr = multierror.Append(multiErr, eris.Wrap(err, "Error found in TrafficShift"))
	}
	if err := t.validateFaultInjection(trafficPolicy.Spec.GetFaultInjection()); err != nil {
		multiErr = multierror.Append(multiErr, eris.Wrap(err, "Error found in FaultInjection"))
	}
	if err := t.validateRequestTimeout(trafficPolicy.Spec.GetRequestTimeout()); err != nil {
		multiErr = multierror.Append(multiErr, eris.Wrap(err, "Error found in RequestTimeout"))
	}
	if err := t.validateRetryPolicy(trafficPolicy.Spec.GetRetries()); err != nil {
		multiErr = multierror.Append(multiErr, eris.Wrap(err, "Error found in RetryPolicy"))
	}
	if err := t.validateCorsPolicy(trafficPolicy.Spec.GetCorsPolicy()); err != nil {
		multiErr = multierror.Append(multiErr, eris.Wrap(err, "Error found in CorsPolicy"))
	}
	if err := t.validateMirror(ctx, trafficPolicy.Spec.GetMirror()); err != nil {
		multiErr = multierror.Append(multiErr, eris.Wrap(err, "Error found in Mirror"))
	}
	return multiErr.ErrorOrNil()
}

func (t *trafficPolicyValidator) validateDestination(ctx context.Context, selector *core_types.Selector) error {
	if selector == nil {
		return NilDestinationRef
	}
	if (selector.GetLabels() != nil || selector.GetNamespaces() != nil || selector.GetCluster().GetValue() != "") &&
		selector.GetRefs() != nil {
		return InvalidDestinationSelector
	} else if selector.GetRefs() != nil {
		// select by refs
		for _, ref := range selector.GetRefs() {
			_, err := t.validateKubeService(ctx, ref)
			if err != nil {
				return err
			}
		}
	} else {
		meshServices, err := t.meshServiceSelector.GetMatchingMeshServices(ctx, selector)
		if err != nil {
			return err
		}
		if len(meshServices) < 1 {
			return DestinationsNotFound(selector)
		}
	}
	return nil
}

// Validate that the TrafficShift destination k8s Service exist
// and if subsets are specified, that they exist on the k8s Service
func (t *trafficPolicyValidator) validateTrafficShift(ctx context.Context, trafficShift *networking_types.MultiDestination) error {
	if trafficShift == nil {
		return nil
	}
	for _, destination := range trafficShift.GetDestinations() {
		meshService, err := t.validateKubeService(ctx, destination.GetDestination())
		if err != nil {
			return err
		}
		if destination.GetSubset() != nil {
			err := t.validateSubsetSelectors(meshService, destination.GetSubset())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *trafficPolicyValidator) validateMirror(ctx context.Context, mirror *networking_types.Mirror) error {
	if mirror == nil {
		return nil
	}
	_, err := t.validateKubeService(ctx, mirror.GetDestination())
	if err != nil {
		return err
	}
	return t.validatePercentage(mirror.GetPercentage())
}

func (t *trafficPolicyValidator) validateRequestTimeout(requestTimeout *types.Duration) error {
	if requestTimeout == nil {
		return nil
	}
	return t.validateDuration(requestTimeout)
}

func (t *trafficPolicyValidator) validateFaultInjection(faultInjection *networking_types.FaultInjection) error {
	if faultInjection == nil {
		return nil
	}
	err := t.validatePercentage(faultInjection.GetPercentage())
	if err != nil {
		return err
	}
	switch injectionType := faultInjection.GetFaultInjectionType().(type) {
	case *networking_types.FaultInjection_Abort_:
		abort := faultInjection.GetAbort()
		switch abortType := abort.GetErrorType().(type) {
		case *networking_types.FaultInjection_Abort_HttpStatus:
			return t.validateHttpStatus(abort.GetHttpStatus())
		default:
			return eris.Errorf("TrafficPolicy.Spec.FaultInjection.Abort.ErrorType has unexpected type %T", abortType)
		}
	case *networking_types.FaultInjection_Delay_:
		delay := faultInjection.GetDelay()
		switch delayType := delay.GetHttpDelayType().(type) {
		case *networking_types.FaultInjection_Delay_FixedDelay:
			return t.validateDuration(delay.GetFixedDelay())
		case *networking_types.FaultInjection_Delay_ExponentialDelay:
			return t.validateDuration(delay.GetExponentialDelay())
		default:
			return eris.Errorf("TrafficPolicy.Spec.FaultInjection.Delay.HTTPDelayType has unexpected type %T", delayType)
		}
	default:
		return eris.Errorf("TrafficPolicy.Spec.FaultInjection.FaultInjectionType has unexpected type %T", injectionType)
	}
}

func (t *trafficPolicyValidator) validateRetryPolicy(retryPolicy *networking_types.RetryPolicy) error {
	if retryPolicy == nil {
		return nil
	}
	if retryPolicy.GetAttempts() < 0 {
		return InvalidRetryPolicyNumAttempts(retryPolicy.GetAttempts())
	}
	return t.validateDuration(retryPolicy.GetPerTryTimeout())
}

func (t *trafficPolicyValidator) validateCorsPolicy(corsPolicy *networking_types.CorsPolicy) error {
	if corsPolicy == nil {
		return nil
	}
	return t.validateDuration(corsPolicy.GetMaxAge())
}

func (t *trafficPolicyValidator) validateHttpStatus(httpStatus int32) error {
	if http.StatusText(int(httpStatus)) == "" {
		return InvalidHttpStatus(httpStatus)
	}
	return nil
}

func (t *trafficPolicyValidator) validatePercentage(percentage float64) error {
	if !(0 <= percentage && percentage <= 100) {
		return InvalidPercentageError(percentage)
	}
	return nil
}

func (t *trafficPolicyValidator) validateDuration(duration *types.Duration) error {
	if duration.GetSeconds() < 0 || (duration.GetSeconds() == 0 && duration.GetNanos() < 1000000) {
		return MinDurationError
	}
	return nil
}

func (t *trafficPolicyValidator) validateKubeService(
	ctx context.Context,
	ref *core_types.ResourceRef,
) (*discovery_v1alpha1.MeshService, error) {
	if ref == nil {
		return nil, NilDestinationRef
	}
	meshService, err := t.meshServiceSelector.GetBackingMeshService(ctx, ref.GetName(), ref.GetNamespace(), ref.GetCluster().GetValue())
	if err != nil {
		return nil, err
	}
	return meshService, nil
}

func (t *trafficPolicyValidator) validateSubsetSelectors(
	meshService *discovery_v1alpha1.MeshService,
	subsetSelectors map[string]string,
) error {
	for subsetKey, subsetValue := range subsetSelectors {
		values, keyExists := meshService.Spec.GetSubsets()[subsetKey]
		found := stringutils.ContainsString(subsetValue, values.GetValues())
		if !keyExists || !found {
			return SubsetSelectorNotFound(meshService, subsetKey, subsetValue)
		}
	}
	return nil
}
