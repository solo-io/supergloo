package traffic_policy_aggregation

import (
	"context"

	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"github.com/solo-io/service-mesh-hub/pkg/reconciliation"
	"github.com/solo-io/service-mesh-hub/pkg/selector"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

func NewAggregationReconciler(
	trafficPolicyClient zephyr_networking.TrafficPolicyClient,
	meshServiceClient zephyr_discovery.MeshServiceClient,
	aggregator Aggregator,
) reconciliation.Reconciler {
	return &aggregationReconciler{
		trafficPolicyClient: trafficPolicyClient,
		meshServiceClient:   meshServiceClient,
		aggregator:          aggregator,
	}
}

type aggregationReconciler struct {
	trafficPolicyClient zephyr_networking.TrafficPolicyClient
	meshServiceClient   zephyr_discovery.MeshServiceClient
	aggregator          Aggregator
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

	// TODO: Get rid of this map
	allMeshServices, serviceToClusterName, err := a.aggregateMeshServices(ctx)
	if err != nil {
		return err
	}

	serviceWithRelevantPolicies := a.aggregator.GroupByMeshService(allValidatedTrafficPolicies, serviceToClusterName)

	trafficPolicyToAllConflicts := map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_ConflictError{}
	serviceToUpdatedStatus := map[*zephyr_discovery.MeshService][]*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{}

	for _, serviceWithPolicies := range serviceWithRelevantPolicies {
		meshService := serviceWithPolicies.MeshService

		newlyComputedMergeablePolicies, trafficPoliciesInConflict := a.determineNewStatuses(
			meshService,
			serviceWithPolicies.TrafficPolicies,
			uniqueStringToValidatedTrafficPolicy,
			allMeshServices,
		)

		serviceToUpdatedStatus[meshService] = newlyComputedMergeablePolicies
		for trafficPolicy, conflicts := range trafficPoliciesInConflict {
			trafficPolicyToAllConflicts[trafficPolicy] = append(trafficPolicyToAllConflicts[trafficPolicy], conflicts...)
		}
	}

	for service, validatedPolicies := range serviceToUpdatedStatus {
		err := a.updateServiceStatusIfNecessary(ctx, service, validatedPolicies)
		if err != nil {
			return err
		}
	}

	for _, policy := range allValidatedTrafficPolicies {
		err := a.updatePolicyStatusIfNecessary(ctx, policy, trafficPolicyToAllConflicts[policy])
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *aggregationReconciler) determineNewStatuses(
	meshService *zephyr_discovery.MeshService,
	policiesForService []*zephyr_networking.TrafficPolicy,
	uniqueStringToNewlyValidatedTrafficPolicy map[string]*zephyr_networking.TrafficPolicy,
	allMeshServices []*zephyr_discovery.MeshService,
) (
	[]*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
	map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_ConflictError,
) {
	var policiesToRecordOnService []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy
	policiesInConflict := map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_ConflictError{}

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

			policiesToRecordOnService = append(policiesToRecordOnService, validatedTrafficPolicyStateToRecord)
		}
	}

	// here we add in the NEW traffic policies- those that we associated with this service, but that we didn't process in the for loop above
	for _, relevantPolicyIter := range policiesForService {
		relevantPolicy := relevantPolicyIter

		// we didn't see this one before, so it must be new
		if !previouslyRecordedNameNamespaces.Has(clients.ToUniqueString(relevantPolicy.ObjectMeta)) {
			policiesToRecordOnService = append(policiesToRecordOnService, &zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				Name:              relevantPolicy.GetName(),
				Namespace:         relevantPolicy.GetNamespace(),
				TrafficPolicySpec: &relevantPolicy.Spec,
			})
		}
	}

	return policiesToRecordOnService, policiesInConflict
}

func (a *aggregationReconciler) aggregateMeshServices(ctx context.Context) ([]*zephyr_discovery.MeshService, map[*zephyr_discovery.MeshService]string, error) {
	meshServiceList, err := a.meshServiceClient.ListMeshService(ctx)
	if err != nil {
		return nil, nil, err
	}

	serviceToClusterName := map[*zephyr_discovery.MeshService]string{}
	allMeshServices := []*zephyr_discovery.MeshService{}
	for _, ms := range meshServiceList.Items {
		meshService := ms

		clusterName, ok := meshService.Labels[constants.COMPUTE_TARGET]
		if !ok {
			return nil, nil, selector.MissingComputeTargetLabel(meshService.GetName())
		}

		serviceToClusterName[&meshService] = clusterName
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
) error {
	shouldUpdateStatus := false
	if len(newConflictErrors) != len(policy.Status.GetConflictErrors()) {
		shouldUpdateStatus = true
	} else {
		for newErrorIndex, newError := range newConflictErrors {
			shouldUpdateStatus = shouldUpdateStatus || !policy.Status.GetConflictErrors()[newErrorIndex].Equal(newError)
		}
	}

	if shouldUpdateStatus {
		policy.Status.ConflictErrors = newConflictErrors
		err := a.trafficPolicyClient.UpdateTrafficPolicyStatus(ctx, policy)
		if err != nil {
			return err
		}
	}

	return nil
}
