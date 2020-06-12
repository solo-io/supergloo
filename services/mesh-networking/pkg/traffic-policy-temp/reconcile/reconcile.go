package reconcile

import (
	"context"

	"github.com/hashicorp/go-multierror"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/common/reconciliation"
	aggregation_framework "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/aggregation/framework"
	translation_framework "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework/snapshot"
	traffic_policy_validation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/validation"
)

func NewValidationReconciler() reconciliation.Reconciler {
	return &reconciler{}
}

type reconciler struct {
	trafficPolicyClient smh_networking.TrafficPolicyClient
	meshServiceClient   smh_discovery.MeshServiceClient

	snapshotReconciler snapshot.TranslationSnapshotReconciler

	validationProcessor  traffic_policy_validation.ValidationProcessor
	aggregationProcessor aggregation_framework.AggregateProcessor
	translationProcessor translation_framework.TranslationProcessor
}

func NewReconciler(
	trafficPolicyClient smh_networking.TrafficPolicyClient,
	meshServiceClient smh_discovery.MeshServiceClient,
	snapshotReconciler snapshot.TranslationSnapshotReconciler,
	validationProcessor traffic_policy_validation.ValidationProcessor,
	aggregationProcessor aggregation_framework.AggregateProcessor,
	translationProcessor translation_framework.TranslationProcessor,
) reconciliation.Reconciler {
	return &reconciler{
		trafficPolicyClient:  trafficPolicyClient,
		meshServiceClient:    meshServiceClient,
		snapshotReconciler:   snapshotReconciler,
		validationProcessor:  validationProcessor,
		aggregationProcessor: aggregationProcessor,
		translationProcessor: translationProcessor,
	}
}

func (*reconciler) GetName() string {
	return "traffic-policy-reconciler"
}

func (v *reconciler) Reconcile(ctx context.Context) error {
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

	trafficPoliciesToUpdate := v.validationProcessor.Process(ctx, allTrafficPolicies, meshServices)

	if objToUpdate, err := v.aggregationProcessor.Process(ctx, allTrafficPolicies); err == nil {
		for _, service := range objToUpdate.MeshServices {
			err := v.meshServiceClient.UpdateMeshServiceStatus(ctx, service)
			if err != nil {
				multierr = multierror.Append(multierr, err)
			}
		}

		// merge traffic policies.
		// add all policies from objToUpdate to trafficPolicies; make sure they are unique;
		// these should be pointers to the same thing; so no need to merge anything.

	} else {
		multierr = multierror.Append(multierr, err)
	}

	if clusterNameToSnapshot, err := v.translationProcessor.Process(ctx); err == nil {
		v.snapshotReconciler.ReconcileAllSnapshots(ctx, clusterNameToSnapshot)
	} else {
		multierr = multierror.Append(multierr, err)
	}

	for _, trafficPolicy := range trafficPoliciesToUpdate {
		err := v.trafficPolicyClient.UpdateTrafficPolicyStatus(ctx, trafficPolicy)
		if err != nil {
			multierr = multierror.Append(multierr, err)
		}
	}
	return multierr
}
