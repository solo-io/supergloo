package aggregation_framework

import (
	"context"

	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"github.com/solo-io/service-mesh-hub/pkg/enum_conversion"
	"github.com/solo-io/service-mesh-hub/pkg/reconciliation"
	traffic_policy_aggregation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/aggregation"
	mesh_translation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/meshes"
)

func NewAggregationReconciler(
	trafficPolicyClient zephyr_networking.TrafficPolicyClient,
	meshServiceClient zephyr_discovery.MeshServiceClient,
	meshClient zephyr_discovery.MeshClient,
	policyCollector traffic_policy_aggregation.PolicyCollector,
	translationValidators map[zephyr_core_types.MeshType]mesh_translation.TranslationValidator,
	inMemoryStatusUpdater traffic_policy_aggregation.InMemoryStatusUpdater,
) reconciliation.Reconciler {
	return &aggregationReconciler{
		trafficPolicyClient:   trafficPolicyClient,
		meshServiceClient:     meshServiceClient,
		meshClient:            meshClient,
		policyCollector:       policyCollector,
		translationValidators: translationValidators,
		inMemoryStatusUpdater: inMemoryStatusUpdater,
	}
}

type aggregationReconciler struct {
	trafficPolicyClient   zephyr_networking.TrafficPolicyClient
	meshServiceClient     zephyr_discovery.MeshServiceClient
	meshClient            zephyr_discovery.MeshClient
	policyCollector       traffic_policy_aggregation.PolicyCollector
	translationValidators map[zephyr_core_types.MeshType]mesh_translation.TranslationValidator
	inMemoryStatusUpdater traffic_policy_aggregation.InMemoryStatusUpdater
}

func (a *aggregationReconciler) GetName() string {
	return "traffic-policy-aggregation"
}

func (a *aggregationReconciler) Reconcile(ctx context.Context) error {
	allTrafficPoliciesList, err := a.trafficPolicyClient.ListTrafficPolicy(ctx)
	if err != nil {
		return err
	}

	var allTrafficPolicies []*zephyr_networking.TrafficPolicy
	for _, tp := range allTrafficPoliciesList.Items {
		trafficPolicy := tp
		allTrafficPolicies = append(allTrafficPolicies, &trafficPolicy)
	}

	allMeshServices, serviceToMetadata, err := a.aggregateMeshServices(ctx)
	if err != nil {
		return err
	}

	trafficPolicyToAllConflicts := map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_ConflictError{}
	trafficPolicyToAllTranslationErrs := map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_TranslatorError{}
	serviceToUpdatedStatus := map[*zephyr_discovery.MeshService][]*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{}

	for _, meshService := range allMeshServices {
		newlyComputedMergeablePolicies, trafficPoliciesInConflict, policyToTranslationErrors, err := a.policyCollector.CollectForService(
			meshService,
			serviceToMetadata[meshService].Mesh,
			a.translationValidators[serviceToMetadata[meshService].MeshType],
			allTrafficPolicies,
		)
		if err != nil {
			return err
		}

		serviceToUpdatedStatus[meshService] = newlyComputedMergeablePolicies
		for trafficPolicy, conflicts := range trafficPoliciesInConflict {
			trafficPolicyToAllConflicts[trafficPolicy] = append(trafficPolicyToAllConflicts[trafficPolicy], conflicts...)
		}

		for trafficPolicy, translationErrors := range policyToTranslationErrors {
			trafficPolicyToAllTranslationErrs[trafficPolicy] = append(trafficPolicyToAllTranslationErrs[trafficPolicy], translationErrors...)
		}
	}

	for service, validatedPolicies := range serviceToUpdatedStatus {
		needsUpdating := a.inMemoryStatusUpdater.UpdateServicePolicies(service, validatedPolicies)
		if needsUpdating {
			err := a.meshServiceClient.UpdateMeshServiceStatus(ctx, service)
			if err != nil {
				return err
			}
		}
	}

	for _, policy := range allTrafficPolicies {
		// Note that this is being called for all policies, regardless of validation status.
		// We don't want knowledge of validation status to leak into this component, and we don't care if
		// invalid policies have their merge/translation errors zeroed out
		needsUpdating := a.inMemoryStatusUpdater.UpdateConflictAndTranslatorErrors(policy, trafficPolicyToAllConflicts[policy], trafficPolicyToAllTranslationErrs[policy])
		if needsUpdating {
			err := a.trafficPolicyClient.UpdateTrafficPolicyStatus(ctx, policy)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *aggregationReconciler) aggregateMeshServices(ctx context.Context) ([]*zephyr_discovery.MeshService, map[*zephyr_discovery.MeshService]*traffic_policy_aggregation.MeshServiceInfo, error) {
	meshServiceList, err := a.meshServiceClient.ListMeshService(ctx)
	if err != nil {
		return nil, nil, err
	}

	serviceToMetadata := map[*zephyr_discovery.MeshService]*traffic_policy_aggregation.MeshServiceInfo{}
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

		serviceToMetadata[&meshService] = &traffic_policy_aggregation.MeshServiceInfo{
			ClusterName: meshForService.Spec.GetCluster().GetName(),
			MeshType:    meshType,
			Mesh:        meshForService,
		}
		allMeshServices = append(allMeshServices, &meshService)
	}

	return allMeshServices, serviceToMetadata, nil
}
