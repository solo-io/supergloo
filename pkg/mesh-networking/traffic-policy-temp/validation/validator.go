package traffic_policy_validation

import (
	"net/http"

	"github.com/gogo/protobuf/types"
	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/stringutils"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
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
	DestinationNotFound = func(ref *smh_core_types.ResourceRef) error {
		return eris.Errorf("No destinations found with ref %s.%s.%s", ref.Name, ref.Namespace, ref.Cluster)
	}
	SubsetSelectorNotFound = func(meshService *smh_discovery.MeshService, subsetKey string, subsetValue string) error {
		return eris.Errorf("Subset selector with key: %s, value: %s not found on k8s service of name: %s, namespace: %s",
			subsetKey, subsetValue, meshService.GetName(), meshService.GetNamespace())
	}
	NilDestinationRef                          = eris.New("Destination reference must be non-nil")
	MinDurationError                           = eris.New("Duration must be >= 1 millisecond")
	OutlierDetectionWithNonEmptySourceSelector = eris.New("OutlierDetection settings require an empty source selector.")
)

func NewValidator(
	resourceSelector selection.BaseResourceSelector,
) Validator {
	return &validator{
		resourceSelector: resourceSelector,
	}
}

type validator struct {
	resourceSelector selection.BaseResourceSelector
}

func (v *validator) ValidateTrafficPolicy(trafficPolicy *smh_networking.TrafficPolicy, allMeshServices []*smh_discovery.MeshService) (*smh_core_types.Status, error) {
	var multiErr *multierror.Error
	if err := v.validateDestination(allMeshServices, trafficPolicy.Spec.GetDestinationSelector()); err != nil {
		multiErr = multierror.Append(multiErr, eris.Wrap(err, "Error found in Destination"))
	}
	if err := v.validateTrafficShift(allMeshServices, trafficPolicy.Spec.GetTrafficShift()); err != nil {
		multiErr = multierror.Append(multiErr, eris.Wrap(err, "Error found in TrafficShift"))
	}
	if err := v.validateFaultInjection(trafficPolicy.Spec.GetFaultInjection()); err != nil {
		multiErr = multierror.Append(multiErr, eris.Wrap(err, "Error found in FaultInjection"))
	}
	if err := v.validateRequestTimeout(trafficPolicy.Spec.GetRequestTimeout()); err != nil {
		multiErr = multierror.Append(multiErr, eris.Wrap(err, "Error found in RequestTimeout"))
	}
	if err := v.validateRetryPolicy(trafficPolicy.Spec.GetRetries()); err != nil {
		multiErr = multierror.Append(multiErr, eris.Wrap(err, "Error found in RetryPolicy"))
	}
	if err := v.validateCorsPolicy(trafficPolicy.Spec.GetCorsPolicy()); err != nil {
		multiErr = multierror.Append(multiErr, eris.Wrap(err, "Error found in CorsPolicy"))
	}
	if err := v.validateMirror(allMeshServices, trafficPolicy.Spec.GetMirror()); err != nil {
		multiErr = multierror.Append(multiErr, eris.Wrap(err, "Error found in Mirror"))
	}
	if err := v.validateOutlierDetection(trafficPolicy.Spec.GetSourceSelector(), trafficPolicy.Spec.GetOutlierDetection()); err != nil {
		multiErr = multierror.Append(multiErr, eris.Wrap(err, "Error found in OutlierDetection"))
	}
	validationErr := multiErr.ErrorOrNil()
	if validationErr == nil {
		return &smh_core_types.Status{
			State: smh_core_types.Status_ACCEPTED,
		}, nil
	} else {
		return &smh_core_types.Status{
			State:   smh_core_types.Status_INVALID,
			Message: validationErr.Error(),
		}, validationErr
	}
}

func (v *validator) validateDestination(allServices []*smh_discovery.MeshService, selector *smh_core_types.ServiceSelector) error {
	if selector == nil {
		return nil
	}
	_, err := v.resourceSelector.FilterMeshServicesByServiceSelector(allServices, selector)
	if err != nil {
		return err
	}
	return nil
}

// Validate that the TrafficShift destination k8s Service exist
// and if subsets are specified, that they exist on the k8s Service
func (v *validator) validateTrafficShift(services []*smh_discovery.MeshService, trafficShift *smh_networking_types.TrafficPolicySpec_MultiDestination) error {
	if trafficShift == nil {
		return nil
	}

	destinationWeightSum := uint32(0)
	for _, destination := range trafficShift.GetDestinations() {
		meshService, err := v.validateKubeService(services, destination.GetDestination())
		if err != nil {
			return err
		}
		if destination.GetSubset() != nil {
			err := v.validateSubsetSelectors(meshService, destination.GetSubset())
			if err != nil {
				return err
			}
		}
		destinationWeightSum += destination.Weight
	}

	// https://github.com/banzaicloud/istio-client-go/blob/98a770729d7b3a60725b53f90318b411cf1cee78/pkg/networking/v1alpha3/virtualservice_types.go#L526
	// if there is more than one destination, the weights must be provided and must add to 100
	if len(trafficShift.GetDestinations()) > 1 && destinationWeightSum != 100 {
		return eris.New("Traffic shift destination weights must add to 100")
	}
	return nil
}

func (v *validator) validateMirror(services []*smh_discovery.MeshService, mirror *smh_networking_types.TrafficPolicySpec_Mirror) error {
	if mirror == nil {
		return nil
	}
	_, err := v.validateKubeService(services, mirror.GetDestination())
	if err != nil {
		return err
	}
	return v.validatePercentage(mirror.GetPercentage())
}

func (v *validator) validateRequestTimeout(requestTimeout *types.Duration) error {
	if requestTimeout == nil {
		return nil
	}
	return v.validateDuration(requestTimeout)
}

func (v *validator) validateFaultInjection(faultInjection *smh_networking_types.TrafficPolicySpec_FaultInjection) error {
	if faultInjection == nil {
		return nil
	}
	err := v.validatePercentage(faultInjection.GetPercentage())
	if err != nil {
		return err
	}
	switch injectionType := faultInjection.GetFaultInjectionType().(type) {
	case *smh_networking_types.TrafficPolicySpec_FaultInjection_Abort_:
		abort := faultInjection.GetAbort()
		switch abortType := abort.GetErrorType().(type) {
		case *smh_networking_types.TrafficPolicySpec_FaultInjection_Abort_HttpStatus:
			return v.validateHttpStatus(abort.GetHttpStatus())
		default:
			return eris.Errorf("TrafficPolicy.Spec.FaultInjection.Abort.ErrorType has unexpected type %T", abortType)
		}
	case *smh_networking_types.TrafficPolicySpec_FaultInjection_Delay_:
		delay := faultInjection.GetDelay()
		switch delayType := delay.GetHttpDelayType().(type) {
		case *smh_networking_types.TrafficPolicySpec_FaultInjection_Delay_FixedDelay:
			return v.validateDuration(delay.GetFixedDelay())
		case *smh_networking_types.TrafficPolicySpec_FaultInjection_Delay_ExponentialDelay:
			return v.validateDuration(delay.GetExponentialDelay())
		default:
			return eris.Errorf("TrafficPolicy.Spec.FaultInjection.Delay.HTTPDelayType has unexpected type %T", delayType)
		}
	default:
		return eris.Errorf("TrafficPolicy.Spec.FaultInjection.FaultInjectionType has unexpected type %T", injectionType)
	}
}

func (v *validator) validateRetryPolicy(retryPolicy *smh_networking_types.TrafficPolicySpec_RetryPolicy) error {
	if retryPolicy == nil {
		return nil
	}
	if retryPolicy.GetAttempts() < 0 {
		return InvalidRetryPolicyNumAttempts(retryPolicy.GetAttempts())
	}
	return v.validateDuration(retryPolicy.GetPerTryTimeout())
}

func (v *validator) validateCorsPolicy(corsPolicy *smh_networking_types.TrafficPolicySpec_CorsPolicy) error {
	if corsPolicy == nil {
		return nil
	}
	return v.validateDuration(corsPolicy.GetMaxAge())
}

func (v *validator) validateHttpStatus(httpStatus int32) error {
	if http.StatusText(int(httpStatus)) == "" {
		return InvalidHttpStatus(httpStatus)
	}
	return nil
}

func (v *validator) validatePercentage(percentage float64) error {
	if !(0 <= percentage && percentage <= 100) {
		return InvalidPercentageError(percentage)
	}
	return nil
}

// Return error if duration < 1ms
func (v *validator) validateDuration(duration *types.Duration) error {
	if duration.GetSeconds() < 0 || (duration.GetSeconds() == 0 && duration.GetNanos() < 1000000) {
		return MinDurationError
	}
	return nil
}

func (v *validator) validateKubeService(
	services []*smh_discovery.MeshService,
	ref *smh_core_types.ResourceRef,
) (*smh_discovery.MeshService, error) {
	if ref == nil {
		return nil, NilDestinationRef
	}
	meshService := v.resourceSelector.FindMeshServiceByRefSelector(services, ref.GetName(), ref.GetNamespace(), ref.GetCluster())

	if meshService == nil {
		return nil, DestinationNotFound(ref)
	}
	return meshService, nil
}

func (v *validator) validateSubsetSelectors(
	meshService *smh_discovery.MeshService,
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

// If OutlierDetection is set, source selector must be nil because outlier detection applies
// to all incoming traffic.
func (v *validator) validateOutlierDetection(
	sourceSelector *smh_core_types.WorkloadSelector,
	outlierDetection *smh_networking_types.TrafficPolicySpec_OutlierDetection,
) error {
	var err error
	if outlierDetection == nil {
		return nil
	}
	if sourceSelector != nil {
		return OutlierDetectionWithNonEmptySourceSelector
	}
	if outlierDetection.GetConsecutiveErrors() < 1 {
		return eris.Errorf(
			"Invalid OutlierDetection consecutive errors: %d, must be > 0",
			outlierDetection.GetConsecutiveErrors(),
		)
	}
	if outlierDetection.GetInterval() != nil {
		if err = v.validateDuration(outlierDetection.GetInterval()); err != nil {
			return eris.Wrap(err, "Invalid OutlierDetection interval")
		}
	}
	if outlierDetection.GetBaseEjectionTime() != nil {
		if err = v.validateDuration(outlierDetection.GetBaseEjectionTime()); err != nil {
			return eris.Wrap(err, "Invalid OutlierDetection base ejection time")
		}
	}
	return nil
}
