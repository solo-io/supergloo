package extensions

import (
	"context"

	xdsv1beta1 "github.com/solo-io/gloo-mesh/pkg/api/xds.agent.enterprise.mesh.gloo.solo.io/v1beta1"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"

	v1beta1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/extensions/v1beta1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/extensions"
)

//go:generate mockgen -source ./istio_networking_extender.go -destination ./mocks/mock_istio_networking_extender.go

// IstioExtender provides a caller-friendly mechanism for the Istio Networking Translator to apply patches supplied by a set of preconfigured v1alpha1.NetworkingExtensionsServer.
type IstioExtender interface {
	// PatchOutputs retrieves from the NetworkingExtensionsServers and applies patches to the outputs for a given Destination
	PatchOutputs(ctx context.Context, inputs input.LocalSnapshot, outputs istio.Builder) error
}

type istioExtender struct {
	// the user should provide an optional list of connection info for extension servers.
	// we create a client for each server and apply them in the order they were specified
	clientset extensions.Clientset
}

func NewIstioExtender(clientset extensions.Clientset) *istioExtender {
	return &istioExtender{clientset: clientset}
}

func (i *istioExtender) PatchOutputs(ctx context.Context, inputs input.LocalSnapshot, outputs istio.Builder) error {
	if i.clientset == nil {
		return nil
	}
	for _, exClient := range i.clientset.GetClients() {
		patches, err := exClient.GetExtensionPatches(ctx, &v1beta1.ExtensionPatchRequest{
			Inputs:  extensions.InputSnapshotToProto(inputs),
			Outputs: OutputsToProto(outputs),
		})
		if err != nil {
			return err
		}
		applyPatches(ctx, outputs, patches.PatchedOutputs)
	}
	return nil
}

// OutputsToProto converts istio.Builder to [generated objects]
// exposed as it is imported in extensions servers
// NOTE: If we add more supported types of v1alpha1.GeneratedObjects, we need to
// update this function to convert them.
func OutputsToProto(outputs istio.Builder) []*v1beta1.GeneratedObject {
	if outputs == nil {
		return nil
	}
	var generatedObjects []*v1beta1.GeneratedObject

	for _, object := range outputs.GetDestinationRules().List() {
		object := object // pike
		generatedObjects = append(generatedObjects, &v1beta1.GeneratedObject{
			Metadata: extensions.ObjectMetaToProto(object.ObjectMeta),
			Type:     &v1beta1.GeneratedObject_DestinationRule{DestinationRule: &object.Spec},
		})
	}

	for _, object := range outputs.GetEnvoyFilters().List() {
		object := object // pike
		generatedObjects = append(generatedObjects, &v1beta1.GeneratedObject{
			Metadata: extensions.ObjectMetaToProto(object.ObjectMeta),
			Type:     &v1beta1.GeneratedObject_EnvoyFilter{EnvoyFilter: &object.Spec},
		})
	}

	for _, object := range outputs.GetServiceEntries().List() {
		object := object // pike
		generatedObjects = append(generatedObjects, &v1beta1.GeneratedObject{
			Metadata: extensions.ObjectMetaToProto(object.ObjectMeta),
			Type:     &v1beta1.GeneratedObject_ServiceEntry{ServiceEntry: &object.Spec},
		})
	}

	for _, object := range outputs.GetVirtualServices().List() {
		object := object // pike
		generatedObjects = append(generatedObjects, &v1beta1.GeneratedObject{
			Metadata: extensions.ObjectMetaToProto(object.ObjectMeta),
			Type:     &v1beta1.GeneratedObject_VirtualService{VirtualService: &object.Spec},
		})
	}

	for _, object := range outputs.GetXdsConfigs().List() {
		object := object // pike
		generatedObjects = append(generatedObjects, &v1beta1.GeneratedObject{
			Metadata: extensions.ObjectMetaToProto(object.ObjectMeta),
			Type:     &v1beta1.GeneratedObject_XdsConfig{XdsConfig: &object.Spec},
		})
	}

	return generatedObjects
}

// OutputsFromProto convert [generated objects] to istio.Builder
// exposed here for use in Server implementations.
// NOTE: If we add more supported types of v1alpha1.GeneratedObjects, we need to
// update this function to convert them.
func OutputsFromProto(ctx context.Context, name string, generated []*v1beta1.GeneratedObject) istio.Builder {
	outputs := istio.NewBuilder(ctx, name)
	applyPatches(ctx, outputs, generated)
	return outputs
}

// NOTE: If we add more supported types of v1alpha1.GeneratedObjects, we need to
// update this function to convert them.
func applyPatches(ctx context.Context, outputs istio.Builder, patches []*v1beta1.GeneratedObject) {
	if patches == nil {
		return
	}
	for _, patchedObject := range patches {
		switch objectType := patchedObject.Type.(type) {
		case *v1beta1.GeneratedObject_DestinationRule:
			contextutils.LoggerFrom(ctx).Debugf("applied patched DestinationRule %v", sets.Key(patchedObject.Metadata))
			outputs.AddDestinationRules(&istionetworkingv1alpha3.DestinationRule{
				ObjectMeta: extensions.ObjectMetaFromProto(patchedObject.Metadata),
				Spec:       *objectType.DestinationRule,
			})
		case *v1beta1.GeneratedObject_EnvoyFilter:
			contextutils.LoggerFrom(ctx).Debugf("applied patched EnvoyFilter %v", sets.Key(patchedObject.Metadata))
			outputs.AddEnvoyFilters(&istionetworkingv1alpha3.EnvoyFilter{
				ObjectMeta: extensions.ObjectMetaFromProto(patchedObject.Metadata),
				Spec:       *objectType.EnvoyFilter,
			})
		case *v1beta1.GeneratedObject_ServiceEntry:
			contextutils.LoggerFrom(ctx).Debugf("applied patched ServiceEntry %v", sets.Key(patchedObject.Metadata))
			outputs.AddServiceEntries(&istionetworkingv1alpha3.ServiceEntry{
				ObjectMeta: extensions.ObjectMetaFromProto(patchedObject.Metadata),
				Spec:       *objectType.ServiceEntry,
			})
		case *v1beta1.GeneratedObject_VirtualService:
			contextutils.LoggerFrom(ctx).Debugf("applied patched VirtualService %v", sets.Key(patchedObject.Metadata))
			outputs.AddVirtualServices(&istionetworkingv1alpha3.VirtualService{
				ObjectMeta: extensions.ObjectMetaFromProto(patchedObject.Metadata),
				Spec:       *objectType.VirtualService,
			})
		case *v1beta1.GeneratedObject_XdsConfig:
			contextutils.LoggerFrom(ctx).Debugf("applied patched XdsConfig %v", sets.Key(patchedObject.Metadata))
			outputs.AddXdsConfigs(&xdsv1beta1.XdsConfig{
				ObjectMeta: extensions.ObjectMetaFromProto(patchedObject.Metadata),
				Spec:       *objectType.XdsConfig,
			})
		default:
			contextutils.LoggerFrom(ctx).DPanicf("unsupported object type %T", objectType)
		}
	}
}
