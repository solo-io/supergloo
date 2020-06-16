package reconcile

import (
	"context"

	"github.com/hashicorp/go-multierror"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	aggregation_framework "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/aggregation/framework"
	translation_framework "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework/snapshot"
	traffic_policy_validation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/validation"
)

type Reconciler struct {
	trafficPolicyClient smh_networking.TrafficPolicyClient
	meshServiceClient   smh_discovery.MeshServiceClient

	snapshotReconciler snapshot.TranslationSnapshotReconciler

	validationProcessor  traffic_policy_validation.ValidationProcessor
	aggregationProcessor aggregation_framework.AggregationProcessor
	translationProcessor translation_framework.TranslationProcessor
}

func NewReconciler(
	trafficPolicyClient smh_networking.TrafficPolicyClient,
	meshServiceClient smh_discovery.MeshServiceClient,
	snapshotReconciler snapshot.TranslationSnapshotReconciler,
	validationProcessor traffic_policy_validation.ValidationProcessor,
	aggregationProcessor aggregation_framework.AggregationProcessor,
	translationProcessor translation_framework.TranslationProcessor,
) *Reconciler {
	return &Reconciler{
		trafficPolicyClient:  trafficPolicyClient,
		meshServiceClient:    meshServiceClient,
		snapshotReconciler:   snapshotReconciler,
		validationProcessor:  validationProcessor,
		aggregationProcessor: aggregationProcessor,
		translationProcessor: translationProcessor,
	}
}

func (*Reconciler) GetName() string {
	return "traffic-policy-reconciler"
}

func (v *Reconciler) Reconcile(ctx context.Context) error {
	var multierr error

	trafficPolicies, err := v.trafficPolicyClient.ListTrafficPolicy(ctx)
	if err != nil {
		return err
	}

	var allTrafficPolicies []*smh_networking.TrafficPolicy
	for _, tp := range trafficPolicies.Items {
		trafficPolicy := tp
		allTrafficPolicies = append(allTrafficPolicies, &trafficPolicy)
	}

	meshServiceList, err := v.meshServiceClient.ListMeshService(ctx)
	if err != nil {
		return err
	}

	var meshServices []*smh_discovery.MeshService
	for _, ms := range meshServiceList.Items {
		meshService := ms
		meshServices = append(meshServices, &meshService)
	}

	// TODO: this works because traffic policies are not copied with the various processors;
	// if we want to relax that constrant, we can consider generating and using the Set object
	// for traffic policies.
	trafficPoliciesToUpdateSet := map[*smh_networking.TrafficPolicy]bool{}

	trafficPoliciesToUpdate := v.validationProcessor.Process(ctx, allTrafficPolicies, meshServices)
	for _, tp := range trafficPoliciesToUpdate {
		trafficPoliciesToUpdateSet[tp] = true
	}

	if objectsToUpdate, err := v.aggregationProcessor.Process(ctx, allTrafficPolicies); err == nil {
		if objectsToUpdate != nil {
			for _, service := range objectsToUpdate.MeshServices {
				err := v.meshServiceClient.UpdateMeshServiceStatus(ctx, service)
				if err != nil {
					multierr = multierror.Append(multierr, err)
				}
			}

			// accumulate traffic policies, so we have only one status update.
			for _, tp := range objectsToUpdate.TrafficPolicies {
				trafficPoliciesToUpdateSet[tp] = true
			}
		}
	} else {
		multierr = multierror.Append(multierr, err)
	}

	for trafficPolicy := range trafficPoliciesToUpdateSet {
		err := v.trafficPolicyClient.UpdateTrafficPolicyStatus(ctx, trafficPolicy)
		if err != nil {
			multierr = multierror.Append(multierr, err)
		}
	}

	if clusterNameToSnapshot, err := v.translationProcessor.Process(ctx, meshServices); err == nil {
		v.snapshotReconciler.ReconcileAllSnapshots(ctx, clusterNameToSnapshot)
	} else {
		multierr = multierror.Append(multierr, err)
	}
	return multierr
}
