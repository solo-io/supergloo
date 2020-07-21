package validation

import (
	"context"
	"sort"

	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/selectorutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
)

// the validator validates user-applied configuration
// and produces a snapshot that is ready for translation (i.e. with accepted policies applied to all the Status of all targeted MeshServices)
// Note that the Validator also updates the statuses of objects contained in the input Snapshot.
// The Input Snapshot's SyncStatuses method should usually be called
// after running the validator.
type Validator interface {
	Validate(ctx context.Context, input input.Snapshot)
}

type validator struct {
	// the validator runs the istioTranslator in order to detect & report translation errors
	istioTranslator istio.Translator
}

func NewValidator(
	istioTranslator istio.Translator,
) Validator {
	return &validator{
		istioTranslator: istioTranslator,
	}
}

func (v *validator) Validate(ctx context.Context, input input.Snapshot) {
	ctx = contextutils.WithLogger(ctx, "validation")
	reporter := newValidationReporter()
	for _, meshService := range input.MeshServices().List() {
		appliedTrafficPolicies := getAppliedTrafficPolicies(input.TrafficPolicies().List(), meshService)
		meshService.Status.AppliedTrafficPolicies = appliedTrafficPolicies

		appliedAccessPolicies := getAppliedAccessPolicies(input.AccessPolicies().List(), meshService)
		meshService.Status.AppliedAccessPolicies = appliedAccessPolicies
	}
	_, err := v.istioTranslator.Translate(ctx, input, reporter)
	if err != nil {
		// should never happen
		contextutils.LoggerFrom(ctx).Errorf("internal error: failed to run istio translator: %v", err)
	}

	// write traffic policy statuses
	for _, trafficPolicy := range input.TrafficPolicies().List() {
		trafficPolicy.Status = v1alpha2.TrafficPolicyStatus{
			ObservedGeneration: trafficPolicy.Generation,
			MeshServices:       map[string]*v1alpha2.ValidationStatus{},
		}
	}

	for _, meshService := range input.MeshServices().List() {
		meshService.Status.ObservedGeneration = meshService.Generation
		meshService.Status.AppliedTrafficPolicies = validateAndReturnAcceptedTrafficPolicies(ctx, input, reporter, meshService)
	}

	// TODO(ilackarms): validate meshworkloads and meshes

}

// this function both validates the status of TrafficPolicies (sets error or accepted state)
// as well as returns a list of accepted traffic policies for the mesh service status
func validateAndReturnAcceptedTrafficPolicies(ctx context.Context, input input.Snapshot, reporter *validationReporter, meshService *discoveryv1alpha2.MeshService) []*discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy {
	var validatedTrafficPolicies []*discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy

	// track accepted index
	var acceptedIndex uint32
	for _, appliedTrafficPolicy := range meshService.Status.AppliedTrafficPolicies {
		errsForTrafficPolicy := reporter.getTrafficPolicyErrors(meshService, appliedTrafficPolicy.Ref)

		trafficPolicy, err := input.TrafficPolicies().Find(appliedTrafficPolicy.Ref)
		if err != nil {
			// should never happen
			contextutils.LoggerFrom(ctx).Errorf("internal error: failed to look up applied traffic policy %v: %v", appliedTrafficPolicy.Ref, err)
			continue
		}

		if len(errsForTrafficPolicy) == 0 {
			trafficPolicy.Status.MeshServices[sets.Key(meshService)] = &v1alpha2.ValidationStatus{
				AcceptanceOrder: acceptedIndex,
				State:           v1alpha2.ValidationState_ACCEPTED,
			}
			validatedTrafficPolicies = append(validatedTrafficPolicies, appliedTrafficPolicy)
			acceptedIndex++
		} else {
			var errMsgs []string
			for _, tpErr := range errsForTrafficPolicy {
				errMsgs = append(errMsgs, tpErr.Error())
			}
			trafficPolicy.Status.MeshServices[sets.Key(meshService)] = &v1alpha2.ValidationStatus{
				State:  v1alpha2.ValidationState_INVALID,
				Errors: errMsgs,
			}
		}
	}

	return validatedTrafficPolicies
}

// the validation reporter uses istioTranslator reports to
// validate individual policies
type validationReporter struct {
	// NOTE(ilackarms): map access should be synchronous (called in a single context),
	// so locking should not be necessary.
	invalidTrafficPolicies map[*discoveryv1alpha2.MeshService]map[string][]error
}

func newValidationReporter() *validationReporter {
	return &validationReporter{invalidTrafficPolicies: map[*discoveryv1alpha2.MeshService]map[string][]error{}}
}

// mark the policy with an error; will be used to filter the policy out of
// the accepted status later
func (v *validationReporter) ReportTrafficPolicy(meshService *discoveryv1alpha2.MeshService, trafficPolicy ezkube.ResourceId, err error) {
	invalidTrafficPoliciesForMeshService := v.invalidTrafficPolicies[meshService]
	if invalidTrafficPoliciesForMeshService == nil {
		invalidTrafficPoliciesForMeshService = map[string][]error{}
	}
	key := sets.Key(trafficPolicy)
	errs := invalidTrafficPoliciesForMeshService[key]
	errs = append(errs, err)
	invalidTrafficPoliciesForMeshService[key] = errs
	v.invalidTrafficPolicies[meshService] = invalidTrafficPoliciesForMeshService
}

func (v *validationReporter) ReportAccessPolicy(meshService *discoveryv1alpha2.MeshService, accessPolicy ezkube.ResourceId, err error) {
	// TODO(ilackarms):
	panic("implement me")
}

func (v *validationReporter) ReportVirtualMesh(mesh *discoveryv1alpha2.Mesh, virtualMesh ezkube.ResourceId, err error) {
	// TODO(ilackarms):
	panic("implement me")
}

func (v *validationReporter) getTrafficPolicyErrors(meshService *discoveryv1alpha2.MeshService, trafficPolicy ezkube.ResourceId) []error {
	invalidTrafficPoliciesForMeshService, ok := v.invalidTrafficPolicies[meshService]
	if !ok {
		return nil
	}
	tpErrors, ok := invalidTrafficPoliciesForMeshService[sets.Key(trafficPolicy)]
	if !ok {
		return nil
	}
	return tpErrors
}

func getAppliedTrafficPolicies(
	trafficPolicies v1alpha2.TrafficPolicySlice,
	meshService *discoveryv1alpha2.MeshService,
) []*discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy {
	var matchingTrafficPolicies v1alpha2.TrafficPolicySlice
	for _, policy := range trafficPolicies {
		if selectorutils.SelectorMatchesService(policy.Spec.DestinationSelector, meshService) {
			matchingTrafficPolicies = append(matchingTrafficPolicies, policy)
		}
	}

	sortTrafficPoliciesByAcceptedDate(meshService, matchingTrafficPolicies)

	var appliedPolicies []*discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy
	for _, policy := range matchingTrafficPolicies {
		policy := policy // pike
		appliedPolicies = append(appliedPolicies, &discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy{
			Ref:                ezkube.MakeObjectRef(policy),
			Spec:               &policy.Spec,
			ObservedGeneration: policy.Generation,
		})
	}
	return appliedPolicies
}

// sort the set of traffic policies in the order in which they were accepted.
// Traffic policies which were accepted first and have not changed (i.e. their observedGeneration is up-to-date) take precedence.
// Next are policies that were previously accepted but whose observedGeneration is out of date. This permits policies which were modified but formerly correct to maintain
// their acceptance status ahead of policies which were unomdified and previously rejected.
// Next will be the policies which have been modified and rejected.
// Finally, policies which are rejected and modified
func sortTrafficPoliciesByAcceptedDate(meshService *discoveryv1alpha2.MeshService, trafficPolicies v1alpha2.TrafficPolicySlice) {
	sort.SliceStable(trafficPolicies, func(i, j int) bool {
		tp1, tp2 := trafficPolicies[i], trafficPolicies[j]

		status1 := tp1.Status.MeshServices[sets.Key(meshService)]
		status2 := tp2.Status.MeshServices[sets.Key(meshService)]

		if status2 == nil {
			// if status is not set, the traffic policy is "pending" for this mesh service
			// and should get sorted after an accepted status.
			return status1 != nil
		}

		switch {
		case status1.State == v1alpha2.ValidationState_ACCEPTED:
			if status2.State != v1alpha2.ValidationState_ACCEPTED {
				// accepted comes before non accepted
				return true
			}

			if tp1UpToDate := isUpToDate(tp1); tp1UpToDate != isUpToDate(tp2) {
				// up to date is validated before modified
				return tp1UpToDate
			}

			// sort by the previous acceptance order
			return status1.AcceptanceOrder < status2.AcceptanceOrder
		case status2.State == v1alpha2.ValidationState_ACCEPTED:
			// accepted comes before non accepted
			return false
		default:
			// neither policy has been accepted, we can simply sort by unique key
			return sets.Key(tp1) < sets.Key(tp2)
		}
	})
}

func isUpToDate(tp *v1alpha2.TrafficPolicy) bool {
	return tp.Status.ObservedGeneration == tp.Generation
}

// Fetch all AccessPolicies applicable to the given MeshService.
// Sorting is not needed because the additive semantics of AccessPolicies does not allow for conflicts.
func getAppliedAccessPolicies(
	accessPolicies v1alpha2.AccessPolicySlice,
	meshService *discoveryv1alpha2.MeshService,
) []*discoveryv1alpha2.MeshServiceStatus_AppliedAccessPolicy {
	var matchingAccessPolicies v1alpha2.AccessPolicySlice
	for _, policy := range accessPolicies {
		if selectorutils.SelectorMatchesService(policy.Spec.DestinationSelector, meshService) {
			matchingAccessPolicies = append(matchingAccessPolicies, policy)
		}
	}

	var appliedPolicies []*discoveryv1alpha2.MeshServiceStatus_AppliedAccessPolicy
	for _, policy := range matchingAccessPolicies {
		policy := policy // pike
		appliedPolicies = append(appliedPolicies, &discoveryv1alpha2.MeshServiceStatus_AppliedAccessPolicy{
			Ref:                ezkube.MakeObjectRef(policy),
			Spec:               &policy.Spec,
			ObservedGeneration: policy.Generation,
		})
	}
	return appliedPolicies
}
