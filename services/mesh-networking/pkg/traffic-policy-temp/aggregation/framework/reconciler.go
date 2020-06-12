package aggregation_framework

import (
	"context"

	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/metadata"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	traffic_policy_aggregation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/aggregation"
	mesh_translation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/translators"
)

//go:generate mockgen -source ./reconciler.go -destination ./mocks/mock_reconciler.go

type AggregateProcessor interface {
	Process(ctx context.Context, allTrafficPolicies []*smh_networking.TrafficPolicy) (*ProcessedObjects, error)
}

type ProcessedObjects struct {
	TrafficPolicies []*smh_networking.TrafficPolicy
	MeshServices    []*smh_discovery.MeshService
}

func NewAggregationProcessor(
	meshServiceReader smh_discovery.MeshServiceReader,
	meshReader smh_discovery.MeshReader,
	policyCollector traffic_policy_aggregation.PolicyCollector,
	translationValidators func(meshType smh_core_types.MeshType) (mesh_translation.TranslationValidator, error),
	inMemoryStatusMutator traffic_policy_aggregation.InMemoryStatusMutator,
) *aggregationProcessor {
	return &aggregationProcessor{
		meshServiceReader:     meshServiceReader,
		meshReader:            meshReader,
		policyCollector:       policyCollector,
		translationValidators: translationValidators,
		inMemoryStatusMutator: inMemoryStatusMutator,
	}
}

type aggregationProcessor struct {
	meshServiceReader     smh_discovery.MeshServiceReader
	meshReader            smh_discovery.MeshReader
	policyCollector       traffic_policy_aggregation.PolicyCollector
	translationValidators func(meshType smh_core_types.MeshType) (mesh_translation.TranslationValidator, error)
	inMemoryStatusMutator traffic_policy_aggregation.InMemoryStatusMutator
}

func (a *aggregationProcessor) Process(ctx context.Context, allTrafficPolicies []*smh_networking.TrafficPolicy) (*ProcessedObjects, error) {

	allMeshServices, serviceToMetadata, err := a.aggregateMeshServices(ctx)
	if err != nil {
		return nil, err
	}

	trafficPolicyToAllConflicts := map[*smh_networking.TrafficPolicy][]*smh_networking_types.TrafficPolicyStatus_ConflictError{}
	trafficPolicyToAllTranslationErrs := map[*smh_networking.TrafficPolicy][]*smh_networking_types.TrafficPolicyStatus_TranslatorError{}
	serviceToUpdatedStatus := map[*smh_discovery.MeshService][]*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{}

	for _, meshService := range allMeshServices {
		validator, err := a.translationValidators(serviceToMetadata[meshService].MeshType)
		if err != nil {
			// intentionally not doing map existence checks here; if it panics, we forgot to implement the validator for this translator
			// TODO: un-panic this
			panic(err)
		}
		collectionResult, err := a.policyCollector.CollectForService(
			meshService,
			allMeshServices,
			serviceToMetadata[meshService].Mesh,
			validator,
			allTrafficPolicies,
		)
		if err != nil {
			return nil, err
		}

		serviceToUpdatedStatus[meshService] = collectionResult.PoliciesToRecordOnService
		for trafficPolicy, conflicts := range collectionResult.PolicyToConflictErrors {
			trafficPolicyToAllConflicts[trafficPolicy] = append(trafficPolicyToAllConflicts[trafficPolicy], conflicts...)
		}

		for trafficPolicy, translationErrors := range collectionResult.PolicyToTranslatorErrors {
			trafficPolicyToAllTranslationErrs[trafficPolicy] = append(trafficPolicyToAllTranslationErrs[trafficPolicy], translationErrors...)
		}
	}

	var objectsToUpdate ProcessedObjects
	for service, validatedPolicies := range serviceToUpdatedStatus {
		needsUpdating := a.inMemoryStatusMutator.MutateServicePolicies(service, validatedPolicies)
		if needsUpdating {
			objectsToUpdate.MeshServices = append(objectsToUpdate.MeshServices, service)
		}
	}

	for _, policy := range allTrafficPolicies {
		// Note that this is being called for all policies, regardless of validation status.
		// We don't want knowledge of validation status to leak into this component, and we don't care if
		// invalid policies have their merge/translation errors zeroed out
		needsUpdating := a.inMemoryStatusMutator.MutateConflictAndTranslatorErrors(policy, trafficPolicyToAllConflicts[policy], trafficPolicyToAllTranslationErrs[policy])
		if needsUpdating {
			objectsToUpdate.TrafficPolicies = append(objectsToUpdate.TrafficPolicies, policy)
		}
	}

	return &objectsToUpdate, nil
}

func (a *aggregationProcessor) aggregateMeshServices(ctx context.Context) ([]*smh_discovery.MeshService, map[*smh_discovery.MeshService]*meshServiceInfo, error) {
	meshServiceList, err := a.meshServiceReader.ListMeshService(ctx)
	if err != nil {
		return nil, nil, err
	}

	serviceToMetadata := map[*smh_discovery.MeshService]*meshServiceInfo{}
	var allMeshServices []*smh_discovery.MeshService
	for _, ms := range meshServiceList.Items {
		meshService := ms

		meshForService, err := a.meshReader.GetMesh(ctx, selection.ResourceRefToObjectKey(meshService.Spec.GetMesh()))
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
