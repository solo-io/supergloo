package workloadutils

import (
	v1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	v1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	"k8s.io/apimachinery/pkg/labels"
)

func FindBackingWorkloads(
	service *v1.DestinationSpec_KubeService,
	meshWorkloads v1alpha2sets.WorkloadSet,
) v1.WorkloadSlice {

	return meshWorkloads.List(func(workload *v1.Workload) bool {
		// TODO(ilackarms): refactor this to support more than just k8s workloads
		// should probably go with a platform-based destination detector (e.g. one for k8s, one for vm, etc.)
		return !isBackingKubeWorkload(service, workload.Spec.GetKubernetes())
	})
}

// Public to be used in enterprise
func FindDestinationForWorkload(
	workload *v1.Workload,
	destinations v1alpha2sets.DestinationSet,
) *v1.Destination {
	matchingDestinations := destinations.List(func(dest *v1.Destination) bool {
		return !isBackingKubeWorkload(dest.Spec.GetKubeService(), workload.Spec.GetKubernetes())
	})
	if len(matchingDestinations) == 0 {
		return nil
	}
	return matchingDestinations[0]
}

func isBackingKubeWorkload(
	service *v1.DestinationSpec_KubeService,
	kubeWorkload *v1.WorkloadSpec_KubernetesWorkload,
) bool {
	if kubeWorkload == nil {
		return false
	}

	workloadRef := kubeWorkload.Controller

	if workloadRef.ClusterName != service.GetRef().GetClusterName() ||
		workloadRef.Namespace != service.GetRef().GetNamespace() {
		return false
	}

	podLabels := kubeWorkload.GetPodLabels()
	selectorLabels := service.WorkloadSelectorLabels

	if len(podLabels) == 0 || len(selectorLabels) == 0 {
		return false
	}

	return labels.SelectorFromSet(selectorLabels).Matches(labels.Set(podLabels))
}
