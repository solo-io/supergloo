package detector

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/utils"
	"github.com/solo-io/skv2/pkg/ezkube"
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

// the TrafficTargetDetector detects TrafficTargets from services
// whose backing pods are injected with a Mesh sidecar.
// If no Mesh is detected, nil is returned
type TrafficTargetDetector interface {
	DetectTrafficTarget(service *corev1.Service, meshWorkloads v1alpha2sets.MeshWorkloadSet) *v1alpha2.TrafficTarget
}

type trafficTargetDetector struct{}

func NewTrafficTargetDetector() TrafficTargetDetector {
	return &trafficTargetDetector{}
}

func (d trafficTargetDetector) DetectTrafficTarget(service *corev1.Service, meshWorkloads v1alpha2sets.MeshWorkloadSet) *v1alpha2.TrafficTarget {
	backingWorkloads := d.findBackingMeshWorkloads(service, meshWorkloads)
	if len(backingWorkloads) == 0 {
		// TODO(ilackarms): we currently only create mesh services for services with backing workloads; this may be problematic when working with external services (not contained inside the mesh)
		return nil
	}

	// all backing workloads should be in the same mesh
	mesh := backingWorkloads[0].Spec.Mesh

	// derive subsets from backing workloads
	subsets := findSubsets(backingWorkloads)

	return &v1alpha2.TrafficTarget{
		ObjectMeta: utils.DiscoveredObjectMeta(service),
		Spec: v1alpha2.TrafficTargetSpec{
			Type: &v1alpha2.TrafficTargetSpec_KubeService_{
				KubeService: &v1alpha2.TrafficTargetSpec_KubeService{
					Ref:                    ezkube.MakeClusterObjectRef(service),
					WorkloadSelectorLabels: service.Spec.Selector,
					Labels:                 service.Labels,
					Ports:                  convertPorts(service),
					Subsets:                subsets,
				},
			},
			Mesh: mesh,
		},
	}
}

func (d trafficTargetDetector) findBackingMeshWorkloads(service *corev1.Service, meshWorkloads v1alpha2sets.MeshWorkloadSet) v1alpha2.MeshWorkloadSlice {
	var backingMeshWorkloads v1alpha2.MeshWorkloadSlice

	for _, workload := range meshWorkloads.List() {
		// TODO(ilackarms): refactor this to support more than just k8s workloads
		// should probably go with a platform-based traffictarget detector (e.g. one for k8s, one for vm, etc.)
		if isBackingKubeWorkload(service, workload.Spec.GetKubernetes()) {
			backingMeshWorkloads = append(backingMeshWorkloads, workload)
		}
	}
	return backingMeshWorkloads
}

func isBackingKubeWorkload(service *corev1.Service, kubeWorkload *v1alpha2.MeshWorkloadSpec_KubernertesWorkload) bool {
	if kubeWorkload == nil {
		return false
	}

	workloadRef := kubeWorkload.Controller

	if workloadRef.ClusterName != service.ClusterName || workloadRef.Namespace != service.Namespace {
		return false
	}

	podLabels := kubeWorkload.GetPodLabels()
	selectorLabels := service.Spec.Selector

	if len(podLabels) == 0 || len(selectorLabels) == 0 {
		return false
	}

	return labels.SelectorFromSet(selectorLabels).Matches(labels.Set(podLabels))
}

// expects a list of just the workloads that back the service you're finding subsets for
func findSubsets(backingWorkloads v1alpha2.MeshWorkloadSlice) map[string]*v1alpha2.TrafficTargetSpec_KubeService_Subset {
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
	subsets := make(map[string]*v1alpha2.TrafficTargetSpec_KubeService_Subset)
	for k, v := range uniqueLabels {
		if v.Len() > 1 {
			subsets[k] = &v1alpha2.TrafficTargetSpec_KubeService_Subset{Values: v.List()}
		}
	}
	if len(subsets) == 0 {
		// important to return nil instead of empty map for asserting equality
		return nil
	}
	return subsets
}

func convertPorts(service *corev1.Service) (ports []*v1alpha2.TrafficTargetSpec_KubeService_KubeServicePort) {
	for _, kubePort := range service.Spec.Ports {
		ports = append(ports, &v1alpha2.TrafficTargetSpec_KubeService_KubeServicePort{
			Port:     uint32(kubePort.Port),
			Name:     kubePort.Name,
			Protocol: string(kubePort.Protocol),
		})
	}
	return ports
}
