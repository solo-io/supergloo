package exploration

import (
	"context"

	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/networking"
	"github.com/solo-io/service-mesh-hub/pkg/selector"
)

var (
	FailedToFindMeshServicesBySelector = func(err error, selector *core_types.ServiceSelector) error {
		return eris.Wrapf(err, "Failed to find services for selector %+v", selector)
	}
	FailedToFindMeshWorkloadsByIdentity = func(err error, selector *core_types.IdentitySelector) error {
		return eris.Wrapf(err, "Failed to find workloads with identity %+v", selector)
	}
	FailedToListAccessControlPolicies = func(err error) error {
		return eris.Wrap(err, "Failed to list access control policies")
	}
	FailedToListTrafficPolicies = func(err error) error {
		return eris.Wrap(err, "Failed to list traffic policies")
	}
)

func NewResourceExplorer(
	trafficPolicyClient zephyr_networking.TrafficPolicyClient,
	accessControlPolicyClient zephyr_networking.AccessControlPolicyClient,

	resourceSelector selector.ResourceSelector,
) ResourceExplorer {
	return &resourceExplorer{
		trafficPolicyClient:       trafficPolicyClient,
		accessControlPolicyClient: accessControlPolicyClient,
		ResourceSelector:          resourceSelector,
	}
}

type resourceExplorer struct {
	trafficPolicyClient       zephyr_networking.TrafficPolicyClient
	accessControlPolicyClient zephyr_networking.AccessControlPolicyClient
	ResourceSelector          selector.ResourceSelector
}

func (r *resourceExplorer) ExploreService(ctx context.Context, kubeResourceIdentifier FullyQualifiedKubeResource) (*ExplorationResult, error) {
	meshServiceForKubeService, err := r.ResourceSelector.GetMeshServiceByRefSelector(ctx, kubeResourceIdentifier.Name, kubeResourceIdentifier.Namespace, kubeResourceIdentifier.ClusterName)
	if err != nil {
		return nil, err
	}

	allAccessControlPolicies, err := r.accessControlPolicyClient.List(ctx)
	if err != nil {
		return nil, FailedToListAccessControlPolicies(err)
	}

	allTrafficControlPolicies, err := r.trafficPolicyClient.List(ctx)
	if err != nil {
		return nil, FailedToListTrafficPolicies(err)
	}

	var relevantAccessControlPolicies []*networking_v1alpha1.AccessControlPolicy
	for _, acp := range allAccessControlPolicies.Items {
		matchingMeshServices, err := r.ResourceSelector.GetMeshServicesByServiceSelector(ctx, acp.Spec.GetDestinationSelector())
		if err != nil {
			return nil, FailedToFindMeshServicesBySelector(err, acp.Spec.GetDestinationSelector())
		}

		for _, matchingMeshService := range matchingMeshServices {
			if clients.SameObject(matchingMeshService.ObjectMeta, meshServiceForKubeService.ObjectMeta) {
				relevantAccessControlPolicies = append(relevantAccessControlPolicies, &acp)

				// as soon as we find our desired service in the list, add this ACP and go on to the next one
				break
			}
		}
	}

	var relevantTrafficPolicies []*networking_v1alpha1.TrafficPolicy
	for _, tp := range allTrafficControlPolicies.Items {
		matchingMeshServices, err := r.ResourceSelector.GetMeshServicesByServiceSelector(ctx, tp.Spec.GetDestinationSelector())
		if err != nil {
			return nil, FailedToFindMeshServicesBySelector(err, tp.Spec.GetDestinationSelector())
		}

		for _, matchingMeshService := range matchingMeshServices {
			if clients.SameObject(matchingMeshService.ObjectMeta, meshServiceForKubeService.ObjectMeta) {
				relevantTrafficPolicies = append(relevantTrafficPolicies, &tp)

				// as soon as we find our desired service in the list, add this TP and go on to the next one
				break
			}
		}
	}

	return &ExplorationResult{
		Policies: &Policies{
			AccessControlPolicies: relevantAccessControlPolicies,
			TrafficPolicies:       relevantTrafficPolicies,
		},
	}, nil
}

func (r *resourceExplorer) ExploreWorkload(ctx context.Context, kubeResourceIdentifier FullyQualifiedKubeResource) (*ExplorationResult, error) {
	meshWorkloadForController, err := r.ResourceSelector.GetMeshWorkloadByRefSelector(ctx, kubeResourceIdentifier.Name, kubeResourceIdentifier.Namespace, kubeResourceIdentifier.ClusterName)
	if err != nil {
		return nil, err
	}

	allAccessControlPolicies, err := r.accessControlPolicyClient.List(ctx)
	if err != nil {
		return nil, FailedToListAccessControlPolicies(err)
	}

	allTrafficControlPolicies, err := r.trafficPolicyClient.List(ctx)
	if err != nil {
		return nil, FailedToListTrafficPolicies(err)
	}

	var relevantAccessControlPolicies []*networking_v1alpha1.AccessControlPolicy
	for _, acp := range allAccessControlPolicies.Items {
		matchingMeshWorkloads, err := r.ResourceSelector.GetMeshWorkloadsByIdentitySelector(ctx, acp.Spec.GetSourceSelector())
		if err != nil {
			return nil, FailedToFindMeshWorkloadsByIdentity(err, acp.Spec.GetSourceSelector())
		}

		for _, matchingMeshWorkload := range matchingMeshWorkloads {
			if clients.SameObject(matchingMeshWorkload.ObjectMeta, meshWorkloadForController.ObjectMeta) {
				relevantAccessControlPolicies = append(relevantAccessControlPolicies, &acp)

				break
			}
		}
	}

	var relevantTrafficPolicies []*networking_v1alpha1.TrafficPolicy
	for _, tp := range allTrafficControlPolicies.Items {
		matchingMeshWorkloads, err := r.ResourceSelector.GetMeshWorkloadsByWorkloadSelector(ctx, tp.Spec.GetSourceSelector())
		if err != nil {
			return nil, FailedToFindMeshServicesBySelector(err, tp.Spec.GetDestinationSelector())
		}

		for _, matchingMeshService := range matchingMeshWorkloads {
			if clients.SameObject(matchingMeshService.ObjectMeta, meshWorkloadForController.ObjectMeta) {
				relevantTrafficPolicies = append(relevantTrafficPolicies, &tp)

				break
			}
		}
	}

	return &ExplorationResult{
		Policies: &Policies{
			AccessControlPolicies: relevantAccessControlPolicies,
			TrafficPolicies:       relevantTrafficPolicies,
		},
	}, nil
}
