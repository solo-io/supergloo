package reconcile


package traffic_policy_validation

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/reconciliation"
	"k8s.io/apimachinery/pkg/types"
)

type trafficPolicyReaderStatusUpdater interface {
	zephyr_networking.TrafficPolicyReader
	zephyr_networking.TrafficPolicyStatusWriter
}

func NewValidationReconciler(
	trafficPolicyClient trafficPolicyReaderStatusUpdater,
	meshServiceReader zephyr_discovery.MeshServiceReader,
	validator Validator,
) reconciliation.Reconciler {
	return &validationLoop{
		trafficPolicyClient: trafficPolicyClient,
		validator:           validator,
		meshServiceReader:   meshServiceReader,
	}
}

type validationLoop struct {
	trafficPolicyClient trafficPolicyReaderStatusUpdater
	meshServiceReader   zephyr_discovery.MeshServiceReader
	validator           Validator
}

func (*validationLoop) GetName() string {
	return "traffic-policy-validation"
}

func (v *validationLoop) Reconcile(ctx context.Context) error {
	logger := contextutils.LoggerFrom(ctx)
	trafficPolicies, err := v.trafficPolicyClient.ListTrafficPolicy(ctx)
	if err != nil {
		return err
	}

	meshServiceList, err := v.meshServiceReader.ListMeshService(ctx)
	if err != nil {
		return err
	}

	var meshServices []*zephyr_discovery.MeshService
	for _, ms := range meshServiceList.Items {
		meshService := ms
		meshServices = append(meshServices, &meshService)
	}

	for _, trafficPolicy := range trafficPolicies.Items {
		newValidationStatus, validationErr := v.validator.ValidateTrafficPolicy(&trafficPolicy, meshServices)
		if validationErr == nil {
			logger.Debugf("Traffic policy %s.%s passed validation", trafficPolicy.GetName(), trafficPolicy.GetNamespace())
		} else {
			logger.Infof("Traffic policy %s.%s failed validation for reason: %+v", trafficPolicy.GetName(), trafficPolicy.GetNamespace(), validationErr)
		}

		if !trafficPolicy.Status.GetValidationStatus().Equal(newValidationStatus) ||
			trafficPolicy.Status.ObservedGeneration != trafficPolicy.Generation {
			trafficPolicy.Status.ObservedGeneration = trafficPolicy.Generation
			trafficPolicy.Status.ValidationStatus = newValidationStatus

			err := v.trafficPolicyClient.UpdateTrafficPolicyStatus(ctx, &trafficPolicy)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func NewAggregationReconciler(
	trafficPolicyClient zephyr_networking.TrafficPolicyClient,
	meshServiceClient zephyr_discovery.MeshServiceClient,
	meshClient zephyr_discovery.MeshClient,
	policyCollector traffic_policy_aggregation.PolicyCollector,
	translationValidators map[zephyr_core_types.MeshType]mesh_translation.TranslationValidator,
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
	trafficPolicyClient   zephyr_networking.TrafficPolicyClient
	meshServiceClient     zephyr_discovery.MeshServiceClient
	meshClient            zephyr_discovery.MeshClient
	policyCollector       traffic_policy_aggregation.PolicyCollector
	translationValidators map[zephyr_core_types.MeshType]mesh_translation.TranslationValidator
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

func (a *aggregationReconciler) aggregateMeshServices(ctx context.Context) ([]*zephyr_discovery.MeshService, map[*zephyr_discovery.MeshService]*meshServiceInfo, error) {
	meshServiceList, err := a.meshServiceClient.ListMeshService(ctx)
	if err != nil {
		return nil, nil, err
	}

	serviceToMetadata := map[*zephyr_discovery.MeshService]*meshServiceInfo{}
	var allMeshServices []*zephyr_discovery.MeshService
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
	Mesh        *zephyr_discovery.Mesh
	MeshType    zephyr_core_types.MeshType
}
