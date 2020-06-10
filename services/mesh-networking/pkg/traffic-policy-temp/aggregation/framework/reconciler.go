package aggregation_framework

import (
	"context"

	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/kube/metadata"
	"github.com/solo-io/service-mesh-hub/pkg/kube/selection"
	"github.com/solo-io/service-mesh-hub/pkg/reconciliation"
	traffic_policy_aggregation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/aggregation"
	mesh_translation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/translators"
)

func NewAggregationReconciler(
	trafficPolicyClient smh_networking.TrafficPolicyClient,
	meshServiceClient smh_discovery.MeshServiceClient,
	meshClient smh_discovery.MeshClient,
	policyCollector traffic_policy_aggregation.PolicyCollector,
	translationValidators map[smh_core_types.MeshType]mesh_translation.TranslationValidator,
	inMemoryStatusMutator traffic_policy_aggregation.InMemoryStatusMutator,
) reconciliation.Reconciler {
	return &aggregationReconciler{
		trafficPolicyClient:   trafficPolicyClient,
		meshServiceClient:     meshServiceClient,
		meshClient:            meshClient,
		policyCollector:       policyCollector,
		translationValidators: translationValidators,
		inMemoryStatusMutator: inMemoryStatusMutator,
	}
}

type aggregationReconciler struct {
	trafficPolicyClient   smh_networking.TrafficPolicyClient
	meshServiceClient     smh_discovery.MeshServiceClient
	meshClient            smh_discovery.MeshClient
	policyCollector       traffic_policy_aggregation.PolicyCollector
	translationValidators map[smh_core_types.MeshType]mesh_translation.TranslationValidator
	inMemoryStatusMutator traffic_policy_aggregation.InMemoryStatusMutator
}

func (a *aggregationReconciler) GetName() string {
	return "traffic-policy-aggregation"
}

func (a *aggregationReconciler) Reconcile(ctx context.Context) error {
	allTrafficPoliciesList, err := a.trafficPolicyClient.ListTrafficPolicy(ctx)
	if err != nil {
		return err
	}

	var allTrafficPolicies []*smh_networking.TrafficPolicy
	for _, tp := range allTrafficPoliciesList.Items {
		trafficPolicy := tp
		allTrafficPolicies = append(allTrafficPolicies, &trafficPolicy)
	}

	allMeshServices, serviceToMetadata, err := a.aggregateMeshServices(ctx)
	if err != nil {
		return err
	}

	trafficPolicyToAllConflicts := map[*smh_networking.TrafficPolicy][]*smh_networking_types.TrafficPolicyStatus_ConflictError{}
	trafficPolicyToAllTranslationErrs := map[*smh_networking.TrafficPolicy][]*smh_networking_types.TrafficPolicyStatus_TranslatorError{}
	serviceToUpdatedStatus := map[*smh_discovery.MeshService][]*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{}

	for _, meshService := range allMeshServices {
		collectionResult, err := a.policyCollector.CollectForService(
			meshService,
			allMeshServices,
			serviceToMetadata[meshService].Mesh,

			// intentionally not doing map existence checks here; if it panics, we forgot to implement the validator for this translator
			a.translationValidators[serviceToMetadata[meshService].MeshType],
			allTrafficPolicies,
		)
		if err != nil {
			return err
		}

		serviceToUpdatedStatus[meshService] = collectionResult.PoliciesToRecordOnService
		for trafficPolicy, conflicts := range collectionResult.PolicyToConflictErrors {
			trafficPolicyToAllConflicts[trafficPolicy] = append(trafficPolicyToAllConflicts[trafficPolicy], conflicts...)
		}

		for trafficPolicy, translationErrors := range collectionResult.PolicyToTranslatorErrors {
			trafficPolicyToAllTranslationErrs[trafficPolicy] = append(trafficPolicyToAllTranslationErrs[trafficPolicy], translationErrors...)
		}
	}

	for service, validatedPolicies := range serviceToUpdatedStatus {
		needsUpdating := a.inMemoryStatusMutator.MutateServicePolicies(service, validatedPolicies)
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
		needsUpdating := a.inMemoryStatusMutator.MutateConflictAndTranslatorErrors(policy, trafficPolicyToAllConflicts[policy], trafficPolicyToAllTranslationErrs[policy])
		if needsUpdating {
			err := a.trafficPolicyClient.UpdateTrafficPolicyStatus(ctx, policy)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *aggregationReconciler) aggregateMeshServices(ctx context.Context) ([]*smh_discovery.MeshService, map[*smh_discovery.MeshService]*meshServiceInfo, error) {
	meshServiceList, err := a.meshServiceClient.ListMeshService(ctx)
	if err != nil {
		return nil, nil, err
	}

	serviceToMetadata := map[*smh_discovery.MeshService]*meshServiceInfo{}
	var allMeshServices []*smh_discovery.MeshService
	for _, ms := range meshServiceList.Items {
		meshService := ms

		meshForService, err := a.meshClient.GetMesh(ctx, selection.ResourceRefToObjectKey(meshService.Spec.GetMesh()))
		if err != nil {
			return nil, nil, err
		}

		meshType, err := metadata.MeshToMeshType(meshForService)
		if err != nil {
			return nil, nil, err
		}

		serviceToMetadata[&meshService] = &meshServiceInfo{
			ClusterName: meshForService.Spec.GetCluster().GetName(),
			MeshType:    meshType,
			Mesh:        meshForService,
		}
		allMeshServices = append(allMeshServices, &meshService)
	}

	return allMeshServices, serviceToMetadata, nil
}

type meshServiceInfo struct {
	ClusterName string
	Mesh        *smh_discovery.Mesh
	MeshType    smh_core_types.MeshType
}
