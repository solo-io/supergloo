package traffic_policy_aggregation

import (
	"context"

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
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (a *aggregationReconciler) Reconcile(ctx context.Context) error {
	allTrafficPolicies, err := a.trafficPolicyClient.ListTrafficPolicy(ctx)
	if err != nil {
		return err
	}

	uniqueStringToValidatedTrafficPolicy := map[string]*zephyr_networking.TrafficPolicy{}
	var allValidatedTrafficPolicies []*zephyr_networking.TrafficPolicy
	for _, tpIter := range allTrafficPolicies.Items {
		trafficPolicy := tpIter

		if trafficPolicy.Status.GetValidationStatus().GetState() == zephyr_core_types.Status_ACCEPTED {
			uniqueStringToValidatedTrafficPolicy[clients.ToUniqueString(trafficPolicy.ObjectMeta)] = &trafficPolicy
			allValidatedTrafficPolicies = append(allValidatedTrafficPolicies, &trafficPolicy)
		}
	}

	allMeshServices, serviceToMetadata, err := a.aggregateMeshServices(ctx)
	if err != nil {
		return err
	}

	serviceWithRelevantPolicies := a.aggregator.GroupByMeshService(allValidatedTrafficPolicies, serviceToMetadata)

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
			allMeshServices,
			translationValidator,
		)

		serviceToUpdatedStatus[meshService] = newlyComputedMergeablePolicies
		for trafficPolicy, conflicts := range trafficPoliciesInConflict {
			trafficPolicyToAllConflicts[trafficPolicy] = append(trafficPolicyToAllConflicts[trafficPolicy], conflicts...)
		}

		for trafficPolicy, translationErrors := range policyToTranslationErrors {
			trafficPolicyToAllTranslationErrs[trafficPolicy] = append(trafficPolicyToAllTranslationErrs[trafficPolicy], translationErrors...)
		}
	}

	for service, validatedPolicies := range serviceToUpdatedStatus {
		err := a.updateServiceStatusIfNecessary(ctx, service, validatedPolicies)
		if err != nil {
			return err
		}
	}

	// here?

	for _, policy := range allValidatedTrafficPolicies {
		err := a.updatePolicyStatusIfNecessary(ctx, policy, trafficPolicyToAllConflicts[policy], trafficPolicyToAllTranslationErrs[policy])
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *aggregationReconciler) determineNewStatuses(
	meshService *zephyr_discovery.MeshService,
	mesh *zephyr_discovery.Mesh,
	policiesForService []*zephyr_networking.TrafficPolicy,
	uniqueStringToNewlyValidatedTrafficPolicy map[string]*zephyr_networking.TrafficPolicy,
	allMeshServices []*zephyr_discovery.MeshService,
	translationValidator mesh_translation.TranslationValidator,
) (
	[]*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
	map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_ConflictError,
	map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_TranslatorError,
) {
	var policiesToRecordOnService []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy
	policiesInConflict := map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_ConflictError{}
	untranslatablePolicies := map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_TranslatorError{}

	// keep track of the entries that were already on the statuses
	previouslyRecordedNameNamespaces := sets.NewString()
	previouslyRecordedPolicies := meshService.Status.GetValidatedTrafficPolicies()

	// here we ensure that everything that was in the list previously is either:
	//   1. maintaining its old state, or
	//   2. taking on the newly-validated state
	// this loop should result in a stable ordering in the list
	for policyIndex, previouslyRecordedPolicyIter := range previouslyRecordedPolicies {
		previouslyRecordedPolicy := previouslyRecordedPolicyIter
		identifierString := clients.ToUniqueString(k8s_meta_types.ObjectMeta{
			Name:      previouslyRecordedPolicy.Name,
			Namespace: previouslyRecordedPolicy.Namespace,
		})

		previouslyRecordedNameNamespaces.Insert(identifierString)

		newlyValidatedPolicyState, validatedPolicyFound := uniqueStringToNewlyValidatedTrafficPolicy[identifierString]

		// if the policy couldn't be validated or if it hasn't changed, keep the old state in the list
		if !validatedPolicyFound || newlyValidatedPolicyState.Spec.Equal(previouslyRecordedPolicy.TrafficPolicySpec) {
			policiesToRecordOnService = append(policiesToRecordOnService, previouslyRecordedPolicy)
		} else {
			// otherwise, there is a newly-valid state to record. Ensure that this new state can in fact be merged in with the rest of the list
			var policiesToMergeWith []*zephyr_networking_types.TrafficPolicySpec
			for i, tpIter := range previouslyRecordedPolicies {
				tp := tpIter

				// exclude the traffic policy that got updated
				if i == policyIndex {
					continue
				}

				policiesToMergeWith = append(policiesToMergeWith, tp.TrafficPolicySpec)
			}

			mergeConflict := a.aggregator.FindMergeConflict(&newlyValidatedPolicyState.Spec, policiesToMergeWith, allMeshServices)

			// this will be either the new one, or the old one if we couldn't merge the new state in with the others
			var validatedTrafficPolicyStateToRecord *zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy
			if mergeConflict != nil {
				policiesInConflict[newlyValidatedPolicyState] = append(policiesInConflict[newlyValidatedPolicyState], mergeConflict)
				validatedTrafficPolicyStateToRecord = previouslyRecordedPolicy
			} else {
				validatedTrafficPolicyStateToRecord = &zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
					Name:              newlyValidatedPolicyState.GetName(),
					Namespace:         newlyValidatedPolicyState.GetNamespace(),
					TrafficPolicySpec: &newlyValidatedPolicyState.Spec,
				}
			}

			// we know that the policies are mergeable; now let's see if they can be translated all together
			dryRunTranslation := append([]*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy(nil), policiesToRecordOnService...)
			dryRunTranslation = append(dryRunTranslation, validatedTrafficPolicyStateToRecord)

			translationErrors := translationValidator.GetTranslationErrors(
				meshService,
				mesh,
				dryRunTranslation,
			)

			if len(translationErrors) == 0 {
				// we're done, record this as validated
				policiesToRecordOnService = append(policiesToRecordOnService, validatedTrafficPolicyStateToRecord)
			} else {
				for _, translationError := range translationErrors {
					// need to pick out the translation policy in question from the error result list
					if clients.SameObject(k8s_meta_types.ObjectMeta{
						Name:      translationError.Policy.GetName(),
						Namespace: translationError.Policy.GetNamespace(),
					}, newlyValidatedPolicyState.ObjectMeta) {
						untranslatablePolicies[newlyValidatedPolicyState] = translationError.TranslatorErrors
					}
				}
			}
		}
	}

	// here we add in the NEW traffic policies- those that we associated with this service, but that we didn't process in the for loop above
	for _, relevantPolicyIter := range policiesForService {
		relevantPolicy := relevantPolicyIter

		// we didn't see this one before, so it must be new
		if !previouslyRecordedNameNamespaces.Has(clients.ToUniqueString(relevantPolicy.ObjectMeta)) {
			var specsToMergeWith []*zephyr_networking_types.TrafficPolicySpec
			for _, validatedPolicyIter := range policiesToRecordOnService {
				validatedPolicy := validatedPolicyIter

				specsToMergeWith = append(specsToMergeWith, validatedPolicy.TrafficPolicySpec)
			}

			mergeConflict := a.aggregator.FindMergeConflict(&relevantPolicy.Spec, specsToMergeWith, allMeshServices)
			if mergeConflict == nil {
				validated := &zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
					Name:              relevantPolicy.GetName(),
					Namespace:         relevantPolicy.GetNamespace(),
					TrafficPolicySpec: &relevantPolicy.Spec,
				}

				dryRunTranslation := append([]*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy(nil), policiesToRecordOnService...)
				dryRunTranslation = append(dryRunTranslation, validated)

				translationErrors := translationValidator.GetTranslationErrors(
					meshService,
					mesh,
					dryRunTranslation,
				)

				if len(translationErrors) == 0 {
					policiesToRecordOnService = append(policiesToRecordOnService, validated)
				} else {
					for _, translationError := range translationErrors {
						// need to pick out the translation policy in question from the error result list
						if clients.SameObject(k8s_meta_types.ObjectMeta{
							Name:      translationError.Policy.GetName(),
							Namespace: translationError.Policy.GetNamespace(),
						}, relevantPolicy.ObjectMeta) {
							untranslatablePolicies[relevantPolicy] = translationError.TranslatorErrors
						}
					}
				}
			} else {
				policiesInConflict[relevantPolicy] = append(policiesInConflict[relevantPolicy], mergeConflict)
			}
		}
	}

	return policiesToRecordOnService, policiesInConflict, untranslatablePolicies
}

func (a *aggregationReconciler) aggregateMeshServices(ctx context.Context) ([]*zephyr_discovery.MeshService, map[*zephyr_discovery.MeshService]*MeshServiceInfo, error) {
	meshServiceList, err := a.meshServiceClient.ListMeshService(ctx)
	if err != nil {
		return nil, nil, err
	}

	serviceToClusterName := map[*zephyr_discovery.MeshService]*MeshServiceInfo{}
	allMeshServices := []*zephyr_discovery.MeshService{}
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

		serviceToClusterName[&meshService] = &MeshServiceInfo{
			ClusterName: meshForService.Spec.GetCluster().GetName(),
			MeshType:    meshType,
			Mesh:        meshForService,
		}
		allMeshServices = append(allMeshServices, &meshService)
	}

	return allMeshServices, serviceToClusterName, nil
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
