package traffic_policy_aggregation

import (
	"context"
	"sort"

	"github.com/rotisserie/eris"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"github.com/solo-io/service-mesh-hub/pkg/enum_conversion"
	"github.com/solo-io/service-mesh-hub/pkg/reconciliation"
	mesh_translation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/meshes"
	"k8s.io/apimachinery/pkg/util/sets"
)

func NewAggregationReconciler(
	trafficPolicyClient zephyr_networking.TrafficPolicyClient,
	meshServiceClient zephyr_discovery.MeshServiceClient,
	meshClient zephyr_discovery.MeshClient,
	aggregator Aggregator,
	translationValidators map[zephyr_core_types.MeshType]mesh_translation.TranslationValidator,
) reconciliation.Reconciler {
	return &aggregationReconciler{
		trafficPolicyClient:   trafficPolicyClient,
		meshServiceClient:     meshServiceClient,
		meshClient:            meshClient,
		aggregator:            aggregator,
		translationValidators: translationValidators,
	}
}

type aggregationReconciler struct {
	trafficPolicyClient   zephyr_networking.TrafficPolicyClient
	meshServiceClient     zephyr_discovery.MeshServiceClient
	meshClient            zephyr_discovery.MeshClient
	aggregator            Aggregator
	translationValidators map[zephyr_core_types.MeshType]mesh_translation.TranslationValidator
}

func (a *aggregationReconciler) GetName() string {
	return "traffic-policy-aggregation"
}

func (a *aggregationReconciler) Reconcile(ctx context.Context) error {
	allTrafficPolicies, err := a.trafficPolicyClient.ListTrafficPolicy(ctx)
	if err != nil {
		return err
	}

	uniqueStringToValidatedTrafficPolicy := map[string]*zephyr_networking.TrafficPolicy{}
	var allValidatedTrafficPolicies []*zephyr_networking.TrafficPolicy
	allTrafficPolicyIds := sets.NewString()
	for _, tpIter := range allTrafficPolicies.Items {
		trafficPolicy := tpIter

		allTrafficPolicyIds.Insert(clients.ToUniqueSingleClusterString(trafficPolicy.ObjectMeta))
		if trafficPolicy.Status.GetValidationStatus().GetState() == zephyr_core_types.Status_ACCEPTED {
			uniqueStringToValidatedTrafficPolicy[clients.ToUniqueSingleClusterString(trafficPolicy.ObjectMeta)] = &trafficPolicy
			allValidatedTrafficPolicies = append(allValidatedTrafficPolicies, &trafficPolicy)
		}
	}

	allMeshServices, serviceToMetadata, err := a.aggregateMeshServices(ctx)
	if err != nil {
		return err
	}

	serviceWithRelevantPolicies, err := a.aggregator.GroupByMeshService(allValidatedTrafficPolicies, allMeshServices)
	if err != nil {
		return err
	}

	trafficPolicyToAllConflicts := map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_ConflictError{}
	trafficPolicyToAllTranslationErrs := map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_TranslatorError{}
	serviceToUpdatedStatus := map[*zephyr_discovery.MeshService][]*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{}

	for _, serviceWithPolicies := range serviceWithRelevantPolicies {
		meshService := serviceWithPolicies.MeshService

		translationValidator, ok := a.translationValidators[serviceToMetadata[meshService].MeshType]
		if !ok {
			return eris.Errorf("Missing translation validator for mesh type %s", serviceToMetadata[meshService].MeshType.String())
		}

		newlyComputedMergeablePolicies, trafficPoliciesInConflict, policyToTranslationErrors := a.determineNewStatuses(
			meshService,
			serviceToMetadata[meshService].Mesh,
			serviceWithPolicies.TrafficPolicies,
			uniqueStringToValidatedTrafficPolicy,
			translationValidator,
			allTrafficPolicyIds,
		)

		serviceToUpdatedStatus[meshService] = newlyComputedMergeablePolicies
		for trafficPolicy, conflicts := range trafficPoliciesInConflict {
			trafficPolicyToAllConflicts[trafficPolicy] = append(trafficPolicyToAllConflicts[trafficPolicy], conflicts...)
		}

		for trafficPolicy, translationErrors := range policyToTranslationErrors {
			trafficPolicyToAllTranslationErrs[trafficPolicy] = append(trafficPolicyToAllTranslationErrs[trafficPolicy], translationErrors...)
		}
	}

	// these two `updateStatus*` methods depend on an idempotent ordering in all the collections coming out of `determineNewStatuses`
	for service, validatedPolicies := range serviceToUpdatedStatus {
		err := a.updateServiceStatusIfNecessary(ctx, service, validatedPolicies)
		if err != nil {
			return err
		}
	}

	for _, policy := range allValidatedTrafficPolicies {
		err := a.updatePolicyStatusIfNecessary(ctx, policy, trafficPolicyToAllConflicts[policy], trafficPolicyToAllTranslationErrs[policy])
		if err != nil {
			return err
		}
	}

	return nil
}

type policyToCheck struct {
	Ref *zephyr_core_types.ResourceRef

	// If this is non-nil, then the traffic policy is both 1. valid, and 2. changed from the last reconcile iteration.
	// That implies that this field will be nil if the policy was invalid.
	UpdatedTrafficPolicy *zephyr_networking.TrafficPolicy

	// if this is non-nil, then we previously recorded a last-known-good state for this policy
	LastKnownGoodSpecState *zephyr_networking_types.TrafficPolicySpec
}

/**
 *  There are three important collections (two in parameters, one on the service) in play here:
 *    - policiesForService: The *new valid* state of the traffic policies that apply to THIS service. These policies will be a subset of the next collection.
 *    - uniqueStringToNewlyValidatedTrafficPolicy: The set of ALL new, valid traffic policies
 *    - previouslyRecordedPolicies: The last-known good state of the policies that applied to this service during a previous reconcile iteration
 */
func (a *aggregationReconciler) determineNewStatuses(
	meshService *zephyr_discovery.MeshService,
	mesh *zephyr_discovery.Mesh,
	policiesForService []*zephyr_networking.TrafficPolicy,
	uniqueStringToNewlyValidatedTrafficPolicy map[string]*zephyr_networking.TrafficPolicy,
	translationValidator mesh_translation.TranslationValidator,
	allTrafficPolicyIds sets.String,
) (
	[]*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
	map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_ConflictError,
	map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_TranslatorError,
) {
	previouslyRecordedTrafficPolicyIDs := sets.NewString()
	anyPolicyChangedSinceLastReconcile := false

	var policiesToCheck []*policyToCheck
	for _, previouslyRecordedPolicyIter := range meshService.Status.GetValidatedTrafficPolicies() {
		previouslyRecordedPolicy := previouslyRecordedPolicyIter
		identifierString := clients.ToUniqueSingleClusterString(clients.ResourceRefToObjectMeta(previouslyRecordedPolicy.GetRef()))

		// if this traffic policy has been deleted, we need to remove this previously-recorded policy from the output of this reconcile loop
		if !allTrafficPolicyIds.Has(identifierString) {
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

	var policiesToRecordOnService []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy
	policiesInConflict := map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_ConflictError{}
	untranslatablePolicies := map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_TranslatorError{}

	// avoid expensive merge and translation validations if nothing changed from the last reconcile iteration
	if !anyPolicyChangedSinceLastReconcile {
		for _, policyIter := range policiesToCheck {
			policiesToRecordOnService = append(policiesToRecordOnService, &zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				Ref:               policyIter.Ref,
				TrafficPolicySpec: policyIter.LastKnownGoodSpecState,
			})
		}
		return policiesToRecordOnService, nil, nil
	}

	// We want to sort those entries with a nil `UpdatedTrafficPolicy` field to the BEGINNING of the list.
	// This is significant- we want to ensure that we only mark those policies that have CHANGED since the
	// last reconcile iteration and are NOW in conflict with other policies to be marked as in conflict, not older
	// ones that are unchanged. We accomplish that doing this sort, which ensures that we accept unchanged
	// policies into the `policiesToRecordOnService` list first (which will all pass validation together), then
	// subsequently we process the changed policies, which may fail and then be marked as in conflict.
	sort.Slice(policiesToCheck, func(i, j int) bool {
		// polices[i] is LESS than policies[j] (i.e., should appear before it in the list) if policies[i] was not updated
		// we don't care if policies[j] was updated or not, it'll get sorted at some point too.
		return policiesToCheck[i].UpdatedTrafficPolicy == nil
	})

	for _, policyToCheckIter := range policiesToCheck {
		policyToProcess := policyToCheckIter

		// see the notes on the sort.Slice call above; because of that ordering, we know that anything
		// with a nil Updated field must by definition be both merge-able and translate-able with everything
		// ELSE with a nil Updated field, so we can avoid doing expensive merge/translate validations on all of those.
		if policyToProcess.UpdatedTrafficPolicy == nil {
			policiesToRecordOnService = append(policiesToRecordOnService, &zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				Ref:               policyToProcess.Ref,
				TrafficPolicySpec: policyToProcess.LastKnownGoodSpecState,
			})
			continue
		}

		var toValidate *zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy
		if policyToProcess.UpdatedTrafficPolicy == nil {
			// nothing was updated from the last reconcile iteration, make sure we can still merge this one in
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

		// fall back to this if `toValidate` does not pass all validations. The `TrafficPolicySpec` field here should always be non-nil
		successfullyValidated := &zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
			Ref:               policyToProcess.Ref,
			TrafficPolicySpec: policyToProcess.LastKnownGoodSpecState, // may be nil
		}

		var mergeabilityCheckCopy []*zephyr_networking_types.TrafficPolicySpec
		for _, tp := range policiesToRecordOnService {
			// don't need to skip any entries in the list in here, since we won't process a traffic policy twice

			mergeabilityCheckCopy = append(mergeabilityCheckCopy, tp.TrafficPolicySpec)
		}

		mergeConflict := a.aggregator.FindMergeConflict(toValidate.TrafficPolicySpec, mergeabilityCheckCopy, meshService)
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

func (a *aggregationReconciler) aggregateMeshServices(ctx context.Context) ([]*zephyr_discovery.MeshService, map[*zephyr_discovery.MeshService]*MeshServiceInfo, error) {
	meshServiceList, err := a.meshServiceClient.ListMeshService(ctx)
	if err != nil {
		return nil, nil, err
	}

	serviceToMetadata := map[*zephyr_discovery.MeshService]*MeshServiceInfo{}
	var allMeshServices []*zephyr_discovery.MeshService
	for _, ms := range meshServiceList.Items {
		meshService := ms

		meshForService, err := a.meshClient.GetMesh(ctx, clients.ResourceRefToObjectKey(meshService.Spec.GetMesh()))
		if err != nil {
			return nil, nil, err
		}

		meshType, err := enum_conversion.MeshToMeshType(meshForService)
		if err != nil {
			return nil, nil, err
		}

		serviceToMetadata[&meshService] = &MeshServiceInfo{
			ClusterName: meshForService.Spec.GetCluster().GetName(),
			MeshType:    meshType,
			Mesh:        meshForService,
		}
		allMeshServices = append(allMeshServices, &meshService)
	}

	return allMeshServices, serviceToMetadata, nil
}

func (a *aggregationReconciler) updateServiceStatusIfNecessary(
	ctx context.Context,
	meshService *zephyr_discovery.MeshService,
	newlyComputedMergeablePolicies []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
) error {
	shouldUpdateStatus := false
	if len(newlyComputedMergeablePolicies) != len(meshService.Status.ValidatedTrafficPolicies) {
		shouldUpdateStatus = true
	} else {
		for i, newlyComputedPolicy := range newlyComputedMergeablePolicies {
			shouldUpdateStatus = shouldUpdateStatus || !meshService.Status.ValidatedTrafficPolicies[i].Equal(newlyComputedPolicy)
		}
	}

	if shouldUpdateStatus {
		meshService.Status.ValidatedTrafficPolicies = newlyComputedMergeablePolicies
		err := a.meshServiceClient.UpdateMeshServiceStatus(ctx, meshService)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *aggregationReconciler) updatePolicyStatusIfNecessary(
	ctx context.Context,
	policy *zephyr_networking.TrafficPolicy,
	newConflictErrors []*zephyr_networking_types.TrafficPolicyStatus_ConflictError,
	newTranslationErrors []*zephyr_networking_types.TrafficPolicyStatus_TranslatorError,
) error {
	shouldUpdateStatus := false
	if len(newConflictErrors) != len(policy.Status.GetConflictErrors()) {
		shouldUpdateStatus = true
	} else {
		for newErrorIndex, newError := range newConflictErrors {
			shouldUpdateStatus = shouldUpdateStatus || !policy.Status.GetConflictErrors()[newErrorIndex].Equal(newError)
		}
	}

	if len(newTranslationErrors) != len(policy.Status.TranslatorErrors) {
		shouldUpdateStatus = true
	} else {
		for newErrorIndex, newError := range newTranslationErrors {
			shouldUpdateStatus = shouldUpdateStatus || !policy.Status.GetTranslatorErrors()[newErrorIndex].Equal(newError)
		}
	}

	if shouldUpdateStatus {
		policy.Status.ConflictErrors = newConflictErrors
		policy.Status.TranslatorErrors = newTranslationErrors

		// the next stage, actual non-dry-run translation, will set this to the proper value
		policy.Status.TranslationStatus = nil

		err := a.trafficPolicyClient.UpdateTrafficPolicyStatus(ctx, policy)
		if err != nil {
			return err
		}
	}

	return nil
}
