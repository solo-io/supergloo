package workloadutils

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"k8s.io/apimachinery/pkg/labels"
)

func FindBackingMeshWorkloads(
	service *v1alpha2.MeshService,
	meshWorkloads v1alpha2sets.MeshWorkloadSet,
) v1alpha2sets.MeshWorkloadSet {

	result := v1alpha2sets.NewMeshWorkloadSet()
	if service.Spec.GetKubeService() == nil {
		return result
	}

	for _, workload := range meshWorkloads.List() {
		// TODO(ilackarms): refactor this to support more than just k8s workloads
		// should probably go with a platform-based meshservice detector (e.g. one for k8s, one for vm, etc.)
		if isBackingKubeWorkload(service.Spec.GetKubeService(), workload.Spec.GetKubernetes()) {
			result.Insert(workload)
		}
	}
	return result
}

func isBackingKubeWorkload(
	service *v1alpha2.MeshServiceSpec_KubeService,
	kubeWorkload *v1alpha2.MeshWorkloadSpec_KubernertesWorkload,
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
