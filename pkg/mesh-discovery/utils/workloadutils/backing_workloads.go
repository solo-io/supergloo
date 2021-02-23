package workloadutils

import (
	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2/sets"
	"k8s.io/apimachinery/pkg/labels"
)

func FindBackingWorkloads(
	service *v1alpha2.DestinationSpec_KubeService,
	meshWorkloads v1alpha2sets.WorkloadSet,
) v1alpha2.WorkloadSlice {

	return meshWorkloads.List(func(workload *v1alpha2.Workload) bool {
		// TODO(ilackarms): refactor this to support more than just k8s workloads
		// should probably go with a platform-based destination detector (e.g. one for k8s, one for vm, etc.)
		return !isBackingKubeWorkload(service, workload.Spec.GetKubernetes())
	})
}

func isBackingKubeWorkload(
	service *v1alpha2.DestinationSpec_KubeService,
	kubeWorkload *v1alpha2.WorkloadSpec_KubernetesWorkload,
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
