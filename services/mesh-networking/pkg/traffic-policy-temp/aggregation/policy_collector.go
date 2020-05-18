package traffic_policy_aggregation

import (
	"sort"

	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	mesh_translation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/meshes"
	"k8s.io/apimachinery/pkg/util/sets"
)

func NewPolicyCollector(trafficPolicyAggregator Aggregator) PolicyCollector {
	return &policyCollector{
		trafficPolicyAggregator: trafficPolicyAggregator,
	}
}

type policyCollector struct {
	trafficPolicyAggregator Aggregator
}

func (p *policyCollector) CollectForService(
	meshService *zephyr_discovery.MeshService,
	mesh *zephyr_discovery.Mesh,
	translationValidator mesh_translation.TranslationValidator,
	allTrafficPolicies []*zephyr_networking.TrafficPolicy,
) (
	policiesToRecordOnService []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
	policyToConflictErrors map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_ConflictError,
	policyToTranslatorErrors map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_TranslatorError,
	err error,
) {
	allTrafficPolicyIds, _, uniqueStringToNewlyValidatedTrafficPolicy, policiesForService, err := p.aggregateTrafficPolicies(
		meshService,
		allTrafficPolicies,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	// build a summary of existing valid policies with their new state and last-known good state
	anyPolicyChangedSinceLastReconcile, policiesToCheck := p.buildPoliciesToCheck(
		meshService,
		allTrafficPolicyIds,
		uniqueStringToNewlyValidatedTrafficPolicy,
		policiesForService,
	)

	// Avoid expensive merge and translation validations if nothing changed from the last reconcile iteration
	if !anyPolicyChangedSinceLastReconcile {
		return meshService.Status.ValidatedTrafficPolicies, nil, nil, nil
	}

	policiesToRecordOnService, policyToConflictErrors, policyToTranslatorErrors = p.determineFinalValidState(
		meshService,
		mesh,
		translationValidator,
		policiesToCheck,
	)

	return policiesToRecordOnService, policyToConflictErrors, policyToTranslatorErrors, nil
}

func (p *policyCollector) determineFinalValidState(
	meshService *zephyr_discovery.MeshService,
	mesh *zephyr_discovery.Mesh,
	translationValidator mesh_translation.TranslationValidator,
	policiesToCheckParam []*policyToCheck,
) (
	policiesToRecordOnService []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
	policyToConflictErrors map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_ConflictError,
	policyToTranslatorErrors map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_TranslatorError,
) {
	// avoid mutating the input parameter
	policiesToCheck := append([]*policyToCheck(nil), policiesToCheckParam...)

	// We want to sort those entries with a nil `UpdatedTrafficPolicy` field to the BEGINNING of the list.
	// This is significant- we want to ensure that we only mark those policies that have CHANGED since the
	// last reconcile iteration and are NOW in conflict with other policies to be marked as in conflict, not older
	// ones that are unchanged. We accomplish that doing this sort, which ensures that we accept unchanged
	// policies into the `policiesToRecordOnService` list first (which will all pass validation together), then
	// subsequently we process the changed policies, which may fail and then be marked as in conflict.
	sort.Slice(policiesToCheck, func(i, j int) bool {
		// policies[i] is LESS than policies[j] (i.e., should appear before it in the list) if policies[i] was not updated
		// we don't care if policies[j] was updated or not, it'll get sorted at some point too.
		// Per the contract on the declaration of this method in the interface, we don't guarantee an ordering beyond this.
		return policiesToCheck[i].UpdatedTrafficPolicy == nil
	})

	policiesInConflict := map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_ConflictError{}
	untranslatablePolicies := map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_TranslatorError{}
	for _, policyToCheckIter := range policiesToCheck {
		policyToProcess := policyToCheckIter

		// see the notes on the sort.Slice call above; because of that ordering, we know that anything
		// with a nil Updated field must by definition be both merge-able and translate-able with everything
		// ELSE with a nil Updated field, so we can avoid doing expensive merge/translate validations on all of those.
		// That has the added benefit of allowing us to avoid reporting errors on policies that have not changed.
		if policyToProcess.UpdatedTrafficPolicy == nil {
			policiesToRecordOnService = append(policiesToRecordOnService, &zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				Ref:               policyToProcess.Ref,
				TrafficPolicySpec: policyToProcess.LastKnownGoodSpecState,
			})
			continue
		}

		var toValidate *zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy
		if policyToProcess.UpdatedTrafficPolicy == nil {
			toValidate = &zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				Ref:               policyToProcess.Ref,
				TrafficPolicySpec: policyToProcess.LastKnownGoodSpecState,
			}
		} else {
			toValidate = &zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				Ref:               policyToProcess.Ref,
				TrafficPolicySpec: &policyToProcess.UpdatedTrafficPolicy.Spec,
			}
		}

		successfullyValidated := &zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
			Ref: policyToProcess.Ref,

			// may be nil, if this is a new policy that we're adding
			TrafficPolicySpec: policyToProcess.LastKnownGoodSpecState,
		}

		var mergeabilityCheckCopy []*zephyr_networking_types.TrafficPolicySpec
		for _, tp := range policiesToRecordOnService {
			// Add everything that we've approved so far (the whole contents of `policiesToRecordOnService`).
			// We don't need to skip any entries in the list in here, since we won't process a traffic policy twice.

			mergeabilityCheckCopy = append(mergeabilityCheckCopy, tp.TrafficPolicySpec)
		}

		mergeConflict := p.trafficPolicyAggregator.FindMergeConflict(toValidate.TrafficPolicySpec, mergeabilityCheckCopy, meshService)
		if mergeConflict != nil {
			policiesInConflict[policyToProcess.UpdatedTrafficPolicy] = append(policiesInConflict[policyToProcess.UpdatedTrafficPolicy], mergeConflict)
		} else {
			// we know they're mergeable now, let's see if they can be translated all together
			mergeableTPs := append([]*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy(nil), policiesToRecordOnService...)
			mergeableTPs = append(mergeableTPs, toValidate)

			translationErrors := translationValidator.GetTranslationErrors(
				meshService,
				mesh,
				mergeableTPs,
			)

			if len(translationErrors) == 0 {
				successfullyValidated = toValidate
			} else {
				for _, translationError := range translationErrors {
					// need to pick out the translation policy in question from the error result list
					if clients.SameObject(clients.ResourceRefToObjectMeta(translationError.Policy.GetRef()), policyToProcess.UpdatedTrafficPolicy.ObjectMeta) {
						untranslatablePolicies[policyToProcess.UpdatedTrafficPolicy] = translationError.TranslatorErrors
					}
				}
			}
		}

		// this field would be nil if we are adding a NEW traffic policy, and there isn't a last-known good state to fall back on
		if successfullyValidated.TrafficPolicySpec != nil {
			policiesToRecordOnService = append(policiesToRecordOnService, successfullyValidated)
		}
	}

	return policiesToRecordOnService, policiesInConflict, untranslatablePolicies
}

func (p *policyCollector) aggregateTrafficPolicies(
	meshService *zephyr_discovery.MeshService,
	allTrafficPolicies []*zephyr_networking.TrafficPolicy,
) (
	allPolicyIds sets.String, // clients.ToUniqueString results for ALL policies, regardless of validation status
	allValidatedTrafficPolicies []*zephyr_networking.TrafficPolicy, // all validated policies
	uniqueStringToValidatedTrafficPolicy map[string]*zephyr_networking.TrafficPolicy, // all validated policies collected into a map
	policiesForService []*zephyr_networking.TrafficPolicy, // just the validated policies that apply to this service
	err error,
) {
	allIds := sets.NewString()
	uniqueStringToValidatedTrafficPolicy = map[string]*zephyr_networking.TrafficPolicy{}

	for _, tpIter := range allTrafficPolicies {
		trafficPolicy := tpIter

		allIds.Insert(clients.ToUniqueSingleClusterString(trafficPolicy.ObjectMeta))
		if trafficPolicy.Status.GetValidationStatus().GetState() == zephyr_core_types.Status_ACCEPTED {
			allValidatedTrafficPolicies = append(allValidatedTrafficPolicies, trafficPolicy)
			uniqueStringToValidatedTrafficPolicy[clients.ToUniqueSingleClusterString(trafficPolicy.ObjectMeta)] = trafficPolicy
		}
	}

	// TODO: Remove list creation here
	serviceWithRelevantPolicies, err := p.trafficPolicyAggregator.GroupByMeshService(allValidatedTrafficPolicies, []*zephyr_discovery.MeshService{meshService})
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return allIds, allValidatedTrafficPolicies, uniqueStringToValidatedTrafficPolicy, serviceWithRelevantPolicies[0].TrafficPolicies, nil
}

// return:
//   - anyPolicyChanged: whether the returned slice differs from the mesh service status at all
//   - policiesToCheck: a summary of policies with their last-known good state that apply to this service, both updated and new policies
//        NOTE: `policiesToCheck` is always the *complete* set of policies applying to the given service
func (*policyCollector) buildPoliciesToCheck(
	meshService *zephyr_discovery.MeshService,
	allTrafficPolicyIds sets.String,
	uniqueStringToNewlyValidatedTrafficPolicy map[string]*zephyr_networking.TrafficPolicy,
	policiesForService []*zephyr_networking.TrafficPolicy,
) (anyPolicyChanged bool, policiesToCheck []*policyToCheck) {
	previouslyRecordedTrafficPolicyIDs := sets.NewString()
	anyPolicyChangedSinceLastReconcile := false

	for _, previouslyRecordedPolicyIter := range meshService.Status.GetValidatedTrafficPolicies() {
		previouslyRecordedPolicy := previouslyRecordedPolicyIter
		identifierString := clients.ToUniqueSingleClusterString(clients.ResourceRefToObjectMeta(previouslyRecordedPolicy.GetRef()))

		// if this traffic policy has been deleted, we need to remove this previously-recorded policy from the output of this reconcile loop
		if !allTrafficPolicyIds.Has(identifierString) {
			anyPolicyChangedSinceLastReconcile = true
			continue
		}

		previouslyRecordedTrafficPolicyIDs.Insert(identifierString)
		newlyValidatedPolicy, newlyValidatedPolicyFound := uniqueStringToNewlyValidatedTrafficPolicy[identifierString]

		var p *policyToCheck
		if !newlyValidatedPolicyFound || newlyValidatedPolicy.Spec.Equal(previouslyRecordedPolicy.TrafficPolicySpec) {
			// if the traffic policy was either 1. not validated, or 2. has not been updated since the last reconcile iteration, keep it the same
			p = &policyToCheck{
				Ref:                    previouslyRecordedPolicy.Ref,
				LastKnownGoodSpecState: previouslyRecordedPolicy.TrafficPolicySpec,
			}
		} else {
			// this policy was updated if a newly-validated state was found
			anyPolicyChangedSinceLastReconcile = anyPolicyChangedSinceLastReconcile || newlyValidatedPolicyFound
			p = &policyToCheck{
				Ref:                    previouslyRecordedPolicy.Ref,
				LastKnownGoodSpecState: previouslyRecordedPolicy.TrafficPolicySpec,
				UpdatedTrafficPolicy:   newlyValidatedPolicy, // this field may be nil if the policy could not be validated
			}
		}

		policiesToCheck = append(policiesToCheck, p)
	}

	// add NEW policies that were not previously recorded
	for _, relevantPolicyIter := range policiesForService {
		relevantPolicy := relevantPolicyIter
		policyId := clients.ToUniqueSingleClusterString(relevantPolicy.ObjectMeta)

		if previouslyRecordedTrafficPolicyIDs.Has(policyId) {
			continue
		}

		anyPolicyChangedSinceLastReconcile = true
		policiesToCheck = append(policiesToCheck, &policyToCheck{
			Ref:                  clients.ObjectMetaToResourceRef(relevantPolicy.ObjectMeta),
			UpdatedTrafficPolicy: relevantPolicy,
		})
	}

	return anyPolicyChangedSinceLastReconcile, policiesToCheck
}

type policyToCheck struct {
	Ref *zephyr_core_types.ResourceRef

	// If this is non-nil, then the traffic policy is both 1. valid, and 2. changed from the last reconcile iteration.
	// That implies that this field will be nil if the policy was invalid.
	UpdatedTrafficPolicy *zephyr_networking.TrafficPolicy

	// if this is non-nil, then we previously recorded a last-known-good state for this policy
	LastKnownGoodSpecState *zephyr_networking_types.TrafficPolicySpec
}
