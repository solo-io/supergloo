package description

import (
	"context"

	"github.com/rotisserie/eris"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
)

var (
	FailedToFindMeshServicesBySelector = func(err error, selector *smh_core_types.ServiceSelector) error {
		return eris.Wrapf(err, "Failed to find services for selector %+v", selector)
	}
	FailedToFindMeshWorkloadsByIdentity = func(err error, selector *smh_core_types.IdentitySelector) error {
		return eris.Wrapf(err, "Failed to find workloads with identity %+v", selector)
	}
	FailedToListAccessControlPolicies = func(err error) error {
		return eris.Wrap(err, "Failed to list access control policies")
	}
	FailedToListTrafficPolicies = func(err error) error {
		return eris.Wrap(err, "Failed to list traffic policies")
	}
)

func NewResourceDescriber(
	trafficPolicyClient smh_networking.TrafficPolicyClient,
	accessControlPolicyClient smh_networking.AccessControlPolicyClient,

	resourceSelector selection.ResourceSelector,
) ResourceDescriber {
	return &resourceDescriber{
		trafficPolicyClient:       trafficPolicyClient,
		accessControlPolicyClient: accessControlPolicyClient,
		ResourceSelector:          resourceSelector,
	}
}

type resourceDescriber struct {
	trafficPolicyClient       smh_networking.TrafficPolicyClient
	accessControlPolicyClient smh_networking.AccessControlPolicyClient
	ResourceSelector          selection.ResourceSelector
}

func (r *resourceDescriber) DescribeService(ctx context.Context, kubeResourceIdentifier FullyQualifiedKubeResource) (*DescriptionResult, error) {
	meshServiceForKubeService, err := r.ResourceSelector.GetAllMeshServiceByRefSelector(ctx, kubeResourceIdentifier.Name, kubeResourceIdentifier.Namespace, kubeResourceIdentifier.ClusterName)
	if err != nil {
		return nil, err
	}

	allAccessControlPolicies, err := r.accessControlPolicyClient.ListAccessControlPolicy(ctx)
	if err != nil {
		return nil, FailedToListAccessControlPolicies(err)
	}

	allTrafficControlPolicies, err := r.trafficPolicyClient.ListTrafficPolicy(ctx)
	if err != nil {
		return nil, FailedToListTrafficPolicies(err)
	}

	var relevantAccessControlPolicies []*smh_networking.AccessControlPolicy
	for _, acpIter := range allAccessControlPolicies.Items {
		acp := acpIter
		matchingMeshServices, err := r.ResourceSelector.GetAllMeshServicesByServiceSelector(ctx, acp.Spec.GetDestinationSelector())
		if err != nil {
			return nil, FailedToFindMeshServicesBySelector(err, acp.Spec.GetDestinationSelector())
		}

		for _, matchingMeshService := range matchingMeshServices {
			if selection.SameObject(matchingMeshService.ObjectMeta, meshServiceForKubeService.ObjectMeta) {
				relevantAccessControlPolicies = append(relevantAccessControlPolicies, &acp)

				// as soon as we find our desired service in the list, add this ACP and go on to the next one
				break
			}
		}
	}

	var relevantTrafficPolicies []*smh_networking.TrafficPolicy
	for _, tpIter := range allTrafficControlPolicies.Items {
		tp := tpIter
		matchingMeshServices, err := r.ResourceSelector.GetAllMeshServicesByServiceSelector(ctx, tp.Spec.GetDestinationSelector())
		if err != nil {
			return nil, FailedToFindMeshServicesBySelector(err, tp.Spec.GetDestinationSelector())
		}

		for _, matchingMeshService := range matchingMeshServices {
			if selection.SameObject(matchingMeshService.ObjectMeta, meshServiceForKubeService.ObjectMeta) {
				relevantTrafficPolicies = append(relevantTrafficPolicies, &tp)

				// as soon as we find our desired service in the list, add this TP and go on to the next one
				break
			}
		}
	}

	return &DescriptionResult{
		Policies: &Policies{
			AccessControlPolicies: relevantAccessControlPolicies,
			TrafficPolicies:       relevantTrafficPolicies,
		},
	}, nil
}

func (r *resourceDescriber) DescribeWorkload(ctx context.Context, kubeResourceIdentifier FullyQualifiedKubeResource) (*DescriptionResult, error) {
	meshWorkloadForController, err := r.ResourceSelector.GetMeshWorkloadByRefSelector(ctx, kubeResourceIdentifier.Name, kubeResourceIdentifier.Namespace, kubeResourceIdentifier.ClusterName)
	if err != nil {
		return nil, err
	}

	allAccessControlPolicies, err := r.accessControlPolicyClient.ListAccessControlPolicy(ctx)
	if err != nil {
		return nil, FailedToListAccessControlPolicies(err)
	}

	allTrafficControlPolicies, err := r.trafficPolicyClient.ListTrafficPolicy(ctx)
	if err != nil {
		return nil, FailedToListTrafficPolicies(err)
	}

	var relevantAccessControlPolicies []*smh_networking.AccessControlPolicy
	for _, acpIter := range allAccessControlPolicies.Items {
		acp := acpIter
		matchingMeshWorkloads, err := r.ResourceSelector.GetMeshWorkloadsByIdentitySelector(ctx, acp.Spec.GetSourceSelector())
		if err != nil {
			return nil, FailedToFindMeshWorkloadsByIdentity(err, acp.Spec.GetSourceSelector())
		}

		for _, matchingMeshWorkload := range matchingMeshWorkloads {
			if selection.SameObject(matchingMeshWorkload.ObjectMeta, meshWorkloadForController.ObjectMeta) {
				relevantAccessControlPolicies = append(relevantAccessControlPolicies, &acp)

				break
			}
		}
	}

	var relevantTrafficPolicies []*smh_networking.TrafficPolicy
	for _, tpIter := range allTrafficControlPolicies.Items {
		tp := tpIter
		matchingMeshWorkloads, err := r.ResourceSelector.GetMeshWorkloadsByWorkloadSelector(ctx, tp.Spec.GetSourceSelector())
		if err != nil {
			return nil, FailedToFindMeshServicesBySelector(err, tp.Spec.GetDestinationSelector())
		}

		for _, matchingMeshService := range matchingMeshWorkloads {
			if selection.SameObject(matchingMeshService.ObjectMeta, meshWorkloadForController.ObjectMeta) {
				relevantTrafficPolicies = append(relevantTrafficPolicies, &tp)

				break
			}
		}
	}

	return &DescriptionResult{
		Policies: &Policies{
			AccessControlPolicies: relevantAccessControlPolicies,
			TrafficPolicies:       relevantTrafficPolicies,
		},
	}, nil
}
