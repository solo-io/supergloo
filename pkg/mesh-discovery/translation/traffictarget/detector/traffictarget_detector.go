package detector

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/utils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/utils/workloadutils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	DiscoveryMeshNameAnnotation      = "discovery.smh.solo.io/mesh-name"
	DiscoveryMeshNamespaceAnnotation = "discovery.smh.solo.io/mesh-namespace"
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
	DetectTrafficTarget(
		service *corev1.Service,
		workloads v1alpha2sets.WorkloadSet,
		meshes v1alpha2sets.MeshSet,
	) *v1alpha2.TrafficTarget
}

type trafficTargetDetector struct {
	ctx context.Context
}

func NewTrafficTargetDetector(ctx context.Context) TrafficTargetDetector {
	return &trafficTargetDetector{ctx: ctx}
}

func (t *trafficTargetDetector) DetectTrafficTarget(
	service *corev1.Service,
	meshWorkloads v1alpha2sets.WorkloadSet,
	meshes v1alpha2sets.MeshSet,
) *v1alpha2.TrafficTarget {

	kubeService := &v1alpha2.TrafficTargetSpec_KubeService{
		Ref:                    ezkube.MakeClusterObjectRef(service),
		WorkloadSelectorLabels: service.Spec.Selector,
		Labels:                 service.Labels,
		Ports:                  t.convertPorts(service),
	}

	var validMesh *v1.ObjectRef

	// TODO: support subsets from services which have been discovered via the annotation
	meshName, ok := service.Annotations[DiscoveryMeshNameAnnotation]
	if ok {
		// If no mesh namespace has been set via annotation, use default
		meshNamespace, ok := service.Annotations[DiscoveryMeshNamespaceAnnotation]
		if !ok {
			meshNamespace = defaults.GetPodNamespace()
		}
		possibleMeshRef := &v1.ObjectRef{
			Name:      meshName,
			Namespace: meshNamespace,
		}
		_, err := meshes.Find(possibleMeshRef)
		if err != nil {
			contextutils.LoggerFrom(t.ctx).Errorf("mesh %s could not be found in ns %s", meshName, meshNamespace)
		} else {
			validMesh = possibleMeshRef
		}
	}

	// if no mesh was found from the annotation, check the workloads
	if validMesh == nil {
		backingWorkloads := workloadutils.FindBackingWorkloads(kubeService, meshWorkloads)
		// if discovery is enabled, do not return
		if len(backingWorkloads) == 0 {
			return nil
		}

		// all backing workloads should be in the same mesh
		validMesh = backingWorkloads[0].Spec.Mesh

		// derive subsets from backing workloads
		kubeService.Subsets = t.findSubsets(backingWorkloads)
	}

	return &v1alpha2.TrafficTarget{
		ObjectMeta: utils.DiscoveredObjectMeta(service),
		Spec: v1alpha2.TrafficTargetSpec{
			Type: &v1alpha2.TrafficTargetSpec_KubeService_{
				KubeService: kubeService,
			},
			Mesh: validMesh,
		},
	}
}

// expects a list of just the workloads that back the service you're finding subsets for
func (t *trafficTargetDetector) findSubsets(backingWorkloads v1alpha2.WorkloadSlice) map[string]*v1alpha2.TrafficTargetSpec_KubeService_Subset {
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

func (t *trafficTargetDetector) convertPorts(service *corev1.Service) (ports []*v1alpha2.TrafficTargetSpec_KubeService_KubeServicePort) {
	for _, kubePort := range service.Spec.Ports {
		ports = append(ports, &v1alpha2.TrafficTargetSpec_KubeService_KubeServicePort{
			Port:     uint32(kubePort.Port),
			Name:     kubePort.Name,
			Protocol: string(kubePort.Protocol),
		})
	}
	return ports
}
