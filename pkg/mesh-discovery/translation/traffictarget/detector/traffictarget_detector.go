package detector

import (
	"context"

	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	discoveryv1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2/sets"
	networkingv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	networkingv1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2/sets"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/utils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/utils/workloadutils"
	"github.com/solo-io/go-utils/contextutils"
	sets2 "github.com/solo-io/skv2/contrib/pkg/sets"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/pointer"
)

const (
	// TODO: allow for specifying specific meshes.
	// Currently this annotation assumes that there is only one mesh per cluster, and therefore the corresponding
	// TrafficTarget will be associated with that mesh.
	DiscoveryMeshAnnotation = "discovery.mesh.gloo.solo.io/enabled"
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
		ctx context.Context,
		service *corev1.Service,
		endpoints corev1sets.EndpointsSet,
		workloads discoveryv1alpha2sets.WorkloadSet,
		meshes discoveryv1alpha2sets.MeshSet,
		virtualMeshes networkingv1alpha2sets.VirtualMeshSet,
	) *v1alpha2.TrafficTarget
}

type trafficTargetDetector struct{}

func NewTrafficTargetDetector() TrafficTargetDetector {
	return &trafficTargetDetector{}
}

func (t *trafficTargetDetector) DetectTrafficTarget(
	ctx context.Context,
	service *corev1.Service,
	endpoints corev1sets.EndpointsSet,
	meshWorkloads discoveryv1alpha2sets.WorkloadSet,
	meshes discoveryv1alpha2sets.MeshSet,
	virtualMeshes networkingv1alpha2sets.VirtualMeshSet,
) *v1alpha2.TrafficTarget {
	kubeService := &v1alpha2.TrafficTargetSpec_KubeService{
		Ref:                    ezkube.MakeClusterObjectRef(service),
		WorkloadSelectorLabels: service.Spec.Selector,
		Labels:                 service.Labels,
		Ports:                  convertPorts(service),
	}

	// If the service is not associated with a mesh, do not create a traffic target
	validMesh := getMeshForKubeService(ctx, service, kubeService, meshWorkloads, meshes)
	if validMesh == nil {
		return nil
	}

	kubeService.Endpoints = getEndpointsForService(ctx, validMesh, endpoints, kubeService, virtualMeshes)

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

func getMeshForKubeService(
	ctx context.Context,
	service *corev1.Service,
	kubeService *v1alpha2.TrafficTargetSpec_KubeService,
	meshWorkloads discoveryv1alpha2sets.WorkloadSet,
	meshes discoveryv1alpha2sets.MeshSet,
) *v1.ObjectRef {

	var validMesh *v1.ObjectRef

	// TODO: support subsets from services which have been discovered via the annotation
	discoveryEnabled, ok := service.Annotations[DiscoveryMeshAnnotation]
	if ok && discoveryEnabled == "true" {

		// Search for mesh which exists on the same cluster as the annotated service
		for _, mesh := range meshes.List() {
			mesh := mesh
			switch typedMesh := mesh.Spec.GetMeshType().(type) {
			case *v1alpha2.MeshSpec_Osm:
				if typedMesh.Osm.GetInstallation().GetCluster() == service.GetClusterName() {
					validMesh = ezkube.MakeObjectRef(mesh)
					break
				}
			case *v1alpha2.MeshSpec_Istio_:
				if typedMesh.Istio.GetInstallation().GetCluster() == service.GetClusterName() {
					validMesh = ezkube.MakeObjectRef(mesh)
					break
				}
			}
		}

		if validMesh == nil {
			contextutils.LoggerFrom(ctx).Errorf(
				"mesh could not be found for annotated service %s",
				sets2.TypedKey(service),
			)
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
		kubeService.Subsets = findSubsets(backingWorkloads)
	}
	return validMesh
}

func getEndpointsForService(
	ctx context.Context,
	validMesh *v1.ObjectRef,
	endpoints corev1sets.EndpointsSet,
	kubeService *v1alpha2.TrafficTargetSpec_KubeService,
	virtualMeshes networkingv1alpha2sets.VirtualMeshSet,
) []*v1alpha2.TrafficTargetSpec_KubeService_EndpointsSubset {
	var result []*v1alpha2.TrafficTargetSpec_KubeService_EndpointsSubset
	// Flat network is enabled for this particular virtual mesh
	// so we will add all of the service endpoints to the traffic target
	if vm := findRelatedVirtualMesh(validMesh, virtualMeshes); vm != nil && vm.Spec.GetFederation().GetFlatNetwork() {
		ep, err := endpoints.Find(kubeService.GetRef())
		if err != nil {
			contextutils.LoggerFrom(ctx).Errorf(
				"endpoints could not be found for kube service %s",
				sets2.TypedKey(kubeService.GetRef()),
			)
		} else {
			for _, epSub := range ep.Subsets {
				sub := &v1alpha2.TrafficTargetSpec_KubeService_EndpointsSubset{}
				for _, addr := range epSub.Addresses {
					sub.IpAddresses = append(sub.IpAddresses, addr.IP)
				}
				for _, port := range epSub.Ports {
					svcPort := &v1alpha2.TrafficTargetSpec_KubeService_KubeServicePort{
						Port:     uint32(port.Port),
						Name:     port.Name,
						Protocol: string(port.Protocol),
					}
					if port.AppProtocol != nil {
						svcPort.AppProtocol = *port.AppProtocol
					}
					sub.Ports = append(sub.Ports, svcPort)
				}
				result = append(result, sub)
			}
		}
	}
	return result
}

func findRelatedVirtualMesh(
	meshRef *v1.ObjectRef,
	virtualMeshes networkingv1alpha2sets.VirtualMeshSet,
) *networkingv1alpha2.VirtualMesh {
	for _, vm := range virtualMeshes.List() {
		for _, vmMeshRef := range vm.Spec.GetMeshes() {
			if vmMeshRef.Equal(meshRef) {
				return vm
			}
		}
	}
	return nil
}

// expects a list of just the workloads that back the service you're finding subsets for
func findSubsets(backingWorkloads v1alpha2.WorkloadSlice) map[string]*v1alpha2.TrafficTargetSpec_KubeService_Subset {
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
			Port:        uint32(kubePort.Port),
			Name:        kubePort.Name,
			Protocol:    string(kubePort.Protocol),
			AppProtocol: pointer.StringPtrDerefOr(kubePort.AppProtocol, ""),
		})
	}
	return ports
}
