package detector

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/utils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/utils/workloadutils"
	"github.com/solo-io/skv2/pkg/ezkube"
	corev1 "k8s.io/api/core/v1"
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
	DetectMeshService(service *corev1.Service, meshWorkloads v1alpha2sets.MeshWorkloadSet) *v1alpha2.MeshService
}

type meshServiceDetector struct{}

func NewMeshServiceDetector() MeshServiceDetector {
	return &meshServiceDetector{}
}

func (m *meshServiceDetector) DetectMeshService(service *corev1.Service, meshWorkloads v1alpha2sets.MeshWorkloadSet) *v1alpha2.MeshService {

	kubeService := &v1alpha2.MeshServiceSpec_KubeService{
		Ref:                    ezkube.MakeClusterObjectRef(service),
		WorkloadSelectorLabels: service.Spec.Selector,
		Labels:                 service.Labels,
		Ports:                  convertPorts(service),
	}

	backingWorkloads := workloadutils.FindBackingMeshWorkloads(kubeService, meshWorkloads)
	if len(backingWorkloads) == 0 {
		// TODO(ilackarms): we currently only create mesh services for services with backing workloads; this may be problematic when working with external services (not contained inside the mesh)
		return nil
	}

	// all backing workloads should be in the same mesh
	mesh := backingWorkloads[0].Spec.Mesh

	// derive subsets from backing workloads
	kubeService.Subsets = m.findSubsets(backingWorkloads)

	return &v1alpha2.MeshService{
		ObjectMeta: utils.DiscoveredObjectMeta(service),
		Spec: v1alpha2.MeshServiceSpec{
			Type: &v1alpha2.MeshServiceSpec_KubeService_{
				KubeService: kubeService,
			},
			Mesh: mesh,
		},
	}
}

// expects a list of just the workloads that back the service you're finding subsets for
func (m *meshServiceDetector) findSubsets(backingWorkloads v1alpha2.MeshWorkloadSlice) map[string]*v1alpha2.MeshServiceSpec_KubeService_Subset {
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
	subsets := make(map[string]*v1alpha2.MeshServiceSpec_KubeService_Subset)
	for k, v := range uniqueLabels {
		if v.Len() > 1 {
			subsets[k] = &v1alpha2.MeshServiceSpec_KubeService_Subset{Values: v.List()}
		}
	}
	if len(subsets) == 0 {
		// important to return nil instead of empty map for asserting equality
		return nil
	}
	return subsets
}

func convertPorts(service *corev1.Service) (ports []*v1alpha2.MeshServiceSpec_KubeService_KubeServicePort) {
	for _, kubePort := range service.Spec.Ports {
		ports = append(ports, &v1alpha2.MeshServiceSpec_KubeService_KubeServicePort{
			Port:     uint32(kubePort.Port),
			Name:     kubePort.Name,
			Protocol: string(kubePort.Protocol),
		})
	}
	return ports
}
