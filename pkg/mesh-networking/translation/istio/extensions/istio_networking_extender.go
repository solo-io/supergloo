package extensions

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/extensions"

	"github.com/solo-io/go-utils/contextutils"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/extensions/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/istio"
	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

//go:generate mockgen -source ./istio_networking_extender.go -destination ./mocks/mock_istio_networking_extender.go

// IstioExtender provides a caller-friendly mechanism for the Istio Networking Translator to apply patches supplied by a set of preconfigured v1alpha1.NetworkingExtensionsServer.
type IstioExtender interface {
	// PatchTrafficTargetOutputs retrieves from the NetworkingExtensionsServers and applies patches to the outputs for a given TrafficTarget
	PatchTrafficTargetOutputs(ctx context.Context, trafficTarget *v1alpha2.TrafficTarget, trafficTargetOutputs istio.Builder) error

	// PatchWorkloadOutputs retrieves from the NetworkingExtensionsServers and applies patches to the outputs for a given Workload
	PatchWorkloadOutputs(ctx context.Context, workload *v1alpha2.Workload, workloadOutputs istio.Builder) error

	// PatchMeshOutputs retrieves from the NetworkingExtensionsServers and applies patches to the outputs for a given Mesh
	PatchMeshOutputs(ctx context.Context, mesh *v1alpha2.Mesh, meshOutputs istio.Builder) error
}

type istioExtensions struct {
	// the user should provide an optional list of connection info for extension servers.
	// we create a client for each server and apply them in the order they were specified
	clientset extensions.Clientset
}

func NewIstioExtensions(clientset extensions.Clientset) *istioExtensions {
	return &istioExtensions{clientset: clientset}
}

func (i *istioExtensions) PatchTrafficTargetOutputs(ctx context.Context, trafficTarget *v1alpha2.TrafficTarget, trafficTargetOutputs istio.Builder) error {
	for _, exClient := range i.clientset.GetClients() {
		patches, err := exClient.GetTrafficTargetPatches(ctx, &v1alpha1.TrafficTargetPatchRequest{
			TrafficTarget: &v1alpha1.TrafficTargetResource{
				Metadata: &trafficTarget.ObjectMeta,
				Spec:     &trafficTarget.Spec,
				Status:   &trafficTarget.Status,
			},
			GeneratedResources: MakeGeneratedResources(trafficTargetOutputs),
		})
		if err != nil {
			return err
		}
		applyPatches(trafficTargetOutputs, patches)
	}
	return nil
}

func (i *istioExtensions) PatchWorkloadOutputs(ctx context.Context, workload *v1alpha2.Workload, workloadOutputs istio.Builder) error {
	for _, exClient := range i.clientset.GetClients() {
		patches, err := exClient.GetWorkloadPatches(ctx, &v1alpha1.WorkloadPatchRequest{
			Workload: &v1alpha1.WorkloadResource{
				Metadata: &workload.ObjectMeta,
				Spec:     &workload.Spec,
				Status:   &workload.Status,
			},
			GeneratedResources: MakeGeneratedResources(workloadOutputs),
		})
		if err != nil {
			return err
		}
		applyPatches(workloadOutputs, patches)
	}
	return nil
}

func (i *istioExtensions) PatchMeshOutputs(ctx context.Context, mesh *v1alpha2.Mesh, meshOutputs istio.Builder) error {
	for _, exClient := range i.clientset.GetClients() {
		patches, err := exClient.GetMeshPatches(ctx, &v1alpha1.MeshPatchRequest{
			Mesh: &v1alpha1.MeshResource{
				Metadata: &mesh.ObjectMeta,
				Spec:     &mesh.Spec,
				Status:   &mesh.Status,
			},
			GeneratedResources: MakeGeneratedResources(meshOutputs),
		})
		if err != nil {
			return err
		}
		applyPatches(meshOutputs, patches)
	}
	return nil
}

// convert istio.Builder to [generated resources]
// exposed as it is imported in extensions servers
func MakeGeneratedResources(outputs istio.Builder) []*v1alpha1.GeneratedResource {
	if outputs == nil {
		return nil
	}
	var generatedResources []*v1alpha1.GeneratedResource

	for _, resource := range outputs.GetDestinationRules().List() {
		resource := resource // pike
		generatedResources = append(generatedResources, &v1alpha1.GeneratedResource{
			Metadata: &resource.ObjectMeta,
			Type:     &v1alpha1.GeneratedResource_DestinationRule{DestinationRule: &resource.Spec},
		})
	}

	for _, resource := range outputs.GetEnvoyFilters().List() {
		resource := resource // pike
		generatedResources = append(generatedResources, &v1alpha1.GeneratedResource{
			Metadata: &resource.ObjectMeta,
			Type:     &v1alpha1.GeneratedResource_EnvoyFilter{EnvoyFilter: &resource.Spec},
		})
	}

	for _, resource := range outputs.GetServiceEntries().List() {
		resource := resource // pike
		generatedResources = append(generatedResources, &v1alpha1.GeneratedResource{
			Metadata: &resource.ObjectMeta,
			Type:     &v1alpha1.GeneratedResource_ServiceEntry{ServiceEntry: &resource.Spec},
		})
	}

	for _, resource := range outputs.GetVirtualServices().List() {
		resource := resource // pike
		generatedResources = append(generatedResources, &v1alpha1.GeneratedResource{
			Metadata: &resource.ObjectMeta,
			Type:     &v1alpha1.GeneratedResource_VirtualService{VirtualService: &resource.Spec},
		})
	}

	return generatedResources
}

// convert [generated resources] to istio.Builder
// exposed here for use in Server implementations.
func MakeOutputs(ctx context.Context, name string, generated []*v1alpha1.GeneratedResource) istio.Builder {
	outputs := istio.NewBuilder(ctx, name)
	for _, resource := range generated {
		switch resourceType := resource.Type.(type) {
		case *v1alpha1.GeneratedResource_DestinationRule:
			outputs.AddDestinationRules(&istionetworkingv1alpha3.DestinationRule{
				ObjectMeta: *resource.Metadata,
				Spec:       *resourceType.DestinationRule,
			})
		case *v1alpha1.GeneratedResource_EnvoyFilter:
			outputs.AddEnvoyFilters(&istionetworkingv1alpha3.EnvoyFilter{
				ObjectMeta: *resource.Metadata,
				Spec:       *resourceType.EnvoyFilter,
			})
		case *v1alpha1.GeneratedResource_ServiceEntry:
			outputs.AddServiceEntries(&istionetworkingv1alpha3.ServiceEntry{
				ObjectMeta: *resource.Metadata,
				Spec:       *resourceType.ServiceEntry,
			})
		case *v1alpha1.GeneratedResource_VirtualService:
			outputs.AddVirtualServices(&istionetworkingv1alpha3.VirtualService{
				ObjectMeta: *resource.Metadata,
				Spec:       *resourceType.VirtualService,
			})
		default:
			contextutils.LoggerFrom(ctx).DPanicf("unsupported resource type %T", resourceType)
		}
	}
	return outputs
}

func applyPatches(outputs istio.Builder, patches *v1alpha1.PatchList) {
	if patches == nil {
		return
	}
	for _, patchedResource := range patches.PatchedResources {
		switch resourceType := patchedResource.Type.(type) {
		case *v1alpha1.GeneratedResource_DestinationRule:
			outputs.AddDestinationRules(&istionetworkingv1alpha3.DestinationRule{
				ObjectMeta: *patchedResource.Metadata,
				Spec:       *resourceType.DestinationRule,
			})
		case *v1alpha1.GeneratedResource_EnvoyFilter:
			outputs.AddEnvoyFilters(&istionetworkingv1alpha3.EnvoyFilter{
				ObjectMeta: *patchedResource.Metadata,
				Spec:       *resourceType.EnvoyFilter,
			})
		case *v1alpha1.GeneratedResource_ServiceEntry:
			outputs.AddServiceEntries(&istionetworkingv1alpha3.ServiceEntry{
				ObjectMeta: *patchedResource.Metadata,
				Spec:       *resourceType.ServiceEntry,
			})
		case *v1alpha1.GeneratedResource_VirtualService:
			outputs.AddVirtualServices(&istionetworkingv1alpha3.VirtualService{
				ObjectMeta: *patchedResource.Metadata,
				Spec:       *resourceType.VirtualService,
			})
		}
	}
}
