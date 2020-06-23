package detector

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/smh/pkg/mesh-discovery/snapshot/translation/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	skippedLabels = sets.NewString(
		"pod-template-hash",
		"service.istio.io/canonical-revision",
	)
)

// the MeshServiceDetector detects MeshServices from services
// whose backing pods are injected with a Mesh sidecar.
// If no Mesh is detected, nil is returned
type MeshServiceDetector interface {
	DetectMeshService(service *corev1.Service, meshWorkloads v1alpha1sets.MeshWorkloadSet) *v1alpha1.MeshService
}

type meshServiceDetector struct {}

func NewMeshServiceDetector() MeshServiceDetector {
	return &meshServiceDetector{}
}

func (d meshServiceDetector) DetectMeshService(service *corev1.Service, meshWorkloads v1alpha1sets.MeshWorkloadSet) *v1alpha1.MeshService {
	backingWorkloads := d.findBackingMeshWorkloads(service, meshWorkloads)
	if len(backingWorkloads) == 0 {
		// TODO(ilackarms): we currently only create mesh services for services with backing workloads; this may be problematic when working with external services (not contained inside the mesh)
		return nil
	}

	// all backing workloads should be in the same mesh
	mesh := backingWorkloads[0].Spec.Mesh

	// derive subsets from backing workloads
	subsets := findSubsets(backingWorkloads)

	return &v1alpha1.MeshService{
		ObjectMeta: utils.DiscoveredObjectMeta(service),
		Spec: v1alpha1.MeshServiceSpec{
			KubeService: &v1alpha1.MeshServiceSpec_KubeService{
				Ref:                    utils.MakeResourceRef(service),
				WorkloadSelectorLabels: service.Spec.Selector,
				Labels:                 service.Labels,
				Ports:                  convertPorts(service),
			},
			Mesh:    mesh,
			Subsets: subsets,
		},
	}
}

func (d meshServiceDetector) findBackingMeshWorkloads(service *corev1.Service, meshWorkloads v1alpha1sets.MeshWorkloadSet) v1alpha1.MeshWorkloadSlice {
	var backingMeshWorkloads v1alpha1.MeshWorkloadSlice

	for _, workload := range meshWorkloads.List() {
		// TODO(ilackarms): refactor this to support more than just k8s workloads
		// should probably go with a platform-based meshservice detector (e.g. one for k8s, one for vm, etc.)
		if isBackingKubeWorkload(service, workload.Spec.GetKubernetes()) {
			backingMeshWorkloads = append(backingMeshWorkloads, workload)
		}
	}
	return backingMeshWorkloads
}

func isBackingKubeWorkload(service *corev1.Service, kubeWorkload *v1alpha1.MeshWorkloadSpec_KubernertesWorkload) bool {
	if kubeWorkload == nil {
		return false
	}

	workloadRef := kubeWorkload.Controller

	if workloadRef.Cluster != service.ClusterName || workloadRef.Namespace != service.Namespace {
		return false
	}


	podLabels := kubeWorkload.GetPodLabels()
	selectorLabels := service.Spec.Selector

	if len(podLabels) == 0 || len(selectorLabels) == 0 {
		return false
	}

	return labels.AreLabelsInWhiteList(selectorLabels, podLabels)
}

// expects a list of just the workloads that back the service you're finding subsets for
func findSubsets(backingWorkloads v1alpha1.MeshWorkloadSlice) map[string]*v1alpha1.MeshServiceSpec_Subset {
	uniqueLabels := make(map[string]sets.String)
	for _, backingWorkload := range backingWorkloads {
		for key, val := range backingWorkload.Spec.GetKubernetes().GetPodLabels() {
			// skip known kubernetes values
			if skippedLabels.Has(key) {
				continue
			}
			existing, ok := uniqueLabels[key]
			if !ok {
				uniqueLabels[key] = sets.NewString(val)
			} else {
				existing.Insert(val)
			}
		}
	}
	/*
		Only select the keys with > 1 value
		The subsets worth noting will be sets of labels which share the same key, but have different values, such as:

			version:
				- v1
				- v2
	*/
	subsets := make(map[string]*v1alpha1.MeshServiceSpec_Subset)
	for k, v := range uniqueLabels {
		if v.Len() > 1 {
			subsets[k] = &v1alpha1.MeshServiceSpec_Subset{Values: v.List()}
		}
	}
	return subsets
}

func convertPorts(service *corev1.Service) (ports []*v1alpha1.MeshServiceSpec_KubeService_KubeServicePort) {
	for _, kubePort := range service.Spec.Ports {
		ports = append(ports, &v1alpha1.MeshServiceSpec_KubeService_KubeServicePort{
			Port:     uint32(kubePort.Port),
			Name:     kubePort.Name,
			Protocol: string(kubePort.Protocol),
		})
	}
	return ports
}
